package bookmaker

import (
	"fmt"
	"regexp"
	"strings"
)

// tocTitles — tiêu đề (đã chuẩn hóa chữ thường) của trang mục lục. "nội dung"
// cũng là tên chương ngầm cho phần đầu tài liệu trước heading đầu tiên (trang
// bìa, trang tên sách) — phần này đọc lên không có ích nên bỏ luôn.
var tocTitles = map[string]bool{
	"mục lục":           true,
	"nội dung":          true,
	"table of contents": true,
	"contents":          true,
	"toc":               true,
}

// tocTitleTrimRe bỏ số mục đầu + dấu câu cuối trước khi so tên tiêu đề.
var tocTitleTrimRe = regexp.MustCompile(`^[\d.\s]+|[\s.:]+$`)

// tocLineRe — dòng mục lục: kết thúc bằng số trang (có thể sau dấu chấm dẫn "....").
var tocLineRe = regexp.MustCompile(`\S.*?[\s.…·_-]\d{1,4}\s*$`)

// Ngưỡng nhận diện mục lục theo nội dung: đủ số dòng và phần lớn dòng kết thúc
// bằng số trang.
const (
	tocMinLines = 4
	tocMinRatio = 0.6
)

// droppedTOC — 1 tiểu mục đã bỏ vì là trang mục lục (để in ra màn hình).
type droppedTOC struct {
	Chapter, Section string
	Lines            int
	Reason           string
}

// isTOCTitle báo tiêu đề có phải kiểu Mục lục / Nội dung / Table of Contents.
func isTOCTitle(title string) bool {
	t := strings.ToLower(strings.TrimSpace(tocTitleTrimRe.ReplaceAllString(title, "")))
	return tocTitles[t]
}

// isTOCBody báo phần lớn dòng của văn bản kết thúc bằng số trang.
func isTOCBody(text string) bool {
	total, hits := 0, 0
	for _, ln := range strings.Split(text, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		total++
		if tocLineRe.MatchString(ln) {
			hits++
		}
	}
	return total >= tocMinLines && float64(hits) >= tocMinRatio*float64(total)
}

// dropTOCSections bỏ các tiểu mục là trang mục lục khỏi book (chương rỗng sau
// khi bỏ cũng bị bỏ). Không bao giờ bỏ hết: nếu mọi tiểu mục đều trông như mục
// lục thì giữ nguyên để không mất cả sách.
func dropTOCSections(book *Book) []droppedTOC {
	var dropped []droppedTOC
	kept := make([]Chapter, 0, len(book.Chapters))
	remaining := 0
	for _, ch := range book.Chapters {
		secs := make([]Section, 0, len(ch.Sections))
		for _, sec := range ch.Sections {
			reason := tocReason(sec)
			if reason == "" {
				secs = append(secs, sec)
				continue
			}
			dropped = append(dropped, droppedTOC{
				Chapter: ch.Title,
				Section: sec.Title,
				Lines:   countNonEmptyLines(sec.Text),
				Reason:  reason,
			})
		}
		remaining += len(secs)
		if len(secs) == 0 {
			continue
		}
		ch.Sections = secs
		kept = append(kept, ch)
	}
	if remaining == 0 {
		return nil
	}
	book.Chapters = kept
	return dropped
}

func countNonEmptyLines(s string) int {
	n := 0
	for _, ln := range strings.Split(s, "\n") {
		if strings.TrimSpace(ln) != "" {
			n++
		}
	}
	return n
}

// formatDroppedTOC — dòng thông báo cho 1 tiểu mục đã bỏ.
func formatDroppedTOC(d droppedTOC) string {
	return fmt.Sprintf("  - «%s» (chương «%s», %d dòng, nhận ra theo %s)", d.Section, d.Chapter, d.Lines, d.Reason)
}
