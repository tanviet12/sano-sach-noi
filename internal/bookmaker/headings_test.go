package bookmaker

import "testing"

func TestDropLeadingSectionNumber(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1.2.3. Tên mục", "Tên mục"},
		{"1. Lời mở đầu", "Lời mở đầu"},
		{"7.5.2.1. QUẢN LÝ THỜI GIAN", "QUẢN LÝ THỜI GIAN"},
		{"4.1 Kế hoạch tuần", "Kế hoạch tuần"},
		{"2.1.Viết liền không cách", "Viết liền không cách"},
		// Không phải số mục → giữ nguyên
		{"3.5 năm kinh nghiệm", "3.5 năm kinh nghiệm"},
		{"2026 là năm bản lề", "2026 là năm bản lề"},
		{"Chương mở đầu", "Chương mở đầu"},
		{"  Tiêu đề có khoảng trắng  ", "Tiêu đề có khoảng trắng"},
	}
	for _, tc := range cases {
		if got := dropLeadingSectionNumber(tc.in); got != tc.want {
			t.Errorf("dropLeadingSectionNumber(%q) = %q, muốn %q", tc.in, got, tc.want)
		}
	}
}

func TestParseHeadingNumbers(t *testing.T) {
	cases := []struct {
		in      string
		keep    bool
		wantErr bool
	}{
		{"", false, false},
		{"drop", false, false},
		{"DROP", false, false},
		{"keep", true, false},
		{"giữ", false, true},
	}
	for _, tc := range cases {
		keep, err := ParseHeadingNumbers(tc.in)
		if (err != nil) != tc.wantErr || keep != tc.keep {
			t.Errorf("ParseHeadingNumbers(%q) = %v, %v", tc.in, keep, err)
		}
	}
}

func TestScript_BodyMultiLevelNumberLines(t *testing.T) {
	in := "Mở đầu.\n\n4.1.2. Kế hoạch tuần\n\n2.5 kg là đủ dùng."
	cases := []struct {
		name string
		n    *Normalizer
		want string
	}{
		{"mặc định bỏ số", &Normalizer{}, "Mở đầu.\n\nKế hoạch tuần\n\n2.5 kg là đủ dùng."},
		{"keep đọc số", &Normalizer{keepHeadingNumbers: true}, "Mở đầu.\n\nbốn chấm một chấm hai. Kế hoạch tuần\n\n2.5 kg là đủ dùng."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.n.script(in); got != tc.want {
				t.Errorf("script = %q, muốn %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeTitle_NoDictionary(t *testing.T) {
	if got := normalizeTitle("1.2. Phân tích SWOT cho KPI"); got != "1.2. Phân tích SWOT cho KPI" {
		t.Errorf("tiêu đề hiển thị không được đổi viết tắt / bỏ số: %q", got)
	}
}
