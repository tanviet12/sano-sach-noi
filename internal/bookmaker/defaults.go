package bookmaker

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	ttsscripts "sano/scripts/tts"
)

// ttsVenvDir — thư mục VieNeu-TTS v3 mặc định (tính từ HOME). Tách khỏi
// ~/VieNeu-TTS để không lẫn với bản v2 cũ: script v3 chạy bằng venv v2 sẽ lỗi.
const ttsVenvDir = "VieNeu-TTS-v3"

// TTSDefaults trả python venv VieNeu-TTS v3 mặc định (~/VieNeu-TTS-v3/.venv) +
// script đọc giọng mặc định (xem findTTSScript).
func TTSDefaults() (python, script string) {
	if home, err := os.UserHomeDir(); err == nil {
		python = venvPython(filepath.Join(home, ttsVenvDir, ".venv"), runtime.GOOS)
	}
	exe, _ := os.Executable()
	cache, _ := os.UserCacheDir()
	return python, findTTSScript(exe, cache)
}

// venvPython trả đường dẫn python trong venv theo hệ điều hành.
func venvPython(venv, goos string) string {
	if goos == "windows" {
		return filepath.Join(venv, "Scripts", "python.exe")
	}
	return filepath.Join(venv, "bin", "python")
}

// ttsScriptRel — vị trí script đọc giọng tính từ gốc repo.
var ttsScriptRel = filepath.Join("scripts", "tts", "audio_gen_batch.py")

// findTTSScript chọn script đọc giọng mặc định khi không truyền --tts-script:
//  1. cạnh file chạy: bin/sano-docx2tts → bin/../scripts/tts/audio_gen_batch.py
//  2. bản nhúng trong chương trình, giải nén vào <cacheDir>/sano/tts-scripts/<mã>
//
// KHÔNG tìm theo thư mục hiện tại: chạy lệnh trong một thư mục lạ có sẵn
// scripts/tts/audio_gen_batch.py sẽ chạy script lạ. Muốn dùng script trong
// repo đang sửa thì truyền --tts-script. Không có gì → trả đường dẫn cạnh file
// chạy để preflight báo lỗi rõ ràng.
func findTTSScript(exe, cacheDir string) string {
	var nextToExe string
	if exe != "" {
		if abs, err := filepath.Abs(filepath.Join(filepath.Dir(exe), "..", ttsScriptRel)); err == nil {
			nextToExe = abs
			if fileExistsTTS(abs) {
				return abs
			}
		}
	}
	if cacheDir != "" {
		if p, err := extractEmbeddedTTSScripts(cacheDir); err == nil {
			return p
		}
	}
	return nextToExe
}

// extractEmbeddedTTSScripts ghi các script bộ đọc nhúng trong chương trình vào
// <cacheDir>/sano/tts-scripts/<sha256 rút gọn> (ghi lại file thiếu / khác nội
// dung), trả đường dẫn audio_gen_batch.py.
func extractEmbeddedTTSScripts(cacheDir string) (string, error) {
	names := ttsscripts.Names()
	h := sha256.New()
	files := make(map[string][]byte, len(names))
	for _, n := range names {
		data, err := ttsscripts.Files.ReadFile(n)
		if err != nil {
			return "", err
		}
		files[n] = data
		h.Write([]byte(n))
		h.Write(data)
	}
	dir := filepath.Join(cacheDir, "sano", "tts-scripts", hex.EncodeToString(h.Sum(nil))[:12])
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	for n, data := range files {
		dst := filepath.Join(dir, n)
		if old, err := os.ReadFile(dst); err == nil && bytes.Equal(old, data) {
			continue
		}
		tmp, err := os.CreateTemp(dir, ".tmp-*")
		if err != nil {
			return "", err
		}
		_, werr := tmp.Write(data)
		cerr := tmp.Close()
		if werr == nil {
			werr = cerr
		}
		if werr == nil {
			werr = os.Rename(tmp.Name(), dst)
		}
		if werr != nil {
			_ = os.Remove(tmp.Name())
			return "", werr
		}
	}
	return filepath.Join(dir, filepath.Base(ttsScriptRel)), nil
}

// SplitTags tách CSV tag, bỏ phần tử rỗng.
func SplitTags(csv string) []string {
	var out []string
	for _, t := range strings.Split(csv, ",") {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// ParseStemSet biến CSV stem (vd "ch03-sec01,ch04-sec02") thành tập tra cứu.
// Trống → nil (nghĩa là render thật toàn bộ, giữ hành vi cũ).
func ParseStemSet(csv string) map[string]bool {
	set := make(map[string]bool)
	for _, s := range strings.Split(csv, ",") {
		if s = strings.TrimSpace(s); s != "" {
			set[s] = true
		}
	}
	if len(set) == 0 {
		return nil
	}
	return set
}

// NewNormalizer dựng bộ luật chuẩn hóa lời đọc: từ điển cách đọc mặc định +
// file TSV của người dùng (trống = chỉ mặc định), giữ hay bỏ số đầu tiêu đề.
func NewNormalizer(pronunciationFile string, keepHeadingNumbers bool) (*Normalizer, error) {
	dict, err := loadPronunciationDict(strings.TrimSpace(pronunciationFile))
	if err != nil {
		return nil, err
	}
	return &Normalizer{dict: dict, keepHeadingNumbers: keepHeadingNumbers}, nil
}

// FirstNonEmpty trả giá trị đầu tiên khác rỗng (đã trim); tất cả rỗng → "Sách nói".
func FirstNonEmpty(vals ...string) string {
	return firstNonEmpty(vals...)
}
