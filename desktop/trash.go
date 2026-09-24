package main

// Xoá sách = chuyển thư mục vào Thùng rác của máy (lấy lại được), không xoá hẳn.
// Hệ điều hành không cho thì dời vào ~/Sano/.da-xoa để người dùng tự dọn.

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// fallbackTrashDir — nơi dời sách khi không chuyển được vào Thùng rác.
const fallbackTrashDir = ".da-xoa"

// moveAside dời src vào thư mục dir, đổi tên nếu trùng. Trả đường dẫn mới.
func moveAside(src, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	base := filepath.Base(src)
	dst := filepath.Join(dir, base)
	for i := 2; ; i++ {
		if _, err := os.Lstat(dst); os.IsNotExist(err) {
			break
		}
		if i > 1000 {
			return "", fmt.Errorf("quá nhiều bản trùng tên %q trong %s", base, dir)
		}
		dst = filepath.Join(dir, base+" "+time.Now().Format("2006-01-02 15.04.05")+" "+strconv.Itoa(i))
	}
	if err := os.Rename(src, dst); err != nil {
		return "", err
	}
	return dst, nil
}

// trashBook chuyển thư mục sách vào Thùng rác; không được thì dời vào
// <sanoRoot>/.da-xoa. Trả nơi đã chuyển tới để báo người dùng.
func trashBook(dir, sanoRoot string) (string, error) {
	if err := moveToSystemTrash(dir); err == nil {
		return "Thùng rác", nil
	}
	dst, err := moveAside(dir, filepath.Join(sanoRoot, fallbackTrashDir))
	if err != nil {
		return "", fmt.Errorf("không chuyển được sách vào Thùng rác: %w", err)
	}
	return dst, nil
}
