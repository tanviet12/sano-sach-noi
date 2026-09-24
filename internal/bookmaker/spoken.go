package bookmaker

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Luật "văn nói": đổi văn viết để đọc bằng mắt thành lời đọc trôi chảy. Chỉ
// thêm luật cho chỗ VieNeu-TTS v3 đọc sai hoặc thiếu; những gì bộ chuẩn hóa của
// v3 đã tự làm đúng (ngoặc đơn → ngắt nhịp, "..." → nghỉ, dấu " - " → dấu
// phẩy, khoảng trắng trước dấu câu, %, ngày tháng, giờ, tiền, số thập phân)
// thì để v3 làm, không viết trùng.

// bulletLineRe — dòng danh sách dấu đầu dòng ("- ý", "• ý", "+ ý"...).
var bulletLineRe = regexp.MustCompile(`^\s*[-–—•●▪◦○∙·+*]\s+(\S.*)$`)

// numberedLineRe — dòng danh sách đánh số một cấp ("1. ý", "2) ý").
var numberedLineRe = regexp.MustCompile(`^\s*(\d{1,2})[.)]\s+(\p{L}.*)$`)

// letteredLineRe — dòng danh sách đánh chữ ("a) ý", "B) ý").
var letteredLineRe = regexp.MustCompile(`^\s*[a-zA-ZđĐ]\)\s+(\S.*)$`)

// ordinalStartRe — ý đã tự mở đầu bằng thứ tự ("Thứ nhất", "Đầu tiên"...).
var ordinalStartRe = regexp.MustCompile(`(?i)^(thứ\s|đầu tiên|trước hết|cuối cùng|sau cùng)`)

// vietOrdinal đọc số thứ tự: 1 → "nhất", 4 → "tư", còn lại đọc như số đếm.
func vietOrdinal(n int) string {
	switch n {
	case 1:
		return "nhất"
	case 4:
		return "tư"
	default:
		return intToViet(n)
	}
}

// listMarkerLines bỏ ký hiệu đầu dòng danh sách, mỗi ý vẫn một dòng. Danh sách
// đánh số đọc thành "Thứ nhất, ...", "Thứ hai, ..." (thứ tự có nghĩa khi nghe);
// ý đã tự có "Thứ nhất..." thì chỉ bỏ số. Danh sách dấu gạch / chấm tròn / chữ
// cái chỉ bỏ ký hiệu.
func listMarkerLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if m := numberedLineRe.FindStringSubmatch(ln); m != nil {
			if ordinalStartRe.MatchString(m[2]) {
				lines[i] = m[2]
				continue
			}
			n, _ := strconv.Atoi(m[1])
			lines[i] = "Thứ " + vietOrdinal(n) + ", " + m[2]
			continue
		}
		if m := bulletLineRe.FindStringSubmatch(ln); m != nil {
			lines[i] = m[1]
			continue
		}
		if m := letteredLineRe.FindStringSubmatch(ln); m != nil {
			lines[i] = m[1]
		}
	}
	return strings.Join(lines, "\n")
}

// ampersandRe — dấu & giữa hai chữ/số (sau khi từ điển đã đọc R&D, P&L...).
var ampersandRe = regexp.MustCompile(`([\p{L}\p{N}])[ \t]*&[ \t]*([\p{L}\p{N}])`)

// expandAmpersands đọc "&" là "và": "XY&Z" → "XY và Z". v3 có lúc đọc "&" trong cụm
// viết tắt thành tiếng Anh "and", nên xử lý trước.
func expandAmpersands(s string) string {
	for i := 0; i < 3; i++ { // "A&B&C": các khớp dính nhau cần vài lượt
		next := ampersandRe.ReplaceAllString(s, "$1 và $2")
		if next == s {
			break
		}
		s = next
	}
	return s
}

// arrowRe — mũi tên trong văn bản (→ ⇒ => ⟶ ➔ ➜ và " -> ").
var arrowRe = regexp.MustCompile(`[ \t]*(?:→|⇒|⟶|➔|➜|=>|[ \t]->)[ \t]*`)

// arrowConnectives — từ nối đứng ngay sau mũi tên: mũi tên chỉ là dấu nối, bỏ
// đi thay vì đọc "dẫn tới" (tránh "dẫn tới dẫn đến", "dẫn tới vì...").
var arrowConnectives = []string{
	"dẫn đến", "dẫn tới", "từ đó", "do đó", "vì vậy", "vì", "mà", "và", "thì", "để",
	"nên", "nơi", "đồng thời", "tức là", "nghĩa là", "chính là", "là", "rồi", "sau đó",
}

