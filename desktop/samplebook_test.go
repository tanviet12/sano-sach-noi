package main

import (
	"os"
	"path/filepath"
	"testing"

	"sano/desktop/internal/library"
)

func TestSeedSampleBook_ChepMotLan(t *testing.T) {
	lib := library.New(t.TempDir())
	if err := seedSampleBook(lib); err != nil {
		t.Fatal(err)
	}
	books, err := lib.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 || books[0].Slug != sampleBookSlug {
		t.Fatalf("thư viện phải có đúng cuốn mẫu, có %+v", books)
	}
	b := books[0]
	if b.Zip == "" || b.Cover == "" || b.Sections != 5 || b.DurationSec < 200 {
		t.Fatalf("cuốn mẫu thiếu zip/bìa/thời lượng: %+v", b)
	}

	// Xoá rồi mở lại app: không chép lại.
	if err := os.RemoveAll(filepath.Join(lib.BooksRoot(), sampleBookSlug)); err != nil {
		t.Fatal(err)
	}
	if err := seedSampleBook(lib); err != nil {
		t.Fatal(err)
	}
	if books, _ := lib.List(); len(books) != 0 {
		t.Fatalf("đã xoá thì không chép lại, còn %d cuốn", len(books))
	}
}

func TestSeedSampleBook_KhongDeCuonDaCo(t *testing.T) {
	lib := library.New(t.TempDir())
	own := filepath.Join(lib.BooksRoot(), sampleBookSlug)
	if err := os.MkdirAll(own, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(own, "metadata.json"), []byte(`{"title":"Của tôi","chapters":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := seedSampleBook(lib); err != nil {
		t.Fatal(err)
	}
	d, err := lib.Get(sampleBookSlug)
	if err != nil || d.Title != "Của tôi" {
		t.Fatalf("cuốn của người dùng bị đè: %+v %v", d, err)
	}
	if _, err := lib.Get(sampleBookSlug + "-2"); err != nil {
		t.Fatalf("cuốn mẫu phải lưu tên khác: %v", err)
	}
}
