package main

// Sách mẫu: một cuốn ngắn dựng sẵn bằng Sano (giọng Hải Đăng), chép vào thư
// viện lần đầu mở app để người mới có sách nghe ngay. Người dùng xoá rồi thì
// không chép lại — đánh dấu bằng một file ẩn ở ~/Sano.

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"sano/desktop/internal/library"
	"sano/internal/bookmaker"
)

//go:embed mau/sach-mau
var sampleBookFS embed.FS

const (
	sampleBookDir   = "mau/sach-mau"
	sampleBookSlug  = "ky-nang-mem-cho-nguoi-tre"
	sampleBookVoice = "Hải Đăng" // giọng đã dùng khi dựng cuốn mẫu (không theo giọng mặc định)
	sampleBookMark  = ".da-chep-sach-mau"
)

// seedSampleBook chép cuốn mẫu vào thư viện nếu chưa từng chép. Lỗi thì không
// đánh dấu, lần mở sau thử lại.
func seedSampleBook(lib *library.Library) error {
	mark := filepath.Join(lib.Root(), sampleBookMark)
	if _, err := os.Stat(mark); err == nil {
		return nil
	}
	work, err := lib.NewWorkDir(sampleBookSlug)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(work)
		}
	}()
	sub, err := fs.Sub(sampleBookFS, sampleBookDir)
	if err != nil {
		return err
	}
	if err := os.CopyFS(work, sub); err != nil {
		return fmt.Errorf("chép sách mẫu: %w", err)
	}
	// Gói zip chuẩn như cuốn tự tạo: thư viện đọc thời lượng từ đây, nút Xuất gói zip dùng lại.
	if _, err := bookmaker.RepackZip(work, filepath.Join(work, "book-"+sampleBookSlug+".zip"), sampleBookVoice, nil); err != nil {
		return fmt.Errorf("đóng gói sách mẫu: %w", err)
	}
	if _, err := lib.Commit(work, sampleBookSlug); err != nil {
		return err
	}
	committed = true
	if err := os.WriteFile(mark, []byte("Sano đã chép sách mẫu vào thư viện. Xoá file này để chép lại.\n"), 0o644); err != nil {
		return fmt.Errorf("đánh dấu đã chép sách mẫu: %w", err)
	}
	return nil
}
