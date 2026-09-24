package setup

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"sano/desktop/internal/tts"
)

// cpuName + ramBytes theo hệ điều hành; không đọc được thì trả rỗng/0 (chỉ để hiện).
func cpuNameUnix(goos string) string {
	switch goos {
	case "darwin":
		return sysctl("machdep.cpu.brand_string")
	case "linux":
		f, err := os.Open("/proc/cpuinfo")
		if err != nil {
			return ""
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			k, v, ok := strings.Cut(sc.Text(), ":")
			if ok && strings.TrimSpace(k) == "model name" {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
}

func ramBytesUnix(goos string) uint64 {
	switch goos {
	case "darwin":
		n, _ := strconv.ParseUint(sysctl("hw.memsize"), 10, 64)
		return n
	case "linux":
		f, err := os.Open("/proc/meminfo")
		if err != nil {
			return 0
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			fields := strings.Fields(sc.Text())
			if len(fields) >= 2 && fields[0] == "MemTotal:" {
				kb, _ := strconv.ParseUint(fields[1], 10, 64)
				return kb * 1024
			}
		}
	}
	return 0
}

func sysctl(name string) string {
	cmd := exec.Command("/usr/sbin/sysctl", "-n", name)
	tts.HideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
