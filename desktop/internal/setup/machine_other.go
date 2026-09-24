//go:build !windows

package setup

func cpuName(goos string) string  { return cpuNameUnix(goos) }
func ramBytes(goos string) uint64 { return ramBytesUnix(goos) }
