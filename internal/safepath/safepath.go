// Package safepath kiểm tên file lấy từ dữ liệu không tin cậy (metadata.json,
// file Word, gói zip người khác gửi) trước khi ghép vào một thư mục.
package safepath

import (
	"path/filepath"
	"strings"
)

// IsPlainName báo name là tên một file nằm ngay trong thư mục: không rỗng,
// không có dấu phân cách / hay \, không phải "." hay "..", không phải đường dẫn
// tuyệt đối hay tên thiết bị (filepath.IsLocal).
func IsPlainName(name string) bool {
	return name != "" && name != "." && name != ".." &&
		!strings.ContainsAny(name, `/\`) && filepath.IsLocal(name)
}

// Within báo full (đã ghép từ dir) vẫn nằm trong dir.
func Within(dir, full string) bool {
	rel, err := filepath.Rel(dir, full)
	return err == nil && rel != "." && filepath.IsLocal(rel)
}
