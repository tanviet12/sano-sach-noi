package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxInstall(t *testing.T) {
	app := filepath.Join(t.TempDir(), "Sano.AppImage")
	os.WriteFile(app, []byte("x"), 0o755)
	if got := linuxInstall(app, "amd64"); got.err != nil || got.suffix != "linux-amd64.AppImage" {
		t.Errorf("AppImage ghi được → %+v", got)
	}
	for name, c := range map[string][2]string{
		"không chạy từ AppImage": {"", "amd64"},
		"đường dẫn tương đối":    {"Sano.AppImage", "amd64"},
		"file không còn":         {app + ".mat", "amd64"},
		"arm64":                  {app, "arm64"},
	} {
		if got := linuxInstall(c[0], c[1]); got.err == nil {
			t.Errorf("%s: phải báo không tự thay được", name)
		}
	}
}

func TestApplyUpdate_AppImage(t *testing.T) {
	dir := t.TempDir()
	app := filepath.Join(dir, "Sano.AppImage")
	os.WriteFile(app, []byte("bản cũ"), 0o755)
	fresh := filepath.Join(t.TempDir(), "Sano-0.2.0-linux-amd64.AppImage")
	os.WriteFile(fresh, []byte("bản mới"), 0o644)
	if err := applyUpdate(linuxInstall(app, "amd64"), fresh, "0.2.0", false); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(app)
	st, _ := os.Stat(app)
	if string(b) != "bản mới" || st.Mode()&0o111 == 0 {
		t.Errorf("sau khi thay: %q, quyền %v", b, st.Mode())
	}
	if left, _ := os.ReadDir(dir); len(left) != 1 {
		t.Errorf("còn file thừa: %v", left)
	}
}
