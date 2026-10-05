package bookmaker

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
)

// ── test logic với nội dung giả (không cần PDFium) ──────────────────────

// line tạo 1 dòng chữ thân bài cỡ 12pt ở độ cao top.
func line(text string, top float64) pdfLine {
	return pdfLine{Text: text, X0: 50, X1: 540, Top: top, Bottom: top + 10, Size: 12}
}

// bodyPage — trang A4 có đầu trang, chân trang (số trang) và các dòng thân bài.
func bodyPage(n int, body ...pdfLine) pdfPage {
	lines := append([]pdfLine{}, body...)
	lines = append(lines,
		pdfLine{Text: "Kỹ năng giao tiếp – Sano", X0: 250, X1: 340, Top: 16, Bottom: 21, Size: 9},
		pdfLine{Text: itoa(n), X0: 296, X1: 300, Top: 821, Bottom: 826, Size: 9},
	)
	return pdfPage{W: 595, H: 842, Lines: lines}
}

var itoa = strconv.Itoa

func TestStripRunningLines(t *testing.T) {
	doc := &pdfDoc{}
	for i := 1; i <= 4; i++ {
		doc.Pages = append(doc.Pages, bodyPage(i, line("Đoạn văn trang này.", 60), line("Dòng tiếp theo.", 76)))
	}
	if n := stripRunningLines(doc); n != 8 {
		t.Fatalf("muốn bỏ 8 dòng (đầu trang + số trang × 4), bỏ %d", n)
	}
	for i, p := range doc.Pages {
		if len(p.Lines) != 2 || !strings.HasPrefix(p.Lines[0].Text, "Đoạn") {
			t.Errorf("trang %d còn %+v", i+1, p.Lines)
		}
	}
}

func TestStripRunningLines_KeepsBigChapterTitles(t *testing.T) {
	doc := &pdfDoc{}
	for i := 1; i <= 3; i++ {
		title := pdfLine{Text: "Chương " + itoa(i), X0: 50, X1: 200, Top: 50, Bottom: 70, Size: 24}
		doc.Pages = append(doc.Pages, bodyPage(i, title, line("Nội dung.", 110), line("Thêm nữa.", 126)))
	}
	stripRunningLines(doc)
	for i, p := range doc.Pages {
		if p.Lines[0].Text != "Chương "+itoa(i+1) {
			t.Errorf("trang %d mất tên chương (chữ to, lặp số): %+v", i+1, p.Lines)
		}
	}
}

func TestJoinPDFLine(t *testing.T) {
	cases := []struct{ a, b, want string }{
		{"", "đầu", "đầu"},
		{"docu-", "mentation is", "documentation is"},
		{"năm 2020-", "2021", "năm 2020- 2021"},
		{"Hà Nội -", "Huế", "Hà Nội - Huế"},
		{"soft\u00ad", "ware", "software"},
		{"một dòng", "dòng sau", "một dòng dòng sau"},
	}
	for _, c := range cases {
		if got := joinPDFLine(c.a, c.b); got != c.want {
			t.Errorf("joinPDFLine(%q, %q) = %q, muốn %q", c.a, c.b, got, c.want)
		}
	}
}

func TestPDFParagraphs(t *testing.T) {
	short := line("hết đoạn một.", 76)
	short.X1 = 200
	p1 := pdfPage{W: 595, H: 842, Lines: []pdfLine{
		line("Đoạn một dòng đầu", 60), short,
		line("Đoạn hai bắt đầu", 92), line("và chưa hết câu", 108),
	}}
	p2 := pdfPage{W: 595, H: 842, Lines: []pdfLine{
		line("nên nối sang trang sau.", 60),
		{Text: "Tiêu đề in đậm", X0: 50, X1: 200, Top: 90, Bottom: 100, Size: 12, Bold: true},
		line("Đoạn ba.", 106),
	}}
	got := pdfParagraphs(&pdfDoc{Pages: []pdfPage{p1, p2}})
	var texts []string
	for _, p := range got {
		texts = append(texts, p.para.text)
	}
	want := []string{
		"Đoạn một dòng đầu hết đoạn một.",
		"Đoạn hai bắt đầu và chưa hết câu nên nối sang trang sau.",
		"Tiêu đề in đậm",
		"Đoạn ba.",
	}
	if strings.Join(texts, "|") != strings.Join(want, "|") {
		t.Errorf("đoạn =\n%q\nmuốn\n%q", texts, want)
	}
	if got[2].para.boldRunes != got[2].para.textRunes || got[2].para.maxSz != 24 {
		t.Errorf("đoạn đậm phải ghi boldRunes/maxSz cho cảnh báo tiêu đề gõ tay: %+v", got[2].para)
	}
}

