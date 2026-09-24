package bookmaker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSampleDocx ghi docx mẫu ra file tạm, trả đường dẫn.
func writeSampleDocx(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.docx")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteSampleDocx(f); err != nil {
		_ = f.Close()
		t.Fatalf("WriteSampleDocx: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseDocx_SampleBook_ExtractsStructure(t *testing.T) {
	book, err := ParseDocx(writeSampleDocx(t))
	if err != nil {
		t.Fatalf("ParseDocx: %v", err)
	}

	if book.Title != "Cẩm Nang Mô Hình Kinh Doanh Mẫu" {
		t.Errorf("title = %q", book.Title)
	}
	if len(book.Chapters) != 3 {
		t.Fatalf("muốn 3 chương, có %d", len(book.Chapters))
	}
	if book.Chapters[0].Title != "Chương I: Nền tảng tư duy" {
		t.Errorf("chương 1 title = %q", book.Chapters[0].Title)
	}
	if len(book.Chapters[0].Sections) != 2 {
		t.Fatalf("chương 1 muốn 2 tiểu mục, có %d", len(book.Chapters[0].Sections))
	}
	if !strings.Contains(book.Chapters[0].Sections[0].Text, "bài toán thực tế") {
		t.Errorf("text tiểu mục 1.1 sai: %q", book.Chapters[0].Sections[0].Text)
	}
}

func TestParseDocx_ExtractsEmbeddedImage(t *testing.T) {
	book, err := ParseDocx(writeSampleDocx(t))
	if err != nil {
		t.Fatalf("ParseDocx: %v", err)
	}
	// Ảnh nằm ở tiểu mục 1 của chương I.
	sec := book.Chapters[0].Sections[0]
	if len(sec.Images) != 1 {
		t.Fatalf("muốn 1 ảnh ở tiểu mục 1.1, có %d", len(sec.Images))
	}
	img := sec.Images[0]
	if img.Name != "img001.png" {
		t.Errorf("tên ảnh = %q", img.Name)
	}
	if len(img.Data) == 0 {
		t.Error("dữ liệu ảnh rỗng")
	}
	// PNG magic header.
	if len(img.Data) < 8 || string(img.Data[1:4]) != "PNG" {
		t.Errorf("dữ liệu không phải PNG hợp lệ: % x", img.Data[:min(8, len(img.Data))])
	}
}

func TestParseDocx_ImplicitSectionForBodyUnderChapter(t *testing.T) {
	book, err := ParseDocx(writeSampleDocx(t))
	if err != nil {
		t.Fatalf("ParseDocx: %v", err)
	}
	// Chương III có 1 đoạn body trước Heading2 đầu tiên → tạo tiểu mục ngầm.
	ch3 := book.Chapters[2]
	if len(ch3.Sections) < 2 {
		t.Fatalf("chương III muốn >=2 tiểu mục (1 ngầm + 1 Heading2), có %d", len(ch3.Sections))
	}
	if !strings.Contains(ch3.Sections[0].Text, "con người") {
		t.Errorf("tiểu mục ngầm chương III sai: %q", ch3.Sections[0].Text)
	}
}

// para là helper dựng docxPara text-only cho test buildBook.
func para(style, text string) docxPara {
	return docxPara{style: style, text: text}
}

func TestMinHeadingLevel(t *testing.T) {
	cases := []struct {
		name  string
		paras []docxPara
		want  int
	}{
		{"H1/H2", []docxPara{para("Title", "T"), para("Heading1", "C1"), para("Heading2", "S1")}, 1},
		{"H2/H3 không H1", []docxPara{para("Title", "T"), para("Heading2", "C1"), para("Heading3", "S1")}, 2},
		{"chỉ text + Title", []docxPara{para("", "abc"), para("Title", "T")}, 0},
		{"chỉ H2", []docxPara{para("Heading2", "C1"), para("Heading2", "C2")}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := minHeadingLevel(tc.paras); got != tc.want {
				t.Errorf("minHeadingLevel = %d, muốn %d", got, tc.want)
			}
		})
	}
}

