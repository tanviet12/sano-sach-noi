package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Tự cập nhật: tải file cài của bản mới (đúng hệ điều hành + kiểu cài), kiểm
// chữ ký ed25519 trên SHA256SUMS rồi kiểm SHA256 của file tải về, sai là từ
// chối. Thay bản đang chạy ở selfupdate_<os>.go. Sách, tiến độ nghe, bộ đọc
// nằm ngoài chỗ cài (~/Sano, thư mục dữ liệu) nên giữ nguyên.
//
// Chuỗi tin cậy: khoá công khai nhúng trong app (updatekey.go) → chữ ký
// SHA256SUMS.sig → SHA256SUMS (có tên file kèm số phiên bản) → file cài. Chỉ
// tải từ trang phát hành của repo; bản mới phải có số phiên bản lớn hơn.
const (
	sumsName          = "SHA256SUMS"
	sigName           = "SHA256SUMS.sig"
	maxSumsBytes      = 64 << 10
	maxSigBytes       = 1 << 10
	maxInstallerBytes = 300 << 20 // bản cài hiện ~10–20 MB; chặn file lạ cỡ lớn
	stallTimeout      = 60 * time.Second
	progressEvery     = 200 * time.Millisecond

	eventUpdateProgress = "update:progress"

	// envUpdateAPI — chỉ để chạy thử cục bộ: trỏ việc hỏi bản mới sang máy chủ
	// thử trên chính máy (127.0.0.1/localhost). Địa chỉ khác bị bỏ qua. Chữ ký
	// vẫn kiểm bằng khoá nhúng trong app như bình thường.
	envUpdateAPI = "SANO_UPDATE_API"
)

var (
	errNoKey        = errors.New("bản này chưa gắn khoá kiểm chữ ký nên không tự cập nhật được — tải bản cài ở trang phát hành")
	errBadSignature = errors.New("chữ ký bản phát hành không hợp lệ — đây không phải bản chính thức của Sano, đã huỷ cập nhật")
	errBadChecksum  = errors.New("file tải về không khớp mã SHA256 đã ký — có thể bị hỏng hoặc bị tráo, đã xoá và huỷ cập nhật")
)

// UpdateStatus — tiến độ tự cập nhật cho giao diện.
type UpdateStatus struct {
	Phase    string `json:"phase"` // idle | downloading | ready | error
	Version  string `json:"version"`
	Done     int64  `json:"done"`
	Total    int64  `json:"total"`
	Verified bool   `json:"verified"` // đã kiểm chữ ký + SHA256 xong, hợp lệ
	Error    string `json:"error"`
	// ApplyOnQuit — người dùng chọn "Khởi động lại sau": thay bản mới khi thoát app.
	ApplyOnQuit bool `json:"applyOnQuit"`
}

// installInfo — Sano đang chạy được cài theo kiểu nào (selfupdate_<os>.go).
type installInfo struct {
	kind   string // dmg | setup | portable | appimage
	path   string // Sano.app | thư mục cài | Sano.exe | file .AppImage
	suffix string // đuôi tên file cài: "macos-universal.dmg", "windows-amd64-setup.exe"...
	err    error  // khác nil: không tự thay được, lý do cho người dùng
}

type updatePlan struct {
	version string
	inst    installInfo
	asset   releaseAsset
	sums    releaseAsset
	sig     releaseAsset
}

type updateJob struct {
	cancel context.CancelFunc
	plan   updatePlan
	file   string // file cài đã tải + kiểm xong
	status UpdateStatus
}

// updateState — phần tự cập nhật của App (khoá riêng, không dính a.mu).
type updateState struct {
	updMu      sync.Mutex
	updRelease *latestRelease
	upd        *updateJob
}

// updateAPI — địa chỉ hỏi bản mới (mặc định GitHub; SANO_UPDATE_API chỉ nhận máy chủ thử cục bộ).
func updateAPI() string {
	if v := strings.TrimSpace(os.Getenv(envUpdateAPI)); v != "" {
		if u, err := url.Parse(v); err == nil && isLoopback(u) {
			return v
		}
		log.Printf("bỏ qua %s=%q: chỉ nhận 127.0.0.1/localhost", envUpdateAPI, v)
	}
	return latestReleaseAPI
}

