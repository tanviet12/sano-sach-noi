package library

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// metadata.json là symlink (ví dụ tới /dev/zero) → bỏ qua cuốn đó, không đọc theo.
func TestGet_MetadataSymlink_Rejected(t *testing.T) {
	base := t.TempDir()
	lib := New(filepath.Join(base, "Sano"))
	makeBook(t, lib, "sach-thu")
	makeBook(t, lib, "sach-la")
	outside := filepath.Join(base, "meta-ngoai.json")
	if err := os.WriteFile(outside, []byte(testMeta), 0o644); err != nil {
		t.Fatal(err)
	}
	meta := filepath.Join(lib.BooksRoot(), "sach-la", "metadata.json")
	if err := os.Remove(meta); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, meta); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.Get("sach-la"); err == nil {
		t.Error("metadata.json là symlink phải bị từ chối")
	}
	if _, err := lib.UpdateInfo("sach-la", Info{Title: "Mới"}); err == nil {
		t.Error("sửa cuốn có metadata.json là symlink phải báo lỗi")
	}
	books, err := lib.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 || books[0].Slug != "sach-thu" {
		t.Errorf("chỉ muốn cuốn sach-thu, got %+v", books)
	}
}

// metadata.json vượt giới hạn → bỏ qua khi liệt kê, báo lỗi khi sửa.
func TestGet_OversizedMetadata_Rejected(t *testing.T) {
	lib := New(t.TempDir())
	makeBook(t, lib, "sach-thu")
	old := maxMetadataBytes
	maxMetadataBytes = int64(len(testMeta) - 1)
	t.Cleanup(func() { maxMetadataBytes = old })
	if _, err := lib.Get("sach-thu"); err == nil {
		t.Error("metadata.json quá lớn phải bị từ chối")
	}
	if _, err := lib.UpdateInfo("sach-thu", Info{Title: "Mới"}); err == nil {
		t.Error("sửa cuốn có metadata.json quá lớn phải báo lỗi")
	}
	books, _ := lib.List()
	if len(books) != 0 {
		t.Errorf("cuốn có metadata.json quá lớn phải bị bỏ qua, got %+v", books)
	}
}

// writeZip ghi gói zip với các mục theo đúng thứ tự (cho phép tên trùng).
func writeZip(t *testing.T, path string, names, bodies []string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for i, name := range names {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(bodies[i])); err != nil {
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

// Nhiều mục chapters.json trùng tên: chỉ đọc mục đầu tiên.
func TestZipDurations_StopsAtFirstChaptersJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "book-x.zip")
	other := `{"chapters":[{"order":1,"sections":[{"order":1,"duration_sec":999}]}]}`
	writeZip(t, p, []string{"chapters.json", "chapters.json", "chapters.json"}, []string{testChapters, other, other})
	got := zipDurations(p)
	if got["1/1"] != 10 || got["1/2"] != 20 || got["2/1"] != 30 {
		t.Errorf("phải chỉ lấy chapters.json đầu tiên, got %v", got)
	}
}

// Gói zip quá nhiều mục: không lặp qua, sửa thông tin báo lỗi.
func TestZip_TooManyEntries_Rejected(t *testing.T) {
	old := maxZipEntries
	maxZipEntries = 3
	t.Cleanup(func() { maxZipEntries = old })
	lib := New(t.TempDir())
	dir, _ := makeFullBook(t, lib, "sach-thu", testMeta) // 6 mục
	if got := zipDurations(filepath.Join(dir, "book-sach-thu.zip")); len(got) != 0 {
		t.Errorf("gói quá nhiều mục phải bị bỏ qua, got %v", got)
	}
	if _, err := lib.UpdateInfo("sach-thu", Info{Title: "Mới"}); err == nil {
		t.Error("sửa gói zip quá nhiều mục phải báo lỗi")
	}
}

// Thư mục mang đuôi gói macOS (X.app...), dấu phẩy hoặc đuôi .{CLSID} không
// được coi là sách: không liệt kê, không mở được.
func TestValidSlug_RejectsBundlesAndShellNames(t *testing.T) {
	bad := []string{
		"evil.app", "Evil.APP", "x.pkg", "x.mpkg", "x.workflow", "x.prefPane", "x.bundle",
		"x.plugin", "x.kext", "x.appex", "x.framework", "x.component", "x.action", "x.saver",
		"x.qlgenerator", "x.mdimporter", "a,b", "a,calc.exe", "x.{645FF040-5081-101B-9F08-00AA002F954E}",
		"", ".an", "a/b", `a\b`,
	}
	for _, s := range bad {
		if validSlug(s) {
			t.Errorf("validSlug(%q) phải từ chối", s)
		}
	}
	for _, s := range []string{"sach-thu", "sach-thu-2", "tuyen-tap.2024", "ban.app-le"} {
		if !validSlug(s) {
			t.Errorf("validSlug(%q) phải nhận", s)
		}
	}
	lib := New(t.TempDir())
	makeBook(t, lib, "sach-thu")
	makeBook(t, lib, "Tro-Choi.app")
	books, err := lib.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 || books[0].Slug != "sach-thu" {
		t.Errorf("thư mục .app không được hiện trong thư viện, got %+v", books)
	}
	if _, err := lib.Dir("Tro-Choi.app"); err == nil {
		t.Error("Dir phải từ chối thư mục .app")
	}
}
