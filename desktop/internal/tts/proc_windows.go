//go:build windows

package tts

import (
	"os/exec"
	"syscall"
)

// createNoWindow — cờ CREATE_NO_WINDOW của Windows.
const createNoWindow = 0x08000000

// HideWindow: app gọi python/uv/ffmpeg thì không bật cửa sổ console.
func HideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
