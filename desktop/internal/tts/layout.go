package tts

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// EnvDataDir đổi thư mục dữ liệu của app (thử cài bộ đọc vào thư mục tạm, hoặc
// máy muốn để bộ đọc ở ổ khác).
const EnvDataDir = "SANO_DATA_DIR"

// RuntimeDirName — thư mục con chứa bộ đọc trong thư mục dữ liệu app. Gỡ bộ đọc =
// xoá đúng thư mục này (xem setup.Uninstall).
const RuntimeDirName = "tts"

// MarkerFile — file đánh dấu thư mục bộ đọc do Sano tạo. Không có file này thì
// Sano không xoá thư mục.
const MarkerFile = ".sano-tts"

// MarkerContent — nội dung file đánh dấu (dòng đầu).
const MarkerContent = "sano-tts: thư mục bộ đọc do phần mềm Sano tạo, gỡ được trong Cài đặt"

// DataDir — thư mục dữ liệu của app theo hệ điều hành:
//
//	macOS   ~/Library/Application Support/Sano
//	Windows %LOCALAPPDATA%\Sano
//	Linux   $XDG_DATA_HOME/sano (mặc định ~/.local/share/sano)
//
// Biến SANO_DATA_DIR (đường dẫn tuyệt đối) được ưu tiên.
func DataDir(getenv func(string) string, goos, home string) (string, error) {
	if d := strings.TrimSpace(getenv(EnvDataDir)); d != "" {
		if !filepath.IsAbs(d) {
			return "", errors.New(EnvDataDir + " phải là đường dẫn tuyệt đối")
		}
		return filepath.Clean(d), nil
	}
	switch goos {
	case "darwin":
		if home == "" {
			return "", errors.New("không xác định được thư mục người dùng")
		}
		return filepath.Join(home, "Library", "Application Support", "Sano"), nil
	case "windows":
		if d := strings.TrimSpace(getenv("LOCALAPPDATA")); d != "" {
			return filepath.Join(d, "Sano"), nil
		}
		if home == "" {
			return "", errors.New("không xác định được thư mục người dùng")
		}
		return filepath.Join(home, "AppData", "Local", "Sano"), nil
	default:
		if d := strings.TrimSpace(getenv("XDG_DATA_HOME")); d != "" && filepath.IsAbs(d) {
			return filepath.Join(d, "sano"), nil
		}
		if home == "" {
			return "", errors.New("không xác định được thư mục người dùng")
		}
		return filepath.Join(home, ".local", "share", "sano"), nil
	}
}

// Layout — vị trí từng thành phần bộ đọc trong <dữ liệu app>/tts. Mọi thứ Sano
// tải về đều nằm trong Root để gỡ sạch được bằng một lần xoá.
type Layout struct {
	Root string
	GOOS string
}

// NewLayout dựng Layout từ thư mục dữ liệu app.
func NewLayout(dataDir, goos string) Layout {
	return Layout{Root: filepath.Join(dataDir, RuntimeDirName), GOOS: goos}
}

func (l Layout) exe(name string) string {
	if l.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// Marker — file đánh dấu thư mục do Sano tạo.
func (l Layout) Marker() string { return filepath.Join(l.Root, MarkerFile) }

// UV — file chạy uv.
func (l Layout) UV() string { return filepath.Join(l.Root, "bin", l.exe("uv")) }

// PythonInstalls — nơi uv đặt Python (UV_PYTHON_INSTALL_DIR).
func (l Layout) PythonInstalls() string { return filepath.Join(l.Root, "python") }

// Cache — cache của uv trong lúc cài (xoá sau khi cài thư viện xong).
func (l Layout) Cache() string { return filepath.Join(l.Root, "uv-cache") }

// VieNeu — mã nguồn VieNeu-TTS đúng commit ghim.
func (l Layout) VieNeu() string { return filepath.Join(l.Root, "vieneu") }

// Venv — môi trường Python của VieNeu (uv sync tạo).
func (l Layout) Venv() string { return filepath.Join(l.VieNeu(), ".venv") }

// Python — python trong venv.
func (l Layout) Python() string { return VenvPython(l.Venv(), l.GOOS) }

// Scripts — script đọc giọng giải nén từ bản nhúng trong app.
func (l Layout) Scripts() string { return filepath.Join(l.Root, "scripts") }

// HFHome — HF_HOME riêng của app: mô hình tải về nằm trong Root.
func (l Layout) HFHome() string { return filepath.Join(l.Root, "hf") }

// FFmpeg — ffmpeg Sano tự tải khi máy chưa có.
func (l Layout) FFmpeg() string { return filepath.Join(l.Root, "ffmpeg", l.exe("ffmpeg")) }

// Downloads — file đang tải dở.
func (l Layout) Downloads() string { return filepath.Join(l.Root, "downloads") }

// Env — biến môi trường khi chạy python của bộ đọc do app cài: mô hình đọc từ
// HF_HOME riêng, in UTF-8 (console Windows), không gửi thống kê.
func (l Layout) Env() []string {
	return []string{
		"HF_HOME=" + l.HFHome(),
		"HF_HUB_CACHE=" + filepath.Join(l.HFHome(), "hub"), // đè HF_HUB_CACHE người dùng đặt sẵn (nếu có)
		"HF_HUB_DISABLE_TELEMETRY=1",
		"PYTHONUTF8=1",
		"PYTHONIOENCODING=utf-8",
	}
}

// EnsureRoot tạo thư mục bộ đọc + file đánh dấu.
func (l Layout) EnsureRoot() error {
	if err := os.MkdirAll(l.Root, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(l.Marker()); err == nil {
		return nil
	}
	return os.WriteFile(l.Marker(), []byte(MarkerContent+"\n"), 0o644)
}
