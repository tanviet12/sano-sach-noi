//go:build !windows

package main

import "os/exec"

// setRawCmdLine chỉ có tác dụng trên Windows.
func setRawCmdLine(*exec.Cmd, string) {}
