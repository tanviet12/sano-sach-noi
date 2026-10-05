package bookmaker

// Từ điển cách đọc của người dùng (wireframe D12): chồng lên bộ chuẩn nhúng sẵn.
// Thứ tự ưu tiên tăng dần: bộ chuẩn → từ điển chung (mọi sách) → từ điển của cuốn.
// Chỉ đổi LỜI ĐỌC; chữ hiện khi nghe (original_text) giữ nguyên.

import (
	"sort"
	"strings"
)

// BookPronunciationsFile — từ điển riêng của cuốn trong thư mục sách; trong gói
// zip là ZipPronunciationsEntry (docs/book-zip-format.md).
const (
	BookPronunciationsFile = "tu-dien.tsv"
	ZipPronunciationsEntry = "pronunciations.tsv"
)

// NewNormalizerWith như NewNormalizer nhưng chồng thêm các lớp từ điển (lớp sau
// ghi đè lớp trước). Cách đọc rỗng = tắt từ đó.
func NewNormalizerWith(keepHeadingNumbers bool, layers ...map[string]string) (*Normalizer, error) {
	d, err := loadPronunciationDict("")
	if err != nil {
		return nil, err
	}
	d.overlay(layers...)
	return &Normalizer{dict: d, keepHeadingNumbers: keepHeadingNumbers}, nil
}

// overlay chồng các lớp từ điển người dùng lên d (lớp sau ghi đè; cách đọc rỗng = tắt).
func (d *pronunciationDict) overlay(layers ...map[string]string) {
	for _, l := range layers {
		for k, v := range l {
			if !ValidPronunciationKey(k) {
				continue
			}
			if v = strings.Join(strings.Fields(v), " "); v == "" {
				delete(d.entries, k)
			} else {
				d.entries[k] = v
			}
		}
	}
}

// ValidPronunciationKey: một "từ" tra được trong từ điển — chữ và số liền nhau,
// có thể nối bằng & (KPI, Nielsen, R&D, B2B). Không nhận cụm nhiều từ.
func ValidPronunciationKey(k string) bool {
	return pronunKeyRe.MatchString(k)
}

// DefaultPronunciations — bộ chuẩn nhúng sẵn (từ → cách đọc).
func DefaultPronunciations() map[string]string {
	m, err := parsePronunciations(defaultPronunciationsTSV, "pronunciations.default.tsv")
	if err != nil {
		panic(err) // file nhúng sai là lỗi lập trình
	}
	return m
}

// ParsePronunciations đọc TSV "<từ><TAB><cách đọc>" (dòng # là chú thích).
func ParsePronunciations(data, source string) (map[string]string, error) {
	return parsePronunciations(data, source)
}

// FormatPronunciations ghi từ điển thành TSV (sắp theo từ, có dòng chú thích đầu).
func FormatPronunciations(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("# Từ điển cách đọc của Sano: <từ trong sách><TAB><đọc là>\n")
	for _, k := range keys {
		b.WriteString(k + "\t" + strings.Join(strings.Fields(m[k]), " ") + "\n")
	}
	return b.String()
}

// CountWord đếm số lần word xuất hiện nguyên từ trong text (cùng cách tách từ
// của từ điển, phân biệt hoa thường).
func CountWord(text, word string) int {
	if word == "" {
		return 0
	}
	n := 0
	for _, tok := range pronunTokenRe.FindAllString(text, -1) {
		if tok == word {
			n++
		}
	}
	return n
}

// CountWordsInDocx đếm từng từ trong chữ của file Word / EPUB / PDF (mọi tiểu mục, cả tiêu đề).
func CountWordsInDocx(path string, words []string) (map[string]int, error) {
	book, err := ParseBook(path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]int, len(words))
	for _, w := range words {
		out[w] = 0
	}
	for _, ch := range book.Chapters {
		for _, sec := range ch.Sections {
			for _, w := range words {
				out[w] += CountWord(sec.Title, w) + CountWord(sec.Text, w)
			}
		}
	}
	return out, nil
}
