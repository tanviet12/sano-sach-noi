package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("0.3.1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "desktop", "frontend")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		built  string
		starts []string
		want   string
	}{
		{"ldflags thắng", "1.2.0", []string{sub}, "1.2.0"},
		{"đọc VERSION ở thư mục cha", "", []string{sub}, "0.3.1"},
		{"bỏ qua gốc rỗng", "", []string{"", sub}, "0.3.1"},
		{"không có gì → dev", "", []string{t.TempDir()}, devVersion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveVersion(tt.built, tt.starts...); got != tt.want {
				t.Errorf("được %q, muốn %q", got, tt.want)
			}
		})
	}
}
