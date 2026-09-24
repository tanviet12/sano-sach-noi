package safepath

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// ErrNotRegular — đường dẫn không phải file thường (symlink, thư mục, thiết bị,
// FIFO...). Thư mục sách người khác gửi có thể chứa symlink trỏ ra ngoài.
var ErrNotRegular = errors.New("không phải file thường")

// IsRegularFile báo path là file thường, không đi theo symlink.
func IsRegularFile(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular()
}

// OpenRegular mở path để đọc, chỉ khi nó là file thường (không phải symlink).
// Kiểm lại sau khi mở để file bị tráo giữa lúc kiểm và lúc mở cũng bị từ chối.
func OpenRegular(path string) (*os.File, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, fmt.Errorf("%q: %w", path, ErrNotRegular)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	after, err := f.Stat()
	if err != nil || !os.SameFile(before, after) {
		_ = f.Close()
		return nil, fmt.Errorf("%q: %w", path, ErrNotRegular)
	}
	return f, nil
}

// ErrTooLarge — file vượt giới hạn đọc.
var ErrTooLarge = errors.New("file quá lớn")

// ReadRegular đọc toàn bộ một file thường (không symlink), tối đa limit byte.
func ReadRegular(path string, limit int64) ([]byte, error) {
	f, err := OpenRegular(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%q: %w", path, ErrTooLarge)
	}
	return data, nil
}
