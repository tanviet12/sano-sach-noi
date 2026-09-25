package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMacInstall(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "Sano.app", "Contents", "MacOS", "Sano")
	info := macInstall(installInfo{kind: "dmg", suffix: "macos-universal.dmg"}, exe)
	if info.err != nil || info.path != filepath.Join(dir, "Sano.app") {
		t.Errorf("cài trong thư mục ghi được → %+v", info)
	}
	for name, p := range map[string]string{
		"chạy từ .dmg":      "/Volumes/Sano/Sano.app/Contents/MacOS/Sano",
		"App Translocation": "/private/var/folders/x/AppTranslocation/ABC/d/Sano.app/Contents/MacOS/Sano",
		"không phải .app":   filepath.Join(dir, "build", "bin", "Sano"),
		"thư mục chỉ đọc":   "/System/Applications/Sano.app/Contents/MacOS/Sano",
	} {
		if got := macInstall(installInfo{}, p); got.err == nil {
			t.Errorf("%s: phải báo không tự thay được", name)
		}
	}
}

func writePlist(t *testing.T, bundle, version string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(bundle, "Contents", "MacOS"), 0o755); err != nil {
		t.Fatal(err)
	}
	plist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict><key>CFBundleShortVersionString</key><string>` + version + `</string></dict></plist>`
	os.WriteFile(filepath.Join(bundle, "Contents", "Info.plist"), []byte(plist), 0o644)
	os.WriteFile(filepath.Join(bundle, "Contents", "MacOS", "Sano"), []byte("x"), 0o755)
}

func TestCheckBundleVersion(t *testing.T) {
	if _, err := exec.LookPath("plutil"); err != nil {
		t.Skip("không có plutil")
	}
	b := filepath.Join(t.TempDir(), "Sano.app")
	writePlist(t, b, "0.2.0")
	if err := checkBundleVersion(b, "0.2.0"); err != nil {
		t.Errorf("đúng phiên bản: %v", err)
	}
	if err := checkBundleVersion(b, "0.2.0-rc.1"); err != nil {
		t.Errorf("bản thử so phần số: %v", err)
	}
	if err := checkBundleVersion(b, "0.3.0"); err == nil || !strings.Contains(err.Error(), "0.3.0") {
		t.Errorf("sai phiên bản phải từ chối, được %v", err)
	}
	os.Remove(filepath.Join(b, "Contents", "MacOS", "Sano"))
	if err := checkBundleVersion(b, "0.2.0"); err == nil {
		t.Error("thiếu file chạy phải từ chối")
	}
}

// Thay bản thật với một .dmg dựng tại chỗ (hdiutil), không mở lại app.
func TestApplyUpdate_MacDmg(t *testing.T) {
	if testing.Short() {
		t.Skip("cần hdiutil")
	}
	if _, err := exec.LookPath("hdiutil"); err != nil {
		t.Skip("không có hdiutil")
	}
	root := t.TempDir()
	stage := filepath.Join(root, "stage")
	writePlist(t, filepath.Join(stage, "Sano.app"), "0.2.0")
	dmg := filepath.Join(root, "Sano-0.2.0-macos-universal.dmg")
	if out, err := exec.Command("hdiutil", "create", "-volname", "Sano", "-srcfolder", stage, "-fs", "HFS+", "-format", "UDZO", "-ov", dmg).CombinedOutput(); err != nil {
		t.Skipf("hdiutil create: %v %s", err, out)
	}

	apps := filepath.Join(root, "Applications")
	cur := filepath.Join(apps, "Sano.app")
	writePlist(t, cur, "0.1.2")
	os.WriteFile(filepath.Join(cur, "Contents", "cu.txt"), []byte("bản cũ"), 0o644)
	inst := macInstall(installInfo{kind: "dmg", suffix: "macos-universal.dmg"}, filepath.Join(cur, "Contents", "MacOS", "Sano"))
	if inst.err != nil {
		t.Fatal(inst.err)
	}

	if err := applyUpdate(inst, dmg, "0.3.0", false); err == nil {
		t.Fatal("dmg ghi 0.2.0 mà đòi 0.3.0: phải từ chối")
	}
	if err := checkBundleVersion(cur, "0.1.2"); err != nil {
		t.Fatalf("từ chối rồi thì bản cũ phải còn nguyên: %v", err)
	}

	if err := applyUpdate(inst, dmg, "0.2.0", false); err != nil {
		t.Fatal(err)
	}
	if err := checkBundleVersion(cur, "0.2.0"); err != nil {
		t.Errorf("chưa thay được bản mới: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cur, "Contents", "cu.txt")); !os.IsNotExist(err) {
		t.Error("file của bản cũ còn sót trong gói mới")
	}
	left, _ := os.ReadDir(apps)
	if len(left) != 1 {
		t.Errorf("thư mục cài còn thừa: %v", left)
	}
}
