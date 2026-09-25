package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// macOS: tải .dmg, gắn (mount) chỉ đọc, chép Sano.app mới cạnh bản đang chạy,
// bỏ cờ quarantine, kiểm số phiên bản trong Info.plist, rồi đổi chỗ hai bản
// (đổi tên trong cùng thư mục, hỏng thì trả lại bản cũ). Mở lại bằng `open -n`.

func detectInstall() installInfo {
	info := installInfo{kind: "dmg", suffix: "macos-universal.dmg"}
	exe, err := os.Executable()
	if err == nil {
		exe, err = filepath.EvalSymlinks(exe)
	}
	if err != nil {
		info.err = fmt.Errorf("không xác định được chỗ cài Sano: %w", err)
		return info
	}
	return macInstall(info, exe)
}

// macInstall kiểm chỗ cài từ đường dẫn file chạy (tách riêng để test).
func macInstall(info installInfo, exe string) installInfo {
	bundle := filepath.Dir(filepath.Dir(filepath.Dir(exe))) // Sano.app/Contents/MacOS/Sano
	switch {
	case !strings.HasSuffix(bundle, ".app") || filepath.Base(filepath.Dir(exe)) != "MacOS":
		info.err = errors.New("Sano không chạy từ gói .app (bản dev?) — tải bản cài ở trang phát hành")
	case strings.Contains(exe, "/AppTranslocation/"):
		info.err = errors.New("macOS đang chạy Sano ở chế độ cách ly (App Translocation) — kéo Sano vào thư mục Applications rồi mở lại để tự cập nhật được")
	case strings.HasPrefix(bundle, "/Volumes/"):
		info.err = errors.New("Sano đang chạy thẳng từ file .dmg — kéo Sano vào thư mục Applications rồi mở lại để tự cập nhật được")
	case !canWriteDir(filepath.Dir(bundle)):
		info.err = fmt.Errorf("không có quyền ghi vào %s — tải bản cài ở trang phát hành và cài đè", filepath.Dir(bundle))
	}
	info.path = bundle
	return info
}

func applyUpdate(inst installInfo, file, version string, relaunch bool) error {
	if inst.kind != "dmg" || inst.err != nil {
		return errors.New("kiểu cài này không tự thay được")
	}
	tmp, err := os.MkdirTemp("", "sano-cap-nhat-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	mnt := filepath.Join(tmp, "dmg")
	if out, err := exec.Command("hdiutil", "attach", "-nobrowse", "-readonly", "-noautoopen", "-mountpoint", mnt, file).CombinedOutput(); err != nil {
		return fmt.Errorf("không mở được file .dmg: %v %s", err, strings.TrimSpace(string(out)))
	}
	defer exec.Command("hdiutil", "detach", mnt, "-force").Run()

	src := filepath.Join(mnt, "Sano.app")
	parent := filepath.Dir(inst.path)
	pid := strconv.Itoa(os.Getpid())
	staged := filepath.Join(parent, ".Sano-moi-"+pid+".app")
	old := filepath.Join(parent, ".Sano-cu-"+pid+".app")
	_ = os.RemoveAll(staged)
	if out, err := exec.Command("ditto", src, staged).CombinedOutput(); err != nil {
		os.RemoveAll(staged)
		return fmt.Errorf("chép bản mới: %v %s", err, strings.TrimSpace(string(out)))
	}
	// File do Sano tải không bị gắn quarantine, nhưng gỡ cho chắc: còn cờ này
	// macOS sẽ hỏi "Vẫn mở" lại hoặc chạy bản mới ở chế độ cách ly.
	_ = exec.Command("xattr", "-dr", "com.apple.quarantine", staged).Run()
	if err := checkBundleVersion(staged, version); err != nil {
		os.RemoveAll(staged)
		return err
	}
	if err := os.Rename(inst.path, old); err != nil {
		os.RemoveAll(staged)
		return fmt.Errorf("không thay được bản đang chạy: %w", err)
	}
	if err := os.Rename(staged, inst.path); err != nil {
		_ = os.Rename(old, inst.path)
		os.RemoveAll(staged)
		return fmt.Errorf("không đặt được bản mới: %w", err)
	}
	// Bản cũ đang chạy đã nạp vào bộ nhớ (giao diện nhúng trong file chạy) nên
	// xoá được ngay.
	_ = os.RemoveAll(old)
	if relaunch {
		return exec.Command("open", "-n", inst.path, "--args", waitFlag, pid).Start()
	}
	return nil
}

// checkBundleVersion: Info.plist của bản mới phải ghi đúng số phiên bản vừa tải.
func checkBundleVersion(bundle, version string) error {
	plist := filepath.Join(bundle, "Contents", "Info.plist")
	out, err := exec.Command("plutil", "-extract", "CFBundleShortVersionString", "raw", "-o", "-", plist).Output()
	if err != nil {
		return fmt.Errorf("bản mới thiếu Info.plist hợp lệ: %w", err)
	}
	want := version
	if i := strings.IndexByte(want, '-'); i >= 0 {
		want = want[:i]
	}
	if got := strings.TrimSpace(string(out)); got != want {
		return fmt.Errorf("bản mới ghi phiên bản %q, cần %q — huỷ cập nhật", got, want)
	}
	if _, err := os.Stat(filepath.Join(bundle, "Contents", "MacOS", "Sano")); err != nil {
		return errors.New("bản mới thiếu file chạy Sano")
	}
	return nil
}

// cleanupAfterUpdate — macOS xoá bản cũ ngay lúc thay, không cần dọn.
func cleanupAfterUpdate() {}
