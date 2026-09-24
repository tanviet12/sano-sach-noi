package safepath

import (
	"path/filepath"
	"testing"
)

func TestIsPlainName(t *testing.T) {
	good := []string{"ch01-sec01.mp3", "cover.png", "img001.bin", "tên có dấu.mp3"}
	bad := []string{"", ".", "..", "../x.mp3", `..\x.mp3`, "a/b.mp3", `a\b.mp3`, "/etc/passwd", `C:\x`}
	for _, n := range good {
		if !IsPlainName(n) {
			t.Errorf("IsPlainName(%q) = false, muốn true", n)
		}
	}
	for _, n := range bad {
		if IsPlainName(n) {
			t.Errorf("IsPlainName(%q) = true, muốn false", n)
		}
	}
}

func TestWithin(t *testing.T) {
	dir := filepath.Join("a", "b")
	if !Within(dir, filepath.Join(dir, "c.png")) {
		t.Error("file con phải nằm trong")
	}
	for _, p := range []string{dir, filepath.Join(dir, "..", "x"), filepath.Join("a", "bx")} {
		if Within(dir, p) {
			t.Errorf("Within(%q, %q) phải false", dir, p)
		}
	}
}
