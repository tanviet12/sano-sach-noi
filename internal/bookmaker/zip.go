package bookmaker

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tcolgate/mp3"

	"sano/internal/safepath"
)

// Định dạng zip chuẩn — SOURCE OF TRUTH: docs/book-zip-format.md.
// Nơi đọc gói zip (phần mềm, máy chủ giai đoạn sau) bám đúng cấu trúc dưới đây.
//
//	book-<slug>.zip
//	├── manifest.json
//	├── chapters.json
//	├── cover.jpg|png
//	└── audio/chNN/secNN.mp3
const zipFormatVersion = 1

// maxMetadataBytes — metadata.json lớn hơn thì từ chối (thư mục có thể của người khác gửi).
const maxMetadataBytes = 8 << 20

// zipManifest — manifest.json: metadata sách + thông tin đóng gói.
type zipManifest struct {
	Title          string   `json:"title"`
	Slug           string   `json:"slug"`
	Author         string   `json:"author,omitempty"`
	Description    string   `json:"description,omitempty"`
	CategorySlug   string   `json:"category_slug,omitempty"`
	Category       string   `json:"category,omitempty"` // tên danh mục hiển thị (vd "Kỹ năng"); category_slug suy ra từ đây
	Tags           []string `json:"tags,omitempty"`
	CoverFilename  string   `json:"cover_filename,omitempty"`
	VoiceID        string   `json:"voice_id,omitempty"`
	Version        int      `json:"version"`
	BuildTimestamp string   `json:"build_timestamp"`
}

// zipChapters — chapters.json: mục lục đa cấp + text song song mỗi tiểu mục.
type zipChapters struct {
	Chapters []zipChapter `json:"chapters"`
}

type zipChapter struct {
	Order    int          `json:"order"`
	Title    string       `json:"title"`
	Sections []zipSection `json:"sections"`
}

