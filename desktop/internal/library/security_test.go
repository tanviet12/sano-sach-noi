package library

import (
	"os"
	"path/filepath"
	"testing"
)

// setMaxZipJSON đặt tạm giới hạn đọc JSON trong gói zip cho test.
func setMaxZipJSON(t *testing.T, n int64) {
	t.Helper()
	old := maxZipJSONBytes
	maxZipJSONBytes = n
	t.Cleanup(func() { maxZipJSONBytes = old })
}

// Symlink trong thư mục sách trỏ ra ngoài ~/Sano không được phục vụ.
func TestResolve_SymlinkOutsideRoot_Rejected(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "Sano")
	lib := New(root)
	makeBook(t, lib, "sach-thu")
	secret := filepath.Join(base, "bi-mat.mp3")
	if err := os.WriteFile(secret, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(base, "ngoai")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "a.mp3"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	bookDir := filepath.Join(lib.BooksRoot(), "sach-thu")
	if err := os.Symlink(secret, filepath.Join(bookDir, "ch09-sec01.mp3")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outDir, filepath.Join(bookDir, "thu-muc")); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"Sach/sach-thu/ch09-sec01.mp3", "Sach/sach-thu/thu-muc/a.mp3"} {
		if got, err := lib.Resolve(rel); err == nil {
			t.Errorf("Resolve(%q) = %q, phải từ chối symlink ra ngoài", rel, got)
		}
	}
	// File thật trong thư viện vẫn phục vụ được (kể cả khi root nằm sau symlink như /var trên macOS).
	if _, err := lib.Resolve("Sach/sach-thu/cover.png"); err != nil {
		t.Errorf("file hợp lệ bị từ chối: %v", err)
	}
	// Symlink trỏ tới file khác trong thư viện: cho phép.
	if err := os.Symlink(filepath.Join(bookDir, "cover.png"), filepath.Join(bookDir, "bia.png")); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.Resolve("Sach/sach-thu/bia.png"); err != nil {
		t.Errorf("symlink trong thư viện bị từ chối: %v", err)
	}
}

// chapters.json quá lớn (gói zip "bom nén") bị bỏ qua, thư viện vẫn mở được.
func TestGet_OversizedChaptersJSON_Skipped(t *testing.T) {
	setMaxZipJSON(t, 16)
	lib := New(t.TempDir())
	makeBook(t, lib, "sach-thu")
	d, err := lib.Get("sach-thu")
	if err != nil {
		t.Fatal(err)
	}
	if d.DurationSec != 0 {
		t.Errorf("chapters.json vượt giới hạn phải bị bỏ qua, DurationSec = %d", d.DurationSec)
	}
}

// manifest.json quá lớn → sửa thông tin báo lỗi, không đọc hết vào RAM, metadata giữ nguyên.
func TestUpdateInfo_OversizedManifest_Errors(t *testing.T) {
	setMaxZipJSON(t, 16)
	lib := New(t.TempDir())
	dir, _ := makeFullBook(t, lib, "sach-thu", testMeta)
	before, _ := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if _, err := lib.UpdateInfo("sach-thu", Info{Title: "Mới"}); err == nil {
		t.Fatal("manifest.json vượt giới hạn phải báo lỗi")
	}
	after, _ := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if string(before) != string(after) {
		t.Error("metadata.json không được đổi khi lỗi")
	}
}
