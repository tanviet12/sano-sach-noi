package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
)

// Linux: chỉ tự thay khi chạy từ file .AppImage (biến APPIMAGE do runtime
// AppImage đặt). Chép bản mới cạnh file đang chạy rồi đổi tên đè lên (tiến
// trình đang chạy vẫn giữ file cũ tới khi thoát), mở bản mới.

func detectInstall() installInfo {
	return linuxInstall(os.Getenv("APPIMAGE"), runtime.GOARCH)
}

// linuxInstall kiểm file AppImage đang chạy (tách riêng để test).
func linuxInstall(appimage, arch string) installInfo {
	info := installInfo{kind: "appimage", path: appimage, suffix: "linux-amd64.AppImage"}
	switch {
	case arch != "amd64":
		info.err = fmt.Errorf("chưa có bản cài cho Linux %s", arch)
	case appimage == "" || !filepath.IsAbs(appimage) || !fileExists(appimage):
		info.err = errors.New("Sano không chạy từ file .AppImage — tải bản mới ở trang phát hành")
	case !canWriteDir(filepath.Dir(appimage)):
		info.err = fmt.Errorf("không có quyền ghi vào %s — tải bản mới ở trang phát hành", filepath.Dir(appimage))
	}
	return info
}

func applyUpdate(inst installInfo, file, version string, relaunch bool) error {
	if inst.kind != "appimage" || inst.err != nil {
		return errors.New("kiểu cài này không tự thay được")
	}
	pid := strconv.Itoa(os.Getpid())
	fresh := filepath.Join(filepath.Dir(inst.path), ".Sano-moi-"+pid+".AppImage")
	if err := copyFile(file, fresh, 0o755); err != nil {
		os.Remove(fresh)
		return fmt.Errorf("chép bản mới: %w", err)
	}
	if err := os.Rename(fresh, inst.path); err != nil {
		os.Remove(fresh)
		return fmt.Errorf("không thay được bản đang chạy: %w", err)
	}
	if !relaunch {
		return nil
	}
	cmd := exec.Command(inst.path, waitFlag, pid)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("không mở được bản mới: %w", err)
	}
	return cmd.Process.Release()
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Chmod(dst, mode)
}

// cleanupAfterUpdate — Linux đổi tên đè nên không còn gì để dọn.
func cleanupAfterUpdate() {}
