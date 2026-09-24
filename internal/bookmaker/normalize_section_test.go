package bookmaker

import "testing"

func TestIntToViet(t *testing.T) {
	cases := map[int]string{
		0: "không", 1: "một", 6: "sáu", 9: "chín",
		10: "mười", 11: "mười một", 15: "mười lăm",
		20: "hai mươi", 21: "hai mươi mốt", 25: "hai mươi lăm",
		100: "một trăm", 105: "một trăm lẻ năm", 111: "một trăm mười một",
	}
	for n, want := range cases {
		if got := intToViet(n); got != want {
			t.Errorf("intToViet(%d) = %q, muốn %q", n, got, want)
		}
	}
}

func TestExpandLeadingSectionNumber(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1. LỜI MỞ ĐẦU CỦA SÁCH", "một. LỜI MỞ ĐẦU CỦA SÁCH"},
		{"1.6. Quản lý thời gian: Lập kế hoạch", "một chấm sáu. Quản lý thời gian: Lập kế hoạch"},
		{"7.5.2.1. QUẢN LÝ THỜI GIAN", "bảy chấm năm chấm hai chấm một. QUẢN LÝ THỜI GIAN"},
		{"11.2.1.2. THÓI QUEN ĐỌC SÁCH", "mười một chấm hai chấm một chấm hai. THÓI QUEN ĐỌC SÁCH"},
		{"Bước đầu tiên: Chuẩn bị tài liệu", "Bước đầu tiên: Chuẩn bị tài liệu"}, // không có số → giữ nguyên
	}
	for _, c := range cases {
		if got := expandLeadingSectionNumber(c.in); got != c.want {
			t.Errorf("expandLeadingSectionNumber(%q) = %q, muốn %q", c.in, got, c.want)
		}
	}
}

func TestExpandGradeSigns(t *testing.T) {
	cases := []struct{ in, want string }{
		{"xếp hạng A-, A hoặc B", "xếp hạng A trừ, A hoặc B"},
		{"đạt mức A- nhưng", "đạt mức A trừ nhưng"},
		{"tăng từ A- lên A", "tăng từ A trừ lên A"},
		{"Kết quả A- → A", "Kết quả A trừ → A"},
		{"đạt B+ rồi", "đạt B cộng rồi"},
		{"kết thúc ở A-", "kết thúc ở A trừ"}, // cuối chuỗi
		// KHÔNG đụng gạch nối " - " trong cụm từ ghép
		{"nhóm khách quen - khách mới", "nhóm khách quen - khách mới"},
		{"vừa đọc - vừa ghi chép", "vừa đọc - vừa ghi chép"},
	}
	for _, c := range cases {
		if got := expandGradeSigns(c.in); got != c.want {
			t.Errorf("expandGradeSigns(%q) = %q, muốn %q", c.in, got, c.want)
		}
	}
}

func TestSpokenTitle_HeadingNumbers(t *testing.T) {
	title := "1.6. Quản lý thời gian: Lập kế hoạch - Ưu tiên việc chính"
	cases := []struct {
		name string
		n    *Normalizer
		want string
	}{
		{"mặc định bỏ số", defaultNormalizer, "Quản lý thời gian, Lập kế hoạch, Ưu tiên việc chính"},
		{"keep đọc số thành chữ", &Normalizer{keepHeadingNumbers: true}, "một chấm sáu. Quản lý thời gian, Lập kế hoạch, Ưu tiên việc chính"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.n.spokenTitle(title); got != tc.want {
				t.Errorf("spokenTitle = %q, muốn %q", got, tc.want)
			}
		})
	}
	if got := spokenSectionTitle(title); got != cases[0].want {
		t.Errorf("spokenSectionTitle phải dùng mặc định bỏ số: %q", got)
	}
}
