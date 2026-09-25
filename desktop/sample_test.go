package main

import (
	"bytes"
	"os"
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
	// File mẫu dạy cách đặt Heading: đúng 3 chương, không có chữ nằm trước
	// chương 1 (phần đó Sano bỏ tick sẵn như trang bìa).
	want := []string{"Chương 1. Đặt tiêu đề để có mục lục", "Chương 2. Viết để nghe", "Chương 3. Nội dung của bạn"}
	if len(book.Chapters) != len(want) {
		t.Fatalf("muốn %d chương, có %d: %+v", len(want), len(book.Chapters), book.Chapters)
	}
	for i, w := range want {
		if book.Chapters[i].Title != w {
			t.Errorf("chương %d = %q, muốn %q", i+1, book.Chapters[i].Title, w)
		}
	}
	if book.Title != "Sách mẫu Sano" {
		t.Errorf("tên sách = %q", book.Title)
	}
}

func TestSaveSampleDocx_LuuDungFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvSampleOut, dir)
	a := &App{}
	path, err := a.SaveSampleDocx()
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, sampleDocxName) {
		t.Fatalf("lưu vào %q", path)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, sampleDocx) {
		t.Fatal("file lưu ra khác file mẫu nhúng")
	}
}
