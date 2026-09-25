// Package ttsscripts nhúng script đọc giọng (scripts/tts/*.py) và file ghim phiên
// bản versions.env vào chương trình Go.
//
// Vì sao đặt ở đây: go:embed chỉ nhúng được file nằm trong cùng thư mục (hoặc
// thư mục con) của package, không đi ngược ra ngoài. Đặt file Go ngay cạnh script
// thì chỉ có MỘT bản script (không chép, không sinh lúc build); phần mềm desktop
// (module sano/desktop, replace sano => ../) import package này và giải nén
// script vào thư mục dữ liệu của app lúc cài bộ đọc.
package ttsscripts

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

// Files — các file bộ đọc cần lúc chạy. requirements.txt và file mẫu không nhúng
// (chỉ để đọc/đối chiếu, xem docs/tts-build-guide.md). Thư mục vieneu/ là
// pyproject.toml + uv.lock Sano chép đè vào mã VieNeu-TTS trước `uv sync`
// (VieNeuProjectFiles); không giải nén cùng script.
//
//go:embed audio_gen.py audio_gen_batch.py models.py models.sha256 versions.env vieneu/pyproject.toml vieneu/uv.lock
var Files embed.FS

// VieNeuProjectFiles — file trong vieneu/ chép vào thư mục mã VieNeu-TTS: bỏ
// giao diện web gradio và ép bản vá bảo mật (scripts/tts/vieneu-overrides.txt,
// dựng bằng scripts/tts/vieneu-lock.sh).
var VieNeuProjectFiles = []string{"pyproject.toml", "uv.lock"}

// VieNeuDir — thư mục con trong Files chứa VieNeuProjectFiles.
const VieNeuDir = "vieneu"

// VersionsFile — tên file ghim phiên bản.
const VersionsFile = "versions.env"

// Pins đọc versions.env đã nhúng.
func Pins() (map[string]string, error) {
	data, err := Files.ReadFile(VersionsFile)
	if err != nil {
		return nil, err
	}
	return ParsePins(data)
}

// ParsePins đọc định dạng KEY=VALUE (bỏ dòng trống và dòng #), cùng quy tắc với
// read_pins() trong models.py.
func ParsePins(data []byte) (map[string]string, error) {
	pins := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("%s dòng %d: thiếu dấu =", VersionsFile, n)
		}
		pins[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return pins, sc.Err()
}

// Names trả tên các file đã nhúng (đã sắp xếp).
func Names() []string {
	entries, _ := fs.ReadDir(Files, ".")
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out
}
