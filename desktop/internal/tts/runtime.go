package tts

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Nguồn của bộ đọc đang dùng.
const (
	SourceEnv    = "env"    // SANO_TTS_PYTHON
	SourceApp    = "app"    // bộ đọc app tự cài trong thư mục dữ liệu
	SourceLegacy = "legacy" // ~/VieNeu-TTS-v3 cài tay theo docs/tts-build-guide.md
)

// Runtime — python + script + biến môi trường để gọi bộ đọc.
type Runtime struct {
	Python     string
	ScriptsDir string
	Env        []string // thêm vào os.Environ() khi chạy python (nil = không thêm)
	Source     string
	// ScriptsErr — lý do không có ScriptsDir (hiện cho người dùng); "" nếu có.
	ScriptsErr string
}

// ResolveOptions — đầu vào của Resolve (tách để test).
type ResolveOptions struct {
	Getenv  func(string) string
	Home    string
	GOOS    string
	Layout  Layout
	Scripts fs.FS    // script nhúng trong app (nil = không giải nén)
	Starts  []string // nơi bắt đầu đi ngược lên tìm scripts/tts (chỉ khi Dev)
	// Dev — bản chạy bằng `wails dev`: cho phép đi ngược lên từ Starts tìm
	// scripts/tts trong repo. Bản phát hành KHÔNG dò theo thư mục hiện tại (mở
	// app từ một thư mục lạ có scripts/tts/ giả sẽ chạy script đó).
	Dev bool
}

// Resolve chọn bộ đọc theo thứ tự:
//
//	python: SANO_TTS_PYTHON → bộ đọc app cài (<dữ liệu app>/tts) → ~/VieNeu-TTS-v3
//	script: SANO_TTS_SCRIPTS → bản nhúng giải nén vào <dữ liệu app>/tts/scripts → scripts/tts trong repo (chỉ bản dev)
//
// Không có bộ đọc nào → trả python của bộ đọc app (chưa có file) để Check báo "Chưa cài".
func Resolve(o ResolveOptions) Runtime {
	rt := Runtime{}
	appPy := o.Layout.Python()
	legacyPy := ""
	if o.Home != "" {
		legacyPy = VenvPython(filepath.Join(o.Home, VenvDir, ".venv"), o.GOOS)
	}
	switch p := strings.TrimSpace(o.Getenv(EnvPython)); {
	case p != "":
		rt.Python, rt.Source = p, SourceEnv
	case o.Layout.Root != "" && fileExists(appPy):
		rt.Python, rt.Source, rt.Env = appPy, SourceApp, o.Layout.Env()
	case legacyPy != "" && fileExists(legacyPy):
		rt.Python, rt.Source = legacyPy, SourceLegacy
	default:
		rt.Python, rt.Source = appPy, SourceApp
		if o.Layout.Root != "" {
			rt.Env = o.Layout.Env()
		}
	}
	rt.ScriptsDir, rt.ScriptsErr = resolveScripts(o)
	return rt
}

// resolveScripts trả thư mục script đọc giọng, hoặc ("", lý do) nếu không có.
func resolveScripts(o ResolveOptions) (string, string) {
	if d := strings.TrimSpace(o.Getenv(EnvScripts)); d != "" && fileExists(filepath.Join(d, "models.py")) {
		return d, ""
	}
	problem := ""
	switch {
	case o.Scripts == nil:
		problem = "Bản này không kèm script đọc giọng. Đặt biến " + EnvScripts + " trỏ tới thư mục scripts/tts."
	case o.Layout.Root == "":
		problem = "Không xác định được thư mục dữ liệu của Sano để giải nén script đọc giọng (kiểm tra thư mục người dùng, hoặc " +
			EnvDataDir + " phải là đường dẫn tuyệt đối)."
	default:
		err := EnsureScripts(o.Layout, o.Scripts)
		if err == nil {
			return o.Layout.Scripts(), ""
		}
		problem = "Không giải nén được script đọc giọng vào " + o.Layout.Scripts() + ": " + err.Error()
	}
	if o.Dev {
		if d := FindScriptsDir(func(string) string { return "" }, o.Starts...); d != "" {
			return d, ""
		}
	}
	return "", problem
}

// EnsureScripts giải nén script nhúng vào <tts>/scripts. Chỉ ghi file khác nội
// dung (gọi mỗi lần kiểm tra bộ đọc mà không tốn công); ghi qua file tạm rồi đổi
// tên để không bao giờ để lại script ghi dở.
func EnsureScripts(l Layout, src fs.FS) error {
	if err := l.EnsureRoot(); err != nil {
		return err
	}
	dir := l.Scripts()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entries, err := fs.ReadDir(src, ".")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := fs.ReadFile(src, e.Name())
		if err != nil {
			return err
		}
		dst := filepath.Join(dir, e.Name())
		if old, err := os.ReadFile(dst); err == nil && bytes.Equal(old, data) {
			continue
		}
		tmp := dst + ".tmp"
		if err := os.WriteFile(tmp, data, 0o644); err != nil {
			return fmt.Errorf("giải nén script %s: %w", e.Name(), err)
		}
		if err := os.Rename(tmp, dst); err != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("giải nén script %s: %w", e.Name(), err)
		}
	}
	return nil
}
