package library

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testManifest = `{"title":"Sách thử","slug":"sach-thu","author":"Tác giả A","voice_id":"Trúc Ly","version":1,"build_timestamp":"2026-09-01T00:00:00Z"}`

// makeFullBook dựng một cuốn có gói zip đầy đủ (manifest, chapters, bìa, mp3).
func makeFullBook(t *testing.T, lib *Library, slug, meta string) (dir string, entries map[string]string) {
	t.Helper()
	dir = filepath.Join(lib.BooksRoot(), slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	entries = map[string]string{
		"manifest.json":        testManifest,
		"chapters.json":        testChapters,
		"cover.png":            "png-bytes",
		"audio/ch01/sec01.mp3": "mp3-1",
		"audio/ch01/sec02.mp3": "mp3-2",
		"audio/ch02/sec01.mp3": "mp3-3",
	}
	f, err := os.Create(filepath.Join(dir, "book-"+slug+".zip"))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for _, name := range []string{"manifest.json", "chapters.json", "cover.png", "audio/ch01/sec01.mp3", "audio/ch01/sec02.mp3", "audio/ch02/sec01.mp3"} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(entries[name])); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return dir, entries
}

func readZip(t *testing.T, path string) (names []string, contents map[string]string) {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	contents = map[string]string{}
	for _, f := range zr.File {
		b, err := readZipJSON(f)
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, f.Name)
		contents[f.Name] = string(b)
	}
	return names, contents
}

func readJSON(t *testing.T, data string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		t.Fatalf("JSON hỏng: %v\n%s", err, data)
	}
	return m
}

func TestGet_OldBookWithoutCategory_Uncategorized(t *testing.T) {
	lib := New(t.TempDir())
	makeBook(t, lib, "sach-cu")
	d, err := lib.Get("sach-cu")
	if err != nil {
		t.Fatal(err)
	}
	if d.Category != "" {
		t.Errorf("sách cũ không có category phải là chưa phân loại, got %q", d.Category)
	}
}

func TestGet_ReadsCategory(t *testing.T) {
	lib := New(t.TempDir())
	meta := `{"title":"A","category":"  Kỹ   năng ","chapters":[]}`
	makeFullBook(t, lib, "a", meta)
	d, err := lib.Get("a")
	if err != nil {
		t.Fatal(err)
	}
	if d.Category != "Kỹ năng" {
		t.Errorf("category = %q, muốn chuẩn hoá thành %q", d.Category, "Kỹ năng")
	}
}

func TestUpdateInfo_UpdatesMetadataAndZipManifest(t *testing.T) {
	lib := New(t.TempDir())
	meta := `{"title":"Sách thử","author":"Tác giả A","narrator":"Trúc Ly","cover":"cover.png","chapters":[{"title":"Chương 1","sections":[{"title":"Mở đầu","file":"ch01-sec01.mp3"}]}]}`
	dir, entries := makeFullBook(t, lib, "sach-thu", meta)
	old := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	if err := os.Chtimes(dir, old, old); err != nil {
		t.Fatal(err)
	}

	d, err := lib.UpdateInfo("sach-thu", Info{Title: "  Tên mới  ", Author: "Tác giả B", Category: "Kỹ năng"})
	if err != nil {
		t.Fatal(err)
	}
	if d.Title != "Tên mới" || d.Author != "Tác giả B" || d.Category != "Kỹ năng" || d.Slug != "sach-thu" {
		t.Errorf("thông tin sau khi sửa sai: %+v", d.Book)
	}

	// metadata.json: đổi 3 trường, giữ các trường khác.
	raw, err := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := readJSON(t, string(raw))
	if m["title"] != "Tên mới" || m["author"] != "Tác giả B" || m["category"] != "Kỹ năng" {
		t.Errorf("metadata.json chưa đổi: %v", m)
	}
	if m["narrator"] != "Trúc Ly" || m["cover"] != "cover.png" || len(m["chapters"].([]any)) != 1 {
		t.Errorf("metadata.json mất trường khác: %v", m)
	}

	// Zip: manifest đổi, slug/phiên bản giữ; mọi file khác giữ nguyên, đúng thứ tự.
	names, got := readZip(t, filepath.Join(dir, "book-sach-thu.zip"))
	want := []string{"manifest.json", "chapters.json", "cover.png", "audio/ch01/sec01.mp3", "audio/ch01/sec02.mp3", "audio/ch02/sec01.mp3"}
	if len(names) != len(want) {
		t.Fatalf("zip có %v, muốn %v", names, want)
	}
	for i, n := range want {
		if names[i] != n {
			t.Errorf("thứ tự entry %d = %q, muốn %q", i, names[i], n)
		}
		if n != "manifest.json" && got[n] != entries[n] {
			t.Errorf("entry %q bị đổi: %q", n, got[n])
		}
	}
	man := readJSON(t, got["manifest.json"])
	if man["title"] != "Tên mới" || man["author"] != "Tác giả B" || man["category"] != "Kỹ năng" || man["category_slug"] != "ky-nang" {
		t.Errorf("manifest chưa đổi: %v", man)
	}
	if man["slug"] != "sach-thu" || man["version"] != float64(1) || man["voice_id"] != "Trúc Ly" || man["build_timestamp"] != "2026-09-01T00:00:00Z" {
		t.Errorf("manifest mất trường khác: %v", man)
	}

	// Không để lại file tạm, giữ giờ thư mục (thứ tự Mới tạo nhất).
	left, _ := filepath.Glob(filepath.Join(dir, ".sua-*"))
	if len(left) != 0 {
		t.Errorf("còn file tạm: %v", left)
	}
	if info, err := os.Stat(dir); err != nil || !info.ModTime().Equal(old) {
		t.Errorf("giờ thư mục phải giữ %v, got %v", old, info.ModTime())
	}
}