func isLoopback(u *url.URL) bool {
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	h := u.Hostname()
	if h == "localhost" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}

// allowedDownload: chỉ tải file của đúng bản version trên trang phát hành của
// repo (GitHub chuyển tiếp sang máy chủ file của nó — xem updateClient). Khi
// chạy thử cục bộ: cùng máy chủ thử.
func allowedDownload(raw, version string) bool {
	api := updateAPI()
	if api == latestReleaseAPI {
		return strings.HasPrefix(raw, releasesPage+"/download/v"+version+"/")
	}
	u, err := url.Parse(raw)
	a, err2 := url.Parse(api)
	return err == nil && err2 == nil && isLoopback(u) && u.Host == a.Host
}

// updateClient: không giới hạn tổng thời gian (file vài chục MB trên mạng chậm)
// nhưng có hạn chờ phản hồi, chống treo khi đang tải (stallTimeout) và chỉ đi
// theo chuyển tiếp https.
func updateClient() *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.ResponseHeaderTimeout = 30 * time.Second
	return &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("chuyển tiếp quá nhiều lần")
			}
			if req.URL.Scheme != "https" && !isLoopback(req.URL) {
				return fmt.Errorf("từ chối chuyển tiếp không mã hoá tới %s", req.URL.Host)
			}
			return nil
		},
	}
}

// planUpdate chọn file cài hợp với máy (theo inst) trong bản phát hành rel.
func planUpdate(rel *latestRelease, inst installInfo) (updatePlan, error) {
	if strings.TrimSpace(updatePublicKeys) == "" {
		return updatePlan{}, errNoKey
	}
	if rel == nil {
		return updatePlan{}, errors.New("chưa kiểm tra bản mới")
	}
	if inst.err != nil {
		return updatePlan{}, inst.err
	}
	p := updatePlan{version: rel.Version, inst: inst}
	want := map[string]*releaseAsset{
		"Sano-" + rel.Version + "-" + inst.suffix: &p.asset,
		sumsName: &p.sums,
		sigName:  &p.sig,
	}
	for _, a := range rel.Assets {
		if dst, ok := want[a.Name]; ok {
			if dst.Name != "" {
				return updatePlan{}, fmt.Errorf("bản phát hành có hai file %s", a.Name)
			}
			*dst = a
		}
	}
	switch {
	case p.asset.Name == "":
		return updatePlan{}, fmt.Errorf("bản %s chưa có file cài cho máy này (%s)", rel.Version, inst.suffix)
	case p.sig.Name == "" || p.sums.Name == "":
		return updatePlan{}, fmt.Errorf("bản %s chưa có chữ ký (%s) nên không tự cập nhật được", rel.Version, sigName)
	case p.asset.Size > maxInstallerBytes:
		return updatePlan{}, fmt.Errorf("file cài %s lớn bất thường (%d MB)", p.asset.Name, p.asset.Size>>20)
	}
	for _, a := range []releaseAsset{p.asset, p.sums, p.sig} {
		if !allowedDownload(a.URL, rel.Version) {
			return updatePlan{}, fmt.Errorf("địa chỉ tải %s không thuộc trang phát hành của Sano", a.Name)
		}
	}
	return p, nil
}

// parsePublicKeys đọc danh sách khoá base64 cách nhau dấu phẩy.
func parsePublicKeys(s string) ([]ed25519.PublicKey, error) {
	var keys []ed25519.PublicKey
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		b, err := base64.StdEncoding.DecodeString(part)
		if err != nil || len(b) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("khoá công khai cập nhật không hợp lệ: %q", part)
		}
		keys = append(keys, ed25519.PublicKey(b))
	}
	if len(keys) == 0 {
		return nil, errNoKey
	}
	return keys, nil
}

