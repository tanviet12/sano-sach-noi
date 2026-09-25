package setup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sano/desktop/internal/tts"
	ttsscripts "sano/scripts/tts"
)

// Kích thước đo trên máy dev (macOS arm64, 09/2026) — chỉ để báo trước dung lượng
// và ước % tiến độ; không dùng để kiểm tính đúng (đã có SHA256).
const (
	ModelBytes    int64 = 608_636_118      // tổng file mô hình ghim (backbone + codec)
	pythonBytes   int64 = 72 << 20         // Python do uv cài (đã giải nén)
	libBytes      int64 = 1_150 << 20      // cache uv + venv lúc uv sync
	RequiredBytes int64 = 2_500 << 20      // cần trống trước khi cài (tính cả cache uv tạm)
	InstalledSize int64 = 1_500 << 20      // dung lượng sau khi cài xong (gồm ffmpeg nếu phải tải)
	DownloadBytes int64 = 1_000 << 20      // tổng tải về (uv + Python + thư viện + mô hình + ffmpeg)
	minFreeBytes  int64 = 300 << 20        // chạy lại khi gần xong vẫn cần tối thiểu chừng này
	verifyWait          = 15 * time.Second // đọc thử một câu: nạp mô hình + đọc (~5 giây trên M-series)
)

// Giới hạn tải về. uv, ffmpeg đã kiểm SHA256 ngay khi tải xong; tarball VieNeu
// chỉ kiểm được tree hash sau khi giải nén, nên giới hạn chặt hơn (bản ghim hiện
// ~3 MB nén, ~6 MB và 165 file sau giải nén).
const (
	maxArtifactBytes int64 = 512 << 20 // uv, ffmpeg (ffmpeg Linux ~125 MB)
	maxVieNeuBytes   int64 = 64 << 20
)

// verifySentence — câu đọc thử ở bước kiểm tra cuối.
const verifySentence = "Xin chào, đây là Sano đang đọc thử giọng tiếng Việt."

// ── Python (uv + Python do uv quản lý) ────────────────────────────────────

func (in *Installer) stepPython(ctx context.Context) error {
	l, pins := in.cfg.Layout, in.cfg.Pins
	uvVer, pyVer := pins["UV_VERSION"], pins["PYTHON_VERSION"]
	uvOK := in.uvReady(ctx, uvVer)
	if uvOK && in.pythonReady(ctx, pyVer) {
		in.skip(StepPython, "Đã có Python "+pyVer)
		return nil
	}
	if !uvOK {
		if err := in.installUV(ctx, uvVer); err != nil {
			return err
		}
	}
	in.running(StepPython, 25, "Tải và cài Python "+pyVer)
	stop := in.watchSize(ctx, StepPython, 25, 99, pythonBytes, nil, l.PythonInstalls())
	_, err := in.run(ctx, l.Root, in.uvEnv(), l.UV(), "python", "install", pyVer, "--no-bin", "--no-registry")
	stop()
	if err != nil {
		return fmt.Errorf("cài Python %s: %w", pyVer, err)
	}
	if !in.pythonReady(ctx, pyVer) {
		return fmt.Errorf("cài Python %s xong nhưng không tìm thấy", pyVer)
	}
	in.done(StepPython, "Python "+pyVer)
	return nil
}

func (in *Installer) installUV(ctx context.Context, ver string) error {
	l := in.cfg.Layout
	a, err := uvArtifact(in.cfg.Pins, in.cfg.GOOS, in.cfg.GOARCH)
	if err != nil {
		return err
	}
	in.running(StepPython, 0, "Tải công cụ cài đặt uv "+ver)
	archive := filepath.Join(l.Downloads(), "uv."+a.Kind)
	_, err = download(ctx, in.cfg.HTTP, a.URL, archive, a.SHA256, maxArtifactBytes, func(d, t int64) {
		in.setPct(StepPython, 25*frac(d, t), "Tải uv "+ver+" · "+progressText(d, t))
	})
	if err != nil {
		return err
	}
	in.log.Printf("uv: SHA256 khớp %s", a.SHA256)
	tmp := l.UV() + ".tmp"
	if a.Kind == "zip" {
		err = extractOneFromZip(archive, a.Member, tmp)
	} else {
		err = extractOneFromTarGz(archive, a.Member, tmp)
	}
	if err == nil {
		err = os.Rename(tmp, l.UV())
	}
	_ = os.Remove(archive)
	if err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("giải nén uv: %w", err)
	}
	if !in.uvReady(ctx, ver) {
		return fmt.Errorf("uv %s tải về không chạy được", ver)
	}
	return nil
}

