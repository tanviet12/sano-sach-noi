package tts

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"
)

func envMap(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestDataDir(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		goos string
		home string
		want string
		err  bool
	}{
		{"biến môi trường", map[string]string{EnvDataDir: "/tmp/sano-thu"}, "darwin", "/Users/a", "/tmp/sano-thu", false},
		{"biến môi trường tương đối", map[string]string{EnvDataDir: "tam"}, "darwin", "/Users/a", "", true},
		{"macOS", nil, "darwin", "/Users/a", "/Users/a/Library/Application Support/Sano", false},
		{"Linux XDG", map[string]string{"XDG_DATA_HOME": "/x/data"}, "linux", "/home/a", "/x/data/sano", false},
		{"Linux mặc định", nil, "linux", "/home/a", "/home/a/.local/share/sano", false},
		{"không HOME", nil, "linux", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DataDir(envMap(tt.env), tt.goos, tt.home)
			if (err != nil) != tt.err {
				t.Fatalf("err = %v, muốn lỗi=%v", err, tt.err)
			}
			if !tt.err && got != filepath.FromSlash(tt.want) {
				t.Errorf("được %q, muốn %q", got, tt.want)
			}
		})
	}
	got, err := DataDir(envMap(map[string]string{"LOCALAPPDATA": `C:\Users\a\AppData\Local`}), "windows", `C:\Users\a`)
	if err != nil || got != filepath.Join(`C:\Users\a\AppData\Local`, "Sano") {
		t.Errorf("Windows: %q %v", got, err)
	}
}

func TestLayout(t *testing.T) {
	l := NewLayout("/d", "windows")
	if l.Root != filepath.Join("/d", "tts") {
		t.Errorf("Root = %q", l.Root)
	}
	if filepath.Base(l.UV()) != "uv.exe" || filepath.Base(l.FFmpeg()) != "ffmpeg.exe" {
		t.Errorf("Windows phải có đuôi .exe: %q %q", l.UV(), l.FFmpeg())
	}
	if l.Python() != filepath.Join("/d", "tts", "vieneu", ".venv", "Scripts", "python.exe") {
		t.Errorf("Python = %q", l.Python())
	}
	env := NewLayout("/d", "linux").Env()
	if !slices.Contains(env, "HF_HOME="+filepath.Join("/d", "tts", "hf")) {
		t.Errorf("Env thiếu HF_HOME: %v", env)
	}
}

