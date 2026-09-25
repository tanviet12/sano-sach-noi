package bookmaker

import (
	"regexp"
	"strconv"
	"strings"
)

// chunkMaxChars — giới hạn ký tự khi chia chunk bằng chunkText. VieNeu-TTS v3
// Turbo tự chia câu (tối đa 256 ký tự/chunk) nên pipeline gửi nguyên văn bản;
// hằng số giữ cho chunkText (dùng trong test).
const chunkMaxChars = 180

// abbreviations — viết tắt → dạng đọc đầy đủ (tiếng Việt có dấu).
// Sắp khóa dài trước để khớp ưu tiên (TP.HCM trước TP.).
var abbreviations = []struct{ from, to string }{
	{"TP.HCM", "Thành phố Hồ Chí Minh"},
	{"TP. HCM", "Thành phố Hồ Chí Minh"},
	{"Tp.HCM", "Thành phố Hồ Chí Minh"},
	{"TPHCM", "Thành phố Hồ Chí Minh"},
	{"PGS.", "Phó giáo sư "},
	{"GS.", "Giáo sư "},
	{"ThS.", "Thạc sĩ "},
	{"TS.", "Tiến sĩ "},
	{"TP.", "Thành phố "},
	{"vd.", "ví dụ"},
	{"VD.", "ví dụ"},
	{"v.v.", "vân vân"},
	{"vv.", "vân vân"},
	{"tr.", "trang "},
}

// romanWords — số La Mã (1..20) → từ tiếng Việt (đọc theo số đếm).
var romanWords = []string{
	"", "một", "hai", "ba", "bốn", "năm", "sáu", "bảy", "tám", "chín", "mười",
	"mười một", "mười hai", "mười ba", "mười bốn", "mười lăm",
	"mười sáu", "mười bảy", "mười tám", "mười chín", "hai mươi",
}

// headingRomanRe khớp "Chương/Phần/Mục/Bài/Chapter <La Mã>" (không phân biệt hoa thường).
// QUAN TRỌNG: KHÔNG dùng \b sau số La Mã — RE2 \b chỉ tính ranh giới ASCII nên
// chỗ chuyển ASCII→ký tự Việt (vd "i" trong "viết" → "ế") bị coi là ranh giới,
// khiến "Bài viết" khớp nhầm "Bài VI" (6) → "Bài sáuết". Thay bằng yêu cầu số La
// Mã PHẢI theo sau bởi khoảng trắng / dấu câu / hết chuỗi (không phải chữ cái).
var headingRomanRe = regexp.MustCompile(`(?i)\b(Chương|Phần|Mục|Bài|Chapter|Phụ lục)\s+([IVXLCDM]+)([\s.,:;)\]!?]|$)`)

// pageNumLineRe khớp dòng chỉ chứa số trang ("12", "- 12 -", "Trang 12").
var pageNumLineRe = regexp.MustCompile(`(?i)^\s*(trang\s+)?[-–—]?\s*\d{1,4}\s*[-–—]?\s*$`)

// controlCharRe khớp ký tự điều khiển (trừ \t \n) + ký tự vô hình hay lẫn khi
// chép từ web / Word (BOM, zero-width space/joiner).
var controlCharRe = regexp.MustCompile("[\x00-\x08\x0b\x0c\x0e-\x1f\x7f\u200B-\u200D\u2060\uFEFF]")

// gradeMinusRe / gradePlusRe khớp ký hiệu xếp loại: 1 chữ HOA đứng riêng dính
// liền dấu -/+ rồi tới dấu cách / dấu câu / mũi tên / hết dòng (vd "A-", "B+").
// Chỉ bắt khi chữ hoa DÍNH LIỀN dấu → KHÔNG đụng gạch nối " - " (vd "khách quen
// - khách mới" giữ nguyên là nhịp nghỉ, không đọc "trừ").
var gradeMinusRe = regexp.MustCompile(`\b([A-Z])-([\s,.;:!?)\]→]|$)`)
var gradePlusRe = regexp.MustCompile(`\b([A-Z])\+([\s,.;:!?)\]→]|$)`)

