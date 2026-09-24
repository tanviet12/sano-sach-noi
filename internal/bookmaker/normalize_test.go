package bookmaker

import (
	"strings"
	"testing"
)

func TestNormalizeReadingScript_ExpandsAbbreviations(t *testing.T) {
	in := "Doanh nghiệp ở TP.HCM tăng trưởng, vd. nhờ chuyển đổi số v.v."
	got := normalizeReadingScript(in)
	if !strings.Contains(got, "Thành phố Hồ Chí Minh") {
		t.Errorf("không mở rộng TP.HCM: %q", got)
	}
	if !strings.Contains(got, "ví dụ") {
		t.Errorf("không mở rộng vd.: %q", got)
	}
	if !strings.Contains(got, "vân vân") {
		t.Errorf("không mở rộng v.v.: %q", got)
	}
	if strings.Contains(got, "TP.") || strings.Contains(got, "vd.") {
		t.Errorf("còn sót viết tắt: %q", got)
	}
}

func TestExpandAbbreviations_NoFalseMatchInsideWord(t *testing.T) {
	cases := []struct {
		name string
		in   string
		bad  string // chuỗi KHÔNG được xuất hiện
		good string // chuỗi PHẢI còn nguyên
	}{
		{
			name: "QTP. không bị nuốt thành QThành phố",
			in:   "Báo cáo gửi phòng QTP. Mọi người cùng đọc.",
			bad:  "Thành phố",
			good: "QTP.",
		},
		{
			name: "TS. trong ARTS. không thành ARTiến sĩ",
			in:   "Ngành ARTS. rất rộng.",
			bad:  "Tiến sĩ",
			good: "ARTS.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeReadingScript(tc.in)
			if strings.Contains(got, tc.bad) {
				t.Errorf("không được chứa %q (match nhầm trong từ): %q", tc.bad, got)
			}
			if !strings.Contains(got, tc.good) {
				t.Errorf("phải giữ nguyên %q: %q", tc.good, got)
			}
		})
	}
}

func TestExpandPronunciations_NATO(t *testing.T) {
	got := normalizeReadingScript("Tổ chức NATO được thành lập năm 1949.")
	if !strings.Contains(got, "Na Tô") {
		t.Errorf("NATO phải đọc thành 'Na Tô': %q", got)
	}
	if strings.Contains(got, "NATO") {
		t.Errorf("không được còn 'NATO' chữ hoa (TTS đánh vần): %q", got)
	}
}

func TestExpandMarketingP_ReadsAsPe(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string // chuỗi PHẢI có
		bad  string // chuỗi KHÔNG được có (rỗng = bỏ qua)
	}{
		{"4P đọc bốn Pê", "Mô hình 4P trong marketing.", "bốn Pê", "phút"},
		{"7P dịch vụ", "Mở rộng thành 7P.", "bảy Pê", "phút"},
		{"3P", "Khung 3P cơ bản.", "ba Pê", "phút"},
		{"4p thường giữ nguyên", "Nghỉ giải lao 4p rồi học.", "4p", "Pê"}, // P thường → không đổi
		{"4PM không đụng", "Họp lúc 4PM chiều nay.", "4PM", "Pê"},         // chữ cái theo sau
		{"4Phương không đụng", "Đi 4Phương tám hướng.", "4Phương", "Pê"},  // chữ cái theo sau
		{"104P mã số không đụng", "Phòng 104P ở tầng một.", "104P", "Pê"}, // số liền trước
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeReadingScript(tc.in)
			if !strings.Contains(got, tc.want) {
				t.Errorf("muốn chứa %q, got %q", tc.want, got)
			}
			if tc.bad != "" && strings.Contains(got, tc.bad) {
				t.Errorf("không được chứa %q, got %q", tc.bad, got)
			}
		})
	}
}

func TestExpandAbbreviations_StillExpandsAtWordStart(t *testing.T) {
	got := normalizeReadingScript("Trụ sở ở TP. Hà Nội, gần GS. Nguyễn.")
	if !strings.Contains(got, "Thành phố Hà Nội") {
		t.Errorf("TP. đầu từ phải mở rộng: %q", got)
	}
	if !strings.Contains(got, "Giáo sư Nguyễn") {
		t.Errorf("GS. đầu từ phải mở rộng: %q", got)
	}
}

func TestNormalizeReadingScript_RomanHeadings(t *testing.T) {
	cases := map[string]string{
		"Chương I: Mở đầu":   "Chương một",
		"Phần II nói về":     "Phần hai",
		"Bài XV tổng kết":    "Bài mười lăm",
		"Chapter IV summary": "Chapter bốn",
	}
	for in, want := range cases {
		got := normalizeReadingScript(in)
		if !strings.Contains(got, want) {
			t.Errorf("normalize(%q) = %q, muốn chứa %q", in, got, want)
		}
	}
}

