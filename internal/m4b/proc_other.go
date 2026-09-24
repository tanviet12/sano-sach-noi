//go:build !windows

package m4b

import "os/exec"

// hideWindow chỉ có tác dụng trên Windows.
func hideWindow(*exec.Cmd) {}
