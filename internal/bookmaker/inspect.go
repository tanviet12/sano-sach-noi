package bookmaker

import (
	"strings"
)

// InspectOptions — cấu hình khi nạp file để xem trước.
type InspectOptions struct {
	KeepHeadingNumbers bool   // đọc số đầu tiêu đề ("một chấm hai") thay vì bỏ
	PronunciationFile  string // TSV bổ sung từ điển cách đọc; trống = mặc định
}

// Outline — mục lục + cảnh báo của một file Word, chưa render gì.
type Outline struct {
	Title          string           `json:"title"`     // tiêu đề trong file (style Title); trống nếu không có
	FileTitle      string           `json:"fileTitle"` // tiêu đề suy từ tên file
	Chapters       []OutlineChapter `json:"chapters"`
	Sections       int              `json:"sections"`
	Chars          int              `json:"chars"` // tổng ký tự lời đọc (không tính mục gợi ý bỏ)
	Warnings       LoadWarnings     `json:"warnings"`
	SampleSentence string           `json:"sampleSentence"` // câu đầu sách, dùng nghe mẫu giọng
}

// OutlineChapter — một chương trong mục lục xem trước.
type OutlineChapter struct {
	Title    string           `json:"title"`
	Sections []OutlineSection `json:"sections"`
}

// OutlineSection — một tiểu mục. Stem là mã gốc dùng cho Options.DropStems,
// Options.ReadingEdits và Preview.
type OutlineSection struct {
	Stem      string `json:"stem"`
	Title     string `json:"title"`
	Chars     int    `json:"chars"`
	Images    int    `json:"images"`
	TOC       bool   `json:"toc"`       // trông như trang mục lục → gợi ý bỏ
	TOCReason string `json:"tocReason"` // "tiêu đề" | "dòng kết thúc bằng số trang"
}

// LoadWarnings — những gì bản đọc không truyền tải được hoặc nên sửa trước khi render.
type LoadWarnings struct {
	Images          int            `json:"images"`
	SkippedImages   int            `json:"skippedImages"` // hình bỏ qua vì quá lớn / vượt số lần chèn
	Tables          int            `json:"tables"`
	FakeHeadings    []string       `json:"fakeHeadings"`
	UnknownAcronyms []AcronymCount `json:"unknownAcronyms"`
}

// AcronymCount — từ viết hoa liền chưa có trong từ điển cách đọc + số lần gặp.
type AcronymCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

// sampleSentenceMax — độ dài tối đa câu nghe mẫu giọng.
const sampleSentenceMax = 200

// Inspect nạp file .docx, trả mục lục (có đánh dấu trang mục lục gợi ý bỏ),
// số ký tự lời đọc từng tiểu mục và cảnh báo lúc nạp. Không ghi file nào.
func Inspect(path string, opt InspectOptions) (*Outline, error) {
	book, err := ParseDocx(path)
	if err != nil {
		return nil, err
	}
	norm, err := NewNormalizer(opt.PronunciationFile, opt.KeepHeadingNumbers)
	if err != nil {
		return nil, err
	}
	out := &Outline{
		Title:     strings.TrimSpace(book.Title),
		FileTitle: titleFromDocxName(path),
		Chapters:  make([]OutlineChapter, 0, len(book.Chapters)),
	}

	tocCount := 0
	for _, ch := range book.Chapters {
		for _, sec := range ch.Sections {
			if tocReason(sec) != "" {
				tocCount++
			}
		}
		out.Sections += len(ch.Sections)
	}
	// Giống dropTOCSections: mọi tiểu mục đều trông như mục lục thì không gợi ý bỏ gì.
	markTOC := tocCount < out.Sections

	var readings []string
	for _, ch := range book.Chapters {
		oc := OutlineChapter{Title: normalizeTitle(ch.Title), Sections: make([]OutlineSection, 0, len(ch.Sections))}
		for _, sec := range ch.Sections {
			reading := sectionReading(norm, sec)
			item := OutlineSection{
				Stem:   sec.Stem,
				Title:  normalizeTitle(sec.Title),
				Chars:  len([]rune(reading)),
				Images: len(sec.Images),
			}
			if markTOC {
				item.TOCReason = tocReason(sec)
				item.TOC = item.TOCReason != ""
			}
			if !item.TOC {
				out.Chars += item.Chars
				readings = append(readings, reading)
				if out.SampleSentence == "" {
					out.SampleSentence = firstSentence(norm.script(strings.TrimSpace(sec.Text)))
				}
			}
			oc.Sections = append(oc.Sections, item)
		}
		out.Chapters = append(out.Chapters, oc)
	}

	out.Warnings = LoadWarnings{
		Images:        book.Stats.Images,
		SkippedImages: book.Stats.SkippedImages + book.Stats.DroppedImages,
		Tables:        book.Stats.Tables,
		FakeHeadings:  append([]string{}, book.Stats.FakeHeadings...),
	}
	for _, a := range unknownAcronyms(readings, norm.dict) {
		out.Warnings.UnknownAcronyms = append(out.Warnings.UnknownAcronyms, AcronymCount(a))
	}
	if out.Warnings.UnknownAcronyms == nil {
		out.Warnings.UnknownAcronyms = []AcronymCount{}
	}
	return out, nil
}

// tocReason: lý do tiểu mục trông như trang mục lục, "" nếu không phải.
func tocReason(sec Section) string {
	switch {
	case isTOCTitle(sec.Title):
		return "tiêu đề"
	case isTOCBody(sec.Text):
		return "dòng kết thúc bằng số trang"
	}
	return ""
}

// firstSentence lấy câu đầu (≤ sampleSentenceMax ký tự, cắt mềm) của lời đọc.
func firstSentence(reading string) string {
	for _, para := range strings.Split(reading, "\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		s := para
		if sents := splitSentences(para); len(sents) > 0 {
			s = sents[0]
		}
		return truncatePreview(s, sampleSentenceMax)
	}
	return ""
}