func (in *Installer) uvReady(ctx context.Context, ver string) bool {
	uv := in.cfg.Layout.UV()
	if !fileExists(uv) {
		return false
	}
	out, err := in.run(ctx, in.cfg.Layout.Root, in.uvEnv(), uv, "--version")
	return err == nil && strings.HasPrefix(strings.TrimSpace(out), "uv "+ver)
}

func (in *Installer) pythonReady(ctx context.Context, ver string) bool {
	l := in.cfg.Layout
	if !fileExists(l.UV()) {
		return false
	}
	_, err := in.run(ctx, l.Root, in.uvEnv(), l.UV(), "python", "find", ver)
	return err == nil
}

// ── Mã VieNeu-TTS + thư viện ─────────────────────────────────────────────

// syncedWant — nội dung dấu "đã sync" (<venv>/.sano-synced). Gồm cả mã băm
// pyproject/uv.lock của Sano: đổi ghi đè thư viện (vd vá bảo mật) thì máy đã cài
// bị coi là cần sync lại (NeedsResync).
func syncedWant(pins map[string]string, projectSum string) string {
	return pins["VIENEU_COMMIT"] + "\n" + pins["PYTHON_VERSION"] + "\n" + pins["UV_VERSION"] + "\n" + projectSum + "\n"
}

// NeedsResync báo bộ đọc app đã cài (có python trong venv) nhưng thư viện chưa
// khớp bản Sano này — cần chạy lại bước cài (chỉ sync thư viện, mô hình giữ nguyên).
func NeedsResync(l tts.Layout, pins map[string]string, scripts fs.FS) bool {
	if !fileExists(l.Python()) {
		return false
	}
	_, sum, err := vieneuProject(scripts)
	if err != nil {
		return false
	}
	return readRaw(filepath.Join(l.Venv(), syncedMarker)) != syncedWant(pins, sum)
}

// vieneuProject đọc pyproject.toml + uv.lock Sano nhúng (scripts/tts/vieneu-project/),
// trả kèm SHA256 chung của hai file.
func vieneuProject(scripts fs.FS) (map[string][]byte, string, error) {
	if scripts == nil {
		return nil, "", errors.New("bản này không kèm pyproject.toml/uv.lock cho VieNeu-TTS")
	}
	files := map[string][]byte{}
	h := sha256.New()
	for _, name := range ttsscripts.VieNeuProjectFiles {
		data, err := fs.ReadFile(scripts, ttsscripts.VieNeuDir+"/"+name)
		if err != nil {
			return nil, "", fmt.Errorf("thiếu %s/%s nhúng trong app: %w", ttsscripts.VieNeuDir, name, err)
		}
		files[name] = data
		fmt.Fprintf(h, "%s %d\n", name, len(data))
		h.Write(data)
	}
	return files, hex.EncodeToString(h.Sum(nil)), nil
}

const (
	commitMarker = ".sano-commit"
	syncedMarker = ".sano-synced"
)

func (in *Installer) stepVieNeu(ctx context.Context) error {
	l, pins := in.cfg.Layout, in.cfg.Pins
	url, commit, treeSHA, err := vieneuTarball(pins)
	if err != nil {
		return err
	}
	short := commit[:7]
	project, projectSum, err := vieneuProject(in.cfg.Scripts)
	if err != nil {
		return err
	}
	wantSynced := syncedWant(pins, projectSum)
	srcOK := readTrim(filepath.Join(l.VieNeu(), commitMarker)) == commit
	if srcOK && fileExists(l.Python()) && readRaw(filepath.Join(l.Venv(), syncedMarker)) == wantSynced {
		in.skip(StepVieNeu, "Đã cài VieNeu-TTS (commit "+short+")")
		return nil
	}
	if !srcOK {
		in.running(StepVieNeu, 0, "Tải mã nguồn VieNeu-TTS (commit "+short+")")
		archive := filepath.Join(l.Downloads(), "vieneu.tar.gz")
		if _, err := download(ctx, in.cfg.HTTP, url, archive, "", maxVieNeuBytes, func(d, t int64) {
			in.setPct(StepVieNeu, 8*frac(d, t), "Tải mã nguồn VieNeu-TTS · "+progressText(d, t))
		}); err != nil {
			return err
		}
		tmp := l.VieNeu() + ".tmp"
		_ = os.RemoveAll(tmp)
		got, err := extractTarGz(ctx, archive, tmp)
		_ = os.Remove(archive)
		if err != nil {
			_ = os.RemoveAll(tmp)
			return fmt.Errorf("giải nén mã VieNeu-TTS: %w", err)
		}
		if got != treeSHA {
			_ = os.RemoveAll(tmp)
			return fmt.Errorf("%w (mã VieNeu-TTS: tree hash %s, cần %s)", ErrChecksum, got[:12], treeSHA[:12])
		}
		in.log.Printf("VieNeu: tree hash khớp %s", treeSHA)
		if err := os.RemoveAll(l.VieNeu()); err != nil {
			return err
		}
		if err := os.Rename(tmp, l.VieNeu()); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(l.VieNeu(), commitMarker), []byte(commit+"\n"), 0o644); err != nil {
			return err
		}
	}

	// Chép pyproject.toml + uv.lock của Sano đè lên bản của VieNeu (sau khi đã
	// kiểm tree hash mã gốc): bỏ giao diện web gradio, ép bản vá bảo mật.
	for name, data := range project {
		if err := os.WriteFile(filepath.Join(l.VieNeu(), name), data, 0o644); err != nil {
			return fmt.Errorf("ghi %s cho VieNeu-TTS: %w", name, err)
		}
	}

	in.running(StepVieNeu, 10, "Cài thư viện Python (onnxruntime, numpy…)")
	stop := in.watchSize(ctx, StepVieNeu, 10, 99, libBytes, func(n int64) string {
		return "Cài thư viện Python · " + humanMB(n)
	}, l.Cache(), l.Venv())
	_, err = in.run(ctx, l.VieNeu(), in.uvEnv(), l.UV(), "sync", "--frozen", "--no-dev", "--python", pins["PYTHON_VERSION"])
	stop()
	if err != nil {
		return fmt.Errorf("cài thư viện của VieNeu-TTS: %w", err)
	}
	if err := os.WriteFile(filepath.Join(l.Venv(), syncedMarker), []byte(wantSynced), 0o644); err != nil {
		return err
	}
	// Cache uv chỉ cần lúc cài (UV_LINK_MODE=copy nên venv không phụ thuộc cache).
	_ = os.RemoveAll(l.Cache())
	in.done(StepVieNeu, "VieNeu-TTS (commit "+short+")")
	return nil
}