// verifySums kiểm chữ ký (base64 của 64 byte ed25519, một dòng) trên đúng nội
// dung SHA256SUMS. Khớp một trong các khoá là đạt.
func verifySums(sums, sig []byte, keys []ed25519.PublicKey) error {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(sig)))
	if err != nil || len(raw) != ed25519.SignatureSize {
		return errBadSignature
	}
	for _, k := range keys {
		if ed25519.Verify(k, sums, raw) {
			return nil
		}
	}
	return errBadSignature
}

// sumFor tìm mã SHA256 (hex thường) của file name trong SHA256SUMS dạng
// "<hash>  <tên>" (sha256sum). Tên xuất hiện hai lần hoặc dòng hỏng → lỗi.
func sumFor(sums []byte, name string) (string, error) {
	found := ""
	sc := bufio.NewScanner(bytes.NewReader(sums))
	for sc.Scan() {
		ln := strings.TrimSpace(sc.Text())
		if ln == "" {
			continue
		}
		hash, file, ok := strings.Cut(ln, " ")
		file = strings.TrimPrefix(strings.TrimSpace(file), "*")
		if !ok || len(hash) != sha256.Size*2 || strings.ToLower(hash) != hash {
			return "", fmt.Errorf("%s có dòng không hợp lệ", sumsName)
		}
		if _, err := hex.DecodeString(hash); err != nil {
			return "", fmt.Errorf("%s có dòng không hợp lệ", sumsName)
		}
		if file != name {
			continue
		}
		if found != "" {
			return "", fmt.Errorf("%s có hai dòng cho %s", sumsName, name)
		}
		found = hash
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("%s đã ký không có %s", sumsName, name)
	}
	return found, nil
}

// fetchSmall tải file nhỏ (SHA256SUMS, chữ ký) vào bộ nhớ, quá max byte là lỗi.
func fetchSmall(ctx context.Context, client *http.Client, raw string, max int64) ([]byte, error) {
	body, _, err := openDownload(ctx, client, raw)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	b, err := io.ReadAll(io.LimitReader(body, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, fmt.Errorf("file %s lớn bất thường", assetFile(raw))
	}
	return b, nil
}

func openDownload(ctx context.Context, client *http.Client, raw string) (io.ReadCloser, int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "Sano-desktop")
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("không tải được %s: %w", assetFile(raw), err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, 0, fmt.Errorf("tải %s: máy chủ trả mã %d", assetFile(raw), resp.StatusCode)
	}
	return resp.Body, resp.ContentLength, nil
}

func assetFile(raw string) string {
	if u, err := url.Parse(raw); err == nil {
		return filepath.Base(u.Path)
	}
	return raw
}

// downloadVerified tải raw về dest, tính SHA256 trong lúc tải, quá max byte
// hoặc treo quá stallTimeout thì dừng. Sai mã → xoá file, errBadChecksum.
// progress(done, total) gọi mỗi khi có dữ liệu (người gọi tự giãn nhịp).
func downloadVerified(ctx context.Context, client *http.Client, raw, dest, wantHex string, max int64, progress func(done, total int64)) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	body, total, err := openDownload(ctx, client, raw)
	if err != nil {
		return err
	}
	defer body.Close()
	if total > max {
		return fmt.Errorf("file %s lớn bất thường (%d MB)", assetFile(raw), total>>20)
	}
	part := dest + ".dang-tai"
	f, err := os.OpenFile(part, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		f.Close()
		if !ok {
			os.Remove(part)
		}
	}()

	stall := time.AfterFunc(stallTimeout, func() { cancel(errors.New("mạng không có dữ liệu quá lâu")) })
	defer stall.Stop()
	h := sha256.New()
	buf := make([]byte, 64<<10)
	var done int64
	for {
		n, rerr := body.Read(buf)
		if n > 0 {
			stall.Reset(stallTimeout)
			done += int64(n)
			if done > max {
				return fmt.Errorf("file %s lớn hơn giới hạn %d MB", assetFile(raw), max>>20)
			}
			h.Write(buf[:n])
			if _, err := f.Write(buf[:n]); err != nil {
				return err
			}
			if progress != nil {
				progress(done, total)
			}
		}
		// Huỷ có hiệu lực ngay cả khi dữ liệu đã nằm sẵn trong bộ đệm.
		if ctx.Err() != nil {
			if c := context.Cause(ctx); c != nil && !errors.Is(c, context.Canceled) {
				return c
			}
			return ctx.Err()
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			if c := context.Cause(ctx); c != nil && !errors.Is(c, context.Canceled) {
				return c
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("tải %s bị ngắt: %w", assetFile(raw), rerr)
		}
	}
	if total > 0 && done != total {
		return fmt.Errorf("tải %s thiếu dữ liệu (%d/%d byte)", assetFile(raw), done, total)
	}
	if hex.EncodeToString(h.Sum(nil)) != wantHex {
		return errBadChecksum
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(part, dest); err != nil {
		return err
	}
	ok = true
	return nil
}

// updateDir — nơi để file cài tải về (thư mục cache của người dùng).
func updateDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "Sano", "cap-nhat"), nil
}

