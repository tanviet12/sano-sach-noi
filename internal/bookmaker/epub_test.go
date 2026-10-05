package bookmaker

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeEpub ghi file .epub tạm từ map đường dẫn → nội dung (mimetype tự thêm).
func writeEpub(t *testing.T, files map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "sach-thu.epub")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	_, _ = w.Write([]byte("application/epub+zip"))
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return p
}

const epubContainer = `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`

func xhtml(body string) string {
	return `<?xml version="1.0" encoding="utf-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head><title>bỏ qua</title><style>p{}</style></head><body>` + body + `</body></html>`
}

// epub3Files — sách EPUB 3: bìa chỉ có ảnh, trang tên sách (h1 trùng dc:title),
// 2 chương h1 có mục h2, chú thích, ảnh, bảng, thực thể HTML.
func epub3Files() map[string]string {
	return map[string]string{
		"META-INF/container.xml": epubContainer,
		"OEBPS/content.opf": `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title>Kỹ năng sống</dc:title><dc:creator>Tác giả</dc:creator></metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="cover" href="Text/cover.xhtml" media-type="application/xhtml+xml"/>
    <item id="title" href="Text/title.xhtml" media-type="application/xhtml+xml"/>
    <item id="c1" href="Text/chuong%201.xhtml" media-type="application/xhtml+xml"/>
    <item id="c2" href="Text/c2.xhtml" media-type="application/xhtml+xml"/>
    <item id="img" href="Images/hinh.png" media-type="image/png"/>
    <item id="coverimg" href="Images/bia.jpg" media-type="image/jpeg"/>
  </manifest>
  <spine><itemref idref="cover"/><itemref idref="nav"/><itemref idref="title"/><itemref idref="c1"/><itemref idref="c2"/></spine>
</package>`,
		"OEBPS/nav.xhtml":        xhtml(`<nav epub:type="toc"><ol><li><a href="Text/chuong%201.xhtml">Chương một</a></li></ol></nav>`),
		"OEBPS/Text/cover.xhtml": xhtml(`<div><img src="../Images/bia.jpg" alt="bìa"/></div>`),
		"OEBPS/Text/title.xhtml": xhtml(`<h1>Kỹ năng sống</h1>`),
		"OEBPS/Text/chuong 1.xhtml": xhtml(`<h1>Chương 1. Giao tiếp</h1>
<p>Lắng&nbsp;nghe trước<a epub:type="noteref" href="#n1">1</a>, nói sau.</p>
<h2>Mục 1.1 Hỏi</h2><p>Đặt câu hỏi <em>mở</em>.</p><p><img src="../Images/hinh.png"/></p>
<figure><img src="../Images/hinh.png"/><figcaption>Hình 1. Sơ đồ</figcaption></figure>
<aside epub:type="footnote" id="n1"><p>Chú thích không đọc.</p></aside>`),
		"OEBPS/Text/c2.xhtml": xhtml(`<h1>Chương 2. Làm việc nhóm</h1><h2>Mục 2.1</h2>
<table><tr><td>Ô một</td><td>Ô hai</td></tr></table><p>Dòng một<br/>dòng hai.</p>`),
		"OEBPS/Images/hinh.png": "PNGDATA",
		"OEBPS/Images/bia.jpg":  "JPGDATA",
	}
}

func TestParseEpub_EPUB3Structure(t *testing.T) {
	book, err := ParseEpub(writeEpub(t, epub3Files()))
	if err != nil {
		t.Fatal(err)
	}
	if book.Title != "Kỹ năng sống" {
		t.Errorf("Title = %q", book.Title)
	}
	if len(book.Chapters) != 2 {
		t.Fatalf("muốn 2 chương (bìa, mục lục, trang tên sách không thành chương), được %d: %+v", len(book.Chapters), book.Chapters)
	}
	c1 := book.Chapters[0]
	if c1.Title != "Chương 1. Giao tiếp" || len(c1.Sections) != 2 {
		t.Fatalf("chương 1 = %q, %d mục", c1.Title, len(c1.Sections))
	}
	if got := c1.Sections[0].Text; got != "Lắng nghe trước, nói sau." {
		t.Errorf("đoạn mở đầu chương = %q (phải bỏ số chú thích, đổi &nbsp;)", got)
	}
	s11 := c1.Sections[1]
	if s11.Title != "Mục 1.1 Hỏi" || s11.Text != "Đặt câu hỏi mở." {
		t.Errorf("mục 1.1 = %q / %q", s11.Title, s11.Text)
	}
	if strings.Contains(s11.Text, "Chú thích") || strings.Contains(s11.Text, "Sơ đồ") {
		t.Errorf("không được đọc chú thích cuối trang hay chú thích hình: %q", s11.Text)
	}
	if len(s11.Images) != 1 || string(s11.Images[0].Data) != "PNGDATA" || s11.Images[0].Name != "img001.png" {
		t.Errorf("ảnh mục 1.1 = %+v", s11.Images)
	}
	s21 := book.Chapters[1].Sections[0]
	if s21.Text != "Ô một\n\nÔ hai\n\nDòng một dòng hai." {
		t.Errorf("mục 2.1 = %q", s21.Text)
	}
	if book.Stats.Tables != 1 || book.Stats.Images != 1 {
		t.Errorf("Stats = %+v (muốn 1 bảng, 1 ảnh; ảnh bìa không tính)", book.Stats)
	}
	if s21.Stem != "ch02-sec01" {
		t.Errorf("Stem = %q", s21.Stem)
	}
}

