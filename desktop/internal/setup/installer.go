// Package setup cài bộ đọc VieNeu-TTS lần mở đầu vào thư mục dữ liệu của app:
// uv (GitHub Release astral-sh/uv) → Python do uv quản lý → mã VieNeu đúng commit
// (tarball GitHub) → thư viện `uv sync --frozen` → mô hình (`models.py fetch`, kiểm
// SHA256) → ffmpeg nếu máy chưa có → đọc thử một câu. Mọi thứ tải về đều ghim
// phiên bản trong scripts/tts/versions.env và kiểm SHA256. Huỷ được giữa chừng;
// chạy lại thì bỏ qua bước đã xong. Gỡ = xoá đúng thư mục <dữ liệu app>/tts.
package setup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"sano/desktop/internal/tts"
)

// Khoá từng dòng tiến độ trên màn cài (khớp giao diện).
const (
	StepPython = "python"
	StepVieNeu = "vieneu"
	StepModels = "models"
	StepFFmpeg = "ffmpeg"
	StepVerify = "verify"
)

// StepState — trạng thái một dòng.
type StepState string

const (
	StatePending StepState = "pending"
	StateRunning StepState = "running"
	StateDone    StepState = "done"
	StateSkipped StepState = "skipped" // đã có sẵn, không phải làm
	StateFailed  StepState = "error"
)

// Step — một dòng tiến độ.
type Step struct {
	Key    string    `json:"key"`
	Label  string    `json:"label"`
	State  StepState `json:"state"`
	Pct    float64   `json:"pct"`
	Detail string    `json:"detail"`
}

// Status — trạng thái lượt cài, đẩy lên giao diện qua sự kiện.
type Status struct {
	Running    bool    `json:"running"`
	Done       bool    `json:"done"`
	Cancelled  bool    `json:"cancelled"`
	Error      string  `json:"error"`
	Hint       string  `json:"hint"`
	Steps      []Step  `json:"steps"`
	ElapsedSec float64 `json:"elapsedSec"`
	EtaSec     float64 `json:"etaSec"` // -1 = chưa ước được
	LogFile    string  `json:"logFile"`
	// Seq tăng dần mỗi lần trạng thái đổi: sự kiện có thể tới giao diện không
	// theo thứ tự (Wails gửi mỗi sự kiện một goroutine), giao diện bỏ bản cũ hơn.
	Seq uint64 `json:"seq"`
}

// Config — đầu vào của Installer.
type Config struct {
	Layout     tts.Layout
	Pins       map[string]string
	Scripts    fs.FS // script đọc giọng nhúng trong app
	GOOS       string
	GOARCH     string
	HTTP       *http.Client
	FindFFmpeg func() string // ffmpeg có sẵn (máy hoặc app đã tải), "" nếu chưa có
	OnProgress func(Status)
}

// statusSeq — số thứ tự trạng thái, chung cho mọi lượt cài (lượt mới luôn lớn hơn).
var statusSeq atomic.Uint64

// Installer chạy một lượt cài. Tạo mới cho mỗi lượt.
type Installer struct {
	cfg   Config
	mu    sync.Mutex
	st    Status
	start time.Time
	log   *log.Logger
	logF  *os.File
}

// New tạo Installer với danh sách bước ban đầu.
func New(cfg Config) *Installer {
	if cfg.HTTP == nil {
		// Không đặt Timeout tổng (tải mô hình lâu); huỷ qua context.
		cfg.HTTP = &http.Client{}
	}
	return &Installer{cfg: cfg, st: InitialStatus(), log: log.New(io.Discard, "", 0)}
}

// InitialStatus — các dòng chưa chạy (giao diện hiện trước khi bấm Cài).
func InitialStatus() Status {
	return Status{EtaSec: -1, Steps: []Step{
		{Key: StepPython, Label: "Python", State: StatePending},
		{Key: StepVieNeu, Label: "Bộ đọc VieNeu-TTS", State: StatePending},
		{Key: StepModels, Label: "Mô hình giọng đọc (" + humanMB(ModelBytes) + ")", State: StatePending},
		{Key: StepFFmpeg, Label: "ffmpeg (ghi file MP3)", State: StatePending},
		{Key: StepVerify, Label: "Kiểm tra đọc thử", State: StatePending},
	}}
}

