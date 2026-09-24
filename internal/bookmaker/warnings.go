package bookmaker

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Ngưỡng nhận diện "tiêu đề gõ tay": đoạn ngắn, gần như toàn bộ in đậm hoặc chữ
// to hơn thân bài, nhưng không dùng style Heading nên không lên mục lục.
const (
	fakeHeadingMaxRunes  = 120
	fakeHeadingBoldRatio = 0.9
	fakeHeadingSzDelta   = 4 // nửa point: to hơn thân bài từ 2pt trở lên
	warnExamples         = 3 // số ví dụ in kèm mỗi cảnh báo
	warnAcronymsShown    = 20
)

// bodyFontSize — cỡ chữ phổ biến nhất của thân bài (tính theo số ký tự), 0 nếu
// tài liệu không ghi cỡ chữ theo run.
func bodyFontSize(paras []docxPara) int {
	weight := map[int]int{}
	for _, p := range paras {
		if p.maxSz > 0 && headingLevel(p.style) == 0 && !isTitleStyle(p.style) {
			weight[p.maxSz] += p.textRunes
		}
	}
	best, bestW := 0, 0
	for sz, w := range weight {
		if w > bestW || (w == bestW && sz < best) {
			best, bestW = sz, w
		}
	}
	return best
}

// looksLikeFakeHeading: đoạn thường (không style Heading) ngắn, không kết thúc
// như câu văn / nhãn ("... :"), và in đậm gần hết hoặc chữ to hơn thân bài.
func looksLikeFakeHeading(p docxPara, text string, bodySz int) bool {
	if headingLevel(p.style) > 0 || isTitleStyle(p.style) || p.textRunes < 2 {
		return false
	}
	if len([]rune(text)) > fakeHeadingMaxRunes || strings.ContainsAny(lastRune(text), ".?!…;,:") {
		return false
	}
	bold := float64(p.boldRunes) >= fakeHeadingBoldRatio*float64(p.textRunes)
	big := bodySz > 0 && p.maxSz >= bodySz+fakeHeadingSzDelta
	return bold || big
}

func lastRune(s string) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) == 0 {
		return ""
	}
	return string(r[len(r)-1])
}

// acronymCount — 1 từ viết hoa liền chưa có trong từ điển + số lần gặp.
type acronymCount struct {
	Word  string
	Count int
}

// romanOnlyRe — số La Mã thuần (II, XIV...): bộ đọc tự đọc, không tính.
var romanOnlyRe = regexp.MustCompile(`^[IVXLCDM]+$`)

// isAcronymCandidate: token chỉ gồm chữ Latin hoa không dấu / số / &, có ≥2 chữ hoa.
func isAcronymCandidate(tok string) bool {
	upper := 0
	for _, r := range tok {
		switch {
		case r >= 'A' && r <= 'Z':
			upper++
		case unicode.IsDigit(r), r == '&':
		default:
			return false
		}
	}
	return upper >= 2 && !romanOnlyRe.MatchString(tok)
}

// unknownAcronyms đếm từ viết hoa liền trong lời đọc (sau khi đã áp từ điển) mà
// từ điển chưa có. Bỏ qua dòng in hoa toàn bộ (tiêu đề in hoa, không phải viết tắt).
func unknownAcronyms(texts []string, dict *pronunciationDict) []acronymCount {
	counts := map[string]int{}
	for _, t := range texts {
		for _, ln := range strings.Split(t, "\n") {
			if isShoutedLine(ln) {
				continue
			}
			for _, tok := range pronunTokenRe.FindAllString(ln, -1) {
				if isAcronymCandidate(tok) && !dict.has(tok) {
					counts[tok]++
				}
			}
		}
	}
	out := make([]acronymCount, 0, len(counts))
	for w, c := range counts {
		out = append(out, acronymCount{Word: w, Count: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Word < out[j].Word
	})
	return out
}

// printLoadWarnings in cảnh báo lúc nạp: những gì bản đọc không truyền tải được
// hoặc người dùng nên sửa trong file Word / từ điển trước khi render.
func printLoadWarnings(w io.Writer, st DocStats, unknown []acronymCount) {
	fmt.Fprintln(w, "\nCảnh báo lúc nạp:")
	fmt.Fprintf(w, "  - %d hình: chưa có mô tả, người nghe sẽ không biết nội dung hình.\n", st.Images)
	if st.SkippedImages > 0 {
		fmt.Fprintf(w, "  - %d hình quá lớn sau khi giải nén: đã bỏ qua, không trích ra.\n", st.SkippedImages)
	}
	if st.DroppedImages > 0 {
		fmt.Fprintf(w, "  - %d lần chèn hình vượt giới hạn %d lần mỗi tài liệu: đã bỏ qua.\n", st.DroppedImages, maxImageRefs)
	}
	fmt.Fprintf(w, "  - %d bảng: nội dung bảng được đọc phẳng từng ô, mất hàng/cột.\n", st.Tables)
	fmt.Fprintf(w, "  - %d đoạn in đậm hoặc chữ to nhưng không dùng style Heading (có thể là tiêu đề gõ tay, không lên mục lục)", len(st.FakeHeadings))
	if n := min(len(st.FakeHeadings), warnExamples); n > 0 {
		fmt.Fprintf(w, ", ví dụ: «%s»", strings.Join(st.FakeHeadings[:n], "», «"))
	}
	fmt.Fprintln(w, ".")
	total := 0
	for _, a := range unknown {
		total += a.Count
	}
	fmt.Fprintf(w, "  - %d từ viết hoa liền chưa có trong từ điển cách đọc (%d lần)", len(unknown), total)
	if len(unknown) > 0 {
		shown := make([]string, 0, warnAcronymsShown)
		for i, a := range unknown {
			if i == warnAcronymsShown {
				shown = append(shown, "...")
				break
			}
			shown = append(shown, fmt.Sprintf("%s ×%d", a.Word, a.Count))
		}
		fmt.Fprintf(w, ": %s. Bổ sung cách đọc bằng --pronunciations <file.tsv> nếu bộ đọc đọc sai", strings.Join(shown, ", "))
	}
	fmt.Fprintln(w, ".")
}
