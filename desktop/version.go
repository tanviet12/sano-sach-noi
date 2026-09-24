package main

import (
	"os"
	"path/filepath"
	"strings"
)

// version gắn lúc build: wails build -ldflags "-X main.version=$(cat VERSION)".
// Rỗng khi chạy `wails dev` → đọc file VERSION ở gốc repo nếu có.
var version = ""

// devVersion hiện khi không có cả ldflags lẫn file VERSION.
const devVersion = "dev"

// resolveVersion: ldflags → file VERSION (đi ngược lên từ các thư mục gốc) → "dev".
func resolveVersion(built string, starts ...string) string {
	if v := strings.TrimSpace(built); v != "" {
		return v
	}
	for _, start := range starts {
		if v := findVersionFile(start); v != "" {
			return v
		}
	}
	return devVersion
}

func findVersionFile(start string) string {
	if start == "" {
		return ""
	}
	dir, err := filepath.Abs(start)
	if err != nil {
		return ""
	}
	for {
		if b, err := os.ReadFile(filepath.Join(dir, "VERSION")); err == nil {
			if v := strings.TrimSpace(string(b)); v != "" {
				return v
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
