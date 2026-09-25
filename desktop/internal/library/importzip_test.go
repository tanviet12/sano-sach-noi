package library

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// fakeMP3 — ID3 + vài byte: đủ qua kiểm chữ ký đầu file (thời lượng đọc ra 0).
func fakeMP3(tag string) []byte {
	return append([]byte("ID3\x04\x00\x00\x00\x00\x00\x00"), []byte(tag)...)
}

var pngBytes = []byte("\x89PNG\r\n\x1a\n0000")

type zipEntry struct {
	name   string
	data   []byte
	store  bool // không nén
	mutate func(*zip.FileHeader)
}

func goodManifest(title string) []byte {
	b, _ := json.Marshal(map[string]any{"title": title, "slug": "bo-qua", "author": "Tác giả A", "voice_id": "Hải Đăng", "cover_filename": "cover.png", "version": 1})
	return b
}

func goodChapters() []byte {
	return []byte(`{"chapters":[
 {"order":1,"title":"Chương 1","sections":[
   {"order":1,"title":"1.1","audio_filename":"audio/ch01/sec01.mp3","duration_sec":10,"original_text":"Câu một. Câu hai.","reading_script":"Một chấm một.\n\nCâu một. Câu hai."},
   {"order":2,"title":"1.2","audio_filename":"audio/ch01/sec02.mp3","duration_sec":20,"original_text":"Câu ba."}]},
 {"order":2,"title":"Chương 2","sections":[
   {"order":1,"title":"2.1","audio_filename":"audio/ch02/sec01.mp3","duration_sec":30,"original_text":"Câu bốn."}]}]}`)
}

func goodEntries(title string) []zipEntry {
	return []zipEntry{
		{name: "manifest.json", data: goodManifest(title)},
		{name: "chapters.json", data: goodChapters()},
		{name: "cover.png", data: pngBytes, store: true},
		{name: "audio/ch01/sec01.mp3", data: fakeMP3("a"), store: true},
		{name: "audio/ch01/sec02.mp3", data: fakeMP3("b"), store: true},
		{name: "audio/ch02/sec01.mp3", data: fakeMP3("c"), store: true},
	}
}

