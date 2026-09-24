package ttsscripts

import (
	"regexp"
	"strings"
	"testing"
)

var hex64 = regexp.MustCompile(`^[0-9a-f]{64}$`)

func TestPins_DuKhoaCanCho(t *testing.T) {
	pins, err := Pins()
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{
		"VIENEU_REPO", "VIENEU_COMMIT", "VIENEU_TREE_SHA256", "UV_VERSION", "PYTHON_VERSION",
		"HF_BACKBONE_REPO", "HF_BACKBONE_REVISION", "HF_BACKBONE_FILES",
		"HF_CODEC_REPO", "HF_CODEC_REVISION", "HF_CODEC_FILES",
	} {
		if pins[k] == "" {
			t.Errorf("versions.env thiếu %s", k)
		}
	}
	for k, v := range pins {
		if strings.HasSuffix(k, "_SHA256") && !hex64.MatchString(v) {
			t.Errorf("%s không phải SHA256 hợp lệ: %q", k, v)
		}
		if strings.HasSuffix(k, "_URL") && !strings.HasPrefix(v, "https://") {
			t.Errorf("%s phải là https: %q", k, v)
		}
	}
	for _, target := range []string{"AARCH64_APPLE_DARWIN", "X86_64_APPLE_DARWIN", "X86_64_PC_WINDOWS_MSVC", "X86_64_UNKNOWN_LINUX_GNU"} {
		if pins["UV_SHA256_"+target] == "" {
			t.Errorf("thiếu SHA256 uv cho %s", target)
		}
	}
}

// Mọi file model ghim trong versions.env phải có dòng SHA256 trong models.sha256.
func TestModelsSHA256_KhopVersions(t *testing.T) {
	pins, err := Pins()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := Files.ReadFile("models.sha256")
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Fields(line)
		if len(f) != 2 || !hex64.MatchString(f[0]) {
			t.Fatalf("dòng sai định dạng: %q", line)
		}
		have[f[1]] = true
	}
	for _, key := range []string{"BACKBONE", "CODEC"} {
		repo, rev := pins["HF_"+key+"_REPO"], pins["HF_"+key+"_REVISION"]
		for _, f := range strings.Split(pins["HF_"+key+"_FILES"], ",") {
			name := repo + "@" + rev + "/" + strings.TrimSpace(f)
			if !have[name] {
				t.Errorf("models.sha256 thiếu %s", name)
			}
		}
	}
}

func TestNames_CoScriptChinh(t *testing.T) {
	names := strings.Join(Names(), ",")
	for _, n := range []string{"audio_gen.py", "audio_gen_batch.py", "models.py", "models.sha256", "versions.env"} {
		if !strings.Contains(names, n) {
			t.Errorf("thiếu %s trong bản nhúng (%s)", n, names)
		}
	}
}

func TestParsePins_LoiThieuDauBang(t *testing.T) {
	if _, err := ParsePins([]byte("# chú thích\nA=1\nKHONGHOP\n")); err == nil {
		t.Error("phải báo lỗi dòng thiếu =")
	}
}
