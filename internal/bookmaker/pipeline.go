package bookmaker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sano/internal/cover"
	"sano/internal/safepath"
)

// outMeta — metadata.json đầu ra (trường chapters nhiều cấp + text song song
// mỗi tiểu mục; xem docs/book-zip-format.md).
type outMeta struct {
	Title       string   `json:"title"`
	Author      string   `json:"author,omitempty"`
	Narrator    string   `json:"narrator,omitempty"`
	Description string   `json:"description,omitempty"`
	Language    string   `json:"language,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Cover       string   `json:"cover,omitempty"`
	Visibility  string   `json:"visibility,omitempty"`
	Category    string   `json:"category,omitempty"`
	// RightsConfirmedAt — lúc người dùng xác nhận có quyền dùng tài liệu này (app desktop, RFC 3339).
	RightsConfirmedAt string       `json:"rights_confirmed_at,omitempty"`
	Chapters          []outChapter `json:"chapters"`
}

type outChapter struct {
	Title    string       `json:"title"`
	Sections []outSection `json:"sections"`
}

type outSection struct {
	Title            string `json:"title"`
	File             string `json:"file"`
	OriginalText     string `json:"original_text,omitempty"`
	ReadingScript    string `json:"reading_script,omitempty"`
	ImageDescription string `json:"image_description,omitempty"`
}

// Options — tham số chạy pipeline docx → output folder (+ gói zip nếu cần).
type Options struct {
	InputDocx         string
	OutputDir         string
	Title             string
	Author            string
	Description       string
	IntroText         string // != "" → chèn mục intro/branding đầu sách (tên mục = tiêu đề sách)
	Category          string
	Visibility        string
	RightsConfirmedAt string // lúc người dùng xác nhận có quyền dùng tài liệu (ghi vào metadata.json)
	Tags              []string
	TTS               TTSConfig
	Norm              *Normalizer     // luật chuẩn hóa lời đọc; nil = mặc định
	DropStems         map[string]bool // loại tiểu mục theo stem GỐC (ch%02d-sec%02d) trước khi đánh số lại
	DropTOC           bool            // tự bỏ trang mục lục (tiêu đề Mục lục/Nội dung… hoặc dòng kết thúc bằng số trang)
	CoverPath         string          // != "" → dùng ảnh bìa này thay bìa tự vẽ
	CoverFirstImage   bool            // không có CoverPath → dùng ảnh đầu sách (mặc định: tự vẽ bìa)
	OutputZip         string          // != "" → đóng gói thêm zip chuẩn ở đường dẫn này
	Logf              func(string, ...any)

	// ReadingEdits — lời đọc người dùng sửa ở bước nghe thử, theo stem GỐC
	// (IntroStem cho lời mở đầu). Chỉ áp khi lời đọc hiện tại còn bắt đầu bằng
	// Edit.From (phần đã nghe thử), phần sau giữ nguyên.
	ReadingEdits map[string]ReadingEdit

	// Progress — nếu != nil: báo tiến độ sau mỗi tiểu mục render xong và khi
	// đóng gói. Gọi từ goroutine đang chạy Run.
	Progress func(Progress)
}

// IntroStem — stem gốc dành cho lời mở đầu (không có trong file Word).
const IntroStem = "intro"

// ReadingEdit — thay phần đầu lời đọc của một tiểu mục.
type ReadingEdit struct {
	From string `json:"from"` // phần đầu lời đọc gốc (đã nghe thử)
	To   string `json:"to"`   // lời đọc thay thế
}

// Các giai đoạn báo trong Progress.Phase.
const (
	PhaseRender  = "render"
	PhasePackage = "package"
	PhaseDone    = "done"
)

// Progress — tiến độ một lượt render.
type Progress struct {
	Phase      string `json:"phase"`
	Done       int    `json:"done"`       // số tiểu mục đã render xong
	Total      int    `json:"total"`      // tổng số tiểu mục
	DoneChars  int    `json:"doneChars"`  // số ký tự lời đọc đã render
	TotalChars int    `json:"totalChars"` // tổng ký tự lời đọc
	Stem       string `json:"stem"`       // tiểu mục vừa xong (stem đầu ra chNN-secNN)
	Title      string `json:"title"`
	ElapsedSec int    `json:"elapsedSec"`
}

// prepared — kết quả bước dựng lời đọc (chưa render audio).
type prepared struct {
	book       *Book
	title      string
	meta       outMeta
	jobs       []ttsJob // theo thứ tự đọc (lời mở đầu, nếu có, ở cuối — giữ như bản CLI cũ)
	origStems  map[string]string
	titles     map[string]string // stem đầu ra → tiêu đề tiểu mục
	firstImage *SectionImage
	written    map[string]bool // ảnh đã ghi ra images/ (mỗi ảnh ghi 1 lần cho cả cuốn)
	dropped    []droppedTOC
	secCount   int
}

// Run thực thi toàn bộ: parse docx → chuẩn hóa → ảnh → TTS → metadata.json
// → (tùy chọn) gói zip. Trả tổng số tiểu mục đã build.
func Run(opts Options) (int, error) {
	return RunContext(context.Background(), opts)
}

// RunContext như Run nhưng hủy được qua ctx (dừng bộ đọc giữa chừng) và báo
// tiến độ qua opts.Progress.
func RunContext(ctx context.Context, opts Options) (int, error) {
	if opts.Logf == nil {
		opts.Logf = func(string, ...any) {}
	}
	p, err := opts.prepare()
	if err != nil {
		return 0, err
	}
	if len(p.dropped) > 0 {
		fmt.Printf("Đã bỏ %d tiểu mục mục lục khỏi bản đọc (tắt bằng --drop-toc=false):\n", len(p.dropped))
		for _, d := range p.dropped {
			fmt.Println(formatDroppedTOC(d))
		}
	}
	meta, jobs, secCount := p.meta, p.jobs, p.secCount

	// Bìa: ảnh user chọn → (tuỳ chọn) ảnh đầu sách → bìa tự vẽ theo tên sách.
	coverName, err := opts.writeCover(p.firstImage, p.title)
	if err != nil {
		return 0, err
	}
	meta.Cover = coverName

	printLoadWarnings(os.Stdout, p.book.Stats, unknownAcronyms(jobTexts(jobs), opts.norm().dict))

	// Pre-render TTS (E12d).
	if err := opts.TTS.preflight(); err != nil {
		return 0, err
	}
	totalChars := 0
	charsOf := make(map[string]int, len(jobs))
	for _, j := range jobs {
		n := len([]rune(j.Text))
		charsOf[j.Stem] = n
		totalChars += n
	}
	start := time.Now()
	done, doneChars := 0, 0
	report := func(phase, stem string) {
		if opts.Progress == nil {
			return
		}
		opts.Progress(Progress{
			Phase: phase, Done: done, Total: len(jobs), DoneChars: doneChars, TotalChars: totalChars,
			Stem: stem, Title: p.titles[stem], ElapsedSec: int(time.Since(start).Seconds()),
		})
	}
	report(PhaseRender, "")
	onDone := func(stem string) {
		done++
		doneChars += charsOf[stem]
		report(PhaseRender, stem)
	}
	if err := opts.TTS.RenderContext(ctx, opts.OutputDir, jobs, onDone); err != nil {
		return 0, err
	}

	// Ghi metadata.json.
	metaPath := filepath.Join(opts.OutputDir, "metadata.json")
	if err := writeJSON(metaPath, meta); err != nil {
		return 0, err
	}
	opts.Logf("✓ metadata.json (%d chương, %d tiểu mục)", len(meta.Chapters), secCount)

	// Đóng gói zip chuẩn (sao lưu / chuyển máy).
	if opts.OutputZip != "" {
		report(PhasePackage, "")
		slug, err := packageBookZip(opts, meta, opts.OutputZip)
		if err != nil {
			return secCount, fmt.Errorf("đóng gói zip: %w", err)
		}
		opts.Logf("✓ zip chuẩn: %s (slug=%s)", opts.OutputZip, slug)
		report(PhaseDone, "")
		return secCount, nil
	}

	report(PhaseDone, "")
	return secCount, nil
}

func (o Options) norm() *Normalizer {
	if o.Norm == nil {
		return defaultNormalizer
	}
	return o.Norm
}

// prepare nạp docx, bỏ tiểu mục theo DropStems / trang mục lục, ghi ảnh ra
// OutputDir/images và dựng lời đọc cho từng tiểu mục. Chưa render audio.
func (opts Options) prepare() (*prepared, error) {
	book, err := ParseDocx(opts.InputDocx)
	if err != nil {
		return nil, err
	}
	norm := opts.norm()
	if opts.Logf == nil {
		opts.Logf = func(string, ...any) {}
	}

	if n := dropSectionsByStem(book, opts.DropStems); n > 0 {
		opts.Logf("đã loại %d tiểu mục theo --drop-stems; còn %d chương (đánh số lại từ 1)", n, len(book.Chapters))
	}
	p := &prepared{book: book, origStems: map[string]string{}, titles: map[string]string{}, written: map[string]bool{}}
	if opts.DropTOC {
		p.dropped = dropTOCSections(book)
	}

	p.title = firstNonEmpty(opts.Title, book.Title, titleFromDocxName(opts.InputDocx))
	if err := os.MkdirAll(filepath.Join(opts.OutputDir, "images"), 0o755); err != nil {
		return nil, fmt.Errorf("tạo thư mục output: %w", err)
	}

	p.meta = outMeta{
		Title:             p.title,
		Author:            opts.Author,
		Narrator:          narratorLabel(opts.TTS),
		Description:       opts.Description,
		Language:          "vi",
		Tags:              opts.Tags,
		Visibility:        opts.Visibility,
		Category:          opts.Category,
		RightsConfirmedAt: opts.RightsConfirmedAt,
	}

	// Intro/branding chiếm chương 1 (nếu có) → chương docx dời xuống ch02+.
	introOffset := 0
	if opts.IntroText != "" {
		introOffset = 1
	}

	for ci, ch := range book.Chapters {
		outCh := outChapter{Title: normalizeTitle(ch.Title)}
		for si, sec := range ch.Sections {
			p.secCount++
			stem := fmt.Sprintf("ch%02d-sec%02d", ci+1+introOffset, si+1)

			// Trích ảnh + sinh image_description (E12b).
			imageDesc, err := opts.processImages(sec, ch.Title, p.written)
			if err != nil {
				return nil, err
			}
			if p.firstImage == nil && len(sec.Images) > 0 {
				p.firstImage = &sec.Images[0]
			}

			original := strings.TrimSpace(sec.Text)
			reading := applyReadingEdit(sectionReading(norm, sec), opts.ReadingEdits[sec.Stem])
			// KHÔNG ghép image_description vào text đọc: hiện chỉ là placeholder +
			// tên file (vd "image1.png") → TTS đọc tên file rất khó chịu. Giữ
			// image_description trong chapters.json (metadata) để dùng/hiển thị sau.

			p.jobs = append(p.jobs, ttsJob{Stem: stem, Text: reading})
			p.origStems[stem] = sec.Stem
			p.titles[stem] = normalizeTitle(sec.Title)
			outCh.Sections = append(outCh.Sections, outSection{
				Title:            normalizeTitle(sec.Title),
				File:             stem + ".mp3",
				OriginalText:     original,
				ReadingScript:    reading,
				ImageDescription: imageDesc,
			})
		}
		p.meta.Chapters = append(p.meta.Chapters, outCh)
	}

	if p.secCount == 0 {
		return nil, fmt.Errorf("docx không sinh được tiểu mục nào")
	}

	// Chèn mục intro/branding làm chương đầu (ch01): tên mục = tiêu đề sách, lời đọc
	// = INTRO_TEXT nguyên văn (KHÔNG prepend tiêu đề để khỏi đọc lặp tên sách).
	if introOffset == 1 {
		introCh, introJob := buildIntroChapter(norm, p.title, opts.IntroText)
		if e, ok := opts.ReadingEdits[IntroStem]; ok {
			introJob.Text = applyReadingEdit(introJob.Text, e)
			introCh.Sections[0].ReadingScript = introJob.Text
		}
		p.meta.Chapters = append([]outChapter{introCh}, p.meta.Chapters...)
		p.jobs = append(p.jobs, introJob)
		p.origStems[introJob.Stem] = IntroStem
		p.titles[introJob.Stem] = introCh.Title
		p.secCount++
	}
	return p, nil
}

// sectionReading dựng lời đọc của một tiểu mục: tiêu đề (đã chuẩn hóa) thành
// một đoạn riêng rồi tới nội dung. Vd "1.6. Quản lý thời gian" → "Quản lý thời
// gian." (hoặc "một chấm sáu. Quản lý thời gian" nếu giữ số), để TTS nghỉ trước
// khi vào nội dung. Không ghép tiêu đề thì audio bỏ qua tiêu đề.
func sectionReading(norm *Normalizer, sec Section) string {
	body := norm.script(strings.TrimSpace(sec.Text))
	spokenTitle := strings.TrimRight(norm.script(norm.spokenTitle(sec.Title)), " .")
	if spokenTitle == "" {
		return body
	}
	if body == "" {
		return spokenTitle + "."
	}
	// Dòng trắng (\n\n) sau tiêu đề → TTS nghỉ tương xứng trước khi vào nội dung.
	return spokenTitle + ".\n\n" + body
}

// applyReadingEdit thay phần đầu e.From của lời đọc bằng e.To. Lời đọc không
// còn bắt đầu bằng e.From (file Word đã đổi) → giữ nguyên, không đoán.
func applyReadingEdit(reading string, e ReadingEdit) string {
	if e.From == "" || !strings.HasPrefix(reading, e.From) {
		return reading
	}
	return strings.TrimSpace(e.To + reading[len(e.From):])
}

// buildIntroChapter dựng chương intro/branding (ch01): tên chương = section =
// tiêu đề sách (đã chuẩn hóa), lời đọc = INTRO_TEXT chuẩn hóa (giữ dòng trắng để
// TTS nghỉ giữa các dòng). KHÔNG prepend tiêu đề vào lời đọc — intro text đã chứa
// tên sách nên prepend sẽ đọc lặp. File audio cố định ch01-sec01.mp3.
func buildIntroChapter(norm *Normalizer, title, introText string) (outChapter, ttsJob) {
	introReading := norm.script(introText)
	t := normalizeTitle(title)
	ch := outChapter{
		Title: t,
		Sections: []outSection{{
			Title:         t,
			File:          "ch01-sec01.mp3",
			OriginalText:  introText,
			ReadingScript: introReading,
		}},
	}
	return ch, ttsJob{Stem: "ch01-sec01", Text: introReading}
}

// dropSectionsByStem loại các tiểu mục có stem GỐC (ch%02d-sec%02d theo thứ tự
// hiện tại của book) nằm trong drop. Chương rỗng sau khi loại cũng bị bỏ. Các
// chương/tiểu mục còn lại sẽ được đánh số lại tuần tự ở vòng build phía sau.
// Trả về số tiểu mục đã loại (0 nếu drop rỗng).
func dropSectionsByStem(book *Book, drop map[string]bool) int {
	if len(drop) == 0 {
		return 0
	}
	removed := 0
	kept := make([]Chapter, 0, len(book.Chapters))
	for ci, ch := range book.Chapters {
		keptSecs := make([]Section, 0, len(ch.Sections))
		for si, sec := range ch.Sections {
			stem := fmt.Sprintf("ch%02d-sec%02d", ci+1, si+1)
			if drop[stem] {
				removed++
				continue
			}
			keptSecs = append(keptSecs, sec)
		}
		if len(keptSecs) == 0 {
			continue // bỏ luôn chương rỗng
		}
		ch.Sections = keptSecs
		kept = append(kept, ch)
	}
	book.Chapters = kept
	return removed
}

// processImages ghi ảnh của tiểu mục ra images/ + sinh mô tả placeholder.
// (E12b — chưa dùng Vision API: chưa có key; mô tả tham chiếu vị trí + tên tệp.)
// written ghi nhớ ảnh đã ghi (theo tên) để một ảnh được tham chiếu nhiều lần
// chỉ ghi đĩa một lần cho cả cuốn; nil → luôn ghi.
func (o Options) processImages(sec Section, chapterTitle string, written map[string]bool) (string, error) {
	if len(sec.Images) == 0 {
		return "", nil
	}
	imagesDir := filepath.Join(o.OutputDir, "images")
	for _, img := range sec.Images {
		// Tên do safeImageName đặt; kiểm lại để không bao giờ ghi ra ngoài images/.
		dst := filepath.Join(imagesDir, img.Name)
		if !safepath.IsPlainName(img.Name) || !safepath.Within(imagesDir, dst) {
			return "", fmt.Errorf("tên ảnh không hợp lệ: %q", img.Name)
		}
		if written[img.Name] {
			continue
		}
		if err := os.WriteFile(dst, img.Data, 0o644); err != nil {
			return "", fmt.Errorf("ghi ảnh %q: %w", dst, err)
		}
		if written != nil {
			written[img.Name] = true
		}
	}
	// KHÔNG nhúng tên file ảnh kỹ thuật (vd image1.png) vào mô tả — mô tả này
	// có thể được đọc/hiển thị, tên file gây nhiễu. Chỉ ghi số lượng hình.
	desc := fmt.Sprintf("Tiểu mục «%s» có %d hình minh họa.",
		strings.TrimSpace(sec.Title), len(sec.Images))
	return desc, nil
}

// writeCover ghi bìa: ưu tiên ảnh do user chỉ định (--cover); rồi ảnh đầu trong
// sách nếu bật --cover-first-image; còn lại tự vẽ bìa theo tên sách (internal/cover,
// cùng thiết kế bìa mặc định trong phần mềm). Trả tên file bìa trong output-dir.
func (o Options) writeCover(first *SectionImage, title string) (string, error) {
	if o.CoverPath != "" {
		ext := strings.ToLower(filepath.Ext(o.CoverPath))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp":
		default:
			return "", fmt.Errorf("ảnh bìa %q định dạng không hỗ trợ (jpg/jpeg/png/webp)", o.CoverPath)
		}
		data, err := os.ReadFile(o.CoverPath)
		if err != nil {
			return "", fmt.Errorf("đọc ảnh bìa %q: %w", o.CoverPath, err)
		}
		name := "cover" + ext
		if err := os.WriteFile(filepath.Join(o.OutputDir, name), data, 0o644); err != nil {
			return "", fmt.Errorf("ghi bìa: %w", err)
		}
		return name, nil
	}
	if o.CoverFirstImage && first != nil && len(first.Data) > 0 {
		name := "cover" + filepath.Ext(first.Name)
		if name == "cover" {
			name = "cover.png"
		}
		if err := os.WriteFile(filepath.Join(o.OutputDir, name), first.Data, 0o644); err != nil {
			return "", fmt.Errorf("ghi bìa: %w", err)
		}
		return name, nil
	}
	f, err := os.Create(filepath.Join(o.OutputDir, "cover.png"))
	if err != nil {
		return "", fmt.Errorf("tạo bìa: %w", err)
	}
	defer func() { _ = f.Close() }()
	if err := cover.WritePNG(f, title, o.Author, coverWidth, coverHeight); err != nil {
		return "", fmt.Errorf("vẽ bìa: %w", err)
	}
	return "cover.png", nil
}

// Bìa tự vẽ tỷ lệ 3:4, khớp khung bìa trong phần mềm.
const (
	coverWidth  = 1200
	coverHeight = 1600
)

// ── helpers ──────────────────────────────────────────────────────────────

func jobTexts(jobs []ttsJob) []string {
	out := make([]string, len(jobs))
	for i, j := range jobs {
		out[i] = j.Text
	}
	return out
}

func narratorLabel(c TTSConfig) string {
	if c.Mode == TTSModeStub {
		return "VieNeu-TTS (stub)"
	}
	return "VieNeu-TTS (" + c.Voice + ")"
}

func titleFromDocxName(p string) string {
	base := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
	base = strings.NewReplacer("-", " ", "_", " ").Replace(base)
	return strings.TrimSpace(base)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return "Sách nói"
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("mã hóa JSON: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("ghi %q: %w", path, err)
	}
	return nil
}
