package m4b

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"sano/internal/safepath"
)

// Track — một file MP3 (một tiểu mục) trong sách, theo thứ tự nghe. Title là
// tên mốc mục lục trong file M4B.
type Track struct {
	Title string
	File  string
}

// Book — dữ liệu để xuất một file M4B.
type Book struct {
	Meta
	Cover  string // ảnh bìa (png/jpg/webp); "" hoặc không đọc được → tự vẽ bìa vuông
	Tracks []Track
}

// dirMeta — phần metadata.json (thư mục đầu ra của bookmaker) cần đọc.
type dirMeta struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Cover    string `json:"cover"`
	Chapters []struct {
		Title    string `json:"title"`
		Sections []struct {
			Title string `json:"title"`
			File  string `json:"file"`
		} `json:"sections"`
	} `json:"chapters"`
}

// maxMetadataBytes — metadata.json lớn hơn thì từ chối.
const maxMetadataBytes = 8 << 20

// FromDir đọc thư mục sách đã render (metadata.json + chNN-secNN.mp3 + bìa) —
// thư mục đầu ra của sano-docx2tts hoặc một cuốn trong thư viện phần mềm.
//
// Thư mục có thể là của người khác gửi: metadata.json, bìa và MP3 chỉ nhận
// file thường (không đi theo symlink ra ngoài thư mục sách).
func FromDir(dir string) (Book, error) {
	data, err := safepath.ReadRegular(filepath.Join(dir, "metadata.json"), maxMetadataBytes)
	if err != nil {
		return Book{}, fmt.Errorf("không đọc được metadata.json trong %q: %w", dir, err)
	}
	var m dirMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return Book{}, fmt.Errorf("metadata.json không hợp lệ: %w", err)
	}
	b := Book{Meta: Meta{Title: cleanTitle(m.Title), Author: cleanTitle(m.Author)}}
	if b.Title == "" {
		b.Title = filepath.Base(dir)
	}
	if safepath.IsPlainName(m.Cover) {
		if p := filepath.Join(dir, m.Cover); safepath.IsRegularFile(p) {
			b.Cover = p
		}
	}
	for _, ch := range m.Chapters {
		for _, sec := range ch.Sections {
			if !safepath.IsPlainName(sec.File) {
				return Book{}, fmt.Errorf("tên file tiểu mục không hợp lệ: %q", sec.File)
			}
			if info, err := os.Lstat(filepath.Join(dir, sec.File)); err == nil && !info.Mode().IsRegular() {
				return Book{}, fmt.Errorf("file tiểu mục %q: %w", sec.File, safepath.ErrNotRegular)
			}
			b.Tracks = append(b.Tracks, Track{
				Title: ChapterTitle(ch.Title, sec.Title, len(ch.Sections)),
				File:  filepath.Join(dir, sec.File),
			})
		}
	}
	if len(b.Tracks) == 0 {
		return Book{}, errors.New("sách không có tiểu mục nào")
	}
	return b, nil
}

// ChapterTitle — tên mốc mục lục của một tiểu mục: tên tiểu mục; chương chỉ có
// một tiểu mục trùng tên chương (hoặc tiểu mục không tên) thì dùng tên chương.
func ChapterTitle(chapter, section string, sections int) string {
	chapter, section = cleanTitle(chapter), cleanTitle(section)
	if section == "" || (sections == 1 && strings.EqualFold(section, chapter)) {
		if chapter != "" {
			return chapter
		}
	}
	return section
}

// cleanTitle gộp khoảng trắng / xuống dòng thành một dấu cách.
func cleanTitle(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// FileName — tên file .m4b an toàn trên mọi hệ điều hành từ tên sách (giữ
// tiếng Việt có dấu, bỏ ký tự cấm như / \ : * ? " < > |).
func FileName(title string) string {
	var b strings.Builder
	for _, r := range cleanTitle(title) {
		switch {
		case strings.ContainsRune(`/\:*?"<>|`, r), unicode.IsControl(r):
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}
	name := strings.Trim(cleanTitle(b.String()), " .")
	if r := []rune(name); len(r) > 120 {
		name = strings.TrimSpace(string(r[:120]))
	}
	if name == "" {
		name = "Sách nói"
	}
	return name + ".m4b"
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