func writeImportZip(t *testing.T, dir string, entries []zipEntry) string {
	t.Helper()
	p := filepath.Join(dir, "goi.zip")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		h := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		if e.store {
			h.Method = zip.Store
		}
		if e.mutate != nil {
			e.mutate(h)
		}
		w, err := zw.CreateHeader(h)
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
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func importAll(t *testing.T, lib *Library, path string, keepBoth bool) (string, error) {
	t.Helper()
	work, slug, err := lib.PrepareImport(context.Background(), path, keepBoth, nil)
	if err != nil {
		return "", err
	}
	return lib.Commit(work, slug)
}

func wantBad(t *testing.T, err error, contains string) {
	t.Helper()
	var b *ErrBadPackage
	if !errors.As(err, &b) {
		t.Fatalf("muốn ErrBadPackage chứa %q, got %v", contains, err)
	}
	if contains != "" && !strings.Contains(b.Reason, contains) {
		t.Fatalf("lý do %q không chứa %q", b.Reason, contains)
	}
}

// Thư viện không được có thư mục tạm sót lại hay file nào ngoài thư mục sách.
func assertNoLeftovers(t *testing.T, lib *Library, allowed ...string) {
	t.Helper()
	entries, _ := os.ReadDir(lib.BooksRoot())
	for _, e := range entries {
		ok := false
		for _, a := range allowed {
			ok = ok || e.Name() == a
		}
		if !ok {
			t.Errorf("còn sót %q trong thư viện", e.Name())
		}
	}
}

func TestImport_RoundTrip(t *testing.T) {
	root := t.TempDir()
	lib := New(root)
	p := writeImportZip(t, t.TempDir(), append(goodEntries("Khởi nghiệp từ số 0"), zipEntry{name: "rac/khong-dung.txt", data: []byte("x")}))

	pv, err := lib.PreviewImport(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(pv.CoverDataURL, "data:image/png;base64,") {
		t.Errorf("thiếu ảnh bìa xem trước: %.40q", pv.CoverDataURL)
	}
	if pv.Title != "Khởi nghiệp từ số 0" || pv.Author != "Tác giả A" || pv.Voice != "Hải Đăng" || pv.Chapters != 2 || pv.Sections != 3 || pv.DurationSec != 60 || !pv.HasCover || pv.ExistingSlug != "" {
		t.Fatalf("xem trước sai: %+v", pv)
	}
	slug, err := importAll(t, lib, p, false)
	if err != nil {
		t.Fatal(err)
	}
	d, err := lib.Get(slug)
	if err != nil {
		t.Fatal(err)
	}
	if d.Title != "Khởi nghiệp từ số 0" || d.Voice != "Hải Đăng" || d.Sections != 3 || d.Cover == "" || d.Zip == "" {
		t.Fatalf("sách nhập sai: %+v", d.Book)
	}
	texts, _ := lib.Texts(slug)
	if texts[0].Text != "Câu một. Câu hai." || !strings.HasPrefix(texts[0].Script, "Một chấm một.") {
		t.Errorf("chữ tiểu mục sai: %+v", texts[0])
	}
	// Gói zip trong thư viện là gói ĐÓNG GÓI LẠI: không mang theo file rác.
	names, _ := readZip(t, filepath.Join(root, filepath.FromSlash(d.Zip)))
	sort.Strings(names)
	for _, n := range names {
		if strings.HasPrefix(n, "rac/") {
			t.Errorf("gói lưu lại vẫn chứa file rác %q", n)
		}
	}
	assertNoLeftovers(t, lib, slug)

	// Nhập lần hai: xem trước báo trùng; giữ cả hai → "(2)", slug-2.
	pv2, err := lib.PreviewImport(p)
	if err != nil || pv2.ExistingSlug != slug {
		t.Fatalf("phải báo trùng %q, got %+v err=%v", slug, pv2, err)
	}
	slug2, err := importAll(t, lib, p, true)
	if err != nil {
		t.Fatal(err)
	}
	d2, _ := lib.Get(slug2)
	if slug2 != slug+"-2" || d2 == nil || d2.Title != "Khởi nghiệp từ số 0 (2)" {
		t.Fatalf("giữ cả hai sai: slug=%q book=%+v", slug2, d2)
	}
}

func TestImport_RejectsPathTricks(t *testing.T) {
	cases := map[string]string{
		"../":        "../../etc/evil.mp3",
		"tuyệt đối":  "/tmp/evil.mp3",
		"gạch ngược": `audio\ch01\sec01.mp3`,
		"đuôi khác":  "audio/ch01/sec01.exe",
		"thư mục lạ": "audio/../ch01/sec01.mp3",
		"ký tự rỗng": "audio/ch01/sec01.mp3\x00",
	}
	for name, audio := range cases {
		t.Run(name, func(t *testing.T) {
			lib := New(t.TempDir())
			ch := strings.Replace(string(goodChapters()), "audio/ch01/sec01.mp3", strings.ReplaceAll(audio, `\`, `\\`), 1)
			ch = strings.Replace(ch, "\x00", `\u0000`, 1)
			es := goodEntries("Sách")
			es[1].data = []byte(ch)
			es = append(es, zipEntry{name: audio, data: fakeMP3("x"), store: true})
			_, err := importAll(t, lib, writeImportZip(t, t.TempDir(), es), false)
			wantBad(t, err, "Tên file âm thanh lạ")
			assertNoLeftovers(t, lib)
		})
	}
}

func TestImport_EvilEntryNamesNeverWritten(t *testing.T) {
	outside := t.TempDir()
	lib := New(filepath.Join(outside, "Sano"))
	es := append(goodEntries("Sách"),
		zipEntry{name: "../../thoat-ra.txt", data: []byte("x")},
		zipEntry{name: "/tuyet-doi.txt", data: []byte("x")},
	)
	slug, err := importAll(t, lib, writeImportZip(t, t.TempDir(), es), false)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{filepath.Join(outside, "thoat-ra.txt"), filepath.Join(filepath.Dir(outside), "thoat-ra.txt"), "/tuyet-doi.txt"} {
		if _, err := os.Stat(p); err == nil {
			t.Fatalf("file ngoài thư viện bị tạo: %s", p)
		}
	}
	files, _ := os.ReadDir(filepath.Join(lib.BooksRoot(), slug))
	for _, f := range files {
		if strings.Contains(f.Name(), "thoat") || strings.Contains(f.Name(), "tuyet") {
			t.Errorf("file lạ trong thư mục sách: %s", f.Name())
		}
	}
}

func TestImport_RejectsMaliciousPackages(t *testing.T) {
	type tc struct {
		name, reason string
		edit         func([]zipEntry) []zipEntry
		setup        func() func()
	}
	cases := []tc{
		{name: "trùng tên", reason: "trùng tên", edit: func(es []zipEntry) []zipEntry {
			return append(es, zipEntry{name: "audio/ch01/sec01.mp3", data: fakeMP3("khac"), store: true})
		}},
		{name: "trùng manifest", reason: "trùng tên", edit: func(es []zipEntry) []zipEntry {
			return append(es, zipEntry{name: "manifest.json", data: goodManifest("Tên khác")})
		}},
		{name: "mã hoá", reason: "mã hoá", edit: func(es []zipEntry) []zipEntry {
			es[3].mutate = func(h *zip.FileHeader) { h.Flags |= 0x1 }
			return es
		}},
		{name: "symlink", reason: "không phải file thường", edit: func(es []zipEntry) []zipEntry {
			es[3].mutate = func(h *zip.FileHeader) { h.SetMode(os.ModeSymlink | 0o777) }
			return es
		}},
		{name: "mp3 giả", reason: "không đúng định dạng", edit: func(es []zipEntry) []zipEntry {
			es[3].data = []byte("<script>alert(1)</script>")
			return es
		}},
		{name: "zip bomb", reason: "nén bất thường", edit: func(es []zipEntry) []zipEntry {
			es[3].data = append(fakeMP3(""), make([]byte, 8<<20)...)
			es[3].store = false
			return es
		}},
		{name: "mp3 quá lớn", reason: "quá lớn", edit: func(es []zipEntry) []zipEntry { return es }, setup: func() func() {
			old := maxImportAudioBytes
			maxImportAudioBytes = 8
			return func() { maxImportAudioBytes = old }
		}},
		{name: "tổng quá lớn", reason: "quá lớn", edit: func(es []zipEntry) []zipEntry { return es }, setup: func() func() {
			old := maxImportTotalBytes
			maxImportTotalBytes = 20
			return func() { maxImportTotalBytes = old }
		}},
		{name: "thiếu chapters", reason: "thiếu manifest.json hoặc chapters.json", edit: func(es []zipEntry) []zipEntry { return append(es[:1], es[2:]...) }},
		{name: "thiếu mp3", reason: "thiếu file âm thanh", edit: func(es []zipEntry) []zipEntry { return es[:5] }},
		{name: "phiên bản lạ", reason: "phiên bản định dạng", edit: func(es []zipEntry) []zipEntry {
			es[0].data = []byte(`{"title":"A","version":2}`)
			return es
		}},
		{name: "không tên", reason: "không có tên", edit: func(es []zipEntry) []zipEntry {
			es[0].data = []byte(`{"title":"  \u0007 ","version":1}`)
			return es
		}},
		{name: "hai mục chung mp3", reason: "dùng chung", edit: func(es []zipEntry) []zipEntry {
			es[1].data = []byte(strings.Replace(string(goodChapters()), "audio/ch01/sec02.mp3", "audio/ch01/sec01.mp3", 1))
			return es
		}},
		{name: "json hỏng", reason: "không đọc được", edit: func(es []zipEntry) []zipEntry {
			es[1].data = []byte(`{"chapters":[`)
			return es
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.setup != nil {
				defer c.setup()()
			}
			lib := New(t.TempDir())
			p := writeImportZip(t, t.TempDir(), c.edit(goodEntries("Sách")))
			_, perr := lib.PreviewImport(p)
			_, err := importAll(t, lib, p, false)
			wantBad(t, err, c.reason)
			if !strings.Contains(c.name, "mp3 giả") { // chữ ký mp3 chỉ kiểm lúc giải nén
				wantBad(t, perr, c.reason)
			}
			assertNoLeftovers(t, lib)
		})
	}
}

func TestImport_CleansStrings(t *testing.T) {
	lib := New(t.TempDir())
	es := goodEntries("Sách‮ hay\x07\n  quá")
	es[1].data = []byte(strings.Replace(string(goodChapters()), `"Câu một. Câu hai."`, `"Câu​ một.\u0000 Dòng\nhai."`, 1))
	slug, err := importAll(t, lib, writeImportZip(t, t.TempDir(), es), false)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := lib.Get(slug)
	if d.Title != "Sách hay quá" {
		t.Errorf("tên chưa làm sạch: %q", d.Title)
	}
	texts, _ := lib.Texts(slug)
	if texts[0].Text != "Câu một. Dòng\nhai." {
		t.Errorf("chữ chưa làm sạch: %q", texts[0].Text)
	}
}

func TestImport_BadCoverIgnored(t *testing.T) {
	lib := New(t.TempDir())
	es := goodEntries("Sách")
	es[0].data = []byte(`{"title":"Sách","version":1,"cover_filename":"cover.svg"}`)
	es[2] = zipEntry{name: "cover.svg", data: []byte(`<svg onload="alert(1)"/>`)}
	slug, err := importAll(t, lib, writeImportZip(t, t.TempDir(), es), false)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := lib.Get(slug)
	if d.Cover != "" {
		t.Errorf("bìa SVG không được nhận, got %q", d.Cover)
	}
	// Bìa .png nhưng nội dung không phải ảnh → không nhập cả gói.
	es2 := goodEntries("Sách khác")
	es2[2].data = []byte("<html>")
	_, err = importAll(t, lib, writeImportZip(t, t.TempDir(), es2), false)
	wantBad(t, err, "không đúng định dạng")
}

func TestImport_CancelCleansUp(t *testing.T) {
	lib := New(t.TempDir())
	p := writeImportZip(t, t.TempDir(), goodEntries("Sách"))
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	_, _, err := lib.PrepareImport(ctx, p, false, func(done, total int) {
		calls++
		cancel()
	})
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("muốn huỷ sau tiểu mục đầu, got err=%v calls=%d", err, calls)
	}
	assertNoLeftovers(t, lib)
}

func TestImport_RejectsNonZipAndSymlink(t *testing.T) {
	lib := New(t.TempDir())
	dir := t.TempDir()
	docx := filepath.Join(dir, "a.docx")
	_ = os.WriteFile(docx, []byte("x"), 0o644)
	_, err := lib.PreviewImport(docx)
	wantBad(t, err, ".zip")

	real := writeImportZip(t, dir, goodEntries("Sách"))
	link := filepath.Join(dir, "lien-ket.zip")
	if err := os.Symlink(real, link); err != nil {
		t.Skip("không tạo được symlink:", err)
	}
	_, err = lib.PreviewImport(link)
	wantBad(t, err, "")

	notZip := filepath.Join(dir, "gia.zip")
	_ = os.WriteFile(notZip, []byte("khong phai zip"), 0o644)
	_, err = lib.PreviewImport(notZip)
	wantBad(t, err, "không phải gói zip")
}

func TestSlugFor_WindowsReservedNames(t *testing.T) {
	for title, want := range map[string]string{"Con": "con-sach", "NUL": "nul-sach", "Com1": "com1-sach", "Con mèo": "con-meo", "": "audiobook"} {
		if got := BookSlug(title); got != want {
			t.Errorf("BookSlug(%q) = %q, muốn %q", title, got, want)
		}
	}
}