// Status trả bản sao trạng thái hiện tại.
func (in *Installer) Status() Status {
	in.mu.Lock()
	defer in.mu.Unlock()
	return in.snapshot()
}

func (in *Installer) snapshot() Status {
	in.st.Seq = statusSeq.Add(1)
	st := in.st
	st.Steps = append([]Step(nil), in.st.Steps...)
	if !in.start.IsZero() {
		st.ElapsedSec = time.Since(in.start).Seconds()
	}
	return st
}

// Run chạy trọn luồng cài. Lỗi → Status.Error + dòng bị lỗi; huỷ → Status.Cancelled.
func (in *Installer) Run(ctx context.Context) (err error) {
	in.mu.Lock()
	in.start = time.Now()
	in.st.Running = true
	in.st.LogFile = filepath.Join(in.cfg.Layout.Root, "install.log")
	in.mu.Unlock()
	in.emit()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("lỗi không mong muốn: %v", r)
			in.log.Printf("panic: %v\n%s", r, debug.Stack())
		}
		in.finish(ctx, err)
	}()

	if err = in.prepare(); err != nil {
		in.failRunning(err)
		return err
	}
	steps := []struct {
		key string
		fn  func(context.Context) error
	}{
		{StepPython, in.stepPython},
		{StepVieNeu, in.stepVieNeu},
		{StepModels, in.stepModels},
		{StepFFmpeg, in.stepFFmpeg},
		{StepVerify, in.stepVerify},
	}
	for _, s := range steps {
		if err = ctx.Err(); err != nil {
			return err
		}
		in.log.Printf("== bước %s", s.key)
		if err = s.fn(ctx); err != nil {
			if ctx.Err() == nil {
				in.update(s.key, func(st *Step) { st.State = StateFailed })
			}
			return err
		}
	}
	return nil
}