func startsWithConnective(s string) bool {
	low := strings.ToLower(strings.TrimSpace(s))
	for _, c := range arrowConnectives {
		if low == c || strings.HasPrefix(low, c+" ") || strings.HasPrefix(low, c+",") {
			return true
		}
	}
	return false
}

// expandArrows đọc mũi tên theo vị trí (v3 đọc → là "đến", => là "sang", ⇒ là
// "suy ra"):
//   - giữa câu: ", dẫn tới" ("mưa lớn → đường ngập" → "mưa lớn, dẫn tới đường ngập")
//   - đầu dòng, sau dấu kết câu / dấu hai chấm: mũi tên là ký hiệu đầu ý hoặc
//     trả lời ("Ai làm? → Trưởng nhóm") → bỏ mũi tên, giữ nhịp nghỉ sẵn có
//   - ngay trước từ nối ("→ dẫn đến", "→ vì") → bỏ mũi tên, thay bằng dấu phẩy
func expandArrows(s string) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		locs := arrowRe.FindAllStringIndex(ln, -1)
		if locs == nil {
			continue
		}
		// Cắt đuôi bằng slice thay vì Reset + ghi lại cả dòng: dòng dài nhiều mũi
		// tên (file cố tình tạo) không làm thời gian tăng theo bình phương.
		b := make([]byte, 0, len(ln)+16*len(locs))
		prev := 0
		for j, loc := range locs {
			b = append(b, ln[prev:loc[0]]...)
			end := len(ln)
			if j+1 < len(locs) {
				end = locs[j+1][0]
			}
			after := ln[loc[1]:end]
			n := len(bytes.TrimRightFunc(b, unicode.IsSpace))
			last := byte(0)
			if n > 0 {
				last = b[n-1]
			}
			switch {
			case n == 0:
				b = b[:0]
			case strings.TrimSpace(after) == "": // mũi tên cuối dòng: không có vế sau
				b = b[:n]
			case strings.IndexByte(".!?:;", last) >= 0:
				b = append(b[:n], ' ')
			case startsWithConnective(after):
				b = append(bytes.TrimRight(b[:n], ",-"), ", "...)
			case strings.IndexByte(",-(", last) >= 0:
				b = append(bytes.TrimRight(b[:n], ",-"), ", dẫn tới "...)
			default:
				b = append(b, ", dẫn tới "...)
			}
			prev = loc[1]
		}
		b = append(b, ln[prev:]...)
		lines[i] = strings.TrimRight(string(b), " ,")
	}
	return strings.Join(lines, "\n")
}

// urlRe — địa chỉ web / email: không đụng dấu "/" bên trong.
var urlRe = regexp.MustCompile(`(?i)\b(?:https?://|www\.)\S+|\S+@\S+\.\S+`)

// slashRe — dấu "/" giữa hai từ hoặc số.
var slashRe = regexp.MustCompile(`([\p{L}\p{N}%]+)[ \t]*/[ \t]*([\p{L}\p{N}]+)`)

// measureSlashRe — "<số hoặc khoảng số> <đơn vị>/<đơn vị>": "2-3 giờ/ngày",
// "5 trang/ngày", "60 km/giờ". Vế trái là đơn vị đo (sau con số) nên "/" nghĩa
// là "mỗi"; "ngày/tháng/năm" không có số đứng trước nên không khớp.
var measureSlashRe = regexp.MustCompile(`(\d[\d.,]*%?(?:[ \t]*[-–][ \t]*\d[\d.,]*%?)?[ \t]+\p{L}+)[ \t]*/[ \t]*(\p{L}+)`)

// numericTokenRe — token toàn số (có thể có dấu chấm/phẩy): để v3 đọc ngày
// tháng, phân số ("12/09/2026", "1/3").
var numericTokenRe = regexp.MustCompile(`^[\d.,]+$`)

// perUnits — đơn vị sau "/" mang nghĩa "mỗi": "300 triệu/năm" → "300 triệu một năm".
var perUnits = map[string]bool{
	"năm": true, "tháng": true, "quý": true, "tuần": true, "ngày": true, "giờ": true,
	"phút": true, "giây": true, "người": true, "lần": true, "lượt": true, "khách": true,
	"đơn": true, "ca": true, "suất": true, "buổi": true, "kỳ": true,
}

// amountWords — từ chỉ lượng đứng trước "/<đơn vị>".
var amountWords = map[string]bool{
	"triệu": true, "tỷ": true, "tỉ": true, "nghìn": true, "ngàn": true, "trăm": true,
	"đồng": true, "đ": true, "k": true, "usd": true, "vnđ": true, "vnd": true, "đô": true,
}