// runUpdateDownload: kiểm chữ ký SHA256SUMS trước (file nhỏ), rồi mới tải
// file cài và so SHA256. Trả đường dẫn file cài đã kiểm.
func runUpdateDownload(ctx context.Context, client *http.Client, plan updatePlan, dir string, progress func(done, total int64)) (string, error) {
	keys, err := parsePublicKeys(updatePublicKeys)
	if err != nil {
		return "", err
	}
	sums, err := fetchSmall(ctx, client, plan.sums.URL, maxSumsBytes)
	if err != nil {
		return "", err
	}
	sig, err := fetchSmall(ctx, client, plan.sig.URL, maxSigBytes)
	if err != nil {
		return "", err
	}
	if err := verifySums(sums, sig, keys); err != nil {
		return "", err
	}
	want, err := sumFor(sums, plan.asset.Name)
	if err != nil {
		return "", err
	}
	// Dọn file của lượt trước (bản cũ tải dở, bản đã cài).
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dest := filepath.Join(dir, plan.asset.Name)
	if err := downloadVerified(ctx, client, plan.asset.URL, dest, want, maxInstallerBytes, progress); err != nil {
		return "", err
	}
	return dest, nil
}

// ── Phần gắn vào giao diện ────────────────────────────────────────────────

// StartUpdate bắt đầu tải bản mới (bản đã báo ở lần CheckUpdate gần nhất).
// Tiến độ qua sự kiện update:progress. Đang tải / đã tải xong thì trả trạng thái.
func (a *App) StartUpdate() (UpdateStatus, error) {
	a.updMu.Lock()
	defer a.updMu.Unlock()
	if j := a.upd; j != nil && (j.status.Phase == "downloading" || j.status.Phase == "ready") {
		return j.status, nil
	}
	plan, err := planUpdate(a.updRelease, detectInstall())
	if err != nil {
		return UpdateStatus{}, err
	}
	dir, err := updateDir()
	if err != nil {
		return UpdateStatus{}, err
	}
	ctx, cancel := context.WithCancel(a.context())
	job := &updateJob{cancel: cancel, plan: plan, status: UpdateStatus{Phase: "downloading", Version: plan.version, Total: plan.asset.Size}}
	a.upd = job
	st := job.status
	go a.runUpdate(ctx, job, dir)
	return st, nil
}

func (a *App) runUpdate(ctx context.Context, job *updateJob, dir string) {
	var last time.Time
	progress := func(done, total int64) {
		if time.Since(last) < progressEvery && done != total {
			return
		}
		last = time.Now()
		a.updMu.Lock()
		job.status.Done = done
		if total > 0 {
			job.status.Total = total
		}
		st := job.status
		a.updMu.Unlock()
		a.emit(eventUpdateProgress, st)
	}
	file, err := runUpdateDownload(ctx, updateClient(), job.plan, dir, progress)
	a.updMu.Lock()
	switch {
	case ctx.Err() != nil:
		job.status = UpdateStatus{Phase: "idle", Version: job.plan.version}
	case err != nil:
		log.Printf("tự cập nhật: %v", err)
		job.status.Phase = "error"
		job.status.Error = err.Error()
	default:
		job.file = file
		job.status.Phase = "ready"
		job.status.Verified = true
		job.status.Done = job.status.Total
	}
	st := job.status
	a.updMu.Unlock()
	job.cancel()
	a.emit(eventUpdateProgress, st)
}

