package main

import "syscall"

const (
	synchronize = 0x00100000
	waitTimeout = 0x00000102
)

func processAlive(pid int) bool {
	h, err := syscall.OpenProcess(synchronize, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)
	ev, _ := syscall.WaitForSingleObject(h, 0)
	return ev == waitTimeout
}
