//go:build windows

package setup

import "golang.org/x/sys/windows"

// freeBytes — dung lượng còn trống (cho người dùng hiện tại) của ổ chứa path.
func freeBytes(path string) (uint64, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var avail, total, free uint64
	if err := windows.GetDiskFreeSpaceEx(p, &avail, &total, &free); err != nil {
		return 0, err
	}
	return avail, nil
}