// choicePairs — cặp lựa chọn đọc "hoặc" ("có/không" → "có hoặc không").
var choicePairs = map[string]bool{
	"có/không": true, "đúng/sai": true, "nam/nữ": true, "online/offline": true,
	"offline/online": true, "được/không": true, "thắng/thua": true, "mua/bán": true,
	"lời/lỗ": true, "lãi/lỗ": true, "tăng/giảm": true, "trước/sau": true,
	"trong/ngoài": true, "yes/no": true, "on/off": true,
}

// expandSlashes đọc dấu "/" theo ngữ cảnh (v3 đọc "/" giữa chữ là "trên"):
//   - số/số (ngày tháng, phân số): giữ để v3 đọc; riêng "24/7" → "24 7"
//   - lượng/đơn vị ("300 triệu/năm", "5%/tháng") → "một": "300 triệu một năm"
//   - số + đơn vị/đơn vị ("2-3 giờ/ngày") → "một": "2-3 giờ một ngày"
//   - cặp lựa chọn ngắn ("có/không", "A/B") → "hoặc"; "và/hoặc" → "và hoặc"
//   - còn lại (không chắc) → dấu phẩy: "sách/tạp chí" → "sách, tạp chí"
//
// Địa chỉ web / email giữ nguyên.
func expandSlashes(s string) string {
	return replaceOutside(s, urlRe, func(seg string) string {
		seg = measureSlashRe.ReplaceAllStringFunc(seg, func(m string) string {
			parts := measureSlashRe.FindStringSubmatch(m)
			if !perUnits[strings.ToLower(parts[2])] {
				return m
			}
			return parts[1] + " một " + parts[2]
		})
		for i := 0; i < 4; i++ { // "a/b/c": các khớp dính nhau cần vài lượt
			next := slashRe.ReplaceAllStringFunc(seg, readSlash)
			if next == seg {
				break
			}
			seg = next
		}
		return seg
	})
}

func readSlash(m string) string {
	parts := slashRe.FindStringSubmatch(m)
	left, right := parts[1], parts[2]
	lowL, lowR := strings.ToLower(left), strings.ToLower(right)
	switch {
	case left == "24" && right == "7":
		return "24 7"
	case numericTokenRe.MatchString(left) && numericTokenRe.MatchString(right):
		return m // ngày tháng / phân số: v3 tự đọc
	case perUnits[lowR] && (quantityTokenRe.MatchString(left) || amountWords[lowL]):
		return left + " một " + right
	case lowL == "và" && lowR == "hoặc":
		return left + " hoặc"
	case choicePairs[lowL+"/"+lowR] || (isSingleUpper(left) && isSingleUpper(right)):
		return left + " hoặc " + right
	default:
		return left + ", " + right
	}
}

// quantityTokenRe — số kèm đơn vị ngắn dính liền: "200", "15k", "7kg", "5%".
var quantityTokenRe = regexp.MustCompile(`^\d[\d.,]*[\p{L}%]{0,3}$`)

func isSingleUpper(s string) bool {
	r := []rune(s)
	return len(r) == 1 && unicode.IsUpper(r[0])
}

// replaceOutside áp fn cho các đoạn KHÔNG khớp re (giữ nguyên đoạn khớp).
func replaceOutside(s string, re *regexp.Regexp, fn func(string) string) string {
	locs := re.FindAllStringIndex(s, -1)
	if locs == nil {
		return fn(s)
	}
	var b strings.Builder
	prev := 0
	for _, loc := range locs {
		b.WriteString(fn(s[prev:loc[0]]))
		b.WriteString(s[loc[0]:loc[1]])
		prev = loc[1]
	}
	b.WriteString(fn(s[prev:]))
	return b.String()
}

// smallRangeRe — khoảng số nhỏ "2-3", "3 - 5" (mỗi vế 1-2 chữ số, không nằm
// trong ngày tháng / số điện thoại / số thập phân).
var smallRangeRe = regexp.MustCompile(`(^|[^\d.,/:-])(\d{1,2})[ \t]*[-–][ \t]*(\d{1,2})($|[^\d/.,:-]|[.,](?:\s|$))`)

// expandSmallRanges đọc khoảng số nhỏ thành "a đến b". v3 hiểu nhầm "4-6 tháng",
// "2-4 năm" là ngày tháng ("ngày bốn tháng sáu tháng"); các khoảng khác v3 đọc
// đúng "đến" nên đổi đồng loạt không làm hỏng chỗ nào.
func expandSmallRanges(s string) string {
	for i := 0; i < 2; i++ { // "2-3, 4-5": khớp dính nhau cần lượt 2
		next := smallRangeRe.ReplaceAllString(s, "${1}${2} đến ${3}${4}")
		if next == s {
			break
		}
		s = next
	}
	return s
}
