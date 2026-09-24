package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDescribeDocx(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "Sách mẫu.DOCX")
	if err := os.WriteFile(doc, []byte("12345"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := describeDocx(doc)
	if err != nil {
		t.Fatalf("lỗi không mong muốn: %v", err)
	}
	if got.Name != "Sách mẫu.DOCX" || got.Size != 5 || got.Path != doc {
		t.Errorf("sai thông tin file: %+v", got)
	}

	if _, err := describeDocx(filepath.Join(dir, "a.pdf")); !errors.Is(err, ErrNotDocx) {
		t.Errorf("file .pdf phải trả ErrNotDocx, được %v", err)
	}
	if _, err := describeDocx(filepath.Join(dir, "khong-co.docx")); err == nil {
		t.Error("file không tồn tại phải báo lỗi")
	}
	folder := filepath.Join(dir, "thu-muc.docx")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := describeDocx(folder); !errors.Is(err, ErrNotDocx) {
		t.Errorf("thư mục phải trả ErrNotDocx, được %v", err)
	}
}
