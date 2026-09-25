package main

// Tài liệu mẫu: file Word mẫu có sẵn kiểu Heading 1/2 và lời hướng dẫn chuẩn bị
// file. Người mới vừa nạp thẳng để thử trọn 6 bước tạo sách, vừa lưu về máy để
// làm theo (thay nội dung của mình vào).

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const sampleDocxName = "Mau-sach-noi-Sano.docx"

// EnvSampleOut — nơi lưu file mẫu khi chạy dev/test (bỏ qua hộp lưu file).
const EnvSampleOut = "SANO_SAMPLE_OUT"

//go:embed mau/Mau-sach-noi-Sano.docx
var sampleDocx []byte

// SampleDocx ghi file Word mẫu vào ~/Sano/.tam rồi trả như file người dùng vừa chọn.
func (a *App) SampleDocx() (*DocxFile, error) {
	dir := filepath.Join(a.lib.Root(), ".tam")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("tạo thư mục tạm: %w", err)
	}
	return writeSampleDocx(filepath.Join(dir, sampleDocxName))
}

func writeSampleDocx(path string) (*DocxFile, error) {
	if err := os.WriteFile(path, sampleDocx, 0o644); err != nil {
		return nil, fmt.Errorf("ghi tài liệu mẫu: %w", err)
	}
	return describeDocx(path)
}

// SaveSampleDocx lưu file Word mẫu về máy (hộp lưu file, mặc định thư mục Tải
// về) rồi mở thư mục, chọn sẵn file. Trả đường dẫn đã lưu; huỷ → "".
func (a *App) SaveSampleDocx() (string, error) {
	path, reveal, err := a.sampleTarget()
	if err != nil || path == "" {
		return "", err
	}
	if err := os.WriteFile(path, sampleDocx, 0o644); err != nil {
		return "", fmt.Errorf("lưu file mẫu: %w", err)
	}
	if reveal {
		_ = openPath(path, true)
	}
	return path, nil
}

func (a *App) sampleTarget() (string, bool, error) {
	if v := strings.TrimSpace(os.Getenv(EnvSampleOut)); v != "" {
		if strings.EqualFold(filepath.Ext(v), ".docx") {
			return v, false, nil
		}
		return filepath.Join(v, sampleDocxName), false, nil
	}
	if a.ctx == nil {
		return "", false, errors.New("ứng dụng chưa khởi động xong")
	}
	path, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:                "Lưu file Word mẫu",
		DefaultDirectory:     downloadsDir(),
		DefaultFilename:      sampleDocxName,
		Filters:              []wruntime.FileFilter{{DisplayName: "File Word (*.docx)", Pattern: "*.docx"}},
		CanCreateDirectories: true,
	})
	if err != nil {
		return "", false, fmt.Errorf("mở hộp lưu file: %w", err)
	}
	if path == "" {
		return "", false, nil
	}
	if !strings.EqualFold(filepath.Ext(path), ".docx") {
		path += ".docx"
	}
	return path, true, nil
}