type zipSection struct {
	Order            int    `json:"order"`
	Title            string `json:"title"`
	AudioFilename    string `json:"audio_filename"` // đường dẫn tương đối trong zip: audio/chNN/secNN.mp3
	DurationSec      int    `json:"duration_sec"`
	OriginalText     string `json:"original_text,omitempty"`
	ReadingScript    string `json:"reading_script,omitempty"`
	ImageDescription string `json:"image_description,omitempty"`
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// packageBookZip đóng gói thư mục output (đã build mp3 + metadata) thành file zip
// chuẩn ở zipPath. KHÔNG đụng DB. Trả slug đã dùng. Validate trước khi ghi.
func packageBookZip(opts Options, meta outMeta, zipPath string) (string, error) {
	slug := docxSlugify(firstNonEmpty(meta.Title))
	if !slugRe.MatchString(slug) {
		return "", fmt.Errorf("slug suy ra %q không hợp lệ (kebab-case)", slug)
	}

	// metadata.json có thể đến từ thư mục sách người khác gửi (--repack-dir):
	// chỉ nhận tên file nằm ngay trong thư mục sách, không đọc file ngoài.
	if meta.Cover != "" && !safepath.IsPlainName(meta.Cover) {
		return "", fmt.Errorf("tên file bìa không hợp lệ: %q", meta.Cover)
	}
	chapters := zipChapters{Chapters: make([]zipChapter, 0, len(meta.Chapters))}
	type audioEntry struct{ srcPath, zipPath string }
	var audios []audioEntry

	for ci, ch := range meta.Chapters {
		zc := zipChapter{Order: ci + 1, Title: ch.Title, Sections: make([]zipSection, 0, len(ch.Sections))}
		for si, sec := range ch.Sections {
			if !safepath.IsPlainName(sec.File) {
				return "", fmt.Errorf("tên file tiểu mục không hợp lệ: %q", sec.File)
			}
			src := filepath.Join(opts.OutputDir, sec.File) // chNN-secNN.mp3 (layout phẳng cũ)
			if _, err := os.Stat(src); err != nil {
				return "", fmt.Errorf("thiếu file audio %q: %w", sec.File, err)
			}
			// Không đi theo symlink trong thư mục sách (có thể trỏ ra ngoài).
			if !safepath.IsRegularFile(src) {
				return "", fmt.Errorf("file audio %q: %w", sec.File, safepath.ErrNotRegular)
			}
			zipAudio := fmt.Sprintf("audio/ch%02d/sec%02d.mp3", ci+1, si+1)
			dur, err := docxMP3DurationSec(src)
			if err != nil {
				opts.Logf("cảnh báo: không đọc được thời lượng %q: %v", src, err)
			}
			zc.Sections = append(zc.Sections, zipSection{
				Order:            si + 1,
				Title:            sec.Title,
				AudioFilename:    zipAudio,
				DurationSec:      dur,
				OriginalText:     sec.OriginalText,
				ReadingScript:    sec.ReadingScript,
				ImageDescription: sec.ImageDescription,
			})
			audios = append(audios, audioEntry{srcPath: src, zipPath: zipAudio})
		}
		chapters.Chapters = append(chapters.Chapters, zc)
	}

	manifest := zipManifest{
		Title:          meta.Title,
		Slug:           slug,
		Author:         meta.Author,
		Description:    meta.Description,
		CategorySlug:   CategorySlug(meta.Category),
		Category:       strings.TrimSpace(meta.Category),
		Tags:           meta.Tags,
		CoverFilename:  coverInZip(meta.Cover),
		VoiceID:        opts.TTS.Voice,
		Version:        zipFormatVersion,
		BuildTimestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err := validateZipContent(manifest, chapters); err != nil {
		return "", fmt.Errorf("validate format: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(zipPath), 0o755); err != nil {
		return "", fmt.Errorf("tạo thư mục zip: %w", err)
	}
	f, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("tạo file zip %q: %w", zipPath, err)
	}
	defer func() { _ = f.Close() }()

	zw := zip.NewWriter(f)
	if err := writeZipJSON(zw, "manifest.json", manifest); err != nil {
		return "", err
	}
	if err := writeZipJSON(zw, "chapters.json", chapters); err != nil {
		return "", err
	}
	if manifest.CoverFilename != "" {
		coverSrc := filepath.Join(opts.OutputDir, meta.Cover)
		if err := writeZipFile(zw, manifest.CoverFilename, coverSrc); err != nil {
			return "", fmt.Errorf("ghi bìa vào zip: %w", err)
		}
	}
	for _, a := range audios {
		if err := writeZipFile(zw, a.zipPath, a.srcPath); err != nil {
			return "", fmt.Errorf("ghi audio %q vào zip: %w", a.zipPath, err)
		}
	}
	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("đóng zip: %w", err)
	}
	return slug, nil
}

// RepackZip đóng gói LẠI zip chuẩn từ thư mục đã render sẵn (mp3 +
// metadata.json), KHÔNG cần docx và KHÔNG render lại TTS. Dùng khi sửa lẻ vài
// tiểu mục — vd render lại 1 đoạn đọc sai — rồi đóng gói lại để upload. Đọc
// metadata.json (outMeta) trong dir, packageBookZip tự đếm lại duration từng mp3.
// voice ghi vào manifest.voice_id (metadata.json không lưu voice nên truyền vào).
func RepackZip(dir, zipPath, voice string, logf func(string, ...any)) (string, error) {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	metaPath := filepath.Join(dir, "metadata.json")
	data, err := safepath.ReadRegular(metaPath, maxMetadataBytes)
	if err != nil {
		return "", fmt.Errorf("đọc metadata.json (cần --repack-dir trỏ thư mục đã render): %w", err)
	}
	var meta outMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return "", fmt.Errorf("parse metadata.json: %w", err)
	}
	opts := Options{
		OutputDir: dir,
		TTS:       TTSConfig{Voice: voice},
		Logf:      logf,
	}
	return packageBookZip(opts, meta, zipPath)
}

