package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWinInstall(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "Sano.exe")
	if got := winInstall(exe, "amd64"); got.err != nil || got.kind != "portable" || got.suffix != "windows-amd64-portable.zip" || got.path != exe {
		t.Errorf("không có uninstall.exe → bản chạy ngay, được %+v", got)
	}
	os.WriteFile(filepath.Join(dir, "uninstall.exe"), []byte("x"), 0o644)
	if got := winInstall(exe, "amd64"); got.err != nil || got.kind != "setup" || got.suffix != "windows-amd64-setup.exe" || got.path != dir {
		t.Errorf("có uninstall.exe → bộ cài, được %+v", got)
	}
	if got := winInstall(exe, "arm64"); got.err == nil {
		t.Error("arm64 chưa có bản cài: phải báo lỗi")
	}
}
