package main

import (
	"path/filepath"
	"testing"

	"sano/internal/bookmaker"
)

func TestWriteSampleDocx_DocDuocMucLuc(t *testing.T) {
	p := filepath.Join(t.TempDir(), sampleDocxName)
	f, err := writeSampleDocx(p)
	if err != nil {
		t.Fatal(err)
	}
	if f.Name != sampleDocxName || f.Size == 0 {
		t.Fatalf("thông tin file sai: %+v", f)
	}
	book, err := bookmaker.ParseDocx(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(book.Chapters) < 2 {
		t.Fatalf("tài liệu mẫu phải có vài chương, có %d", len(book.Chapters))
	}
}