// CancelUpdate dừng lượt tải đang chạy (file tải dở bị xoá).
func (a *App) CancelUpdate() {
	a.updMu.Lock()
	defer a.updMu.Unlock()
	if a.upd != nil && a.upd.status.Phase == "downloading" {
		a.upd.cancel()
	}
}

// UpdateStatus trả tiến độ tự cập nhật hiện tại.
func (a *App) UpdateStatus() UpdateStatus {
	a.updMu.Lock()
	defer a.updMu.Unlock()
	if a.upd == nil {
		return UpdateStatus{Phase: "idle"}
	}
	return a.upd.status
}

// updateBlocked — lý do chưa được thay bản lúc này: thoát app sẽ làm hỏng
// việc đang chạy (render, cài bộ đọc, xuất M4B, gỡ bộ đọc).
func (a *App) updateBlocked() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch {
	case a.job != nil && a.job.status.Running:
		return errors.New("đang render sách — cập nhật sau khi render xong")
	case a.setup != nil && a.setup.inst.Status().Running:
		return errors.New("đang cài bộ đọc — cập nhật sau khi cài xong")
	case a.m4b != nil && a.m4b.status.Running:
		return errors.New("đang xuất M4B — cập nhật sau khi xuất xong")
	case a.uninstalling:
		return errUninstalling
	}
	return nil
}

func (a *App) readyUpdate() (*updateJob, error) {
	a.updMu.Lock()
	defer a.updMu.Unlock()
	if a.upd == nil || a.upd.status.Phase != "ready" || a.upd.file == "" {
		return nil, errors.New("chưa tải xong bản mới")
	}
	return a.upd, nil
}

// ApplyUpdate thay bản mới rồi thoát; bản mới tự mở lại. Đang render / cài bộ
// đọc / xuất M4B thì từ chối.
func (a *App) ApplyUpdate() error {
	if err := a.updateBlocked(); err != nil {
		return err
	}
	job, err := a.readyUpdate()
	if err != nil {
		return err
	}
	if err := applyUpdate(job.plan.inst, job.file, job.plan.version, true); err != nil {
		log.Printf("thay bản mới: %v", err)
		return err
	}
	if a.ctx != nil {
		wruntime.Quit(a.ctx)
	}
	return nil
}

// ApplyUpdateOnQuit — "Khởi động lại sau": thay bản mới khi người dùng thoát app.
func (a *App) ApplyUpdateOnQuit() (UpdateStatus, error) {
	job, err := a.readyUpdate()
	if err != nil {
		return UpdateStatus{}, err
	}
	a.updMu.Lock()
	defer a.updMu.Unlock()
	job.status.ApplyOnQuit = true
	return job.status, nil
}

// applyOnShutdown chạy lúc thoát app: đã hẹn thay bản thì thay (không mở lại).
func (a *App) applyOnShutdown() {
	job, err := a.readyUpdate()
	if err != nil || !job.status.ApplyOnQuit {
		return
	}
	if err := applyUpdate(job.plan.inst, job.file, job.plan.version, false); err != nil {
		log.Printf("thay bản mới lúc thoát: %v", err)
	}
}

// canWriteDir: tạo thử một file tạm trong dir (quyền ghi thật, kể cả ACL Windows).
func canWriteDir(dir string) bool {
	f, err := os.CreateTemp(dir, ".sano-kiem-ghi-*")
	if err != nil {
		return false
	}
	name := f.Name()
	f.Close()
	os.Remove(name)
	return true
}