// EPUB 2: mục lục NCX, trang không có thẻ h (tiêu đề gõ bằng <p class>), mục trỏ neo #id.
func TestParseEpub_EPUB2NCXHeadings(t *testing.T) {
	files := map[string]string{
		"META-INF/container.xml": epubContainer,
		"OEBPS/content.opf": `<package xmlns="http://www.idpf.org/2007/opf" version="2.0">
  <metadata><dc:title xmlns:dc="http://purl.org/dc/elements/1.1/">Sách cũ</dc:title></metadata>
  <manifest>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    <item id="p1" href="p1.html" media-type="application/xhtml+xml"/>
  </manifest>
  <spine toc="ncx"><itemref idref="p1"/></spine>
</package>`,
		"OEBPS/toc.ncx": `<ncx><navMap>
  <navPoint id="a"><navLabel><text>Phần một</text></navLabel><content src="p1.html"/>
    <navPoint id="b"><navLabel><text>Mở đầu</text></navLabel><content src="p1.html#s2"/></navPoint>
  </navPoint>
</navMap></ncx>`,
		"OEBPS/p1.html": xhtml(`<p class="title">Phần một</p><p>Lời dẫn.</p><p id="s2" class="sub">Mở đầu</p><p>Nội dung mở đầu.</p>`),
	}
	book, err := ParseEpub(writeEpub(t, files))
	if err != nil {
		t.Fatal(err)
	}
	if book.Title != "Sách cũ" || len(book.Chapters) != 1 {
		t.Fatalf("Title %q, %d chương", book.Title, len(book.Chapters))
	}
	ch := book.Chapters[0]
	if ch.Title != "Phần một" || len(ch.Sections) != 2 {
		t.Fatalf("chương = %q, mục = %+v", ch.Title, ch.Sections)
	}
	if ch.Sections[0].Text != "Lời dẫn." {
		t.Errorf("không được đọc lặp tiêu đề: %q", ch.Sections[0].Text)
	}
	if ch.Sections[1].Title != "Mở đầu" || ch.Sections[1].Text != "Nội dung mở đầu." {
		t.Errorf("mục 2 = %q / %q", ch.Sections[1].Title, ch.Sections[1].Text)
	}
}

func TestParseEpub_DRM(t *testing.T) {
	enc := func(alg string) string {
		return `<encryption xmlns="urn:oasis:names:tc:opendocument:xmlns:container" xmlns:enc="http://www.w3.org/2001/04/xmlenc#">
<enc:EncryptedData><enc:EncryptionMethod Algorithm="` + alg + `"/>
<enc:CipherData><enc:CipherReference URI="OEBPS/Text/c2.xhtml"/></enc:CipherData></enc:EncryptedData></encryption>`
	}
	files := epub3Files()
	files["META-INF/encryption.xml"] = enc("http://www.w3.org/2001/04/xmlenc#aes128-cbc")
	if _, err := ParseEpub(writeEpub(t, files)); !errors.Is(err, ErrProtectedFile) {
		t.Errorf("nội dung mã hoá phải trả ErrProtectedFile, được %v", err)
	}
	files["META-INF/encryption.xml"] = enc("http://www.idpf.org/2008/embedding")
	if _, err := ParseEpub(writeEpub(t, files)); err != nil {
		t.Errorf("font làm rối theo chuẩn IDPF không phải DRM: %v", err)
	}
	delete(files, "META-INF/encryption.xml")
	files["META-INF/rights.xml"] = "<rights/>"
	if _, err := ParseEpub(writeEpub(t, files)); !errors.Is(err, ErrProtectedFile) {
		t.Errorf("rights.xml (Adobe DRM) phải trả ErrProtectedFile, được %v", err)
	}
}

func TestParseEpub_Invalid(t *testing.T) {
	if _, err := ParseEpub(writeEpub(t, map[string]string{"a.txt": "x"})); err == nil || !strings.Contains(err.Error(), "không phải EPUB") {
		t.Errorf("thiếu container.xml: %v", err)
	}
	files := epub3Files()
	for k := range files {
		if strings.HasPrefix(k, "OEBPS/Text/") {
			files[k] = xhtml("")
		}
	}
	if _, err := ParseEpub(writeEpub(t, files)); err == nil {
		t.Error("EPUB không có chữ phải báo lỗi")
	}
}

func TestParseEpub_TextLimit(t *testing.T) {
	old := maxEpubTextBytes
	maxEpubTextBytes = 200
	defer func() { maxEpubTextBytes = old }()
	if _, err := ParseEpub(writeEpub(t, epub3Files())); err == nil || !strings.Contains(err.Error(), "quá lớn") {
		t.Errorf("vượt giới hạn dung lượng phải báo lỗi, được %v", err)
	}
}

func TestParseBook_DispatchAndInspect(t *testing.T) {
	for _, c := range []struct {
		name string
		want bool
	}{{"a.docx", true}, {"a.EPUB", true}, {"a.Pdf", true}, {"a.txt", false}, {"a.doc", false}} {
		if got := IsSourceFile(c.name); got != c.want {
			t.Errorf("IsSourceFile(%q) = %v", c.name, got)
		}
	}
	p := writeEpub(t, epub3Files())
	out, err := Inspect(p, InspectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if out.Title != "Kỹ năng sống" || out.FileTitle != "sach thu" || len(out.Chapters) != 2 || out.Chars == 0 {
		t.Errorf("Inspect = %+v", out)
	}
	counts, err := CountWordsInDocx(p, []string{"nghe"})
	if err != nil || counts["nghe"] != 1 {
		t.Errorf("CountWordsInDocx = %v, %v", counts, err)
	}
}
