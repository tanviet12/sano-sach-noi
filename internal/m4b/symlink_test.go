package m4b

import (
	"os"
	"path/filepath"
	"testing"
)

const symMeta = `{"title":"X","cover":"cover.png","chapters":[{"title":"A","sections":[{"title":"A","file":"ch01-sec01.mp3"}]}]}`

func symDir(t *testing.T) (base, dir, secret string) {
	t.Helper()
	base = t.TempDir()
	dir = filepath.Join(base, "sach")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	secret = filepath.Join(base, "bi-mat")
	if err := os.WriteFile(secret, []byte("bí mật"), 0o644); err != nil {
		t.Fatal(err)
	}
	return base, dir, secret
}

func write(t *testing.T, p, body string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// MP3 là symlink ra ngoài → từ chối xuất.
func TestFromDir_SymlinkTrack_Rejected(t *testing.T) {
	_, dir, secret := symDir(t)
	write(t, filepath.Join(dir, "metadata.json"), symMeta)
	if err := os.Symlink(secret, filepath.Join(dir, "ch01-sec01.mp3")); err != nil {
		t.Fatal(err)
	}
	if _, err := FromDir(dir); err == nil {
		t.Fatal("MP3 là symlink phải bị từ chối")
	}
}

// Bìa là symlink → bỏ qua bìa (tự vẽ), không đọc file ngoài.
func TestFromDir_SymlinkCover_Ignored(t *testing.T) {
	_, dir, secret := symDir(t)
	write(t, filepath.Join(dir, "metadata.json"), symMeta)
	write(t, filepath.Join(dir, "ch01-sec01.mp3"), "x")
	if err := os.Symlink(secret, filepath.Join(dir, "cover.png")); err != nil {
		t.Fatal(err)
	}
	b, err := FromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if b.Cover != "" {
		t.Errorf("bìa symlink phải bị bỏ qua, got %q", b.Cover)
	}
}

// metadata.json là symlink → từ chối.
func TestFromDir_SymlinkMetadata_Rejected(t *testing.T) {
	base, dir, _ := symDir(t)
	outside := filepath.Join(base, "meta.json")
	write(t, outside, symMeta)
	write(t, filepath.Join(dir, "ch01-sec01.mp3"), "x")
	if err := os.Symlink(outside, filepath.Join(dir, "metadata.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := FromDir(dir); err == nil {
		t.Fatal("metadata.json là symlink phải bị từ chối")
	}
}