// marketingPRe khớp mô hình marketing "<số>P" (4P, 7P, 3P...): 1-2 chữ số dính
// liền "P" HOA, rồi tới ranh giới không phải chữ cái. VieNeu-TTS đọc "4P" thành
// "bốn phút" (tiếng Việt "15p" = phút) → ép đọc "bốn Pê". CHỈ bắt P HOA: "4p"
// thường giữ nguyên (vd "nghỉ 4p" = 4 phút thật). KHÔNG dùng \b (RE2 \b chỉ tính
// ranh giới ASCII → sai với chữ Việt theo sau, xem headingRomanRe). Nhóm 1 giữ ký
// tự ranh giới trước số (đầu chuỗi / không phải chữ-số), nhóm 3 giữ ranh giới sau.
var marketingPRe = regexp.MustCompile(`(^|[^\p{L}\d])(\d{1,2})Ps?([^\p{L}]|$)`)

// loneLetterPRe khớp chữ "P" HOA đứng riêng ("chữ P đầu tiên", "P thứ hai là
// giá"): trước/sau không phải chữ cái, chữ số hay "&" (P&L đã có trong từ điển).
// VieNeu đánh vần chữ P thành "phê" → ép "Pê" như cách đọc tên chữ cái.
var loneLetterPRe = regexp.MustCompile(`(^|[^\p{L}\d&])P([^\p{L}\d&]|$)`)

// quoteStripper bỏ dấu ngoặc kép khỏi kịch bản đọc — VieNeu đọc dấu " thành
// "dấu ngoặc kép". Word thường tự đổi " thẳng thành nháy cong “ ”. Bỏ ký tự
// ngoặc, GIỮ nội dung bên trong (vd “phải chọn” → phải chọn).
var quoteStripper = strings.NewReplacer(
	"\"", "", // " thẳng (U+0022)
	"“", "", // " trái (U+201C)
	"”", "", // " phải (U+201D)
	"„", "", // „ (U+201E)
)

// Normalizer — bộ luật chuẩn hóa lời đọc (không dùng AI, chạy offline).
// Giữ cấu hình người dùng chọn qua cờ: từ điển cách đọc, giữ/bỏ số đầu tiêu đề.
type Normalizer struct {
	dict               *pronunciationDict // viết tắt → cách đọc (mặc định nhúng sẵn + --pronunciations)
	keepHeadingNumbers bool               // true = đọc số đầu tiêu đề ("một chấm hai"); false = bỏ
}

// defaultNormalizer — cấu hình mặc định (từ điển nhúng sẵn), dùng khi không
// truyền cấu hình riêng (test, intro).
var defaultNormalizer = &Normalizer{dict: mustDefaultPronunciationDict()}

// normalizeReadingScript chuẩn hóa bằng cấu hình mặc định (xem Normalizer.script).
func normalizeReadingScript(original string) string {
	return defaultNormalizer.script(original)
}

// script chuẩn hóa văn bản gốc thành kịch bản đọc:
//   - loại ký tự điều khiển + dòng số trang
//   - dòng bắt đầu bằng số mục nhiều cấp ("4.1.2. Tên") → bỏ số / đọc số theo --heading-numbers
//   - dòng danh sách: bỏ ký hiệu "-", "•", "a)"; "1." / "1)" → "Thứ nhất, ..."
//   - bỏ dấu ngoặc kép (VieNeu đọc thành "dấu ngoặc kép")
//   - mở rộng viết tắt (vd. → ví dụ, TP.HCM → Thành phố Hồ Chí Minh)
//   - đọc viết tắt theo từ điển cách đọc (vd "KPI" → "ca pê i")
//   - văn nói: "&" → "và", mũi tên → "dẫn tới", "/" giữa chữ theo ngữ cảnh,
//     khoảng số nhỏ "2-3" → "2 đến 3" (xem spoken.go)
//   - số La Mã trong tiêu đề chương → chữ ("Chương I" → "Chương một")
//   - ký hiệu xếp loại A-/B+ → "A trừ"/"B cộng"
//   - gộp khoảng trắng thừa nhưng giữ xuống dòng giữa các đoạn
//
// KHÔNG đổi nội dung chính của sách (PRD §5) — original_text giữ riêng.
func (n *Normalizer) script(original string) string {
	s := controlCharRe.ReplaceAllString(original, "")
	s = quoteStripper.Replace(s)
	s = dropPageNumberLines(s)
	s = n.sectionNumberLines(s)
	s = listMarkerLines(s)
	s = expandAbbreviations(s)
	s = n.dict.apply(s)
	s = expandAmpersands(s)
	s = expandArrows(s)
	s = expandSlashes(s)
	s = expandSmallRanges(s)
	s = expandHeadingRomans(s)
	s = expandGradeSigns(s)
	s = expandMarketingP(s)
	s = expandLoneLetterP(s)
	s = strings.TrimSpace(collapseSpacesKeepLines(s))
	return s
}

