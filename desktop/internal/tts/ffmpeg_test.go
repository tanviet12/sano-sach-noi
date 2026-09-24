package tts

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func noPath(string) (string, error) { return "", errors.New("không có") }

func TestFindFFmpeg_EnvFirst(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ffmpeg")
	if err := os.WriteFile(p, []byte("#"), 0o755); err != nil {
		t.Fatal(err)
	}
	env := func(k string) string {
		if k == EnvFFmpeg {
			return p
		}
		return ""
	}
	if got := FindFFmpeg(env, noPath, "darwin"); got != p {
		t.Errorf("muốn %q, got %q", p, got)
	}
}

func TestFindFFmpeg_FromPath(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ffmpeg")
	look := func(string) (string, error) { return p, nil }
	if got := FindFFmpeg(func(string) string { return "" }, look, "linux"); got != p {
		t.Errorf("muốn %q, got %q", p, got)
	}
}

// ffmpeg app tự tải (đã kiểm SHA256) được chọn trước bản trong PATH.
func TestFindFFmpeg_AppDownloadedBeforePath(t *testing.T) {
	dir := t.TempDir()
	app := filepath.Join(dir, "app-ffmpeg")
	onPath := filepath.Join(dir, "path-ffmpeg")
	for _, p := range []string{app, onPath} {
		if err := os.WriteFile(p, []byte("#"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	look := func(string) (string, error) { return onPath, nil }
	if got := FindFFmpeg(func(string) string { return "" }, look, "windows", app); got != app {
		t.Errorf("muốn ffmpeg app tải %q, got %q", app, got)
	}
}

// Vị trí dự phòng chỉ gồm nơi cần quyền quản trị / trình quản lý gói để ghi:
// không có thư mục ở gốc ổ C (người dùng thường tạo được).
func TestFFmpegFallbacks_NoUserWritableWindowsPaths(t *testing.T) {
	for _, p := range ffmpegFallbacks["windows"] {
		if !strings.HasPrefix(strings.ToLower(p), `c:\program files\`) {
			t.Errorf("vị trí dự phòng Windows %q người dùng thường có thể ghi", p)
		}
	}
}

func TestFindFFmpeg_Missing_Empty(t *testing.T) {
	if got := FindFFmpeg(func(string) string { return "" }, noPath, "plan9"); got != "" {
		t.Errorf("không có ffmpeg phải trả rỗng, got %q", got)
	}
}