func TestBookFromPDF_Outline(t *testing.T) {
	doc := &pdfDoc{
		Title: "Microsoft Word - ban-thao.docx",
		Pages: []pdfPage{
			{W: 595, H: 842, Lines: []pdfLine{{Text: "Sách mẫu", X0: 50, X1: 300, Top: 60, Bottom: 90, Size: 30}}},
			{W: 595, H: 842, Lines: []pdfLine{
				{Text: "Chương 1. Mở đầu", X0: 50, X1: 300, Top: 60, Bottom: 80, Size: 20},
				line("Nội dung chương một đủ dài để đọc thành lời, có thêm vài câu cho đủ chữ.", 100),
				line("Câu thứ hai của chương một, viết thêm cho đủ độ dài cần thiết.", 116),
			}},
			{W: 595, H: 842, Lines: []pdfLine{line("Trang không có dòng tiêu đề, bookmark trỏ vào đây.", 60)}},
		},
		Outline: []pdfOutline{
			{Title: "Chương 1. Mở đầu", Level: 1, Page: 1},
			{Title: "Mục 1.1", Level: 2, Page: 2},
			{Title: "Bookmark hỏng", Level: 1, Page: 99},
		},
	}
	book, err := bookFromPDF(doc)
	if err != nil {
		t.Fatal(err)
	}
	if book.Title != "" {
		t.Errorf("tên sách rác từ metadata phải bỏ, được %q", book.Title)
	}
	if len(book.Chapters) != 2 || book.Chapters[1].Title != "Chương 1. Mở đầu" {
		t.Fatalf("chương = %+v", book.Chapters)
	}
	ch := book.Chapters[1]
	if len(ch.Sections) != 2 || ch.Sections[1].Title != "Mục 1.1" || !strings.HasPrefix(ch.Sections[1].Text, "Trang không có dòng tiêu đề") {
		t.Errorf("mục = %+v", ch.Sections)
	}
	for _, n := range book.Stats.Notes {
		if strings.Contains(n.Text, "bookmark") {
			t.Errorf("có bookmark thì không cảnh báo thiếu bookmark: %q", n.Text)
		}
	}
	if doc.Pages[1].Lines[0].Text != "Chương 1. Mở đầu" {
		t.Error("bookFromPDF không được sửa nội dung gốc (dùng lại từ cache)")
	}
}

func TestBookFromPDF_FontSizeHeadings(t *testing.T) {
	var pages []pdfPage
	pages = append(pages, pdfPage{W: 595, H: 842, Lines: []pdfLine{{Text: "Tên Sách To", X0: 50, X1: 400, Top: 300, Bottom: 340, Size: 36}}})
	for c := 1; c <= 2; c++ {
		pages = append(pages, pdfPage{W: 595, H: 842, Lines: []pdfLine{
			{Text: "Chương " + itoa(c), X0: 50, X1: 300, Top: 60, Bottom: 84, Size: 22},
			line("Đoạn mở đầu chương, đủ chữ để làm thân bài chuẩn.", 100),
			{Text: "Mục nhỏ", X0: 50, X1: 200, Top: 130, Bottom: 146, Size: 16},
			line("Nội dung mục nhỏ, cũng là thân bài cỡ chữ mười hai.", 160),
		}})
	}
	book, err := bookFromPDF(&pdfDoc{Pages: pages})
	if err != nil {
		t.Fatal(err)
	}
	if book.Title != "Tên Sách To" {
		t.Errorf("tên sách trên trang bìa = %q", book.Title)
	}
	if len(book.Chapters) != 2 || book.Chapters[0].Title != "Chương 1" || len(book.Chapters[0].Sections) != 2 || book.Chapters[0].Sections[1].Title != "Mục nhỏ" {
		t.Fatalf("chương/mục đoán theo cỡ chữ sai: %+v", book.Chapters)
	}
	if len(book.Stats.Notes) == 0 || !strings.Contains(book.Stats.Notes[0].Text, "không có mục lục (bookmark)") {
		t.Errorf("phải cảnh báo thiếu bookmark: %+v", book.Stats.Notes)
	}
}