func writeExe(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("#"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestResolve_ThuTuPython(t *testing.T) {
	home := t.TempDir()
	l := NewLayout(t.TempDir(), "linux")
	base := ResolveOptions{Getenv: noEnv, Home: home, GOOS: "linux", Layout: l}

	// Chưa có gì → trỏ về bộ đọc app (chưa cài)
	rt := Resolve(base)
	if rt.Python != l.Python() || rt.Source != SourceApp {
		t.Errorf("chưa cài: %+v", rt)
	}

	// Có ~/VieNeu-TTS-v3 → dùng bản cài tay, không thêm env
	legacy := VenvPython(filepath.Join(home, VenvDir, ".venv"), "linux")
	writeExe(t, legacy)
	rt = Resolve(base)
	if rt.Python != legacy || rt.Source != SourceLegacy || rt.Env != nil {
		t.Errorf("legacy: %+v", rt)
	}

	// App đã cài → ưu tiên hơn bản cài tay, có HF_HOME riêng
	writeExe(t, l.Python())
	rt = Resolve(base)
	if rt.Python != l.Python() || rt.Source != SourceApp || len(rt.Env) == 0 {
		t.Errorf("app: %+v", rt)
	}

	// SANO_TTS_PYTHON thắng tất cả
	o := base
	o.Getenv = envMap(map[string]string{EnvPython: "/opt/py"})
	rt = Resolve(o)
	if rt.Python != "/opt/py" || rt.Source != SourceEnv {
		t.Errorf("env: %+v", rt)
	}
}

// Bản phát hành không dò scripts/tts theo thư mục hiện tại, kể cả khi không xác
// định được thư mục dữ liệu: báo lỗi rõ. Chỉ bản dev mới dò.
func TestResolve_ReleaseDoesNotWalkCwd(t *testing.T) {
	cwd := t.TempDir()
	evil := filepath.Join(cwd, "scripts", "tts")
	if err := os.MkdirAll(evil, 0o755); err != nil {
		t.Fatal(err)
	}
	writeExe(t, filepath.Join(evil, "models.py"))
	src := fstest.MapFS{"models.py": {Data: []byte("print('m')")}}
	noRoot := Layout{GOOS: "linux"} // thư mục dữ liệu không xác định được

	rt := Resolve(ResolveOptions{Getenv: noEnv, GOOS: "linux", Layout: noRoot, Scripts: src, Starts: []string{cwd}})
	if rt.ScriptsDir != "" {
		t.Fatalf("bản phát hành không được dùng script trong thư mục hiện tại: %q", rt.ScriptsDir)
	}
	if rt.ScriptsErr == "" {
		t.Error("phải có lý do rõ ràng khi không có script")
	}
	st := Check(t.Context(), Runtime{Python: writeExeRet(t), ScriptsErr: rt.ScriptsErr}, func(context.Context, []string, string, ...string) (string, error) {
		return "Python 3.12.0", nil
	})
	if st.Detail != rt.ScriptsErr {
		t.Errorf("Check phải hiện lý do, Detail = %q", st.Detail)
	}

	rt = Resolve(ResolveOptions{Getenv: noEnv, GOOS: "linux", Layout: noRoot, Scripts: src, Starts: []string{cwd}, Dev: true})
	if rt.ScriptsDir != evil {
		t.Errorf("bản dev phải dò được scripts/tts trong repo, got %q", rt.ScriptsDir)
	}
}

func writeExeRet(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "python")
	writeExe(t, p)
	return p
}

func TestResolve_ScriptNhung(t *testing.T) {
	l := NewLayout(t.TempDir(), "linux")
	src := fstest.MapFS{
		"models.py":    {Data: []byte("print('m')")},
		"versions.env": {Data: []byte("A=1\n")},
	}
	rt := Resolve(ResolveOptions{Getenv: noEnv, GOOS: "linux", Layout: l, Scripts: src})
	if rt.ScriptsDir != l.Scripts() {
		t.Fatalf("ScriptsDir = %q, muốn %q", rt.ScriptsDir, l.Scripts())
	}
	got, err := os.ReadFile(filepath.Join(l.Scripts(), "models.py"))
	if err != nil || string(got) != "print('m')" {
		t.Fatalf("script giải nén sai: %q %v", got, err)
	}
	if _, err := os.Stat(l.Marker()); err != nil {
		t.Errorf("thiếu file đánh dấu: %v", err)
	}

	// Bản nhúng đổi → ghi đè
	src["models.py"] = &fstest.MapFile{Data: []byte("print('moi')")}
	if err := EnsureScripts(l, src); err != nil {
		t.Fatal(err)
	}
	got, _ = os.ReadFile(filepath.Join(l.Scripts(), "models.py"))
	if string(got) != "print('moi')" {
		t.Errorf("không ghi đè bản mới: %q", got)
	}

	// SANO_TTS_SCRIPTS vẫn ưu tiên (dev)
	dev := t.TempDir()
	writeExe(t, filepath.Join(dev, "models.py"))
	rt = Resolve(ResolveOptions{Getenv: envMap(map[string]string{EnvScripts: dev}), GOOS: "linux", Layout: l, Scripts: src})
	if rt.ScriptsDir != dev {
		t.Errorf("SANO_TTS_SCRIPTS phải thắng: %q", rt.ScriptsDir)
	}
}
