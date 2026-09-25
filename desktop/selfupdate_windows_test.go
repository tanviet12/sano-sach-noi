package main

import (
	"archive/zip"
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

func TestSetupCmdLine(t *testing.T) {
	file := `C:\Users\Nguyen Van A\AppData\Local\Sano\cap-nhat\Sano-0.2.0-windows-amd64-setup.exe`
	dir := `C:\Users\Nguyen Van A\AppData\Local\Programs\Sano`
	want := `"` + file + `" /S /sano-mo-lai /D=` + dir
	if got := setupCmdLine(file, dir, true); got != want {
		t.Errorf("mở lại:\n được %s\n muốn %s", got, want)
	}
	if got := setupCmdLine(file, dir, false); got != `"`+file+`" /S /D=`+dir {
		t.Errorf("không mở lại: %s", got)
	}
}

// Thay bản chạy ngay: Sano.exe cũ → .cu, Sano.exe mới lấy từ .zip.
func TestApplyUpdate_Portable(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "Sano.exe")
	os.WriteFile(exe, []byte("bản cũ"), 0o755)
	zipFile := filepath.Join(t.TempDir(), "Sano-0.2.0-windows-amd64-portable.zip")
	f, _ := os.Create(zipFile)
	zw := zip.NewWriter(f)
	w, _ := zw.Create("Sano.exe")
	w.Write([]byte("bản mới"))
	zw.Close()
	f.Close()
	if err := applyUpdate(winInstall(exe, "amd64"), zipFile, "0.2.0", false); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "bản mới" {
		t.Errorf("Sano.exe = %q", b)
	}
	if b, _ := os.ReadFile(exe + ".cu"); string(b) != "bản cũ" {
		t.Errorf("Sano.exe.cu = %q", b)
	}
}