func TestUpdateInfo_ClearCategoryAndAuthor_RemovesKeys(t *testing.T) {
	lib := New(t.TempDir())
	dir, _ := makeFullBook(t, lib, "b", `{"title":"B","author":"X","category":"Kỹ năng","chapters":[]}`)
	if _, err := lib.UpdateInfo("b", Info{Title: "B", Author: " ", Category: "  "}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "metadata.json"))
	m := readJSON(t, string(raw))
	if _, ok := m["category"]; ok {
		t.Errorf("bỏ danh mục phải xoá khoá category: %v", m)
	}
	if _, ok := m["author"]; ok {
		t.Errorf("bỏ tác giả phải xoá khoá author: %v", m)
	}
	_, got := readZip(t, filepath.Join(dir, "book-b.zip"))
	man := readJSON(t, got["manifest.json"])
	if _, ok := man["category_slug"]; ok {
		t.Errorf("manifest phải xoá category_slug: %v", man)
	}
}

func TestUpdateInfo_EmptyTitle_Error(t *testing.T) {
	lib := New(t.TempDir())
	makeFullBook(t, lib, "c", `{"title":"C","chapters":[]}`)
	if _, err := lib.UpdateInfo("c", Info{Title: "   "}); !errors.Is(err, ErrEmptyTitle) {
		t.Errorf("tên trống phải lỗi ErrEmptyTitle, got %v", err)
	}
	if _, err := lib.UpdateInfo("khong-co", Info{Title: "X"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("sách không có phải lỗi ErrNotFound, got %v", err)
	}
}

func TestUpdateInfo_ZipWithoutManifest_OnlyMetadata(t *testing.T) {
	lib := New(t.TempDir())
	makeBook(t, lib, "sach-thu") // zip chỉ có chapters.json
	zipPath := filepath.Join(lib.BooksRoot(), "sach-thu", "book-sach-thu.zip")
	before, _ := os.ReadFile(zipPath)
	d, err := lib.UpdateInfo("sach-thu", Info{Title: "Mới", Category: "Sức khoẻ"})
	if err != nil {
		t.Fatal(err)
	}
	if d.Title != "Mới" || d.Category != "Sức khoẻ" {
		t.Errorf("metadata chưa đổi: %+v", d.Book)
	}
	after, _ := os.ReadFile(zipPath)
	if string(before) != string(after) {
		t.Error("zip không có manifest phải giữ nguyên")
	}
}

func TestUpdateInfo_MergesDuplicateCategory(t *testing.T) {
	lib := New(t.TempDir())
	makeFullBook(t, lib, "a", `{"title":"A","category":"Kỹ năng","chapters":[]}`)
	makeFullBook(t, lib, "b", `{"title":"B","chapters":[]}`)
	d, err := lib.UpdateInfo("b", Info{Title: "B", Category: "  kỹ   NĂNG "})
	if err != nil {
		t.Fatal(err)
	}
	if d.Category != "Kỹ năng" {
		t.Errorf("danh mục trùng phải gộp về cách viết đã có %q, got %q", "Kỹ năng", d.Category)
	}
	// Sửa cách viết của cuốn duy nhất trong danh mục: không bị gộp về chính nó.
	d, err = lib.UpdateInfo("a", Info{Title: "A", Category: "KỸ NĂNG"})
	if err != nil {
		t.Fatal(err)
	}
	if d.Category != "Kỹ năng" {
		t.Errorf("cuốn b vẫn dùng %q nên phải gộp, got %q", "Kỹ năng", d.Category)
	}
}

func TestMergeCategory(t *testing.T) {
	existing := []string{"Kỹ năng", "Tài chính"}
	tests := map[string]string{
		"":             "",
		"   ":          "",
		"tài CHÍNH":    "Tài chính",
		" Sức   khoẻ ": "Sức khoẻ",
		"Kỹ năng mềm":  "Kỹ năng mềm",
	}
	for in, want := range tests {
		if got := MergeCategory(in, existing); got != want {
			t.Errorf("MergeCategory(%q) = %q, muốn %q", in, got, want)
		}
	}
	long := "Một danh mục có tên rất rất dài vượt quá giới hạn cho phép của Sano"
	if got := NormalizeCategory(long); len([]rune(got)) > MaxCategoryLen {
		t.Errorf("danh mục dài phải cắt còn %d ký tự, got %d", MaxCategoryLen, len([]rune(got)))
	}
}

func TestDefault_EnvHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvHome, dir)
	lib, err := Default()
	if err != nil || lib.Root() != dir {
		t.Errorf("SANO_HOME phải đổi thư mục gốc: %v %v", lib, err)
	}
}