// normalizeTitle chuẩn hóa tiêu đề để HIỂN THỊ (metadata chương/tiểu mục): chỉ
// dọn ký tự, bỏ ngoặc kép, mở rộng viết tắt dạng có dấu chấm, số La Mã, ký hiệu
// xếp loại. KHÔNG áp từ điển cách đọc và luật văn nói — "SWOT" trên màn hình
// vẫn là "SWOT", không thành "suốt".
func normalizeTitle(title string) string {
	s := controlCharRe.ReplaceAllString(title, "")
	s = quoteStripper.Replace(s)
	s = expandAbbreviations(s)
	s = expandHeadingRomans(s)
	s = expandGradeSigns(s)
	s = expandMarketingP(s)
	return strings.TrimSpace(collapseSpacesKeepLines(s))
}

// expandGradeSigns đọc ký hiệu xếp loại "A-" → "A trừ", "B+" → "B cộng".
// Chỉ tác động chữ HOA đứng riêng dính liền dấu (xem gradeMinusRe), giữ nguyên
// gạch nối " - " trong cụm từ ghép.
func expandGradeSigns(s string) string {
	s = gradeMinusRe.ReplaceAllString(s, "$1 trừ$2")
	s = gradePlusRe.ReplaceAllString(s, "$1 cộng$2")
	return s
}

// expandMarketingP đổi mô hình "<số>P" → "<số chữ> Pê" (vd "4P" → "bốn Pê") để
// VieNeu không đọc nhầm thành "phút". Chỉ tác động P HOA dính liền số (xem
// marketingPRe). Dùng intToViet đọc số nguyên thành chữ tiếng Việt.
func expandMarketingP(s string) string {
	return marketingPRe.ReplaceAllStringFunc(s, func(m string) string {
		parts := marketingPRe.FindStringSubmatch(m)
		if len(parts) != 4 {
			return m
		}
		n, _ := strconv.Atoi(parts[2])
		return parts[1] + intToViet(n) + " Pê" + parts[3]
	})
}

// expandLoneLetterP đổi chữ "P" HOA đứng riêng thành "Pê" (xem loneLetterPRe).
// Chạy 2 lượt vì ký tự ranh giới bị ăn khi hai chữ P đứng sát ("P P").
func expandLoneLetterP(s string) string {
	for i := 0; i < 2; i++ {
		s = loneLetterPRe.ReplaceAllString(s, "${1}Pê${2}")
	}
	return s
}

// collapseSpacesKeepLines gộp khoảng trắng/tab liền nhau TRONG mỗi dòng thành 1
// space, và GIỮ ranh giới đoạn dạng DÒNG TRẮNG (\n\n): gộp ≥2 dòng trắng liên
// tiếp thành đúng 1, bỏ dòng trắng đầu/cuối. Dòng trắng giữa các đoạn để VieNeu
// (split theo đoạn) chèn nghỉ tương xứng sau tiêu đề + giữa các đoạn; gộp hết
// thành 1 dòng khiến TTS đọc liền tù tì.
func collapseSpacesKeepLines(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	prevBlank := true // coi đầu chuỗi như dòng trắng → bỏ dòng trắng dẫn đầu
	for _, ln := range lines {
		ln = strings.Join(strings.Fields(ln), " ")
		if ln == "" {
			if prevBlank {
				continue // gộp nhiều dòng trắng liên tiếp thành 1
			}
			out = append(out, "")
			prevBlank = true
			continue
		}
		out = append(out, ln)
		prevBlank = false
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1] // bỏ dòng trắng cuối
	}
	return strings.Join(out, "\n")
}

