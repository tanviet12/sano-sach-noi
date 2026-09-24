package tts

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func noEnv(string) string { return "" }

func TestVenvPython(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{"darwin", filepath.Join("v", "bin", "python")},
		{"linux", filepath.Join("v", "bin", "python")},
		{"windows", filepath.Join("v", "Scripts", "python.exe")},
	}
	for _, tt := range tests {
		if got := VenvPython("v", tt.goos); got != tt.want {
			t.Errorf("VenvPython(%q) = %q, muốn %q", tt.goos, got, tt.want)
		}
	}
}

func TestFindScriptsDir(t *testing.T) {
	root := t.TempDir()
	scripts := filepath.Join(root, "scripts", "tts")
	if err := os.MkdirAll(scripts, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scripts, "models.py"), []byte("#"), 0o644); err != nil {
		t.Fatal(err)
	}
	deep := filepath.Join(root, "desktop", "build", "bin", "Sano.app", "Contents", "MacOS")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}

	if got := FindScriptsDir(noEnv, "", deep); got != scripts {
		t.Errorf("đi ngược lên từ .app: được %q, muốn %q", got, scripts)
	}
	envDir := func(k string) string {
		if k == EnvScripts {
			return scripts
		}
		return ""
	}
	if got := FindScriptsDir(envDir); got != scripts {
		t.Errorf("biến môi trường: được %q", got)
	}
	if got := FindScriptsDir(noEnv, t.TempDir()); got != "" {
		t.Errorf("không có scripts phải trả rỗng, được %q", got)
	}
}

func TestCheck(t *testing.T) {
	dir := t.TempDir()
	py := filepath.Join(dir, "python")
	if err := os.WriteFile(py, []byte(""), 0o755); err != nil {
		t.Fatal(err)
	}

	ok := func(_ context.Context, _ []string, _ string, args ...string) (string, error) {
		if args[0] == "--version" {
			return "Python 3.12.4", nil
		}
		return "Đủ model ghim trong /cache", nil
	}
	missingModel := func(_ context.Context, _ []string, _ string, args ...string) (string, error) {
		if args[0] == "--version" {
			return "Python 3.12.4", nil
		}
		return "Thiếu:\n  a/b", errors.New("exit status 1")
	}
	brokenPy := func(context.Context, []string, string, ...string) (string, error) {
		return "", errors.New("exec format error")
	}

	tests := []struct {
		name      string
		python    string
		scripts   string
		run       Runner
		ready     bool
		pyFound   bool
		msgSubstr string
	}{
		{"không có python", filepath.Join(dir, "khong-co"), dir, ok, false, false, "Chưa cài"},
		{"python rỗng", "", dir, ok, false, false, "Chưa cài"},
		{"python hỏng", py, dir, brokenPy, false, true, "không chạy được"},
		{"thiếu script", py, "", ok, false, true, "Thiếu script"},
		{"thiếu mô hình", py, dir, missingModel, false, true, "mô hình"},
		{"sẵn sàng", py, dir, ok, true, true, "Sẵn sàng"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := Check(context.Background(), Runtime{Python: tt.python, ScriptsDir: tt.scripts}, tt.run)
			if st.Ready != tt.ready || st.PythonFound != tt.pyFound {
				t.Errorf("ready=%v pythonFound=%v, muốn %v %v (%+v)", st.Ready, st.PythonFound, tt.ready, tt.pyFound, st)
			}
			if !strings.Contains(st.Message, tt.msgSubstr) {
				t.Errorf("message %q không chứa %q", st.Message, tt.msgSubstr)
			}
		})
	}

	st := Check(context.Background(), Runtime{Python: py, ScriptsDir: dir}, ok)
	if st.PythonVersion != "3.12.4" {
		t.Errorf("PythonVersion = %q", st.PythonVersion)
	}
}
