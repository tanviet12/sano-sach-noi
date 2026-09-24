package bookmaker

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessImages_NoFilenameInDescription(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	o := Options{OutputDir: dir}
	sec := Section{
		Title:  "1.6. Lập kế hoạch",
		Images: []SectionImage{{Name: "image1.png", Data: []byte("x")}, {Name: "image2.png", Data: []byte("y")}},
	}
	desc, err := o.processImages(sec, "Chương 1", nil)
	if err != nil {
		t.Fatalf("processImages lỗi: %v", err)
	}
	if strings.Contains(strings.ToLower(desc), ".png") || strings.Contains(strings.ToLower(desc), "tệp") {
		t.Errorf("mô tả KHÔNG được chứa tên file ảnh: %q", desc)
	}
	if !strings.Contains(desc, "2 hình") {
		t.Errorf("mô tả nên ghi số lượng hình: %q", desc)
	}
}

func TestWriteCover_CustomCoverPath_CopiesWithExt(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bia-goc.png")
	if err := os.WriteFile(src, []byte("\x89PNG\r\n\x1a\nFAKE"), 0o644); err != nil {
		t.Fatalf("ghi ảnh nguồn: %v", err)
	}
	o := Options{OutputDir: dir, CoverPath: src}
	// truyền first != nil để chắc chắn --cover được ưu tiên hơn ảnh đầu sách
	name, err := o.writeCover(&SectionImage{Name: "image1.jpg", Data: []byte("xx")}, "Sách thử")
	if err != nil {
		t.Fatalf("writeCover lỗi: %v", err)
	}
	if name != "cover.png" {
		t.Errorf("muốn cover.png, got %q", name)
	}
	if _, err := os.Stat(filepath.Join(dir, "cover.png")); err != nil {
		t.Errorf("file cover.png chưa được ghi: %v", err)
	}
}

func TestWriteCover_UnsupportedExt_Errors(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bia.gif")
	if err := os.WriteFile(src, []byte("GIF89a"), 0o644); err != nil {
		t.Fatalf("ghi ảnh nguồn: %v", err)
	}
	o := Options{OutputDir: dir, CoverPath: src}
	if _, err := o.writeCover(nil, "Sách thử"); err == nil {
		t.Error("muốn lỗi với định dạng .gif không hỗ trợ, nhưng không có lỗi")
	}
}

func TestBuildIntroChapter_TitleAndNoTitlePrepend(t *testing.T) {
	title := "Kỹ năng giao tiếp — Lắng nghe (Phần 1)"
	intro := "Bạn đang nghe sách nói được tạo bằng Sano.\n\nKỹ năng giao tiếp, Lắng nghe, Phần 1.\n\nTác giả: Nguyễn Văn A."
	ch, job := buildIntroChapter(defaultNormalizer, title, intro)

	wantTitle := normalizeTitle(title)
	if ch.Title != wantTitle {
		t.Errorf("tên chương intro phải = tiêu đề sách %q, got %q", wantTitle, ch.Title)
	}
	if len(ch.Sections) != 1 {
		t.Fatalf("intro phải có đúng 1 tiểu mục, got %d", len(ch.Sections))
	}
	sec := ch.Sections[0]
	if sec.File != "ch01-sec01.mp3" || job.Stem != "ch01-sec01" {
		t.Errorf("intro phải là ch01-sec01, got file=%q stem=%q", sec.File, job.Stem)
	}
	// KHÔNG prepend tiêu đề: lời đọc bắt đầu bằng câu xưng nguồn, không phải tên sách.
	if !strings.HasPrefix(sec.ReadingScript, "Bạn đang nghe") {
		t.Errorf("lời đọc intro không được prepend tiêu đề; got mở đầu %q", sec.ReadingScript[:min(40, len(sec.ReadingScript))])
	}
	// Giữ dòng trắng giữa 3 dòng intro → có \n\n để TTS nghỉ.
	if !strings.Contains(sec.ReadingScript, "\n\n") {
		t.Errorf("lời đọc intro phải giữ dòng trắng giữa các dòng: %q", sec.ReadingScript)
	}
	// original_text giữ nguyên intro gốc.
	if sec.OriginalText != intro {
		t.Errorf("original_text intro phải giữ nguyên text gốc")
	}
}

