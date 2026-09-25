package bookmaker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultVoice — giọng mặc định (khớp DEFAULT_VOICE của scripts/tts/audio_gen.py).
const DefaultVoice = "Hải Đăng"

// Voice — một giọng preset của VieNeu-TTS.
type Voice struct {
	Name     string `json:"name"`     // tên truyền vào --voice
	Desc     string `json:"desc"`     // vd "Nữ · Bắc · Phong cách tự nhiên"
	Featured bool   `json:"featured"` // bộ đọc đánh dấu ⭐
}

// voiceLineRe khớp dòng "   • ⭐ Trúc Ly — Nữ · Bắc · Phong cách tự nhiên" của --list-voices.
var voiceLineRe = regexp.MustCompile(`^\s*•\s*(⭐\s*)?(.+?)\s+—\s+(.+?)\s*$`)

// ListVoices hỏi bộ đọc danh sách giọng (audio_gen_batch.py --list-voices).
// env: biến môi trường thêm cho python (như TTSConfig.Env), nil nếu không cần.
func ListVoices(ctx context.Context, python, script string, env []string) ([]Voice, error) {
	if !fileExistsTTS(python) {
		return nil, fmt.Errorf("không tìm thấy python của bộ đọc %q", python)
	}
	if !fileExistsTTS(script) {
		return nil, fmt.Errorf("không tìm thấy script đọc giọng %q", script)
	}
	cmd := exec.CommandContext(ctx, python, script, "--list-voices")
	cmd.Env = append(append(os.Environ(), env...), "PYTHONPATH="+filepath.Dir(script), "PYTHONUNBUFFERED=1")
	hideWindow(cmd)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("lấy danh sách giọng: %w\n%s", err, truncateBytes(out.Bytes(), 800))
	}
	voices := parseVoices(out.String())
	if len(voices) == 0 {
		return nil, fmt.Errorf("bộ đọc không trả giọng nào")
	}
	return voices, nil
}

func parseVoices(s string) []Voice {
	var out []Voice
	for _, ln := range strings.Split(s, "\n") {
		m := voiceLineRe.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		out = append(out, Voice{Name: m[2], Desc: m[3], Featured: m[1] != ""})
	}
	return out
}