// ── Mô hình ──────────────────────────────────────────────────────────────

const verifiedMarker = ".sano-verified"

func (in *Installer) stepModels(ctx context.Context) error {
	l, pins := in.cfg.Layout, in.cfg.Pins
	py, models := l.Python(), filepath.Join(l.Scripts(), "models.py")
	want := pins["HF_BACKBONE_REVISION"] + "\n" + pins["HF_CODEC_REVISION"] + "\n"
	marker := filepath.Join(l.HFHome(), verifiedMarker)
	env := in.pyEnv("HF_HUB_DISABLE_XET=1", "HF_HUB_DISABLE_PROGRESS_BARS=1")

	if _, err := in.run(ctx, l.Root, env, py, models, "check"); err == nil {
		if readRaw(marker) == want {
			in.skip(StepModels, "Đủ mô hình, đã kiểm SHA256")
			return nil
		}
		in.running(StepModels, 90, "Kiểm SHA256 mô hình đã có")
		if _, err := in.run(ctx, l.Root, env, py, models, "verify"); err == nil {
			return in.modelsDone(marker, want)
		}
	}

	hub := filepath.Join(l.HFHome(), "hub")
	have := DirSize(hub)
	in.running(StepModels, 100*float64(have)/float64(ModelBytes), "Tải mô hình từ Hugging Face")
	start := time.Now()
	stop := in.watchSize(ctx, StepModels, 100*float64(have)/float64(ModelBytes), 99, ModelBytes-have, func(n int64) string {
		done := have + n
		if sec := time.Since(start).Seconds(); sec > 3 && n > 0 {
			rate := float64(n) / sec
			in.setEta((float64(ModelBytes-done))/rate + verifyWait.Seconds())
		}
		return "Tải mô hình · " + humanMB(done) + " / " + humanMB(ModelBytes)
	}, hub)
	_, err := in.run(ctx, l.Root, env, py, models, "fetch")
	stop()
	in.setEta(-1)
	if err != nil {
		return fmt.Errorf("tải mô hình giọng đọc: %w", err)
	}
	return in.modelsDone(marker, want)
}

func (in *Installer) modelsDone(marker, want string) error {
	if err := os.WriteFile(marker, []byte(want), 0o644); err != nil {
		return err
	}
	in.done(StepModels, "Đủ mô hình, đã kiểm SHA256")
	return nil
}

// ── ffmpeg ────────────────────────────────────────────────────────────────