// dropPageNumberLines bỏ các dòng chỉ chứa số trang.
func dropPageNumberLines(s string) string {
	lines := strings.Split(s, "\n")
	out := lines[:0]
	for _, ln := range lines {
		if pageNumLineRe.MatchString(ln) {
			continue
		}
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}

// abbrevRes — regex cho từng viết tắt, CHỈ khớp khi đứng đầu từ (đầu chuỗi hoặc
// sau ký tự KHÔNG phải chữ cái). Tránh nuốt nhầm khi viết tắt nằm trong từ khác:
// vd "STP." chứa "TP." không được biến thành "SThành phố". Nhóm 1 giữ lại ký tự
// ranh giới phía trước.
var abbrevRes = buildAbbrevRes()

func buildAbbrevRes() []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(abbreviations))
	for i, a := range abbreviations {
		out[i] = regexp.MustCompile(`(^|[^\p{L}])` + regexp.QuoteMeta(a.from))
	}
	return out
}

// expandAbbreviations thay viết tắt bằng dạng đọc đầy đủ (chỉ ở đầu từ).
func expandAbbreviations(s string) string {
	for i, a := range abbreviations {
		s = abbrevRes[i].ReplaceAllString(s, "${1}"+a.to)
	}
	return s
}

// expandHeadingRomans đổi số La Mã sau từ khóa chương thành chữ tiếng Việt.
func expandHeadingRomans(s string) string {
	return headingRomanRe.ReplaceAllStringFunc(s, func(m string) string {
		parts := headingRomanRe.FindStringSubmatch(m)
		if len(parts) != 4 {
			return m
		}
		n := romanToInt(parts[2])
		if n <= 0 || n >= len(romanWords) {
			return m // ngoài phạm vi map → giữ nguyên
		}
		return parts[1] + " " + romanWords[n] + parts[3] // giữ lại dấu phân tách sau số
	})
}