func TestBookFromPDF_Quality(t *testing.T) {
	text := strings.Repeat("Người Việt Nam có câu học ăn học nói. ", 4)
	good := pdfPage{W: 595, H: 842, Lines: []pdfLine{line(text, 60), line(text, 76)}}
	tcvn := pdfPage{W: 595, H: 842, Lines: []pdfLine{line(strings.Repeat("Ng­êi ViÖt Nam cã c©u häc ¨n. ", 6), 60)}}
	scan := pdfPage{W: 595, H: 842, Images: 1}

	// Ít trang lỗi phông + 1 trang scan: vẫn nạp, có cảnh báo.
	book, err := bookFromPDF(&pdfDoc{Pages: []pdfPage{good, good, good, good, tcvn, scan}})
	if err != nil {
		t.Fatal(err)
	}
	var garble, empty *LoadNote
	for i, n := range book.Stats.Notes {
		switch {
		case strings.Contains(n.Text, "lỗi phông"):
			garble = &book.Stats.Notes[i]
		case strings.Contains(n.Text, "không có chữ"):
			empty = &book.Stats.Notes[i]
		}
	}
	if garble == nil || !strings.Contains(garble.Text, "trang 5") {
		t.Errorf("thiếu cảnh báo lỗi phông kèm số trang: %+v", book.Stats.Notes)
	}
	if empty == nil || !strings.Contains(empty.Text, "1/6 trang") || empty.Severe {
		t.Errorf("thiếu cảnh báo trang scan (nhẹ): %+v", book.Stats.Notes)
	}

	// Lỗi phông nhiều: cảnh báo nặng.
	book, err = bookFromPDF(&pdfDoc{Pages: []pdfPage{tcvn, tcvn}})
	if err != nil {
		t.Fatal(err)
	}
	severe := false
	for _, n := range book.Stats.Notes {
		severe = severe || (n.Severe && strings.Contains(n.Text, "lỗi phông"))
	}
	if !severe {
		t.Errorf("lỗi phông nhiều phải là cảnh báo nặng: %+v", book.Stats.Notes)
	}

	// Gần như toàn bộ lỗi phông: từ chối.
	junk := pdfPage{W: 595, H: 842, Lines: []pdfLine{line(strings.Repeat("\uFFFD\uE001\uFFFDa ", 60), 60)}}
	if _, err := bookFromPDF(&pdfDoc{Pages: []pdfPage{junk, junk}}); !errors.Is(err, ErrPDFGarbled) {
		t.Errorf("chữ rác gần hết phải trả ErrPDFGarbled, được %v", err)
	}
	// Không có chữ: từ chối.
	if _, err := bookFromPDF(&pdfDoc{Pages: []pdfPage{scan, scan}}); !errors.Is(err, ErrPDFNoText) {
		t.Errorf("PDF toàn ảnh phải trả ErrPDFNoText, được %v", err)
	}
}

func TestBadRune_VietnameseUnicodeIsClean(t *testing.T) {
	for _, r := range "Người Việt Nam có câu “học ăn, học nói” – © 2026, 25°C, «ghi chú» ĐÀ NẴNG ỹ ữ ặ" {
		if badRune(r) {
			t.Errorf("chữ Việt Unicode %q bị coi là lỗi phông", r)
		}
	}
	for _, r := range "Öä¬¨ªµ¸" { // ViÖt (TCVN3), Vieät (VNI), ¬ ¨ ª … thay ơ ă ê
		if !badRune(r) {
			t.Errorf("ký tự phông cũ %q không bị nhận ra", r)
		}
	}
	if badRune('\u00ad') {
		t.Error("gạch nối mềm (Word chèn khi ngắt từ) không phải lỗi phông")
	}
}

func TestMultiColumnPages(t *testing.T) {
	var lines []pdfLine
	for i := 0; i < 8; i++ {
		lines = append(lines,
			pdfLine{Text: "cột trái", X0: 50, X1: 280, Top: float64(60 + 16*i), Size: 12},
			pdfLine{Text: "cột phải", X0: 310, X1: 540, Top: float64(60 + 16*i), Size: 12})
	}
	one := pdfPage{W: 595, H: 842, Lines: []pdfLine{}}
	for i := 0; i < 12; i++ {
		one.Lines = append(one.Lines, line("một cột", float64(60+16*i)))
	}
	got := multiColumnPages(&pdfDoc{Pages: []pdfPage{one, {W: 595, H: 842, Lines: lines}}})
	if len(got) != 1 || got[0] != 2 {
		t.Errorf("trang nhiều cột = %v, muốn [2]", got)
	}
}

