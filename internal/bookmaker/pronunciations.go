package bookmaker

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// defaultPronunciationsTSV — từ điển cách đọc mặc định (viết tắt kinh doanh /
// công nghệ phổ biến), nhúng vào binary để chạy offline, không cần file ngoài.
//
//go:embed pronunciations.default.tsv
var defaultPronunciationsTSV string

// pronunciationDict — viết tắt → cách đọc. Khớp nguyên từ, phân biệt hoa thường.
type pronunciationDict struct {
	entries map[string]string
}

// pronunKeyRe — dạng hợp lệ của 1 viết tắt trong từ điển: chữ/số, có thể nối
// bằng & (R&D, P&L, B2B). Cũng là regex tách "từ" khi tra từ điển trong văn bản.
var pronunKeyRe = regexp.MustCompile(`^[\p{L}\p{N}]+(?:&[\p{L}\p{N}]+)*$`)

// pronunTokenRe tách các "từ" trong văn bản để tra từ điển (cùng dạng khóa).
var pronunTokenRe = regexp.MustCompile(`[\p{L}\p{N}]+(?:&[\p{L}\p{N}]+)*`)

// parsePronunciations đọc định dạng TSV: "<viết tắt>\t<cách đọc>", dòng # là chú
// thích, dòng trắng bỏ qua. Cách đọc rỗng = tắt mục đó (value ""), để lớp ghi đè
// có thể bỏ 1 mục mặc định. source dùng cho thông báo lỗi.
func parsePronunciations(data, source string) (map[string]string, error) {
	out := map[string]string{}
	sc := bufio.NewScanner(strings.NewReader(data))
	line := 0
	for sc.Scan() {
		line++
		raw := strings.TrimRight(sc.Text(), "\r")
		if strings.TrimSpace(raw) == "" || strings.HasPrefix(strings.TrimSpace(raw), "#") {
			continue
		}
		key, val, ok := strings.Cut(raw, "\t")
		if !ok {
			return nil, fmt.Errorf("%s dòng %d: thiếu dấu TAB giữa viết tắt và cách đọc", source, line)
		}
		key = strings.TrimSpace(key)
		val = strings.Join(strings.Fields(val), " ")
		if !pronunKeyRe.MatchString(key) {
			return nil, fmt.Errorf("%s dòng %d: viết tắt %q không hợp lệ (chỉ chữ, số và &)", source, line, key)
		}
		out[key] = val
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("đọc %s: %w", source, err)
	}
	return out, nil
}

// loadPronunciationDict dựng từ điển: mặc định nhúng sẵn, rồi (nếu có) file của
// người dùng ghi đè / bổ sung. Mục có cách đọc rỗng bị bỏ khỏi từ điển.
func loadPronunciationDict(userFile string) (*pronunciationDict, error) {
	base, err := parsePronunciations(defaultPronunciationsTSV, "pronunciations.default.tsv")
	if err != nil {
		return nil, err
	}
	if userFile != "" {
		data, err := os.ReadFile(userFile)
		if err != nil {
			return nil, fmt.Errorf("đọc từ điển cách đọc %q: %w", userFile, err)
		}
		user, err := parsePronunciations(string(data), userFile)
		if err != nil {
			return nil, err
		}
		for k, v := range user {
			base[k] = v
		}
	}
	d := &pronunciationDict{entries: map[string]string{}}
	for k, v := range base {
		if v != "" {
			d.entries[k] = v
		}
	}
	return d, nil
}

// mustDefaultPronunciationDict — từ điển mặc định; file nhúng sai là lỗi lập trình.
func mustDefaultPronunciationDict() *pronunciationDict {
	d, err := loadPronunciationDict("")
	if err != nil {
		panic(err)
	}
	return d
}

// has báo viết tắt đã có trong từ điển chưa.
func (d *pronunciationDict) has(key string) bool {
	if d == nil {
		return false
	}
	_, ok := d.entries[key]
	return ok
}

// keys trả danh sách viết tắt đã sắp xếp (phục vụ test / in thông tin).
func (d *pronunciationDict) keys() []string {
	out := make([]string, 0, len(d.entries))
	for k := range d.entries {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// apply thay viết tắt bằng cách đọc, xét từng dòng. Dòng viết hoa toàn bộ (tiêu
// đề in hoa) được bỏ qua để không đổi nhầm chữ tiếng Việt viết hoa (vd "AI" là
// đại từ "ai" trong câu in hoa).
func (d *pronunciationDict) apply(s string) string {
	if d == nil || len(d.entries) == 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if isShoutedLine(ln) {
			continue
		}
		lines[i] = pronunTokenRe.ReplaceAllStringFunc(ln, func(tok string) string {
			if v, ok := d.entries[tok]; ok {
				return v
			}
			return tok
		})
	}
	return strings.Join(lines, "\n")
}

// isShoutedLine: dòng có ≥3 từ và ≥80% chữ cái là chữ hoa (tiêu đề in hoa).
func isShoutedLine(s string) bool {
	if len(strings.Fields(s)) < 3 {
		return false
	}
	letters, upper := 0, 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			letters++
			if unicode.IsUpper(r) {
				upper++
			}
		}
	}
	return letters > 0 && upper*5 >= letters*4
}
