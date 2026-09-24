package bookmaker

import (
	"os"
	"path/filepath"
	"testing"
)

// Thư mục sách (người khác gửi) có MP3 hoặc bìa là symlink ra ngoài → đóng gói
// lại phải từ chối, không chép file ngoài vào gói zip.
func TestRepackZip_RejectsSymlinkedSources(t *testing.T) {
	const meta = `{"title":"T","cover":"cover.png","chapters":[{"title":"C","sections":[{"title":"S","file":"ch01-sec01.mp3"}]}]}`
	for _, link := range []string{"ch01-sec01.mp3", "cover.png"} {
		t.Run(link, func(t *testing.T) {
			base := t.TempDir()
			dir := filepath.Join(base, "sach")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			secret := filepath.Join(base, "bi-mat")
			if err := os.WriteFile(secret, []byte("bí mật"), 0o644); err != nil {
				t.Fatal(err)
			}
			for _, f := range []string{"ch01-sec01.mp3", "cover.png"} {
				p := filepath.Join(dir, f)
				var err error
				if f == link {
					err = os.Symlink(secret, p)
				} else {
					err = os.WriteFile(p, []byte("x"), 0o644)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(meta), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := RepackZip(dir, filepath.Join(base, "out.zip"), "v", nil); err == nil {
				t.Fatalf("%s là symlink: phải từ chối", link)
			}
		})
	}
}

// metadata.json là symlink ra ngoài thư mục sách → không đọc theo, từ chối.
func TestRepackZip_RejectsSymlinkedMetadata(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "sach")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(base, "meta-ngoai.json")
	const meta = `{"title":"T","chapters":[{"title":"C","sections":[{"title":"S","file":"ch01-sec01.mp3"}]}]}`
	if err := os.WriteFile(outside, []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ch01-sec01.mp3"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "metadata.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := RepackZip(dir, filepath.Join(base, "out.zip"), "v", nil); err == nil {
		t.Fatal("metadata.json là symlink: phải từ chối")
	}
}