func TestBuildBook_BaseHeadingLevel(t *testing.T) {
	cases := []struct {
		name      string
		paras     []docxPara
		wantTitle string
		// wantChapters: tiêu đề chương → số tiểu mục mong đợi.
		wantChapters []struct {
			title    string
			sections int
		}
	}{
		{
			name: "H1 chương / H2 tiểu mục (giữ hành vi cũ)",
			paras: []docxPara{
				para("Title", "T"),
				para("Heading1", "C1"),
				para("Heading2", "S1"),
				para("", "b1"),
			},
			wantTitle: "T",
			wantChapters: []struct {
				title    string
				sections int
			}{{"C1", 1}},
		},
		{
			name: "H2 chương / H3 tiểu mục (không có H1)",
			paras: []docxPara{
				para("Title", "Sổ tay mẫu"),
				para("Heading2", "Lời mở đầu"),
				para("", "Mở đầu nội dung"),
				para("Heading2", "Chương 1"),
				para("Heading3", "Tiểu mục 1.1"),
				para("", "noi dung 1.1"),
				para("Heading3", "Tiểu mục 1.2"),
				para("", "noi dung 1.2"),
				para("Heading2", "Chương 2"),
				para("", "noi dung chuong 2"),
			},
			wantTitle: "Sổ tay mẫu",
			wantChapters: []struct {
				title    string
				sections int
			}{{"Lời mở đầu", 1}, {"Chương 1", 2}, {"Chương 2", 1}},
		},
		{
			name:      "không heading → 1 chương Nội dung",
			paras:     []docxPara{para("", "abc")},
			wantTitle: "",
			wantChapters: []struct {
				title    string
				sections int
			}{{"Nội dung", 1}},
		},
		{
			name: "chỉ 1 cấp heading (toàn H2) → mỗi H2 là chương",
			paras: []docxPara{
				para("Heading2", "C1"),
				para("", "b1"),
				para("Heading2", "C2"),
				para("", "b2"),
			},
			wantTitle: "",
			wantChapters: []struct {
				title    string
				sections int
			}{{"C1", 1}, {"C2", 1}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			book := buildBook(tc.paras, nil, nil)
			if book.Title != tc.wantTitle {
				t.Errorf("title = %q, muốn %q", book.Title, tc.wantTitle)
			}
			if len(book.Chapters) != len(tc.wantChapters) {
				t.Fatalf("số chương = %d, muốn %d", len(book.Chapters), len(tc.wantChapters))
			}
			for i, wc := range tc.wantChapters {
				if book.Chapters[i].Title != wc.title {
					t.Errorf("chương %d title = %q, muốn %q", i, book.Chapters[i].Title, wc.title)
				}
				if len(book.Chapters[i].Sections) != wc.sections {
					t.Errorf("chương %d (%q) có %d tiểu mục, muốn %d",
						i, wc.title, len(book.Chapters[i].Sections), wc.sections)
				}
			}
		})
	}
}

func TestIsCaptionStyle(t *testing.T) {
	cases := map[string]bool{
		"ImageCaption":    true,
		"CaptionedFigure": true,
		"Caption":         true,
		"TableCaption":    true,
		"Heading1":        false,
		"Title":           false,
		"Normal":          false,
		"":                false,
	}
	for in, want := range cases {
		if got := isCaptionStyle(in); got != want {
			t.Errorf("isCaptionStyle(%q) = %v, muốn %v", in, got, want)
		}
	}
}

func TestBuildBook_DropsCaptionText(t *testing.T) {
	paras := []docxPara{
		para("Heading2", "Chương 1"),
		para("ImageCaption", "Hình 1 — Sơ đồ quy trình mẫu"),
		para("", "Nội dung thật của chương."),
	}
	book := buildBook(paras, nil, nil)
	if len(book.Chapters) != 1 {
		t.Fatalf("muốn 1 chương, có %d", len(book.Chapters))
	}
	var full string
	for _, s := range book.Chapters[0].Sections {
		full += s.Text + "\n"
	}
	if strings.Contains(full, "Hình 1") {
		t.Errorf("caption 'Hình 1' không được vào lời đọc, text = %q", full)
	}
	if !strings.Contains(full, "Nội dung thật") {
		t.Errorf("nội dung thật bị mất, text = %q", full)
	}
}

func TestHeadingLevel(t *testing.T) {
	cases := map[string]int{
		"Heading1":  1,
		"Heading 2": 2,
		"heading3":  3,
		"Title":     0,
		"":          0,
		"Normal":    0,
	}
	for in, want := range cases {
		if got := headingLevel(in); got != want {
			t.Errorf("headingLevel(%q) = %d, muốn %d", in, got, want)
		}
	}
}