func TestPDFTitle(t *testing.T) {
	for in, want := range map[string]string{
		"Kỹ năng giao tiếp":         "Kỹ năng giao tiếp",
		"Microsoft Word - abc.docx": "",
		"untitled":                  "",
		"ban-thao-cuoi.pdf":         "",
		"  Tâm  lý   tích cực ":     "Tâm lý tích cực",
	} {
		if got := pdfTitle(in); got != want {
			t.Errorf("pdfTitle(%q) = %q, muốn %q", in, got, want)
		}
	}
}

// ── PDF thật qua PDFium (testdata/pdf) ──────────────────────────────────

func TestParsePDF_RealFiles(t *testing.T) {
	book, err := ParseBook("testdata/pdf/outline.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if book.Title != "Kỹ năng giao tiếp" || len(book.Chapters) != 3 {
		t.Fatalf("Title %q, %d chương", book.Title, len(book.Chapters))
	}
	ch := book.Chapters[0]
	if ch.Title != "Chương 1. Lắng nghe phần 1" || len(ch.Sections) != 3 || ch.Sections[2].Title != "Mục 1.2 Đặt câu hỏi" {
		t.Fatalf("chương 1 = %+v", ch)
	}
	sec := ch.Sections[2].Text
	if strings.Count(sec, "\n\n") != 4 || strings.Contains(sec, "Sano") || strings.Contains(sec, "– Sano") {
		t.Errorf("mục 1.2 phải có 5 đoạn, không lẫn đầu trang:\n%s", sec)
	}
	if !strings.Contains(sec, "hiệu quả hơn.\n\nĐoạn văn thứ 2") {
		t.Errorf("đoạn cắt ngang hai trang phải được nối lại:\n%s", sec)
	}

	out, err := Inspect("testdata/pdf/outline.pdf", InspectOptions{})
	if err != nil || out.FileTitle != "outline" || out.Chars == 0 || len(out.Warnings.Notes) != 0 {
		t.Errorf("Inspect = %+v, %v", out, err)
	}

	book, err = ParseBook("testdata/pdf/tcvn3.pdf")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range book.Stats.Notes {
		found = found || (n.Severe && strings.Contains(n.Text, "lỗi phông"))
	}
	if !found {
		t.Errorf("tcvn3.pdf phải có cảnh báo nặng lỗi phông: %+v", book.Stats.Notes)
	}

	if _, err := ParseBook("testdata/pdf/scanned.pdf"); !errors.Is(err, ErrPDFNoText) {
		t.Errorf("scanned.pdf: %v", err)
	}
	for _, f := range []string{"testdata/pdf/password.pdf", "testdata/pdf/nocopy.pdf"} {
		if _, err := ParseBook(f); !errors.Is(err, ErrProtectedFile) {
			t.Errorf("%s phải trả ErrProtectedFile, được %v", f, err)
		}
	}
}

func TestParsePDF_Invalid(t *testing.T) {
	p := t.TempDir() + "/hong.pdf"
	if err := os.WriteFile(p, []byte("%PDF-1.4 không phải PDF thật"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseBook(p); err == nil {
		t.Error("PDF hỏng phải báo lỗi")
	}
	// Sau file hỏng, bộ đọc PDF vẫn đọc được file khác (bản sao để không trúng cache).
	data, err := os.ReadFile("testdata/pdf/outline.pdf")
	if err != nil {
		t.Fatal(err)
	}
	cp := t.TempDir() + "/ban-sao.pdf"
	if err := os.WriteFile(cp, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if b, err := ParseBook(cp); err != nil || len(b.Chapters) != 3 {
		t.Errorf("đọc PDF sau file hỏng: %v", err)
	}
	old := maxPDFBytes
	maxPDFBytes = 10
	defer func() { maxPDFBytes = old }()
	if _, err := ParseBook("testdata/pdf/tcvn3.pdf"); !errors.Is(err, ErrPDFTooLarge) {
		t.Errorf("vượt dung lượng phải trả ErrPDFTooLarge, được %v", err)
	}
}
