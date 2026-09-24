package tts

import (
	"path/filepath"
	"strings"
)

// EnvFFmpeg cho phép chỉ định ffmpeg ở vị trí khác.
const EnvFFmpeg = "SANO_FFMPEG"

// ffmpegFallbacks — nơi ffmpeg hay nằm nhưng không có trong PATH khi mở app từ
// Finder/Launchpad (PATH của app macOS chỉ có /usr/bin:/bin:/usr/sbin:/sbin).
// Chỉ gồm thư mục của trình quản lý gói / cần quyền quản trị để ghi. KHÔNG thêm
// kiểu C:\ffmpeg\bin: mọi tài khoản Windows đều tạo được thư mục ở gốc ổ C, người
// dùng khác trên cùng máy có thể đặt ffmpeg giả ở đó.
var ffmpegFallbacks = map[string][]string{
	"darwin":  {"/opt/homebrew/bin/ffmpeg", "/usr/local/bin/ffmpeg", "/opt/local/bin/ffmpeg"},
	"linux":   {"/usr/bin/ffmpeg", "/usr/local/bin/ffmpeg", "/snap/bin/ffmpeg"},
	"windows": {`C:\Program Files\ffmpeg\bin\ffmpeg.exe`},
}

// FindFFmpeg tìm ffmpeg: biến SANO_FFMPEG → các đường dẫn thêm (ffmpeg app tự
// tải, đã kiểm SHA256 — ưu tiên hơn mọi bản có sẵn trên máy) → PATH → các vị
// trí cài phổ biến (Homebrew, MacPorts, apt, snap, Program Files). Không thấy → "".
func FindFFmpeg(getenv func(string) string, lookPath func(string) (string, error), goos string, extra ...string) string {
	if p := strings.TrimSpace(getenv(EnvFFmpeg)); p != "" && fileExists(p) {
		return p
	}
	for _, p := range extra {
		if p != "" && fileExists(p) {
			return p
		}
	}
	if p, err := lookPath("ffmpeg"); err == nil && p != "" {
		if abs, err := filepath.Abs(p); err == nil {
			return abs
		}
		return p
	}
	for _, p := range ffmpegFallbacks[goos] {
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// FFmpegHint — gợi ý cài ffmpeg theo hệ điều hành, hiện khi thiếu.
func FFmpegHint(goos string) string {
	switch goos {
	case "darwin":
		return "Cài bằng Homebrew: brew install ffmpeg (hoặc đặt biến " + EnvFFmpeg + " trỏ tới file ffmpeg)."
	case "windows":
		return "Tải ffmpeg cho Windows (gyan.dev), giải nén rồi thêm thư mục bin vào PATH (hoặc đặt biến " + EnvFFmpeg + ")."
	default:
		return "Cài bằng trình quản lý gói, vd: sudo apt install ffmpeg (hoặc đặt biến " + EnvFFmpeg + ")."
	}
}
