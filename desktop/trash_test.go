package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Đường dẫn thư mục sách (tên do người khác đặt, có thể chứa dấu nháy cong ’
// PowerShell coi là đóng chuỗi) không bao giờ được ghép vào script PowerShell:
// script là hằng số, đường dẫn chỉ đi qua biến môi trường.
func TestRecycleBinCommand_PathOnlyInEnv(t *testing.T) {
	path := `C:\Users\a\Sano\Sach\a’;Start-Process calc;’`
	cmd := recycleBinCommand(path)
	script := cmd.Args[len(cmd.Args)-1]
	if script != recycleBinScript {
		t.Fatalf("script phải là hằng số recycleBinScript, got %q", script)
	}
	for _, a := range cmd.Args {
		if strings.Contains(a, "Start-Process") || strings.Contains(a, "Sach") {
			t.Fatalf("đường dẫn lọt vào tham số dòng lệnh: %q", a)
		}
	}
	if !strings.Contains(recycleBinScript, "$env:"+trashPathEnv) {
		t.Fatalf("script phải đọc đường dẫn từ $env:%s", trashPathEnv)
	}
	found := 0
	for _, e := range cmd.Env {
		if e == trashPathEnv+"="+path {
			found++
		}
	}
	if found != 1 {
		t.Fatalf("muốn đúng 1 biến %s=<đường dẫn> trong Env, có %d", trashPathEnv, found)
	}
}

func TestMoveAside_DoiTenKhiTrung(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "rac")
	for i := 0; i < 2; i++ {
		src := filepath.Join(root, "sach-mau")
		if err := os.MkdirAll(src, 0o755); err != nil {
			t.Fatal(err)
		}
		dst, err := moveAside(src, bin)
		if err != nil {
			t.Fatalf("lần %d: %v", i, err)
		}
		if _, err := os.Stat(src); !os.IsNotExist(err) {
			t.Fatalf("lần %d: thư mục gốc vẫn còn", i)
		}
		if _, err := os.Stat(dst); err != nil {
			t.Fatalf("lần %d: không thấy %s", i, dst)
		}
	}
	entries, _ := os.ReadDir(bin)
	if len(entries) != 2 {
		t.Fatalf("muốn 2 bản trong thùng rác, có %d", len(entries))
	}
}