func mkTestBook() *Book {
	return &Book{Chapters: []Chapter{
		{Title: "Nội dung", Sections: []Section{{Title: "Nội dung"}}},
		{Title: "MỤC LỤC", Sections: []Section{{Title: "MỤC LỤC"}}},
		{Title: "Mở đầu", Sections: []Section{{Title: "A"}, {Title: "B"}}},
	}}
}

func TestDropSectionsByStem_DropsFrontMatterAndEmptyChapters(t *testing.T) {
	b := mkTestBook()
	removed := dropSectionsByStem(b, map[string]bool{"ch01-sec01": true, "ch02-sec01": true})
	if removed != 2 {
		t.Fatalf("muốn loại 2 tiểu mục, got %d", removed)
	}
	if len(b.Chapters) != 1 {
		t.Fatalf("hai chương rỗng phải bị bỏ, muốn còn 1 chương, got %d", len(b.Chapters))
	}
	if b.Chapters[0].Title != "Mở đầu" || len(b.Chapters[0].Sections) != 2 {
		t.Errorf("chương còn lại sai: %+v", b.Chapters[0])
	}
}

func TestDropSectionsByStem_EmptyDrop_NoChange(t *testing.T) {
	b := mkTestBook()
	if n := dropSectionsByStem(b, nil); n != 0 {
		t.Errorf("drop nil không được loại gì, got %d", n)
	}
	if len(b.Chapters) != 3 {
		t.Errorf("drop nil phải giữ nguyên 3 chương, got %d", len(b.Chapters))
	}
}

func TestDropSectionsByStem_PartialSectionDrop_KeepsChapter(t *testing.T) {
	b := mkTestBook()
	// loại 1 trong 2 tiểu mục của chương 3 → chương vẫn còn
	removed := dropSectionsByStem(b, map[string]bool{"ch03-sec01": true})
	if removed != 1 {
		t.Fatalf("muốn loại 1 tiểu mục, got %d", removed)
	}
	if len(b.Chapters) != 3 {
		t.Fatalf("không chương nào rỗng, muốn giữ 3 chương, got %d", len(b.Chapters))
	}
	last := b.Chapters[2]
	if len(last.Sections) != 1 || last.Sections[0].Title != "B" {
		t.Errorf("chương 3 phải còn đúng tiểu mục B, got %+v", last.Sections)
	}
}

func TestWriteCover_NoCover_DrawsGeneratedPNG(t *testing.T) {
	dir := t.TempDir()
	o := Options{OutputDir: dir, Author: "Nguyễn Văn A"}
	// có ảnh đầu sách nhưng KHÔNG bật CoverFirstImage → vẫn tự vẽ bìa
	name, err := o.writeCover(&SectionImage{Name: "image1.jpg", Data: []byte("xx")}, "Kỹ năng giao tiếp")
	if err != nil {
		t.Fatalf("writeCover lỗi: %v", err)
	}
	if name != "cover.png" {
		t.Fatalf("muốn cover.png, got %q", name)
	}
	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("mở bìa: %v", err)
	}
	defer f.Close()
	cfg, err := png.DecodeConfig(f)
	if err != nil {
		t.Fatalf("bìa không phải PNG hợp lệ: %v", err)
	}
	if cfg.Width != coverWidth || cfg.Height != coverHeight {
		t.Errorf("kích thước bìa = %dx%d, muốn %dx%d", cfg.Width, cfg.Height, coverWidth, coverHeight)
	}
}

func TestWriteCover_FirstImageFlag_UsesBookImage(t *testing.T) {
	dir := t.TempDir()
	o := Options{OutputDir: dir, CoverFirstImage: true}
	name, err := o.writeCover(&SectionImage{Name: "image1.jpg", Data: []byte("xx")}, "Sách thử")
	if err != nil {
		t.Fatalf("writeCover lỗi: %v", err)
	}
	if name != "cover.jpg" {
		t.Errorf("muốn cover.jpg (ảnh đầu sách), got %q", name)
	}
}