func TestNormalizeReadingScript_DropsPageNumbers(t *testing.T) {
	in := "Nội dung chương một.\n12\nNội dung tiếp theo.\n- 13 -\nTrang 14"
	got := normalizeReadingScript(in)
	if strings.Contains(got, "12") || strings.Contains(got, "13") || strings.Contains(got, "14") {
		t.Errorf("còn số trang: %q", got)
	}
	if !strings.Contains(got, "Nội dung chương một") || !strings.Contains(got, "Nội dung tiếp theo") {
		t.Errorf("mất nội dung thật: %q", got)
	}
}

func TestNormalizeReadingScript_StripsDoubleQuotes(t *testing.T) {
	in := "Cô giáo dặn “phải chọn” và \"làm đúng\"."
	got := normalizeReadingScript(in)
	if strings.ContainsAny(got, "\"“”„") {
		t.Errorf("còn dấu ngoặc kép trong kịch bản đọc: %q", got)
	}
	if !strings.Contains(got, "phải chọn") || !strings.Contains(got, "làm đúng") {
		t.Errorf("mất nội dung bên trong ngoặc: %q", got)
	}
}

func TestCollapseSpacesKeepLines_KeepsParagraphBlankLine(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "gộp nhiều dòng trắng thành 1, giữ ranh giới đoạn",
			in:   "Dòng  một   nhiều   space\n\n\nDòng hai\n   \nDòng ba",
			want: "Dòng một nhiều space\n\nDòng hai\n\nDòng ba",
		},
		{
			name: "dòng kề nhau (\\n đơn) giữ nguyên \\n đơn",
			in:   "Đoạn một.\nĐoạn hai.",
			want: "Đoạn một.\nĐoạn hai.",
		},
		{
			name: "bỏ dòng trắng đầu/cuối",
			in:   "\n\nNội dung.\n\n",
			want: "Nội dung.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := collapseSpacesKeepLines(tc.in); got != tc.want {
				t.Errorf("collapseSpacesKeepLines = %q, muốn %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeReadingScript_PreservesParagraphBreaks(t *testing.T) {
	in := "Đoạn một.\nĐoạn hai.\nĐoạn ba."
	got := normalizeReadingScript(in)
	if n := strings.Count(got, "\n"); n != 2 {
		t.Errorf("phải giữ 2 xuống dòng giữa các đoạn, got %d: %q", n, got)
	}
}

func TestExpandHeadingRomans_NoFalseMatchOnVietnameseWord(t *testing.T) {
	// "Bài viết" KHÔNG được biến thành "Bài sáuết" (vi != La Mã VI ở đây)
	cases := map[string]string{
		"Bài viết mới trên trang": "viết", // phải còn nguyên "viết"
		"Phần mềm quản lý":        "mềm",
		"Mục lục đầu sách":        "lục",
	}
	for in, mustContain := range cases {
		got := normalizeReadingScript(in)
		if !strings.Contains(got, mustContain) {
			t.Errorf("normalize(%q) = %q — làm hỏng từ %q", in, got, mustContain)
		}
	}
	// nhưng số La Mã đứng riêng vẫn đổi đúng
	if got := normalizeReadingScript("Phần V là phần cuối"); !strings.Contains(got, "Phần năm") {
		t.Errorf("số La Mã đứng riêng phải đổi: %q", got)
	}
}

func TestRomanToInt(t *testing.T) {
	cases := map[string]int{"I": 1, "IV": 4, "IX": 9, "XV": 15, "XX": 20, "ABC": 0}
	for in, want := range cases {
		if got := romanToInt(in); got != want {
			t.Errorf("romanToInt(%q) = %d, muốn %d", in, got, want)
		}
	}
}

func TestChunkText_RespectsMaxChars(t *testing.T) {
	long := strings.Repeat("Câu mẫu có độ dài vừa phải để kiểm tra chia chunk. ", 20)
	chunks := chunkText(long, chunkMaxChars)
	if len(chunks) < 2 {
		t.Fatalf("muốn nhiều chunk, có %d", len(chunks))
	}
	for i, c := range chunks {
		if len(c) > chunkMaxChars {
			t.Errorf("chunk[%d] dài %d > %d: %q", i, len(c), chunkMaxChars, c)
		}
		if strings.TrimSpace(c) == "" {
			t.Errorf("chunk[%d] rỗng", i)
		}
	}
}

func TestChunkText_SplitsOverlongSentence(t *testing.T) {
	// 1 "câu" không có dấu kết, dài hơn maxChars → phải chia theo từ.
	sent := strings.Repeat("từ ", 200) // ~800 ký tự, không dấu chấm
	chunks := chunkText(sent, chunkMaxChars)
	if len(chunks) < 2 {
		t.Fatalf("câu dài phải chia >=2 chunk, có %d", len(chunks))
	}
	for i, c := range chunks {
		if len(c) > chunkMaxChars {
			t.Errorf("chunk[%d] dài %d > %d", i, len(c), chunkMaxChars)
		}
	}
}

func TestChunkText_Empty(t *testing.T) {
	if got := chunkText("   ", chunkMaxChars); got != nil {
		t.Errorf("input rỗng phải trả nil, có %v", got)
	}
}
