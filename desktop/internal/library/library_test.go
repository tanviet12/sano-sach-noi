package library

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

const testMeta = `{"title":"Sách thử","author":"Tác giả A","cover":"cover.png","chapters":[
 {"title":"Chương 1","sections":[{"title":"Mở đầu","file":"ch01-sec01.mp3"},{"title":"Phần hai","file":"ch01-sec02.mp3"}]},
 {"title":"Chương 2","sections":[{"title":"Kết","file":"ch02-sec01.mp3"}]}]}`

const testChapters = `{"chapters":[
 {"order":1,"sections":[{"order":1,"duration_sec":10},{"order":2,"duration_sec":20}]},
 {"order":2,"sections":[{"order":1,"duration_sec":30}]}]}`

func makeBook(t *testing.T, lib *Library, slug string) {
	t.Helper()
	dir := filepath.Join(lib.BooksRoot(), slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"metadata.json": testMeta, "cover.png": "png"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	f, err := os.Create(filepath.Join(dir, "book-"+slug+".zip"))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("chapters.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(testChapters)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestList_EmptyLibrary_ReturnsEmpty(t *testing.T) {
	books, err := New(t.TempDir()).List()
	if err != nil || books == nil || len(books) != 0 {
		t.Fatalf("muốn danh sách rỗng, got %v %v", books, err)
	}
}

func TestGet_ReadsMetadataAndDurations(t *testing.T) {
	lib := New(t.TempDir())
	makeBook(t, lib, "sach-thu")
	d, err := lib.Get("sach-thu")
	if err != nil {
		t.Fatal(err)
	}
	if d.Title != "Sách thử" || d.Chapters != 2 || d.Sections != 3 || d.DurationSec != 60 {
		t.Errorf("thông tin sách sai: %+v", d.Book)
	}
	if d.Cover != "Sach/sach-thu/cover.png" || d.Zip != "Sach/sach-thu/book-sach-thu.zip" {
		t.Errorf("đường dẫn bìa/zip sai: %q %q", d.Cover, d.Zip)
	}
	if d.Tracks[1].Title != "Phần hai" || d.Tracks[1].DurationSec != 20 || d.Tracks[2].Chapter != "Chương 2" {
		t.Errorf("tiểu mục sai: %+v", d.Tracks)
	}
}

func TestList_SkipsWorkDirsAndBrokenBooks(t *testing.T) {
	lib := New(t.TempDir())
	makeBook(t, lib, "sach-thu")
	if _, err := lib.NewWorkDir("dang-lam"); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(lib.BooksRoot(), "thu-muc-la"), 0o755); err != nil {
		t.Fatal(err)
	}
	books, err := lib.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 || books[0].Slug != "sach-thu" {
		t.Errorf("chỉ muốn 1 cuốn sach-thu, got %+v", books)
	}
}

func TestCommit_NeverOverwritesExistingBook(t *testing.T) {
	lib := New(t.TempDir())
	makeBook(t, lib, "sach-thu")
	work, err := lib.NewWorkDir("sach-thu")
	if err != nil {
		t.Fatal(err)
	}
	slug, err := lib.Commit(work, "sach-thu")
	if err != nil {
		t.Fatal(err)
	}
	if slug != "sach-thu-2" {
		t.Errorf("trùng tên phải thành sach-thu-2, got %q", slug)
	}
	if _, err := lib.Get("sach-thu"); err != nil {
		t.Errorf("cuốn cũ phải còn nguyên: %v", err)
	}
}

func TestResolve_BlocksTraversal(t *testing.T) {
	lib := New(t.TempDir())
	for _, bad := range []string{"../x", "Sach/../../x", "", "."} {
		if _, err := lib.Resolve(bad); err == nil {
			t.Errorf("Resolve(%q) phải lỗi", bad)
		}
	}
	got, err := lib.Resolve("Sach/a/ch01-sec01.mp3")
	if err != nil || got != filepath.Join(lib.Root(), "Sach", "a", "ch01-sec01.mp3") {
		t.Errorf("Resolve hợp lệ sai: %q %v", got, err)
	}
}

func TestGet_InvalidSlug_NotFound(t *testing.T) {
	lib := New(t.TempDir())
	for _, bad := range []string{"", "..", ".dang-lam-x", "a/b"} {
		if _, err := lib.Get(bad); err == nil {
			t.Errorf("Get(%q) phải lỗi", bad)
		}
	}
}

func TestSize_CongFileThuongKhongTheoSymlink(t *testing.T) {
	lib := New(t.TempDir())
	if n, err := lib.Size(); err != nil || n != 0 {
		t.Fatalf("thư viện chưa có thư mục: %d %v", n, err)
	}
	dir := filepath.Join(lib.BooksRoot(), "sach-a")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "a.mp3"), make([]byte, 1000), 0o644)
	os.WriteFile(filepath.Join(dir, "metadata.json"), make([]byte, 24), 0o644)
	outside := filepath.Join(t.TempDir(), "lon.bin")
	os.WriteFile(outside, make([]byte, 1<<20), 0o644)
	if err := os.Symlink(outside, filepath.Join(dir, "link.bin")); err != nil {
		t.Fatal(err)
	}
	if n, err := lib.Size(); err != nil || n != 1024 {
		t.Errorf("Size = %d %v, muốn 1024", n, err)
	}
}
