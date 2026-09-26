package library

// Nhập sách từ gói zip người khác gửi (wireframe D6). Gói zip là dữ liệu KHÔNG
// tin cậy, nên:
//   - chỉ đọc đúng các mục cần (manifest.json, chapters.json, bìa, mp3 mà
//     chapters.json trỏ tới, tên khớp mẫu cố định) — không bao giờ giải nén theo
//     tên ghi trong zip, nên không có đường dẫn "../" hay tuyệt đối nào lọt ra ngoài;
//   - từ chối gói có tên trùng (xem trước một đằng, giải nén một nẻo), mục mã hoá,
//     quá nhiều mục, file quá lớn hoặc nén bất thường (zip bomb); đếm lại số byte
//     thật khi giải nén thay vì tin kích thước ghi trong zip;
//   - kiểm chữ ký đầu file: mp3 phải là mp3, bìa phải là PNG/JPEG/WebP (không SVG);
//   - bỏ ký tự điều khiển, cắt độ dài mọi chuỗi; mã sách tự suy từ tên, không lấy
//     manifest.slug;
//   - giải nén vào thư mục tạm ẩn, file tạo mới (O_EXCL), xong mới đưa vào thư
//     viện; lỗi/huỷ giữa chừng thì xoá sạch; đóng gói LẠI zip sạch, không giữ zip gốc.

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"sano/internal/bookmaker"
	"sano/internal/safepath"
)

// Giới hạn khi nhập (biến để test chỉnh được).
var (
	maxImportZipBytes   int64  = 3 << 30   // file zip
	maxImportTotalBytes int64  = 3 << 30   // tổng dung lượng giải nén
	maxImportAudioBytes int64  = 512 << 20 // một file mp3
	maxImportCoverBytes int64  = 10 << 20
	maxManifestBytes    int64  = 1 << 20
	maxImportChapters          = 1000
	maxImportSections          = 5000
	maxImportRatio      uint64 = 20 // mp3/ảnh hầu như không nén được: tỉ lệ cao = đáng ngờ
	maxJSONRatio        uint64 = 50
)

const (
	maxSectionTitleLen = 300     // tên chương / tiểu mục
	maxVoiceLen        = 60      // tên giọng
	maxTextLen         = 1 << 20 // chữ một tiểu mục
)

var (
	audioNameRe = regexp.MustCompile(`^audio/ch[0-9]{2,4}/sec[0-9]{2,4}\.mp3$`)
	coverNames  = []string{"cover.jpg", "cover.jpeg", "cover.png", "cover.webp"}
)

// ErrBadPackage — gói không phải gói sách Sano hợp lệ (thông điệp nói rõ vì sao).
type ErrBadPackage struct{ Reason string }

func (e *ErrBadPackage) Error() string { return e.Reason }

func bad(format string, a ...any) error { return &ErrBadPackage{Reason: fmt.Sprintf(format, a...)} }

