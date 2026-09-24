package main

import (
	"os/exec"
	"syscall"
)

// setRawCmdLine đặt dòng lệnh nguyên văn cho tiến trình con (explorer tự tách
// tham số, không theo quy ước ngoặc kép của Go).
func setRawCmdLine(cmd *exec.Cmd, line string) {
	if line == "" {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: line}
}
