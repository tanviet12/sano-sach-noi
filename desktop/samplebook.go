package main

// Sách mẫu: vài cuốn ngắn dựng sẵn bằng Sano, chép vào thư viện lần đầu mở app
// để người mới có sách nghe ngay. Mỗi cuốn chép một lần: người dùng xoá rồi thì
// không chép lại — ghi slug đã chép vào một file ẩn ở ~/Sano. Bản sau thêm cuốn
// mẫu mới thì máy đã cài vẫn nhận cuốn mới.

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"sano/desktop/internal/library"
	"sano/internal/bookmaker"
)

//go:embed mau/sach-mau
var sampleBookFS embed.FS

type sampleBook struct {
	slug  string // cũng là tên thư mục trong mau/sach-mau
	voice string // giọng đã dùng khi dựng cuốn này (không theo giọng mặc định)
}

// Thứ tự hiện trong thư viện (mới tạo nhất trước): cuốn đầu nằm trên cùng.
var sampleBooks = []sampleBook{
	{"gioi-thieu-sano", "Trúc Ly"},
	{"nghe-de-nho", "Hải Đăng"},
	{"tiem-ca-phe-thu-hai", "Trúc Ly"},
}

const (
	sampleBookDir  = "mau/sach-mau"
	sampleBookMark = ".sach-mau-da-chep" // mỗi dòng một slug đã chép
)

// seedSampleBooks chép các cuốn mẫu chưa từng chép. Cuốn lỗi thì không ghi
// nhận, lần mở sau thử lại; các cuốn khác vẫn chép.
func seedSampleBooks(lib *library.Library) error {
	mark := filepath.Join(lib.Root(), sampleBookMark)
	done := readSeeded(mark)
	now := time.Now()
	var errs []string
	for i, sb := range sampleBooks {
		if slices.Contains(done, sb.slug) {
			continue
		}
		// Thư viện xếp theo giờ sửa thư mục: lùi mỗi cuốn một giây để giữ đúng thứ tự.
		if err := seedSampleBook(lib, sb, now.Add(-time.Duration(i)*time.Second)); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", sb.slug, err))
			continue
		}
		done = append(done, sb.slug)
		if err := os.WriteFile(mark, []byte(strings.Join(done, "\n")+"\n"), 0o644); err != nil {
			return fmt.Errorf("ghi nhận đã chép sách mẫu: %w", err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("chép sách mẫu: %s", strings.Join(errs, "; "))
	}
	return nil
}

func readSeeded(mark string) []string {
	data, err := os.ReadFile(mark)
	if err != nil {
		return nil
	}
	var out []string
	for _, l := range strings.Split(string(data), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

func seedSampleBook(lib *library.Library, sb sampleBook, at time.Time) error {
	work, err := lib.NewWorkDir(sb.slug)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(work)
		}
	}()
	sub, err := fs.Sub(sampleBookFS, path.Join(sampleBookDir, sb.slug))
	if err != nil {
		return err
	}
	if err := os.CopyFS(work, sub); err != nil {
		return fmt.Errorf("chép: %w", err)
	}
	// Gói zip chuẩn như cuốn tự tạo: thư viện đọc thời lượng từ đây, nút Xuất gói zip dùng lại.
	if _, err := bookmaker.RepackZip(work, filepath.Join(work, "book-"+sb.slug+".zip"), sb.voice, nil); err != nil {
		return fmt.Errorf("đóng gói: %w", err)
	}
	final, err := lib.Commit(work, sb.slug)
	if err != nil {
		return err
	}
	committed = true
	_ = os.Chtimes(filepath.Join(lib.BooksRoot(), final), at, at)
	return nil
}
