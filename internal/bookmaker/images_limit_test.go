package bookmaker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repeatImageDoc dựng document.xml: 2 chương, mỗi chương 1 tiểu mục; đoạn đầu
// tham chiếu các rel trong refs1, đoạn sau tham chiếu refs2.
func repeatImageDoc(refs1, refs2 []string) string {
	blips := func(refs []string) string {
		var b strings.Builder
		for _, r := range refs {
			b.WriteString(`<w:r><w:drawing><a:blip r:embed="` + r + `"/></w:drawing></w:r>`)
		}
		return b.String()
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
  xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<w:body>
<w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>Chương một</w:t></w:r></w:p>
<w:p><w:r><w:t>Đoạn một.</w:t></w:r>` + blips(refs1) + `</w:p>
<w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>Chương hai</w:t></w:r></w:p>
<w:p><w:r><w:t>Đoạn hai.</w:t></w:r>` + blips(refs2) + `</w:p>
</w:body></w:document>`
}

func repeat(s string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = s
	}
	return out
}

// Một ảnh tham chiếu lặp lại: chỉ gắn 1 lần mỗi tiểu mục, ghi đĩa 1 lần cả cuốn.
func TestParseDocx_RepeatedImage_OncePerSection(t *testing.T) {
	base := t.TempDir()
	docx := filepath.Join(base, "lap.docx")
	writeZip(t, docx, []zipEntry{
		{"word/document.xml", []byte(repeatImageDoc(repeat("rId3", 500), repeat("rId3", 50)))},
		{"word/_rels/document.xml.rels", []byte(bombRels)},
		{"word/media/c.png", samplePNG()},
	})
	book, err := ParseDocx(docx)
	if err != nil {
		t.Fatal(err)
	}
	for ci, ch := range book.Chapters {
		if n := len(ch.Sections[0].Images); n != 1 {
			t.Errorf("chương %d: ảnh lặp phải gắn 1 lần, có %d", ci+1, n)
		}
	}
	if book.Stats.Images != 2 {
		t.Errorf("Stats.Images = %d, muốn 2", book.Stats.Images)
	}

	// Ghi ra images/: ảnh đã ghi ở tiểu mục trước không ghi lại.
	out := filepath.Join(base, "out")
	if err := os.MkdirAll(filepath.Join(out, "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	o := Options{OutputDir: out}
	written := map[string]bool{}
	if _, err := o.processImages(book.Chapters[0].Sections[0], "C1", written); err != nil {
		t.Fatal(err)
	}
	img := filepath.Join(out, "images", book.Chapters[0].Sections[0].Images[0].Name)
	if err := os.Remove(img); err != nil {
		t.Fatalf("ảnh phải được ghi lần đầu: %v", err)
	}
	if _, err := o.processImages(book.Chapters[1].Sections[0], "C2", written); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(img); !os.IsNotExist(err) {
		t.Error("ảnh đã ghi không được ghi lại lần hai")
	}
}

// Vượt giới hạn tổng số lần gắn ảnh mỗi tài liệu → bỏ phần dư, kèm cảnh báo.
func TestParseDocx_TooManyImageRefs_Dropped(t *testing.T) {
	old := maxImageRefs
	maxImageRefs = 2
	t.Cleanup(func() { maxImageRefs = old })
	docx := filepath.Join(t.TempDir(), "nhieu.docx")
	writeZip(t, docx, []zipEntry{
		{"word/document.xml", []byte(repeatImageDoc([]string{"rId1", "rId2"}, []string{"rId3", "rId1"}))},
		{"word/_rels/document.xml.rels", []byte(bombRels)},
		{"word/media/a.png", samplePNG()},
		{"word/media/b.png", samplePNG()},
		{"word/media/c.png", samplePNG()},
	})
	book, err := ParseDocx(docx)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(book.Chapters[0].Sections[0].Images); n != 2 {
		t.Errorf("tiểu mục 1 phải giữ 2 ảnh, có %d", n)
	}
	if n := len(book.Chapters[1].Sections[0].Images); n != 0 {
		t.Errorf("tiểu mục 2 vượt giới hạn phải bỏ ảnh, có %d", n)
	}
	if book.Stats.DroppedImages != 2 {
		t.Errorf("DroppedImages = %d, muốn 2", book.Stats.DroppedImages)
	}
	var sb strings.Builder
	printLoadWarnings(&sb, book.Stats, nil)
	if !strings.Contains(sb.String(), "vượt giới hạn") {
		t.Errorf("thiếu cảnh báo vượt giới hạn: %s", sb.String())
	}
}
