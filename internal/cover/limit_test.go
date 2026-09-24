package cover

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// Tên sách / tác giả cực dài (một "từ" không dấu cách) vẫn vẽ bìa nhanh.
func TestRender_HugeTitle_Fast(t *testing.T) {
	title := strings.Repeat("Ư", 200000)
	author := strings.Repeat("A", 200000)
	start := time.Now()
	if _, err := Render(title, author, 1200, 1600); err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("vẽ bìa tên cực dài mất %v, quá chậm", d)
	}
}

// truncate trả tiền tố + "…" vừa khung, dài nhất có thể.
func TestTruncate_FitsAndIsPrefix(t *testing.T) {
	if err := loadFonts(); err != nil {
		t.Fatal(err)
	}
	f, err := face(serifFont, 60)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	s := "Tuyển tập truyện ngắn hay nhất mọi thời đại của văn học Việt Nam"
	for _, tracking := range []float64{0, 4} {
		got := truncate(f, s, 500, tracking)
		if advance(f, got, tracking) > 500 {
			t.Errorf("tracking=%v: %q rộng quá khung", tracking, got)
		}
		body := strings.TrimSuffix(got, "…")
		if !strings.HasSuffix(got, "…") || !strings.HasPrefix(s, body) {
			t.Errorf("tracking=%v: %q không phải tiền tố + …", tracking, got)
		}
		// Thêm một ký tự nữa thì không còn vừa (đã lấy dài nhất).
		next := []rune(s)[utf8.RuneCountInString(body)]
		if next != ' ' && advance(f, body+string(next)+"…", tracking) <= 500 {
			t.Errorf("tracking=%v: %q chưa phải tiền tố dài nhất", tracking, got)
		}
	}
	if got := truncate(f, "Ngắn", 500, 0); got != "Ngắn" {
		t.Errorf("chuỗi vừa khung phải giữ nguyên, got %q", got)
	}
}