// ImportPreview — thông tin hiện ở hộp xem trước.
type ImportPreview struct {
	Path        string `json:"path"`
	FileName    string `json:"fileName"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Category    string `json:"category"`
	Series      string `json:"series"`
	Volume      int    `json:"volume"`
	Voice       string `json:"voice"`
	Chapters    int    `json:"chapters"`
	Sections    int    `json:"sections"`
	DurationSec int    `json:"durationSec"`
	SizeBytes   int64  `json:"sizeBytes"`
	HasCover    bool   `json:"hasCover"`
	// CoverDataURL — ảnh bìa (đã kiểm là PNG/JPEG/WebP thật) để xem trước; trống nếu không có.
	CoverDataURL string `json:"coverDataUrl"`
	// Trùng: thư viện đã có cuốn cùng mã sách.
	ExistingSlug      string `json:"existingSlug"`
	ExistingTitle     string `json:"existingTitle"`
	ExistingCreatedAt string `json:"existingCreatedAt"`
}

// importPlan — nội dung gói đã kiểm, dùng cho cả xem trước và giải nén.
type importPlan struct {
	title, author, category, voice string
	series                         string
	volume                         int
	chapters                       []planChapter
	cover                          *zip.File
	coverExt                       string
	durationSec                    int
	sections                       int
	size                           int64
}

type planChapter struct {
	title    string
	sections []planSection
}

type planSection struct {
	title, original, script string
	audio                   *zip.File
}

// windowsReserved — tên thiết bị Windows, không đặt được làm tên thư mục ("Con" → "con").
var windowsReserved = regexp.MustCompile(`^(con|prn|aux|nul|com[0-9]|lpt[0-9])$`)

// BookSlug — mã sách (tên thư mục) suy từ tên sách, dùng chung cho tạo sách và nhập sách.
func BookSlug(title string) string {
	s := bookmaker.Slugify(title)
	if s == "" || !validSlug(s) {
		s = "sach"
	}
	if windowsReserved.MatchString(s) {
		s += "-sach"
	}
	return s
}

// PreviewImport kiểm gói zip ở path và trả thông tin xem trước. Không ghi gì.
func (l *Library) PreviewImport(path string) (*ImportPreview, error) {
	zr, size, err := openImportZip(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()
	p, err := planImport(zr, size)
	if err != nil {
		return nil, err
	}
	pv := &ImportPreview{
		Path: path, FileName: filepath.Base(path), Title: p.title, Author: p.author, Category: p.category, Series: p.series, Volume: p.volume, Voice: p.voice,
		Chapters: len(p.chapters), Sections: p.sections, DurationSec: p.durationSec, SizeBytes: size, HasCover: p.cover != nil,
	}
	if p.cover != nil {
		if data, err := readSmallEntry(p.cover, maxImportCoverBytes); err == nil && isImage(data) {
			mime := map[string]string{".png": "image/png", ".webp": "image/webp"}[p.coverExt]
			if mime == "" {
				mime = "image/jpeg"
			}
			pv.CoverDataURL = "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
		}
	}
	if d, err := l.Get(BookSlug(p.title)); err == nil {
		pv.ExistingSlug, pv.ExistingTitle, pv.ExistingCreatedAt = d.Slug, d.Title, d.CreatedAt
	}
	return pv, nil
}

// ImportProgress — tiến độ giải nén (số tiểu mục đã chép / tổng).
type ImportProgress func(done, total int)

// PrepareImport kiểm lại gói (file có thể đổi sau lúc xem trước) rồi giải nén vào
// thư mục tạm ẩn trong thư viện, ghi metadata.json và đóng gói lại zip sạch.
// Trả thư mục tạm, mã sách để Commit, và mã cuốn đang có trùng (existing, rỗng
// nếu không trùng). Trùng + !replace → giữ cả hai ("Tên (2)", slug-2); trùng +
// replace → giữ mã cũ, người gọi chuyển cuốn cũ vào Thùng rác rồi Commit.
// Lỗi / huỷ (ctx) → thư mục tạm đã bị xoá.
func (l *Library) PrepareImport(ctx context.Context, path string, replace bool, progress ImportProgress) (workDir, slug, existing string, err error) {
	zr, size, err := openImportZip(path)
	if err != nil {
		return "", "", "", err
	}
	defer func() { _ = zr.Close() }()
	p, err := planImport(zr, size)
	if err != nil {
		return "", "", "", err
	}

	slug = BookSlug(p.title)
	title := p.title
	if _, serr := os.Stat(filepath.Join(l.BooksRoot(), slug)); serr == nil {
		existing = slug
	}
	if existing != "" && !replace {
		// Giữ cả hai: tên và mã sách theo số thứ tự chưa dùng ("Tên (2)", slug-2).
		for i := 2; i < 1000; i++ {
			s := fmt.Sprintf("%s-%d", slug, i)
			if _, err := os.Stat(filepath.Join(l.BooksRoot(), s)); errors.Is(err, os.ErrNotExist) {
				slug, title = s, clip(fmt.Sprintf("%s (%d)", p.title, i), maxTitleLen+8)
				break
			}
		}
	}

	workDir, err = l.NewWorkDir(slug)
	if err != nil {
		return "", "", "", err
	}
	tmp := workDir // `return "", "", "", err` gán rỗng cho workDir trước khi defer chạy
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tmp)
		}
	}()

	var written int64
	meta := importMeta{Title: title, Author: p.author, Category: p.category, Language: "vi"}
	// Bộ sách: giữ tên bộ (gộp cách viết đã có); trùng số tập với cuốn đang có thì lấy tập kế tiếp.
	except := ""
	if replace {
		except = existing
	}
	if sr, vol, perr := l.placeInSeries(p.series, p.volume, except); perr == nil {
		meta.Series, meta.SeriesVolume = sr, vol
	} else if errors.Is(perr, ErrVolumeTaken) {
		meta.Series, meta.SeriesVolume, _ = l.placeInSeries(p.series, 0, except)
	}
	if p.voice != "" {
		meta.Narrator = "VieNeu-TTS (" + p.voice + ")"
	}
	if p.cover != nil {
		name := "cover" + p.coverExt
		if err = extractEntry(p.cover, filepath.Join(workDir, name), maxImportCoverBytes, &written, isImage); err != nil {
			return "", "", "", err
		}
		meta.Cover = name
	}
	done := 0
	for ci, ch := range p.chapters {
		mc := importMetaChapter{Title: ch.title}
		for si, sec := range ch.sections {
			if err = ctx.Err(); err != nil {
				return "", "", "", err
			}
			file := fmt.Sprintf("ch%02d-sec%02d.mp3", ci+1, si+1)
			if err = extractEntry(sec.audio, filepath.Join(workDir, file), maxImportAudioBytes, &written, isMP3); err != nil {
				return "", "", "", err
			}
			mc.Sections = append(mc.Sections, importMetaSection{Title: sec.title, File: file, OriginalText: sec.original, ReadingScript: sec.script})
			done++
			if progress != nil {
				progress(done, p.sections)
			}
		}
		meta.Chapters = append(meta.Chapters, mc)
	}

	data, err := marshalIndent(meta)
	if err != nil {
		return "", "", "", err
	}
	if err = writeNewFile(filepath.Join(workDir, "metadata.json"), bytes.NewReader(data), int64(len(data))+1, &written); err != nil {
		return "", "", "", err
	}
	if _, err = bookmaker.RepackZip(workDir, filepath.Join(workDir, "book-"+bookmaker.Slugify(title)+".zip"), p.voice, nil); err != nil {
		return "", "", "", fmt.Errorf("đóng gói lại sách: %w", err)
	}
	if err = ctx.Err(); err != nil { // huỷ trong lúc đóng gói lại
		return "", "", "", err
	}
	return workDir, slug, existing, nil
}

// CancelPrepared xoá thư mục tạm của một lượt nhập chưa đưa vào thư viện.
func (l *Library) CancelPrepared(workDir string) {
	if workDir != "" && strings.HasPrefix(filepath.Base(workDir), workPrefix) && safepath.Within(l.BooksRoot(), workDir) {
		_ = os.RemoveAll(workDir)
	}
}

// importMeta — metadata.json của cuốn nhập (cùng khuôn bookmaker outMeta).
type importMeta struct {
	Title    string `json:"title"`
	Author   string `json:"author,omitempty"`
	Narrator string `json:"narrator,omitempty"`
	Language string `json:"language,omitempty"`
	Cover    string `json:"cover,omitempty"`
	Category string `json:"category,omitempty"`
	Series   string `json:"series,omitempty"`
	// SeriesVolume — số tập trong bộ (cùng khoá với bookmaker outMeta).
	SeriesVolume int                 `json:"series_volume,omitempty"`
	Chapters     []importMetaChapter `json:"chapters"`
}

type importMetaChapter struct {
	Title    string              `json:"title"`
	Sections []importMetaSection `json:"sections"`
}

type importMetaSection struct {
	Title         string `json:"title"`
	File          string `json:"file"`
	OriginalText  string `json:"original_text,omitempty"`
	ReadingScript string `json:"reading_script,omitempty"`
}

// openImportZip mở gói: phải là file thường (không symlink, không thiết bị), đuôi
// .zip, không quá lớn. Đọc zip trên CHÍNH file đã kiểm (không mở lại theo đường
// dẫn, tránh bị tráo file giữa lúc kiểm và lúc đọc).
func openImportZip(path string) (*importZip, int64, error) {
	if !strings.EqualFold(filepath.Ext(path), ".zip") {
		return nil, 0, bad("Chỉ nhập được gói sách .zip (file Word thì vào Tạo sách mới).")
	}
	f, err := safepath.OpenRegular(path)
	if err != nil {
		return nil, 0, bad("Không mở được file: %v", err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, bad("Không đọc được file: %v", err)
	}
	if info.Size() > maxImportZipBytes {
		_ = f.Close()
		return nil, 0, bad("Gói quá lớn (%d MB). Sano nhận gói tới %d MB.", info.Size()>>20, maxImportZipBytes>>20)
	}
	zr, err := zip.NewReader(f, info.Size())
	if err != nil {
		_ = f.Close()
		return nil, 0, bad("File không phải gói zip hợp lệ hoặc đã hỏng khi gửi.")
	}
	return &importZip{Reader: zr, f: f}, info.Size(), nil
}

// importZip — zip.Reader kèm file đang mở.
type importZip struct {
	*zip.Reader
	f *os.File
}

func (z *importZip) Close() error { return z.f.Close() }

// planImport đọc và kiểm toàn bộ cấu trúc gói (chưa giải nén mp3).
func planImport(zr *importZip, zipSize int64) (*importPlan, error) {
	if len(zr.File) > maxZipEntries {
		return nil, bad("Gói có quá nhiều file (%d).", len(zr.File))
	}
	byName := map[string]*zip.File{}
	for _, f := range zr.File {
		name := f.Name
		if !utf8.ValidString(name) {
			continue // tên lạ: không bao giờ được dùng, bỏ qua
		}
		if _, dup := byName[name]; dup {
			return nil, bad("Gói có hai file trùng tên %q, có thể đã bị sửa. Không nhập.", clip(name, 80))
		}
		byName[name] = f
	}
	get := func(name string) *zip.File { return byName[name] }

	mf := get("manifest.json")
	cf := get("chapters.json")
	if mf == nil || cf == nil {
		return nil, bad("Gói thiếu manifest.json hoặc chapters.json nên không phải gói sách của Sano, hoặc file bị hỏng khi gửi. Nhờ người gửi xuất lại bằng nút “Xuất gói zip”.")
	}
	mdata, err := readSmallEntry(mf, maxManifestBytes)
	if err != nil {
		return nil, err
	}
	var man struct {
		Title    string `json:"title"`
		Author   string `json:"author"`
		Category string `json:"category"`
		Series   string `json:"series"`
		Volume   int    `json:"series_volume"`
		Cover    string `json:"cover_filename"`
		Voice    string `json:"voice_id"`
		Version  int    `json:"version"`
	}
	if err := json.Unmarshal(mdata, &man); err != nil {
		return nil, bad("manifest.json không đọc được (%v).", err)
	}
	if man.Version != 1 {
		return nil, bad("Gói tạo bằng phiên bản định dạng %d, Sano này chỉ đọc bản 1. Hãy cập nhật Sano.", man.Version)
	}
	p := &importPlan{
		title:    cleanLine(man.Title, maxTitleLen),
		author:   cleanLine(man.Author, maxTitleLen),
		category: NormalizeCategory(cleanLine(man.Category, maxTitleLen)),
		voice:    cleanLine(man.Voice, maxVoiceLen),
		series:   NormalizeSeries(cleanLine(man.Series, maxTitleLen)),
	}
	if p.series != "" && man.Volume > 0 && man.Volume <= MaxVolume {
		p.volume = man.Volume
	}
	if p.title == "" {
		return nil, bad("Gói không có tên sách.")
	}

	cdata, err := readSmallEntry(cf, maxZipJSONBytes)
	if err != nil {
		return nil, err
	}
	var cj struct {
		Chapters []struct {
			Order    int    `json:"order"`
			Title    string `json:"title"`
			Sections []struct {
				Order         int    `json:"order"`
				Title         string `json:"title"`
				Audio         string `json:"audio_filename"`
				DurationSec   int    `json:"duration_sec"`
				OriginalText  string `json:"original_text"`
				ReadingScript string `json:"reading_script"`
			} `json:"sections"`
		} `json:"chapters"`
	}
	if err := json.Unmarshal(cdata, &cj); err != nil {
		return nil, bad("chapters.json không đọc được (%v).", err)
	}
	if len(cj.Chapters) == 0 || len(cj.Chapters) > maxImportChapters {
		return nil, bad("Số chương không hợp lệ (%d).", len(cj.Chapters))
	}
	sort.SliceStable(cj.Chapters, func(i, j int) bool { return cj.Chapters[i].Order < cj.Chapters[j].Order })

	used := map[string]bool{}
	var total uint64
	for _, ch := range cj.Chapters {
		pc := planChapter{title: cleanLine(ch.Title, maxSectionTitleLen)}
		secs := ch.Sections
		sort.SliceStable(secs, func(i, j int) bool { return secs[i].Order < secs[j].Order })
		for _, s := range secs {
			p.sections++
			if p.sections > maxImportSections {
				return nil, bad("Gói có quá nhiều tiểu mục (hơn %d).", maxImportSections)
			}
			if !audioNameRe.MatchString(s.Audio) {
				return nil, bad("Tên file âm thanh lạ trong gói: %q.", clip(s.Audio, 80))
			}
			if used[s.Audio] {
				return nil, bad("Hai tiểu mục dùng chung file %q.", s.Audio)
			}
			used[s.Audio] = true
			af := get(s.Audio)
			if af == nil {
				return nil, bad("Gói thiếu file âm thanh %q.", s.Audio)
			}
			if err := checkEntry(af, maxImportAudioBytes); err != nil {
				return nil, err
			}
			total += af.UncompressedSize64
			if s.DurationSec > 0 && s.DurationSec < 24*3600 {
				p.durationSec += s.DurationSec
			}
			pc.sections = append(pc.sections, planSection{
				title: cleanLine(s.Title, maxSectionTitleLen), original: cleanText(s.OriginalText), script: cleanText(s.ReadingScript), audio: af,
			})
		}
		if len(pc.sections) > 0 {
			p.chapters = append(p.chapters, pc)
		}
	}
	if p.sections == 0 {
		return nil, bad("Gói không có tiểu mục nào.")
	}

	// Bìa: tên theo manifest nếu hợp lệ, không thì cover.* — thiếu / hỏng thì bỏ qua bìa.
	names := coverNames
	if c := strings.ToLower(man.Cover); c != "" && safepath.IsPlainName(c) {
		names = append([]string{man.Cover}, coverNames...)
	}
	for _, n := range names {
		if f := get(n); f != nil && checkEntry(f, maxImportCoverBytes) == nil {
			ext := strings.ToLower(filepath.Ext(n))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" {
				p.cover, p.coverExt = f, ext
				total += f.UncompressedSize64
				break
			}
		}
	}
	if total > uint64(maxImportTotalBytes) {
		return nil, bad("Gói giải nén ra quá lớn (%d MB).", total>>20)
	}
	// Cả gói: mp3/ảnh gần như không nén được → tổng giải nén không thể gấp nhiều lần file zip.
	if total > uint64(zipSize)*maxImportRatio+1<<20 {
		return nil, bad("Gói nén bất thường, có thể là gói độc. Không nhập.")
	}
	// Bom zip chồng lấn: nhiều mục cùng trỏ vào một vùng dữ liệu nén (từng mục vẫn
	// qua kiểm tỉ lệ). Mọi mục dùng tới phải nằm ở vùng riêng, không đè lên nhau.
	usedFiles := []*zip.File{mf, cf}
	if p.cover != nil {
		usedFiles = append(usedFiles, p.cover)
	}
	for _, ch := range p.chapters {
		for _, sec := range ch.sections {
			usedFiles = append(usedFiles, sec.audio)
		}
	}
	if err := checkNoOverlap(usedFiles); err != nil {
		return nil, err
	}
	p.size = int64(total)
	return p, nil
}

// checkNoOverlap: vùng dữ liệu nén của các mục không được chồng lên nhau.
func checkNoOverlap(files []*zip.File) error {
	type span struct{ from, to int64 }
	spans := make([]span, 0, len(files))
	for _, f := range files {
		off, err := f.DataOffset()
		if err != nil {
			return bad("Gói hỏng (%s).", clip(f.Name, 80))
		}
		spans = append(spans, span{off, off + int64(f.CompressedSize64)})
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].from < spans[j].from })
	for i := 1; i < len(spans); i++ {
		if spans[i].from < spans[i-1].to {
			return bad("Gói có các file chồng dữ liệu lên nhau, có thể là gói độc. Không nhập.")
		}
	}
	return nil
}

// checkEntry kiểm một mục trước khi đọc: không mã hoá, không phải thư mục/symlink,
// kích thước và tỉ lệ nén hợp lý.
func checkEntry(f *zip.File, limit int64) error {
	if f.Flags&0x1 != 0 {
		return bad("Gói có file bị mã hoá (%s). Không nhập.", clip(f.Name, 80))
	}
	if m := f.Mode(); m.IsDir() || m&os.ModeSymlink != 0 || m&os.ModeType != 0 {
		return bad("Mục %q trong gói không phải file thường.", clip(f.Name, 80))
	}
	if f.UncompressedSize64 > uint64(limit) {
		return bad("File %q trong gói quá lớn.", clip(f.Name, 80))
	}
	if f.CompressedSize64 > 0 && f.UncompressedSize64/f.CompressedSize64 > maxImportRatio && f.UncompressedSize64 > 1<<20 {
		return bad("File %q nén bất thường, có thể là gói độc. Không nhập.", clip(f.Name, 80))
	}
	return nil
}

func readSmallEntry(f *zip.File, limit int64) ([]byte, error) {
	if f.Flags&0x1 != 0 || f.UncompressedSize64 > uint64(limit) {
		return nil, bad("%s trong gói không hợp lệ hoặc quá lớn.", f.Name)
	}
	// Chữ thật nén cỡ 3–8 lần; JSON rác lặp lại ("[{},{},…]") nén cả nghìn lần và
	// giải mã ra hàng triệu phần tử → tốn bộ nhớ.
	if f.UncompressedSize64 > 1<<20 && f.CompressedSize64 > 0 && f.UncompressedSize64/f.CompressedSize64 > maxJSONRatio {
		return nil, bad("%s trong gói nén bất thường, có thể là gói độc. Không nhập.", f.Name)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, bad("Không đọc được %s trong gói.", f.Name)
	}
	defer func() { _ = rc.Close() }()
	data, err := io.ReadAll(io.LimitReader(rc, limit+1))
	if err != nil || int64(len(data)) > limit {
		return nil, bad("Không đọc được %s trong gói (hỏng hoặc quá lớn).", f.Name)
	}
	return data, nil
}

// extractEntry chép một mục ra file mới dst, đếm byte thật (không tin kích thước
// ghi trong zip), kiểm chữ ký đầu file.
func extractEntry(f *zip.File, dst string, limit int64, written *int64, sniff func([]byte) bool) error {
	if err := checkEntry(f, limit); err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		return bad("Không đọc được %s trong gói.", f.Name)
	}
	defer func() { _ = rc.Close() }()
	head := make([]byte, 12)
	n, _ := io.ReadFull(rc, head)
	if !sniff(head[:n]) {
		return bad("File %q trong gói không đúng định dạng.", clip(f.Name, 80))
	}
	body := io.MultiReader(bytes.NewReader(head[:n]), rc)
	if err := writeNewFile(dst, body, limit, written); err != nil {
		return err
	}
	return nil
}

// writeNewFile tạo file MỚI (O_EXCL, không đi theo symlink có sẵn) và chép tối đa
// limit byte; vượt limit hoặc vượt tổng maxImportTotalBytes → lỗi.
func writeNewFile(dst string, src io.Reader, limit int64, written *int64) error {
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("ghi %s: %w", filepath.Base(dst), err)
	}
	budget := limit
	if left := maxImportTotalBytes - *written; left < budget {
		budget = left
	}
	n, err := io.Copy(out, io.LimitReader(src, budget+1))
	cerr := out.Close()
	*written += n
	if err != nil {
		return bad("Gói hỏng khi giải nén %s: %v", filepath.Base(dst), err)
	}
	if n > budget {
		return bad("Dữ liệu giải nén vượt giới hạn, có thể là gói độc. Không nhập.")
	}
	return cerr
}

func isMP3(b []byte) bool {
	return len(b) >= 3 && (string(b[:3]) == "ID3" || (b[0] == 0xFF && b[1]&0xE0 == 0xE0))
}

func isImage(b []byte) bool {
	switch {
	case len(b) >= 8 && string(b[:8]) == "\x89PNG\r\n\x1a\n":
		return true
	case len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF:
		return true
	case len(b) >= 12 && string(b[:4]) == "RIFF" && string(b[8:12]) == "WEBP":
		return true
	}
	return false
}

// cleanLine: một dòng, bỏ ký tự điều khiển / định dạng vô hình, gộp khoảng trắng, cắt độ dài.
func cleanLine(s string, n int) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return ' '
		}
		return r
	}, strings.ToValidUTF8(s, ""))
	return clip(strings.Join(strings.Fields(s), " "), n)
}

// cleanText: giữ xuống dòng, bỏ ký tự điều khiển khác, cắt độ dài.
func cleanText(s string) string {
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, strings.ToValidUTF8(s, ""))
	return clip(strings.TrimSpace(s), maxTextLen)
}
