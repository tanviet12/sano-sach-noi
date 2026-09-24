//go:build !darwin && !windows

package main

import "os/exec"

// moveToSystemTrash dùng `gio trash` (có sẵn trên GNOME và phần lớn desktop Linux).
func moveToSystemTrash(path string) error {
	return exec.Command("gio", "trash", "--", path).Run()
}
