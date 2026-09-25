// Package tts tìm và kiểm tra bộ đọc VieNeu-TTS trên máy người dùng: python của
// venv VieNeu, script scripts/tts của Sano, mô hình ghim, ffmpeg; và vị trí từng
// thành phần trong thư mục dữ liệu app (Layout). Việc tải/cài nằm ở package setup.
package tts

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// VenvDir — thư mục VieNeu-TTS v3 cài tay (tính từ HOME). Cùng quy ước với
// cmd/sano-docx2tts (~/VieNeu-TTS-v3/.venv). Giữ để máy đã cài theo
// docs/tts-build-guide.md dùng tiếp được; Sano không sửa gì trong đó.
const VenvDir = "VieNeu-TTS-v3"

// EnvPython cho phép chỉ định python khác mặc định (vd venv ở ổ khác).
const EnvPython = "SANO_TTS_PYTHON"

// EnvScripts cho phép chỉ định thư mục chứa models.py + audio_gen_batch.py (dev).
const EnvScripts = "SANO_TTS_SCRIPTS"

// checkTimeout — giới hạn cho mỗi lệnh python (khởi động python lần đầu có thể chậm).
const checkTimeout = 30 * time.Second

// Status là kết quả kiểm tra bộ đọc, trả thẳng lên giao diện.
type Status struct {
	Ready         bool   `json:"ready"`
	Python        string `json:"python"`
	PythonFound   bool   `json:"pythonFound"`
	PythonVersion string `json:"pythonVersion"`
	ScriptsDir    string `json:"scriptsDir"`
	ModelsOK      bool   `json:"modelsOk"`
	FFmpeg        string `json:"ffmpeg"`
	Source        string `json:"source"`  // env | app | legacy
	DataDir       string `json:"dataDir"` // thư mục bộ đọc app cài (<dữ liệu app>/tts)
	// Update — bộ đọc app cài chạy được (Ready) nhưng thư viện Python chưa khớp
	// bản Sano này (vd bản vá bảo mật): giao diện mời chạy lại màn cài để sync.
	Update  bool   `json:"update"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

// VenvPython trả đường dẫn python trong venv theo hệ điều hành.
func VenvPython(venv, goos string) string {
	if goos == "windows" {
		return filepath.Join(venv, "Scripts", "python.exe")
	}
	return filepath.Join(venv, "bin", "python")
}

// FindScriptsDir tìm thư mục scripts/tts (có models.py) theo thứ tự: biến môi
// trường SANO_TTS_SCRIPTS, rồi đi ngược lên từ từng thư mục gốc (thư mục hiện
// tại, thư mục chứa file chạy). Đi ngược lên để chạy được cả khi `wails dev`
// (cwd = desktop/) lẫn bản .app build trong repo (desktop/build/bin/Sano.app/...).
func FindScriptsDir(getenv func(string) string, starts ...string) string {
	if d := strings.TrimSpace(getenv(EnvScripts)); d != "" {
		if fileExists(filepath.Join(d, "models.py")) {
			return d
		}
	}
	for _, start := range starts {
		if start == "" {
			continue
		}
		dir, err := filepath.Abs(start)
		if err != nil {
			continue
		}
		for {
			cand := filepath.Join(dir, "scripts", "tts")
			if fileExists(filepath.Join(cand, "models.py")) {
				return cand
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return ""
}

// Runner chạy một lệnh (thêm env vào biến môi trường hiện có) và trả output gộp
// stdout+stderr. Tách ra để test.
type Runner func(ctx context.Context, env []string, name string, args ...string) (string, error)

// ExecRunner chạy lệnh thật, không bật cửa sổ console trên Windows.
func ExecRunner(ctx context.Context, env []string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	HideWindow(cmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

// Check kiểm tra python + script + mô hình. Không bao giờ panic; lỗi nằm trong Status.
func Check(ctx context.Context, rt Runtime, run Runner) Status {
	st := Status{Python: rt.Python, ScriptsDir: rt.ScriptsDir, Source: rt.Source}

	if rt.Python == "" || !fileExists(rt.Python) {
		st.Message = "Chưa cài bộ đọc"
		st.Detail = "Không thấy Python của VieNeu-TTS tại " + orDash(rt.Python)
		return st
	}
	st.PythonFound = true

	vctx, cancel := context.WithTimeout(ctx, checkTimeout)
	ver, err := run(vctx, rt.Env, rt.Python, "--version")
	cancel()
	if err != nil {
		st.Message = "Python của bộ đọc không chạy được"
		st.Detail = firstLine(ver, err)
		return st
	}
	st.PythonVersion = strings.TrimSpace(strings.TrimPrefix(ver, "Python "))

	if rt.ScriptsDir == "" {
		st.Message = "Thiếu script đọc giọng của Sano"
		st.Detail = "Không thấy thư mục scripts/tts (có models.py). Đặt biến " + EnvScripts + " trỏ tới thư mục đó."
		if rt.ScriptsErr != "" {
			st.Detail = rt.ScriptsErr
		}
		return st
	}

	mctx, cancel := context.WithTimeout(ctx, checkTimeout)
	out, err := run(mctx, rt.Env, rt.Python, filepath.Join(rt.ScriptsDir, "models.py"), "check")
	cancel()
	if err != nil {
		st.Message = "Chưa tải đủ mô hình giọng đọc"
		if errors.Is(mctx.Err(), context.DeadlineExceeded) {
			st.Message = "Kiểm tra mô hình quá lâu"
		}
		st.Detail = firstLine(out, err)
		return st
	}
	st.ModelsOK = true
	st.Ready = true
	st.Message = "Sẵn sàng"
	st.Detail = out
	return st
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// firstLine lấy tối đa vài dòng đầu của output (hoặc lỗi nếu output rỗng).
func firstLine(out string, err error) string {
	out = strings.TrimSpace(out)
	if out == "" && err != nil {
		return err.Error()
	}
	lines := strings.Split(out, "\n")
	if len(lines) > 4 {
		lines = lines[:4]
	}
	return strings.Join(lines, "\n")
}
