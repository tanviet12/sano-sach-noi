package cover

import (
	"bytes"
	"image/png"
	"testing"
)

// Chỉ số bảng màu PHẢI khớp BookCover.vue (cùng FNV-1a + cùng thứ tự bảng màu)
// để bìa trên web và file ảnh trong gói zip cùng một màu.
func TestPaletteIndex_MatchesWeb(t *testing.T) {
	cases := map[string]int{
		"Kỹ năng mềm cho người trẻ": 6,
		"Lãnh đạo cho quản lý mới":  4,
		"Tài chính cá nhân cơ bản":  1,
		"Khởi nghiệp từ số 0":       2,
		"Kỹ năng giao tiếp":         7,
	}
	for title, want := range cases {
		if got := PaletteIndex(title); got != want {
			t.Errorf("PaletteIndex(%q) = %d, muốn %d (lệch với web)", title, got, want)
		}
	}
}

func TestWritePNG_Sizes(t *testing.T) {
	cases := []struct {
		name          string
		title, author string
		w, h          int
	}{
		{"bìa 3:4 cho gói zip", "Kỹ năng giao tiếp", "Nguyễn Văn A", 1200, 1600},
		{"bìa vuông cho M4B", "Kỹ năng giao tiếp", "Nguyễn Văn A", 1400, 1400},
		{"không tác giả", "Sổ tay quản lý thời gian", "", 600, 800},
		{"tên rất dài tự thu nhỏ", "Một cuốn sách có cái tên rất rất dài để kiểm tra việc tự ngắt dòng và thu nhỏ chữ khi không đủ chỗ trên bìa", "Tác giả có tên cũng khá là dài", 600, 800},
		{"tên rỗng", "", "", 300, 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := WritePNG(&buf, tc.title, tc.author, tc.w, tc.h); err != nil {
				t.Fatalf("WritePNG lỗi: %v", err)
			}
			img, err := png.Decode(&buf)
			if err != nil {
				t.Fatalf("PNG không hợp lệ: %v", err)
			}
			if b := img.Bounds(); b.Dx() != tc.w || b.Dy() != tc.h {
				t.Errorf("kích thước = %dx%d, muốn %dx%d", b.Dx(), b.Dy(), tc.w, tc.h)
			}
		})
	}
}

func TestRender_TooSmall_Errors(t *testing.T) {
	if _, err := Render("x", "", 10, 10); err == nil {
		t.Error("muốn lỗi khi kích thước quá nhỏ")
	}
}
