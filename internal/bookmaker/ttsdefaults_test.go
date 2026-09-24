package bookmaker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVenvPython(t *testing.T) {
	venv := filepath.Join("home", ".venv")
	tests := []struct {
		goos string
		want string
	}{
		{"linux", filepath.Join(venv, "bin", "python")},
		{"darwin", filepath.Join(venv, "bin", "python")},
		{"windows", filepath.Join(venv, "Scripts", "python.exe")},
	}
	for _, tt := range tests {
		if got := venvPython(venv, tt.goos); got != tt.want {
			t.Errorf("venvPython(%q) = %q, want %q", tt.goos, got, tt.want)
		}
	}
}

func TestFindTTSScriptNextToBinary(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, ttsScriptRel)
	if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, []byte("#"), 0o644); err != nil {
		t.Fatal(err)
	}
	// CWD = thư mục test (không có scripts/tts) → phải tìm theo vị trí binary.
	exe := filepath.Join(root, "bin", "sano-docx2tts")
	got := findTTSScript(exe, t.TempDir())
	want, _ := filepath.Abs(script)
	if got != want {
		t.Errorf("findTTSScript = %q, want %q", got, want)
	}
}

// Chạy sano-docx2tts trong một thư mục lạ có sẵn scripts/tts/audio_gen_batch.py
// không được chạy script đó khi không truyền --tts-script.
func TestFindTTSScript_IgnoresScriptInCWD(t *testing.T) {
	cwd := t.TempDir()
	planted := filepath.Join(cwd, ttsScriptRel)
	if err := os.MkdirAll(filepath.Dir(planted), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planted, []byte("import os; os.system('x')"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(cwd)
	got := findTTSScript("", t.TempDir())
	if real, _ := filepath.EvalSymlinks(got); got == planted || real == planted || filepath.Dir(got) == filepath.Dir(planted) {
		t.Fatalf("findTTSScript dùng script trong thư mục hiện tại: %q", got)
	}
}

// Không có script cạnh file chạy (vd `go run`) → dùng bản nhúng, giải nén vào
// thư mục cache, nội dung khớp script trong repo; chạy lại không lỗi.
func TestFindTTSScript_FallsBackToEmbedded(t *testing.T) {
	cache := t.TempDir()
	got := findTTSScript("", cache)
	rel, err := filepath.Rel(cache, got)
	if err != nil || !filepath.IsLocal(rel) || filepath.Base(got) != "audio_gen_batch.py" {
		t.Fatalf("findTTSScript = %q, muốn bản nhúng trong %q", got, cache)
	}
	want, err := os.ReadFile(filepath.Join("..", "..", ttsScriptRel))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(got)
	if err != nil || string(data) != string(want) {
		t.Fatalf("bản nhúng khác script trong repo (err=%v)", err)
	}
	for _, n := range []string{"models.py", "audio_gen.py", "versions.env"} {
		if !fileExistsTTS(filepath.Join(filepath.Dir(got), n)) {
			t.Errorf("thiếu %s cạnh script", n)
		}
	}
	if again := findTTSScript("", cache); again != got {
		t.Errorf("lần 2 = %q, muốn %q", again, got)
	}
}
