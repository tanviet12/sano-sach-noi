package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
)

// Windows có hai kiểu cài:
//   - Bộ cài NSIS (có uninstall.exe cạnh Sano.exe): chạy bộ cài mới im lặng
//     (/S) vào đúng thư mục đang cài rồi thoát app. Bộ cài tự đợi Sano.exe
//     thoát hẳn rồi mới chép, xong thì mở lại app (cờ /sano-mo-lai, xem
//     build/windows/installer/project.nsi).
//   - Bản chạy ngay (.zip): Windows cho đổi tên file .exe đang chạy → đổi
//     Sano.exe thành Sano.exe.cu, đặt Sano.exe mới vào chỗ, mở bản mới (bản mới
//     đợi bản cũ thoát rồi xoá .cu).

const (
	createNewProcessGroup = 0x00000200
	detachedProcess       = 0x00000008
)

func detectInstall() installInfo {
	exe, err := os.Executable()
	if err == nil {
		exe, err = filepath.EvalSymlinks(exe)
	}
	if err != nil {
		return installInfo{err: fmt.Errorf("không xác định được chỗ cài Sano: %w", err)}
	}
	return winInstall(exe, runtime.GOARCH)
}

// winInstall chọn kiểu cài theo thư mục chứa Sano.exe (tách riêng để test).
func winInstall(exe, arch string) installInfo {
	dir := filepath.Dir(exe)
	if arch != "amd64" {
		return installInfo{err: fmt.Errorf("chưa có bản cài cho Windows %s", arch)}
	}
	if fileExists(filepath.Join(dir, "uninstall.exe")) {
		info := installInfo{kind: "setup", path: dir, suffix: "windows-amd64-setup.exe"}
		if !canWriteDir(dir) {
			info.err = fmt.Errorf("không có quyền ghi vào %s — tải bản cài ở trang phát hành và cài đè", dir)
		}
		return info
	}
	info := installInfo{kind: "portable", path: exe, suffix: "windows-amd64-portable.zip"}
	if !canWriteDir(dir) {
		info.err = fmt.Errorf("không có quyền ghi vào %s — tải bản mới ở trang phát hành", dir)
	}
	return info
}

func applyUpdate(inst installInfo, file, version string, relaunch bool) error {
	if inst.err != nil {
		return inst.err
	}
	pid := strconv.Itoa(os.Getpid())
	switch inst.kind {
	case "setup":
		cmd := exec.Command(file)
		cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: setupCmdLine(file, inst.path, relaunch), CreationFlags: createNewProcessGroup | detachedProcess}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("không chạy được bộ cài: %w", err)
		}
		return cmd.Process.Release()
	case "portable":
		exe := inst.path
		fresh := exe + ".moi"
		if err := extractExe(file, fresh); err != nil {
			os.Remove(fresh)
			return err
		}
		old := exe + ".cu"
		_ = os.Remove(old)
		if err := os.Rename(exe, old); err != nil {
			os.Remove(fresh)
			return fmt.Errorf("không thay được bản đang chạy: %w", err)
		}
		if err := os.Rename(fresh, exe); err != nil {
			_ = os.Rename(old, exe)
			os.Remove(fresh)
			return fmt.Errorf("không đặt được bản mới: %w", err)
		}
		if !relaunch {
			return nil // .cu xoá ở lần mở sau
		}
		cmd := exec.Command(exe, waitFlag, pid)
		cmd.Dir = filepath.Dir(exe)
		cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNewProcessGroup}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("không mở được bản mới: %w", err)
		}
		return cmd.Process.Release()
	}
	return errors.New("kiểu cài này không tự thay được")
}

// setupCmdLine — dòng lệnh chạy bộ cài NSIS im lặng. /D= phải đứng cuối, không
// có ngoặc kép kể cả khi đường dẫn có dấu cách (quy ước NSIS).
func setupCmdLine(file, dir string, relaunch bool) string {
	line := syscall.EscapeArg(file) + " /S"
	if relaunch {
		line += " /sano-mo-lai"
	}
	return line + " /D=" + dir
}

// cleanupAfterUpdate xoá Sano.exe.cu còn lại sau khi thay bản chạy ngay.
func cleanupAfterUpdate() {
	if exe, err := os.Executable(); err == nil {
		_ = os.Remove(exe + ".cu")
	}
}
