package main

// Tài liệu mẫu: người mới cài chưa có file Word vẫn đi trọn 6 bước tạo sách.

import (
	"fmt"
	"os"
	"path/filepath"

	"sano/internal/bookmaker"
)

const sampleDocxName = "tai-lieu-mau.docx"

// SampleDocx ghi file Word mẫu (3 chương, có tiêu đề, đoạn văn, 1 hình) vào
// ~/Sano/.tam rồi trả như file người dùng vừa chọn.
func (a *App) SampleDocx() (*DocxFile, error) {
	dir := filepath.Join(a.lib.Root(), ".tam")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("tạo thư mục tạm: %w", err)
	}
	return writeSampleDocx(filepath.Join(dir, sampleDocxName))
}

func writeSampleDocx(path string) (*DocxFile, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("ghi tài liệu mẫu: %w", err)
	}
	if err := bookmaker.WriteSampleDocx(f); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("ghi tài liệu mẫu: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("ghi tài liệu mẫu: %w", err)
	}
	return describeDocx(path)
}