// prepare: tạo thư mục + file đánh dấu, giải nén script, mở nhật ký, kiểm đĩa trống.
func (in *Installer) prepare() error {
	l := in.cfg.Layout
	if err := tts.EnsureScripts(l, in.cfg.Scripts); err != nil {
		return fmt.Errorf("tạo thư mục bộ đọc %s: %w", l.Root, err)
	}
	f, err := os.OpenFile(filepath.Join(l.Root, "install.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err == nil {
		in.logF = f
		in.log = log.New(f, "", log.LstdFlags)
		in.log.Printf("===== cài bộ đọc (%s/%s) vào %s", in.cfg.GOOS, in.cfg.GOARCH, l.Root)
	}
	free, err := freeBytes(l.Root)
	if err != nil {
		return nil // không đo được thì bỏ qua, không chặn cài
	}
	need := RequiredBytes - DirSize(l.Root)
	if need < minFreeBytes {
		need = minFreeBytes
	}
	if int64(free) < need {
		return fmt.Errorf("ổ đĩa chỉ còn trống %s, cần khoảng %s để cài bộ đọc", humanGB(int64(free)), humanGB(need))
	}
	return nil
}

func (in *Installer) finish(ctx context.Context, err error) {
	in.mu.Lock()
	in.st.Running = false
	switch {
	case err == nil:
		in.st.Done = true
		in.st.EtaSec = 0
	case errors.Is(ctx.Err(), context.Canceled):
		in.st.Cancelled = true
		for i := range in.st.Steps {
			if in.st.Steps[i].State == StateRunning {
				in.st.Steps[i].State = StatePending
				in.st.Steps[i].Detail = "Đã huỷ"
			}
		}
	default:
		in.st.Error = err.Error()
		in.st.Hint = hintFor(err)
	}
	in.mu.Unlock()
	if err != nil {
		in.log.Printf("kết thúc: %v", err)
	} else {
		in.log.Printf("kết thúc: xong")
	}
	in.cleanupPartial()
	if in.logF != nil {
		_ = in.logF.Close()
	}
	in.emit()
}

// cleanupPartial xoá file tải dở + thư mục giải nén dở (huỷ hoặc lỗi). File
// .incomplete của mô hình giữ lại để lần sau tải tiếp.
func (in *Installer) cleanupPartial() {
	l := in.cfg.Layout
	_ = os.RemoveAll(l.Downloads())
	_ = os.RemoveAll(l.VieNeu() + ".tmp")
	_ = os.RemoveAll(filepath.Join(l.Root, "verify"))
	// Cache uv chỉ cần khi chưa cài xong thư viện (lệnh uv khác cũng tạo lại vài KB).
	if fileExists(filepath.Join(l.Venv(), syncedMarker)) {
		_ = os.RemoveAll(l.Cache())
	}
}

// hintFor — gợi ý sửa theo loại lỗi.
func hintFor(err error) string {
	msg := err.Error()
	switch {
	case errors.Is(err, ErrChecksum):
		return "File tải về bị sai (mạng chập chờn hoặc bị can thiệp). Bấm Thử lại; lặp lại nhiều lần thì báo lỗi cho Sano."
	case strings.Contains(msg, "ổ đĩa") || strings.Contains(msg, "no space left"):
		return "Giải phóng thêm dung lượng ổ đĩa rồi bấm Thử lại."
	case errors.Is(err, ErrNetwork) || looksOffline(msg):
		return "Kiểm tra kết nối mạng rồi bấm Thử lại — phần đã tải xong được giữ lại."
	}
	return "Bấm Thử lại. Vẫn lỗi thì gửi file nhật ký cài đặt khi báo lỗi."
}

// looksOffline nhận lỗi mạng trong output của uv / huggingface_hub.
func looksOffline(msg string) bool {
	m := strings.ToLower(msg)
	for _, s := range []string{"failed to fetch", "error sending request", "dns error", "timed out", "connectionerror", "max retries exceeded", "network is unreachable", "temporary failure in name resolution"} {
		if strings.Contains(m, s) {
			return true
		}
	}
	return false
}

// ── tiện ích cập nhật trạng thái ──────────────────────────────────────────

func (in *Installer) update(key string, fn func(*Step)) {
	in.mu.Lock()
	for i := range in.st.Steps {
		if in.st.Steps[i].Key == key {
			fn(&in.st.Steps[i])
		}
	}
	in.mu.Unlock()
	in.emit()
}

// running đánh dấu dòng đang chạy, đặt lại % (bắt đầu bước hoặc chuyển bước con).
func (in *Installer) running(key string, pct float64, detail string) {
	in.update(key, func(s *Step) { s.State, s.Pct, s.Detail = StateRunning, clampPct(pct), detail })
}

// setPct cập nhật % (chỉ tăng) + dòng chi tiết.
func (in *Installer) setPct(key string, pct float64, detail string) {
	in.update(key, func(s *Step) {
		if p := clampPct(pct); p > s.Pct {
			s.Pct = p
		}
		if detail != "" {
			s.Detail = detail
		}
	})
}

func (in *Installer) done(key, detail string) {
	in.update(key, func(s *Step) { s.State, s.Pct, s.Detail = StateDone, 100, detail })
}

func (in *Installer) skip(key, detail string) {
	in.log.Printf("bỏ qua %s: %s", key, detail)
	in.update(key, func(s *Step) { s.State, s.Pct, s.Detail = StateSkipped, 100, detail })
}

// failRunning — lỗi trước khi vào bước nào (thư mục, đĩa đầy): gắn vào dòng đầu.
func (in *Installer) failRunning(err error) {
	in.update(StepPython, func(s *Step) { s.State, s.Detail = StateFailed, err.Error() })
}

func (in *Installer) setEta(sec float64) {
	in.mu.Lock()
	in.st.EtaSec = sec
	in.mu.Unlock()
}

func (in *Installer) emit() {
	if in.cfg.OnProgress == nil {
		return
	}
	in.cfg.OnProgress(in.Status())
}

func clampPct(p float64) float64 {
	switch {
	case p < 0:
		return 0
	case p > 100:
		return 100
	}
	return p
}

// ── chạy lệnh ─────────────────────────────────────────────────────────────

// baseEnv — môi trường hiện tại bỏ các biến làm lệch uv/python/HF (venv đang
// bật, cấu hình uv riêng của máy, PYTHONPATH...), để cài ra đúng như ghim.
func baseEnv() []string {
	drop := []string{"UV_", "VIRTUAL_ENV", "CONDA_PREFIX", "PYTHONHOME", "PYTHONPATH", "PYTHONSTARTUP", "HF_"}
	var out []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		skip := false
		for _, p := range drop {
			if strings.HasPrefix(strings.ToUpper(k), p) {
				skip = true
				break
			}
		}
		if !skip {
			out = append(out, kv)
		}
	}
	return out
}

