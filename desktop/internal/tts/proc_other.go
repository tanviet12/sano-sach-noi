//go:build !windows

package tts

import "os/exec"

// HideWindow chỉ có tác dụng trên Windows.
func HideWindow(*exec.Cmd) {}
