package bookmaker

import (
	"archive/zip"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// zipEntry — 1 file trong docx tổng hợp dùng cho test bảo mật.
type zipEntry struct {
	name string
	data []byte
}

// writeZip ghi các entry thành file zip (docx) ở path.
func writeZip(t *testing.T, path string, entries []zipEntry) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for _, e := range entries {
		w, err := zw.Create(e.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(e.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

const evilDocXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
  xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<w:body>
<w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>Chương một</w:t></w:r></w:p>
<w:p><w:r><w:t>Nội dung có hình.</w:t></w:r>
<w:r><w:drawing><a:blip r:embed="rId1"/></w:drawing></w:r>
<w:r><w:drawing><a:blip r:embed="rId2"/></w:drawing></w:r>
<w:r><w:drawing><a:blip r:embed="rId3"/></w:drawing></w:r>
<w:r><w:drawing><a:blip r:embed="rId3"/></w:drawing></w:r>
</w:p>
</w:body></w:document>`

const evilRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/..\..\..\evil.bat"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../../evil"/>
<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/image1.PNG"/>
</Relationships>`

// Tên ảnh lấy từ file Word (rels Target chứa ..\ hoặc ../) không được quyết
// định tên file ghi ra đĩa: Sano tự đặt tên img%03d<ext>, mọi file nằm trong
// OutputDir/images.
func TestPrepare_ImageNamesFromDocxCannotEscapeOutputDir(t *testing.T) {
	base := t.TempDir()
	docx := filepath.Join(base, "doc.docx")
	writeZip(t, docx, []zipEntry{
		{"word/document.xml", []byte(evilDocXML)},
		{"word/_rels/document.xml.rels", []byte(evilRels)},
		{`word/media/..\..\..\evil.bat`, []byte("@echo pwned")},
		{"../evil", []byte("pwned")},
		{"word/media/image1.PNG", samplePNG()},
	})
	out := filepath.Join(base, "out", "sach")
	o := Options{InputDocx: docx, OutputDir: out, TTS: TTSConfig{Mode: TTSModeStub}}
	if _, err := o.prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}

	// Không có file nào ngoài OutputDir (ngoài chính file docx).
	err := filepath.Walk(base, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || p == docx {
			return err
		}
		rel, _ := filepath.Rel(out, p)
		if !filepath.IsLocal(rel) {
			t.Errorf("file nằm ngoài thư mục sách: %s", p)
		}
		if strings.Contains(strings.ToLower(filepath.Base(p)), "evil") {
			t.Errorf("tên file lấy từ tài liệu: %s", p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(filepath.Join(out, "images"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	want := []string{"img001.bin", "img002.bin", "img003.png"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Errorf("ảnh trong images/ = %v, muốn %v", names, want)
	}
}

func TestSafeImageName(t *testing.T) {
	re := regexp.MustCompile(`^img\d{3}\.[a-z]+$`)
	cases := map[string]string{
		"image1.png":        "img001.png",
		"ẢNH.JPEG":          "img001.jpeg",
		`..\..\x.bat`:       "img001.bin",
		"../../x":           "img001.bin",
		"a.tif":             "img001.tif",
		"noext":             "img001.bin",
		`C:\Windows\x.webp`: "img001.webp",
	}
	for src, want := range cases {
		got := safeImageName(1, src)
		if got != want || !re.MatchString(got) {
			t.Errorf("safeImageName(%q) = %q, muốn %q", src, got, want)
		}
	}
}

func TestProcessImages_RejectsUnsafeName(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	o := Options{OutputDir: dir}
	for _, bad := range []string{"../x.png", `..\x.png`, "", "..", "a/b.png"} {
		sec := Section{Title: "T", Images: []SectionImage{{Name: bad, Data: []byte("x")}}}
		if _, err := o.processImages(sec, "C", nil); err == nil {
			t.Errorf("processImages(%q) phải lỗi", bad)
		}
	}
}

// metadata.json của thư mục sách người khác gửi trỏ file ra ngoài thư mục →
// --repack-dir phải từ chối, không đọc file đó vào gói zip.
func TestRepackZip_RejectsPathOutsideDir(t *testing.T) {
	cases := map[string]string{
		"file ..":        `{"title":"T","chapters":[{"title":"C","sections":[{"title":"S","file":"../secret.txt"}]}]}`,
		"file \\":        `{"title":"T","chapters":[{"title":"C","sections":[{"title":"S","file":"..\\secret.txt"}]}]}`,
		"file tuyệt đối": `{"title":"T","chapters":[{"title":"C","sections":[{"title":"S","file":"/etc/hosts"}]}]}`,
		"cover ..":       `{"title":"T","cover":"../secret.png","chapters":[{"title":"C","sections":[{"title":"S","file":"ch01-sec01.mp3"}]}]}`,
	}
	for name, meta := range cases {
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			dir := filepath.Join(base, "sach")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			for _, f := range []string{filepath.Join(base, "secret.txt"), filepath.Join(base, "secret.png"), filepath.Join(dir, "ch01-sec01.mp3")} {
				if err := os.WriteFile(f, []byte("bí mật"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(meta), 0o644); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(base, "out.zip")
			if _, err := RepackZip(dir, out, "v", nil); err == nil {
				t.Fatal("phải từ chối đường dẫn ra ngoài thư mục sách")
			}
			if _, err := os.Stat(out); err == nil {
				t.Error("không được tạo gói zip")
			}
		})
	}
}

// setLimits đặt tạm giới hạn giải nén nhỏ cho test, trả lại giá trị cũ khi xong.
func setLimits(t *testing.T, xml, img, total int64) {
	t.Helper()
	oldX, oldI, oldT := maxXMLBytes, maxImageBytes, maxTotalImageBytes
	maxXMLBytes, maxImageBytes, maxTotalImageBytes = xml, img, total
	t.Cleanup(func() { maxXMLBytes, maxImageBytes, maxTotalImageBytes = oldX, oldI, oldT })
}

const bombRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/a.png"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/b.png"/>
<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/c.png"/>
</Relationships>`

// Ảnh vượt giới hạn (từng ảnh hoặc tổng) bị bỏ kèm cảnh báo, cuốn sách vẫn nạp được.
func TestParseDocx_OversizedImages_SkippedWithWarning(t *testing.T) {
	setLimits(t, 1<<20, 1000, 1500)
	path := filepath.Join(t.TempDir(), "bomb.docx")
	writeZip(t, path, []zipEntry{
		{"word/document.xml", []byte(evilDocXML)}, // rId1, rId2, rId3, rId3
		{"word/_rels/document.xml.rels", []byte(bombRels)},
		{"word/media/a.png", make([]byte, 5000)}, // quá giới hạn 1 ảnh
		{"word/media/b.png", make([]byte, 900)},  // giữ
		{"word/media/c.png", make([]byte, 900)},  // tổng 1800 > 1500 → bỏ
	})
	book, err := ParseDocx(path)
	if err != nil {
		t.Fatalf("ParseDocx: %v", err)
	}
	imgs := book.Chapters[0].Sections[0].Images
	if len(imgs) != 1 || len(imgs[0].Data) != 900 {
		t.Fatalf("muốn giữ đúng 1 ảnh (b.png), có %d", len(imgs))
	}
	if book.Stats.SkippedImages != 2 {
		t.Errorf("SkippedImages = %d, muốn 2 (a.png, c.png)", book.Stats.SkippedImages)
	}
}

// document.xml bung vượt giới hạn → lỗi rõ ràng.
func TestParseDocx_OversizedDocumentXML_Errors(t *testing.T) {
	setLimits(t, 4<<10, 1<<20, 1<<20)
	path := filepath.Join(t.TempDir(), "bomb.docx")
	big := strings.Replace(evilDocXML, "Nội dung có hình.", strings.Repeat("a", 64<<10), 1)
	writeZip(t, path, []zipEntry{{"word/document.xml", []byte(big)}})
	_, err := ParseDocx(path)
	if err == nil || !strings.Contains(err.Error(), "quá lớn") {
		t.Fatalf("muốn lỗi 'quá lớn', got %v", err)
	}
}

// Entry khai báo kích thước nhỏ nhưng bung ra lớn hơn vẫn chỉ đọc tới giới hạn.
func TestReadZipFile_LyingHeader_Limited(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lie.zip")
	writeZip(t, path, []zipEntry{{"x.xml", make([]byte, 10000)}})
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	zr.File[0].UncompressedSize64 = 10 // header nói dối
	data, err := readZipFile(&zr.Reader, "x.xml", 100)
	if err == nil || len(data) > 101 {
		t.Fatalf("muốn lỗi vượt giới hạn, got len=%d err=%v", len(data), err)
	}
}