// uvEnv — biến môi trường cho uv: mọi thứ nằm trong thư mục bộ đọc, không đọc
// cấu hình uv của máy, không ghi Python ra ~/.local/bin.
func (in *Installer) uvEnv() []string {
	l := in.cfg.Layout
	return append(baseEnv(),
		"UV_CACHE_DIR="+l.Cache(),
		"UV_PYTHON_INSTALL_DIR="+l.PythonInstalls(),
		"UV_PYTHON_BIN_DIR="+filepath.Join(l.PythonInstalls(), "bin"),
		"UV_PYTHON_PREFERENCE=only-managed",
		"UV_PYTHON_DOWNLOADS=automatic",
		"UV_NO_CONFIG=1",
		"UV_LINK_MODE=copy", // xoá cache sau khi cài không ảnh hưởng venv
		"UV_NO_PROGRESS=1",
	)
}

// pyEnv — biến môi trường cho python của bộ đọc app cài.
func (in *Installer) pyEnv(extra ...string) []string {
	env := append(baseEnv(), in.cfg.Layout.Env()...)
	env = append(env, "PYTHONPATH="+in.cfg.Layout.Scripts(), "PYTHONUNBUFFERED=1")
	return append(env, extra...)
}

// run chạy lệnh, ghi output vào nhật ký, trả vài dòng cuối nếu lỗi.
func (in *Installer) run(ctx context.Context, dir string, env []string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	tts.HideWindow(cmd)
	cmd.WaitDelay = 5 * time.Second
	var tail tailWriter
	w := io.MultiWriter(&tail, in.log.Writer())
	cmd.Stdout, cmd.Stderr = w, w
	in.log.Printf("$ %s %s", filepath.Base(name), strings.Join(args, " "))
	err := cmd.Run()
	if err != nil {
		if ctx.Err() != nil {
			return tail.String(), ctx.Err()
		}
		return tail.String(), fmt.Errorf("%s: %w\n%s", filepath.Base(name), err, tail.String())
	}
	return tail.String(), nil
}

// tailWriter giữ ~20 dòng cuối output để báo lỗi.
type tailWriter struct {
	mu    sync.Mutex
	lines []string
	cur   strings.Builder
}

func (t *tailWriter) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, b := range p {
		if b == '\n' || b == '\r' {
			if t.cur.Len() > 0 {
				t.lines = append(t.lines, t.cur.String())
				if len(t.lines) > 20 {
					t.lines = t.lines[1:]
				}
				t.cur.Reset()
			}
			continue
		}
		t.cur.WriteByte(b)
	}
	return len(p), nil
}

func (t *tailWriter) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	lines := t.lines
	if t.cur.Len() > 0 {
		lines = append(append([]string(nil), lines...), t.cur.String())
	}
	return strings.Join(lines, "\n")
}

// watchSize báo % theo dung lượng thư mục tăng dần (bước chạy lệnh ngoài không
// báo tiến độ). base = dung lượng lúc bắt đầu; cap = % tối đa khi chưa xong.
func (in *Installer) watchSize(ctx context.Context, key string, from, to float64, expected int64, detail func(int64) string, paths ...string) func() {
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				in.log.Printf("watchSize panic: %v", r)
			}
		}()
		base := DirSize(paths...)
		t := time.NewTicker(700 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				got := DirSize(paths...) - base
				frac := float64(got) / float64(expected)
				if frac > 0.97 {
					frac = 0.97
				}
				d := ""
				if detail != nil {
					d = detail(got)
				}
				in.setPct(key, from+(to-from)*frac, d)
			}
		}
	}()
	return func() { cancel(); wg.Wait() }
}