func (in *Installer) stepFFmpeg(ctx context.Context) error {
	if p := in.findFFmpeg(); p != "" {
		in.skip(StepFFmpeg, "Có sẵn trên máy: "+p)
		return nil
	}
	l := in.cfg.Layout
	a, err := ffmpegArtifact(in.cfg.Pins, in.cfg.GOOS, in.cfg.GOARCH)
	if err != nil {
		return fmt.Errorf("%v. %s", err, tts.FFmpegHint(in.cfg.GOOS))
	}
	in.running(StepFFmpeg, 0, "Tải ffmpeg")
	archive := filepath.Join(l.Downloads(), "ffmpeg."+a.Kind)
	if _, err := download(ctx, in.cfg.HTTP, a.URL, archive, a.SHA256, maxArtifactBytes, func(d, t int64) {
		in.setPct(StepFFmpeg, 90*frac(d, t), "Tải ffmpeg · "+progressText(d, t))
	}); err != nil {
		return err
	}
	in.log.Printf("ffmpeg: SHA256 khớp %s", a.SHA256)
	tmp := l.FFmpeg() + ".tmp"
	switch a.Kind {
	case "zip":
		err = extractOneFromZip(archive, a.Member, tmp)
	case "tar.xz":
		err = extractOneFromTarXz(archive, a.Member, tmp)
	default:
		err = gunzipTo(archive, tmp)
	}
	_ = os.Remove(archive)
	if err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("giải nén ffmpeg: %w", err)
	}
	in.setPct(StepFFmpeg, 95, "Kiểm tra ffmpeg")
	out, err := in.run(ctx, l.Root, baseEnv(), tmp, "-hide_banner", "-h", "encoder=libmp3lame")
	if err != nil || !strings.Contains(out, "Encoder libmp3lame") {
		_ = os.Remove(tmp)
		if err == nil {
			err = fmt.Errorf("ffmpeg tải về không có bộ mã MP3 (libmp3lame)")
		}
		return err
	}
	if err := os.Rename(tmp, l.FFmpeg()); err != nil {
		return err
	}
	in.done(StepFFmpeg, "Đã tải ffmpeg (bản dựng tĩnh GPL)")
	return nil
}

func (in *Installer) findFFmpeg() string {
	if in.cfg.FindFFmpeg == nil {
		return ""
	}
	return in.cfg.FindFFmpeg()
}

// ── Kiểm tra cuối: mô hình + ffmpeg + đọc thử một câu ─────────────────────

func (in *Installer) stepVerify(ctx context.Context) error {
	l := in.cfg.Layout
	py := l.Python()
	in.running(StepVerify, 0, "Kiểm tra mô hình")
	if _, err := in.run(ctx, l.Root, in.pyEnv(), py, filepath.Join(l.Scripts(), "models.py"), "check"); err != nil {
		return fmt.Errorf("kiểm tra mô hình: %w", err)
	}
	ff := in.findFFmpeg()
	if ff == "" {
		return fmt.Errorf("không tìm thấy ffmpeg. %s", tts.FFmpegHint(in.cfg.GOOS))
	}
	if _, err := in.run(ctx, l.Root, baseEnv(), ff, "-hide_banner", "-version"); err != nil {
		return fmt.Errorf("ffmpeg không chạy được: %w", err)
	}

	in.running(StepVerify, 15, "Đọc thử một câu")
	dir := filepath.Join(l.Root, "verify")
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	if err := os.WriteFile(filepath.Join(dir, "thu.txt"), []byte(verifySentence+"\n"), 0o644); err != nil {
		return err
	}
	start := time.Now()
	stop := in.watchTime(ctx, StepVerify, 15, 97, verifyWait)
	_, err := in.run(ctx, dir, in.pyEnv(), py, filepath.Join(l.Scripts(), "audio_gen_batch.py"), "thu.txt")
	stop()
	if err != nil {
		return fmt.Errorf("đọc thử: %w", err)
	}
	info, err := os.Stat(filepath.Join(dir, "thu_full.wav"))
	if err != nil || info.Size() < 10_000 {
		return fmt.Errorf("đọc thử không ra file âm thanh")
	}
	in.done(StepVerify, fmt.Sprintf("Đọc thử được (%.0f giây)", time.Since(start).Seconds()))
	return nil
}

// watchTime tăng % theo thời gian (bước không đo được tiến độ).
func (in *Installer) watchTime(ctx context.Context, key string, from, to float64, expected time.Duration) func() {
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				in.log.Printf("watchTime panic: %v", r)
			}
		}()
		start := time.Now()
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				f := time.Since(start).Seconds() / expected.Seconds()
				if f > 0.97 {
					f = 0.97
				}
				in.setPct(key, from+(to-from)*f, "")
			}
		}
	}()
	return func() { cancel(); wg.Wait() }
}

// ── tiện ích ─────────────────────────────────────────────────────────────

func frac(done, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(done) / float64(total)
}

func progressText(done, total int64) string {
	if total <= 0 {
		return humanMB(done)
	}
	return humanMB(done) + " / " + humanMB(total)
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func readRaw(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}

func readTrim(p string) string { return strings.TrimSpace(readRaw(p)) }
