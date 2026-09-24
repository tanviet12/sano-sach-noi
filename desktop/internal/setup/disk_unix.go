//go:build !windows

package setup

import "syscall"

// freeBytes — dung lượng còn trống (cho người dùng thường) của ổ chứa path.
func freeBytes(path string) (uint64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, err
	}
	return uint64(st.Bavail) * uint64(st.Bsize), nil //nolint:unconvert // kiểu trường khác nhau giữa macOS và Linux
}