// romanToInt chuyển số La Mã → int (0 nếu không hợp lệ).
func romanToInt(s string) int {
	vals := map[byte]int{'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100, 'D': 500, 'M': 1000}
	s = strings.ToUpper(s)
	total, prev := 0, 0
	for i := len(s) - 1; i >= 0; i-- {
		v, ok := vals[s[i]]
		if !ok {
			return 0
		}
		if v < prev {
			total -= v
		} else {
			total += v
			prev = v
		}
	}
	return total
}

// leadingSectionNumRe khớp tiền tố số mục ở đầu tiêu đề: "1.", "1.6.",
// "7.5.2.1." — nhóm số (có thể nhiều cấp ngăn bởi dấu chấm) + dấu chấm tùy chọn.
var leadingSectionNumRe = regexp.MustCompile(`^\s*(\d+(?:\.\d+)*)\.?\s*`)

var vietUnits = []string{"không", "một", "hai", "ba", "bốn", "năm", "sáu", "bảy", "tám", "chín"}

// intToViet đọc số nguyên 0..999 thành chữ tiếng Việt (đủ cho số mục/chương).
func intToViet(n int) string {
	switch {
	case n < 0 || n > 999:
		return strconv.Itoa(n) // ngoài phạm vi → giữ số
	case n < 10:
		return vietUnits[n]
	case n < 100:
		chuc, donvi := n/10, n%10
		s := "mười"
		if chuc > 1 {
			s = vietUnits[chuc] + " mươi"
		}
		switch {
		case donvi == 0:
			return s
		case donvi == 1 && chuc > 1:
			return s + " mốt"
		case donvi == 5:
			return s + " lăm"
		default:
			return s + " " + vietUnits[donvi]
		}
	default:
		tram, rem := n/100, n%100
		s := vietUnits[tram] + " trăm"
		switch {
		case rem == 0:
			return s
		case rem < 10:
			return s + " lẻ " + vietUnits[rem]
		default:
			return s + " " + intToViet(rem)
		}
	}
}

// expandLeadingSectionNumber đổi tiền tố số mục đầu tiêu đề thành chữ đọc, các
// cấp nối bằng "chấm": "1.6. Lập kế hoạch" → "một chấm sáu. Lập kế hoạch". Phần chữ sau
// giữ nguyên; tiêu đề không bắt đầu bằng số → trả nguyên (đã trim).
func expandLeadingSectionNumber(title string) string {
	m := leadingSectionNumRe.FindStringSubmatch(title)
	if m == nil {
		return strings.TrimSpace(title)
	}
	nums := strings.Split(m[1], ".")
	parts := make([]string, 0, len(nums))
	for _, ns := range nums {
		n, _ := strconv.Atoi(ns)
		parts = append(parts, intToViet(n))
	}
	spoken := strings.Join(parts, " chấm ")
	rest := strings.TrimSpace(title[len(m[0]):])
	if rest == "" {
		return spoken
	}
	return spoken + ". " + rest
}

// spokenSectionTitle dựng tiêu đề mục để đọc theo cấu hình mặc định (bỏ số mục).
func spokenSectionTitle(title string) string {
	return defaultNormalizer.spokenTitle(title)
}

// spokenTitle dựng tiêu đề mục để đọc: bỏ số mục đầu (mặc định) hoặc đọc số
// thành chữ (--heading-numbers keep), đổi ":" và gạch nối giữa cụm thành dấu
// phẩy cho TTS ngắt nghỉ tự nhiên.
func (n *Normalizer) spokenTitle(title string) string {
	t := dropLeadingSectionNumber(title)
	if n.keepHeadingNumbers {
		t = expandLeadingSectionNumber(title)
	}
	t = strings.ReplaceAll(t, ":", ",")
	t = strings.ReplaceAll(t, " - ", ", ")
	t = strings.ReplaceAll(t, " – ", ", ")
	t = strings.ReplaceAll(t, " — ", ", ")
	return t
}

// chunkText chia văn bản thành các chunk <= maxChars, ưu tiên ranh giới câu,
// sau đó tới ranh giới từ. Trả ít nhất 1 chunk khi input không rỗng.
func chunkText(s string, maxChars int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	sentences := splitSentences(s)
	var chunks []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(cur.String()))
			cur.Reset()
		}
	}
	for _, sent := range sentences {
		if len(sent) > maxChars {
			flush()
			chunks = append(chunks, splitByWords(sent, maxChars)...)
			continue
		}
		if cur.Len()+len(sent)+1 > maxChars {
			flush()
		}
		if cur.Len() > 0 {
			cur.WriteByte(' ')
		}
		cur.WriteString(sent)
	}
	flush()
	return chunks
}

// splitSentences tách câu theo dấu kết câu, giữ dấu.
func splitSentences(s string) []string {
	var out []string
	var cur strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		cur.WriteRune(r)
		if r == '.' || r == '!' || r == '?' {
			// Kết câu khi theo sau là khoảng trắng hoặc hết chuỗi.
			if i+1 >= len(runes) || runes[i+1] == ' ' || runes[i+1] == '\n' {
				if t := strings.TrimSpace(cur.String()); t != "" {
					out = append(out, t)
				}
				cur.Reset()
			}
		}
	}
	if t := strings.TrimSpace(cur.String()); t != "" {
		out = append(out, t)
	}
	return out
}

// splitByWords chia 1 câu dài thành các mẩu <= maxChars theo ranh giới từ.
func splitByWords(s string, maxChars int) []string {
	words := strings.Fields(s)
	var out []string
	var cur strings.Builder
	for _, w := range words {
		if cur.Len()+len(w)+1 > maxChars && cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
		if cur.Len() > 0 {
			cur.WriteByte(' ')
		}
		cur.WriteString(w)
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
