package bookmaker

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Giá trị cờ --heading-numbers.
const (
	HeadingNumbersDrop = "drop"
	HeadingNumbersKeep = "keep"
)

// ParseHeadingNumbers đổi giá trị cờ --heading-numbers thành "giữ số hay không".
func ParseHeadingNumbers(v string) (keep bool, err error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", HeadingNumbersDrop:
		return false, nil
	case HeadingNumbersKeep:
		return true, nil
	default:
		return false, fmt.Errorf("--heading-numbers không hợp lệ %q (keep|drop)", v)
	}
}

// sectionNumPrefixRe khớp số mục đầu dòng: nhóm 1 = các cấp số, nhóm 2 = phần
// ngăn cách (có dấu chấm cuối hay chỉ khoảng trắng), nhóm 3 = chữ cái đầu tiên
// của phần tên. Việc có coi là số mục hay không do splitSectionNumber quyết.
var sectionNumPrefixRe = regexp.MustCompile(`^\s*(\d{1,3}(?:\.\d{1,3})*)(\.\s*|\s+)(\p{L})`)

// splitSectionNumber tách số mục đầu dòng, trả phần tên (từ chữ cái đầu) và
// ok=true. Chỉ coi là số mục khi có dấu chấm cuối ("1.", "1.2.") hoặc nhiều cấp
// rồi tới chữ HOA ("1.2 Tên"). Nhờ vậy "3.5 năm kinh nghiệm" hay "2026 là
// năm..." không bị coi là số mục.
func splitSectionNumber(s string) (rest string, ok bool) {
	m := sectionNumPrefixRe.FindStringSubmatchIndex(s)
	if m == nil {
		return "", false
	}
	nums, sep, first := s[m[2]:m[3]], s[m[4]:m[5]], []rune(s[m[6]:m[7]])[0]
	if !strings.Contains(sep, ".") && !(strings.Contains(nums, ".") && unicode.IsUpper(first)) {
		return "", false
	}
	return strings.TrimSpace(s[m[6]:]), true
}

// dropLeadingSectionNumber bỏ số mục đầu tiêu đề: "1.2.3. Tên" → "Tên".
// Tiêu đề không có số mục → trả nguyên (đã trim).
func dropLeadingSectionNumber(title string) string {
	if rest, ok := splitSectionNumber(title); ok {
		return rest
	}
	return strings.TrimSpace(title)
}

// multiLevelLineRe — dòng thân bài bắt đầu bằng số mục NHIỀU cấp (tiêu đề gõ tay
// không dùng style Heading, vd "4.1.2. Tên mục"). Số một cấp "1." đầu dòng là
// danh sách đánh số, xử lý ở luật danh sách.
var multiLevelLineRe = regexp.MustCompile(`^\s*\d{1,3}(?:\.\d{1,3})+(?:\.\s*|\s+)\p{L}`)

// sectionNumberLines áp chính sách --heading-numbers cho dòng thân bài bắt đầu
// bằng số mục nhiều cấp: bỏ số (mặc định) hoặc đọc số thành chữ.
func (n *Normalizer) sectionNumberLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if !multiLevelLineRe.MatchString(ln) {
			continue
		}
		rest, ok := splitSectionNumber(ln)
		switch {
		case !ok:
		case n.keepHeadingNumbers:
			lines[i] = expandLeadingSectionNumber(ln)
		default:
			lines[i] = rest
		}
	}
	return strings.Join(lines, "\n")
}
