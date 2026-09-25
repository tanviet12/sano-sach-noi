package main

import (
	"os"
	"path/filepath"
	"testing"

	"sano/desktop/internal/library"
)

func TestSeedSampleBooks_ChepMotLan(t *testing.T) {
	lib := library.New(t.TempDir())
	if err := seedSampleBooks(lib); err != nil {
		t.Fatal(err)
	}
	books, err := lib.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != len(sampleBooks) {
		t.Fatalf("thư viện phải có đủ %d cuốn mẫu, có %+v", len(sampleBooks), books)
	}
	for i, b := range books {
		if b.Slug != sampleBooks[i].slug {
			t.Fatalf("cuốn thứ %d phải là %s (đúng thứ tự), gặp %s", i, sampleBooks[i].slug, b.Slug)
		}
		if b.Zip == "" || b.Cover == "" || b.Sections == 0 || b.DurationSec < 60 {
			t.Fatalf("cuốn mẫu thiếu zip/bìa/thời lượng: %+v", b)
		}
	}

	// Xoá một cuốn rồi mở lại app: không chép lại cuốn đó.
	gone := sampleBooks[1].slug
	if err := os.RemoveAll(filepath.Join(lib.BooksRoot(), gone)); err != nil {
		t.Fatal(err)
	}
	if err := seedSampleBooks(lib); err != nil {
		t.Fatal(err)
	}
	if books, _ := lib.List(); len(books) != len(sampleBooks)-1 {
		t.Fatalf("đã xoá thì không chép lại, còn %d cuốn", len(books))
	}
	if _, err := lib.Get(gone); err == nil {
		t.Fatalf("%s đã xoá mà bị chép lại", gone)
	}
}

func TestSeedSampleBooks_ChiChepCuonMoi(t *testing.T) {
	lib := library.New(t.TempDir())
	// Máy đã chép cuốn đầu từ bản trước (người dùng đã xoá nó).
	if err := os.MkdirAll(lib.Root(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lib.Root(), sampleBookMark), []byte(sampleBooks[0].slug+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := seedSampleBooks(lib); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.Get(sampleBooks[0].slug); err == nil {
		t.Fatal("cuốn đã ghi nhận chép rồi không được chép lại")
	}
	if books, _ := lib.List(); len(books) != len(sampleBooks)-1 {
		t.Fatalf("phải chép các cuốn còn lại, có %d cuốn", len(books))
	}
}

func TestSeedSampleBooks_KhongDeCuonDaCo(t *testing.T) {
	lib := library.New(t.TempDir())
	slug := sampleBooks[0].slug
	own := filepath.Join(lib.BooksRoot(), slug)
	if err := os.MkdirAll(own, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(own, "metadata.json"), []byte(`{"title":"Của tôi","chapters":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := seedSampleBooks(lib); err != nil {
		t.Fatal(err)
	}
	d, err := lib.Get(slug)
	if err != nil || d.Title != "Của tôi" {
		t.Fatalf("cuốn của người dùng bị đè: %+v %v", d, err)
	}
	if _, err := lib.Get(slug + "-2"); err != nil {
		t.Fatalf("cuốn mẫu phải lưu tên khác: %v", err)
	}
}
