//go:build !windows

package bookmaker

import "os/exec"

// hideWindow chỉ có tác dụng trên Windows.
func hideWindow(*exec.Cmd) {}