// validateZipContent kiểm tra ràng buộc format trước khi đóng gói (docs/book-zip-format.md §6).
func validateZipContent(m zipManifest, c zipChapters) error {
	if strings.TrimSpace(m.Title) == "" {
		return errors.New("manifest.title bắt buộc")
	}
	if !slugRe.MatchString(m.Slug) {
		return fmt.Errorf("manifest.slug %q không phải kebab-case", m.Slug)
	}
	if m.Version != zipFormatVersion {
		return fmt.Errorf("manifest.version phải = %d", zipFormatVersion)
	}
	if len(c.Chapters) == 0 {
		return errors.New("chapters rỗng")
	}
	total := 0
	for _, ch := range c.Chapters {
		for _, s := range ch.Sections {
			if strings.TrimSpace(s.AudioFilename) == "" {
				return fmt.Errorf("section «%s» thiếu audio_filename", s.Title)
			}
			total++
		}
	}
	if total == 0 {
		return errors.New("không có tiểu mục nào")
	}
	return nil
}

// coverInZip chuẩn hóa tên bìa trong zip: giữ đuôi, đặt tên cover.<ext>.
func coverInZip(coverName string) string {
	if strings.TrimSpace(coverName) == "" {
		return ""
	}
	ext := strings.ToLower(filepath.Ext(coverName))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		ext = ".png"
	}
	return "cover" + ext
}

func writeZipJSON(zw *zip.Writer, name string, v any) error {
	w, err := zw.Create(name)
	if err != nil {
		return fmt.Errorf("tạo entry %q: %w", name, err)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("ghi JSON %q: %w", name, err)
	}
	return nil
}

// writeZipFile chép srcPath vào zip. Chỉ nhận file thường: thư mục sách đóng
// gói lại (--repack-dir) có thể là của người khác gửi, chứa symlink ra ngoài.
func writeZipFile(zw *zip.Writer, name, srcPath string) error {
	src, err := safepath.OpenRegular(srcPath)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, src)
	return err
}

// docxSlugify — gấp dấu tiếng Việt + kebab-case (slug thư mục + gói zip).
var docxVietFold = strings.NewReplacer(
	"à", "a", "á", "a", "ả", "a", "ã", "a", "ạ", "a",
	"ă", "a", "ằ", "a", "ắ", "a", "ẳ", "a", "ẵ", "a", "ặ", "a",
	"â", "a", "ầ", "a", "ấ", "a", "ẩ", "a", "ẫ", "a", "ậ", "a",
	"è", "e", "é", "e", "ẻ", "e", "ẽ", "e", "ẹ", "e",
	"ê", "e", "ề", "e", "ế", "e", "ể", "e", "ễ", "e", "ệ", "e",
	"ì", "i", "í", "i", "ỉ", "i", "ĩ", "i", "ị", "i",
	"ò", "o", "ó", "o", "ỏ", "o", "õ", "o", "ọ", "o",
	"ô", "o", "ồ", "o", "ố", "o", "ổ", "o", "ỗ", "o", "ộ", "o",
	"ơ", "o", "ờ", "o", "ớ", "o", "ở", "o", "ỡ", "o", "ợ", "o",
	"ù", "u", "ú", "u", "ủ", "u", "ũ", "u", "ụ", "u",
	"ư", "u", "ừ", "u", "ứ", "u", "ử", "u", "ữ", "u", "ự", "u",
	"ỳ", "y", "ý", "y", "ỷ", "y", "ỹ", "y", "ỵ", "y",
	"đ", "d",
)

func docxSlugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = docxVietFold.Replace(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "audiobook"
	}
	return out
}

// CategorySlug suy slug danh mục (kebab-case, gấp dấu) từ tên danh mục người
// dùng đặt, vd "Kỹ năng" → "ky-nang". Tên đã là slug thì giữ nguyên. Tên rỗng
// hoặc không có chữ/số nào → "".
func CategorySlug(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	s := docxSlugify(name)
	if s == "audiobook" && !strings.Contains(docxVietFold.Replace(strings.ToLower(name)), "audiobook") {
		return ""
	}
	return s
}

// docxMP3DurationSec đọc thời lượng MP3 (giây) pure-Go — không cần ffprobe.
func docxMP3DurationSec(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()

	dec := mp3.NewDecoder(f)
	var total time.Duration
	var frame mp3.Frame
	skipped := 0
	for {
		if err := dec.Decode(&frame, &skipped); err != nil {
			break // EOF hoặc rác giữa chừng → trả phần đã cộng
		}
		total += frame.Duration()
	}
	return int(math.Round(total.Seconds())), nil
}
