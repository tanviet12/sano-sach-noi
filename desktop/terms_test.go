package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTerms_GhiRoiDocLai(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", termsFileName)
	if st := readTerms(p); st.AcceptedVersion != 0 {
		t.Fatalf("chưa đồng ý phải là 0, có %d", st.AcceptedVersion)
	}
	now := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	if _, err := writeTerms(p, 1, now); err != nil {
		t.Fatal(err)
	}
	st := readTerms(p)
	if st.AcceptedVersion != 1 || st.AcceptedAt != "2026-09-24T10:00:00Z" {
		t.Fatalf("đọc lại sai: %+v", st)
	}
	if _, err := writeTerms(p, 0, now); err == nil {
		t.Fatal("phiên bản 0 phải báo lỗi")
	}
}

func TestTerms_FileHongCoiNhuChuaDongY(t *testing.T) {
	p := filepath.Join(t.TempDir(), termsFileName)
	if err := os.WriteFile(p, []byte("{hỏng"), 0o644); err != nil {
		t.Fatal(err)
	}
	if st := readTerms(p); st.AcceptedVersion != 0 {
		t.Fatalf("file hỏng phải coi như chưa đồng ý, có %+v", st)
	}
}
