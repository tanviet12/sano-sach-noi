//go:build windows

package m4b

import (
	"os/exec"
	"syscall"
)

// createNoWindow — cờ CREATE_NO_WINDOW của Windows.
const createNoWindow = 0x08000000

// hideWindow: phần mềm desktop gọi ffmpeg thì không bật cửa sổ console.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
