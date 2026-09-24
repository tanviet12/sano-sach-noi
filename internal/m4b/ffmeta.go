package m4b

import (
	"strconv"
	"strings"
)

// Chapter — một mốc mục lục trong file M4B (mili giây tính từ đầu sách).
type Chapter struct {
	Title   string
	StartMs int64
	EndMs   int64
}

// Meta — thông tin chung ghi vào file M4B.
type Meta struct {
	Title   string
	Author  string
	Comment string
}

// DefaultComment — dòng ghi chú mặc định trong file M4B.
const DefaultComment = "Tạo bằng Sano"

// genre + media_type=2 (thẻ iTunes "stik" = Audiobook): Apple Books và các app
// nghe nhận ra đây là sách nói.
const (
	genreAudiobook = "Audiobook"
	mediaTypeBook  = "2"
)

// escapeMeta thoát ký tự đặc biệt của định dạng ffmetadata: '=', ';', '#', '\'
// và xuống dòng đều phải có '\' đứng trước.
func escapeMeta(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\\', '=', ';', '#', '\n':
			b.WriteByte('\\')
		case '\r':
			continue // bỏ CR, xuống dòng chỉ giữ LF
		}
		b.WriteRune(r)
	}
	return b.String()
}

// FFMetadata dựng nội dung file ffmetadata: thẻ chung (tên sách, tác giả,
// thể loại, ghi chú) + mục lục chương, dùng làm đầu vào `-f ffmetadata` của
// ffmpeg (-map_metadata / -map_chapters).
func FFMetadata(m Meta, chapters []Chapter) string {
	var b strings.Builder
	b.WriteString(";FFMETADATA1\n")
	tag := func(k, v string) {
		if strings.TrimSpace(v) == "" {
			return
		}
		b.WriteString(k + "=" + escapeMeta(v) + "\n")
	}
	tag("title", m.Title)
	tag("album", m.Title)
	tag("artist", m.Author)
	tag("album_artist", m.Author)
	tag("genre", genreAudiobook)
	tag("media_type", mediaTypeBook)
	comment := m.Comment
	if comment == "" {
		comment = DefaultComment
	}
	tag("comment", comment)
	for _, c := range chapters {
		b.WriteString("\n[CHAPTER]\nTIMEBASE=1/1000\n")
		b.WriteString("START=" + strconv.FormatInt(c.StartMs, 10) + "\n")
		b.WriteString("END=" + strconv.FormatInt(c.EndMs, 10) + "\n")
		b.WriteString("title=" + escapeMeta(c.Title) + "\n")
	}
	return b.String()
}
