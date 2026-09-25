// Package library quản lý thư mục sách đã tạo trên máy: ~/Sano/Sach/<slug>/.
//
// Mỗi cuốn là một thư mục đầu ra của bookmaker: MP3 từng tiểu mục
// (chNN-secNN.mp3), lời đọc .txt, bìa, metadata.json và gói zip chuẩn
// book-<slug>.zip. Thư viện chỉ đọc các file đó, không có cơ sở dữ liệu riêng.
package library

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DirName — thư mục gốc của Sano trong HOME; BooksDir — thư mục chứa các cuốn.
const (
	DirName  = "Sano"
	BooksDir = "Sach"
	// workPrefix — thư mục đang render dở (ẩn, bỏ qua khi liệt kê).
	workPrefix = ".dang-lam-"
)

// Library trỏ tới thư mục gốc ~/Sano.
type Library struct {
	root string
}

// New tạo Library với thư mục gốc root (thường là ~/Sano).
func New(root string) *Library {
	return &Library{root: root}
}

// EnvHome — biến môi trường đổi thư mục gốc thay cho ~/Sano (dùng khi phát
// triển / kiểm thử với thư viện tạm, không đụng sách thật).
const EnvHome = "SANO_HOME"

// Default: $SANO_HOME nếu có, không thì ~/Sano.
func Default() (*Library, error) {
	if dir := strings.TrimSpace(os.Getenv(EnvHome)); dir != "" {
		return New(dir), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("không tìm thấy thư mục người dùng: %w", err)
	}
	return New(filepath.Join(home, DirName)), nil
}

// Root trả thư mục gốc (~/Sano).
func (l *Library) Root() string { return l.root }

// BooksRoot trả thư mục chứa các cuốn (~/Sano/Sach).
func (l *Library) BooksRoot() string { return filepath.Join(l.root, BooksDir) }

// Book — một cuốn trong thư viện. Các đường dẫn file tương đối với Root().
type Book struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Category    string `json:"category"` // tên danh mục; trống = chưa phân loại
	Cover       string `json:"cover"`    // tương đối với Root(); trống nếu không có
	Zip         string `json:"zip"`      // tương đối với Root(); trống nếu không có
	Chapters    int    `json:"chapters"`
	Sections    int    `json:"sections"`
	DurationSec int    `json:"durationSec"`
	Voice       string `json:"voice"`     // giọng đọc (manifest.voice_id trong gói zip); trống nếu không rõ
	CreatedAt   string `json:"createdAt"` // RFC3339
}

// Track — một tiểu mục phát được.
type Track struct {
	Chapter     string `json:"chapter"`
	Title       string `json:"title"`
	File        string `json:"file"` // tương đối với Root()
	DurationSec int    `json:"durationSec"`
}

// Detail — cuốn sách kèm danh sách tiểu mục để phát.
type Detail struct {
	Book
	Tracks []Track `json:"tracks"`
}

// metadata — phần metadata.json (bookmaker outMeta) thư viện cần đọc.
type metadata struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Category string `json:"category"`
	Cover    string `json:"cover"`
	Chapters []struct {
		Title    string `json:"title"`
		Sections []struct {
			Title string `json:"title"`
			File  string `json:"file"`
		} `json:"sections"`
	} `json:"chapters"`
}

// ErrNotFound — không có cuốn với slug này.
var ErrNotFound = errors.New("không tìm thấy sách")

// List liệt kê các cuốn đã tạo, mới nhất trước. Thư mục chưa có → danh sách rỗng.
func (l *Library) List() ([]Book, error) {
	entries, err := os.ReadDir(l.BooksRoot())
	if errors.Is(err, os.ErrNotExist) {
		return []Book{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("đọc thư viện: %w", err)
	}
	books := []Book{}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		d, err := l.Get(e.Name())
		if err != nil {
			continue // thư mục lạ / hỏng: bỏ qua, không làm hỏng cả thư viện
		}
		books = append(books, d.Book)
	}
	sort.SliceStable(books, func(i, j int) bool { return books[i].CreatedAt > books[j].CreatedAt })
	return books, nil
}

// Size — tổng dung lượng file thường trong thư mục sách (không đi theo symlink).
func (l *Library) Size() (int64, error) {
	var total int64
	err := filepath.WalkDir(l.BooksRoot(), func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.Type().IsRegular() {
			if info, err := d.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total, err
}

// Get đọc một cuốn theo slug.
func (l *Library) Get(slug string) (*Detail, error) {
	if !validSlug(slug) {
		return nil, ErrNotFound
	}
	dir := filepath.Join(l.BooksRoot(), slug)
	data, err := readJSONFile(filepath.Join(dir, "metadata.json"))
	if errors.Is(err, errJSONTooLarge) || errors.Is(err, errNotRegular) {
		return nil, fmt.Errorf("đọc metadata.json của %q: %w", slug, err)
	}
	if err != nil {
		return nil, ErrNotFound
	}
	var m metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("đọc metadata.json của %q: %w", slug, err)
	}
	rel := func(name string) string { return filepath.ToSlash(filepath.Join(BooksDir, slug, name)) }

	d := &Detail{Book: Book{Slug: slug, Title: m.Title, Author: m.Author, Category: NormalizeCategory(m.Category), Chapters: len(m.Chapters)}}
	if info, err := os.Stat(dir); err == nil {
		d.CreatedAt = info.ModTime().UTC().Format(time.RFC3339)
	}
	if m.Cover != "" && fileExists(filepath.Join(dir, m.Cover)) {
		d.Cover = rel(m.Cover)
	}
	zipName := findZip(dir)
	durations := map[string]int{}
	if zipName != "" {
		d.Zip = rel(zipName)
		durations, d.Voice = zipInfo(filepath.Join(dir, zipName))
	}
	for ci, ch := range m.Chapters {
		for si, sec := range ch.Sections {
			dur := durations[strconv.Itoa(ci+1)+"/"+strconv.Itoa(si+1)]
			d.Tracks = append(d.Tracks, Track{Chapter: ch.Title, Title: sec.Title, File: rel(sec.File), DurationSec: dur})
			d.DurationSec += dur
		}
	}
	d.Sections = len(d.Tracks)
	if d.Tracks == nil {
		d.Tracks = []Track{}
	}
	return d, nil
}

// Dir trả thư mục của một cuốn.
func (l *Library) Dir(slug string) (string, error) {
	if !validSlug(slug) {
		return "", ErrNotFound
	}
	dir := filepath.Join(l.BooksRoot(), slug)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "", ErrNotFound
	}
	return dir, nil
}

// NewWorkDir tạo thư mục render tạm (ẩn) trong ~/Sano/Sach cho slug.
func (l *Library) NewWorkDir(slug string) (string, error) {
	if err := os.MkdirAll(l.BooksRoot(), 0o755); err != nil {
		return "", fmt.Errorf("tạo thư mục thư viện: %w", err)
	}
	dir, err := os.MkdirTemp(l.BooksRoot(), workPrefix+slug+"-")
	if err != nil {
		return "", fmt.Errorf("tạo thư mục render: %w", err)
	}
	return dir, nil
}

// Commit chuyển thư mục render xong vào thư viện dưới slug chưa dùng (slug,
// slug-2, slug-3...). Không bao giờ ghi đè cuốn đã có. Trả slug đã dùng.
func (l *Library) Commit(workDir, slug string) (string, error) {
	for i := 1; i < 1000; i++ {
		name := slug
		if i > 1 {
			name = slug + "-" + strconv.Itoa(i)
		}
		dst := filepath.Join(l.BooksRoot(), name)
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		if err := os.Rename(workDir, dst); err != nil {
			return "", fmt.Errorf("lưu sách vào thư viện: %w", err)
		}
		// MkdirTemp tạo quyền 0700; thư mục sách để như thư mục thường.
		if err := os.Chmod(dst, 0o755); err != nil {
			return "", fmt.Errorf("đặt quyền thư mục sách: %w", err)
		}
		return name, nil
	}
	return "", fmt.Errorf("quá nhiều cuốn trùng tên %q", slug)
}

// Resolve đổi đường dẫn tương đối với Root() (dạng a/b/c) thành đường dẫn thật,
// chặn thoát ra ngoài Root().
func (l *Library) Resolve(rel string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(rel, "/")))
	if clean == "." || strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", errors.New("đường dẫn không hợp lệ")
	}
	full := filepath.Join(l.root, clean)
	r, err := filepath.Rel(l.root, full)
	if err != nil || strings.HasPrefix(r, "..") {
		return "", errors.New("đường dẫn không hợp lệ")
	}
	// Kiểm theo chữ chưa đủ: symlink trong thư mục sách (người khác gửi) có thể
	// trỏ ra ngoài ~/Sano. Giải symlink cả file lẫn root rồi kiểm lại. File chưa
	// có → trả đường dẫn như cũ (nơi gọi tự kiểm tồn tại).
	real, err := filepath.EvalSymlinks(full)
	if errors.Is(err, os.ErrNotExist) {
		return full, nil
	}
	if err != nil {
		return "", errors.New("đường dẫn không hợp lệ")
	}
	realRoot, err := filepath.EvalSymlinks(l.root)
	if err != nil {
		return "", errors.New("đường dẫn không hợp lệ")
	}
	if r, err := filepath.Rel(realRoot, real); err != nil || !filepath.IsLocal(r) {
		return "", errors.New("đường dẫn không hợp lệ")
	}
	return full, nil
}

// Rel đổi đường dẫn thật trong Root() thành dạng tương đối a/b/c.
func (l *Library) Rel(full string) (string, error) {
	r, err := filepath.Rel(l.root, full)
	if err != nil || strings.HasPrefix(r, "..") {
		return "", fmt.Errorf("%q nằm ngoài thư mục Sano", full)
	}
	return filepath.ToSlash(r), nil
}

// bundleExts — đuôi thư mục mà macOS coi là gói (ứng dụng, trình cài...):
// "mở" thư mục như vậy là chạy nó. Thư mục sách không bao giờ mang các đuôi này.
var bundleExts = map[string]bool{
	".app": true, ".pkg": true, ".mpkg": true, ".workflow": true, ".prefpane": true,
	".bundle": true, ".plugin": true, ".kext": true, ".appex": true, ".framework": true,
	".component": true, ".action": true, ".saver": true, ".qlgenerator": true, ".mdimporter": true,
}

// validSlug nhận tên thư mục sách. Slug do app tạo chỉ gồm a-z, 0-9, "-";
// thư mục người khác gửi thì chặn thêm: ẩn, chứa dấu phân cách, đuôi gói
// macOS, dấu phẩy (explorer tách tham số theo dấu phẩy) và đuôi .{CLSID}
// (Windows coi là thư mục đặc biệt của shell).
func validSlug(s string) bool {
	if s == "" || strings.HasPrefix(s, ".") || strings.ContainsAny(s, `/\,`) || s != filepath.Base(s) {
		return false
	}
	ext := strings.ToLower(filepath.Ext(s))
	if bundleExts[ext] {
		return false
	}
	if strings.HasPrefix(ext, ".{") && strings.HasSuffix(ext, "}") {
		return false
	}
	return true
}

func findZip(dir string) string {
	matches, _ := filepath.Glob(filepath.Join(dir, "book-*.zip"))
	if len(matches) == 0 {
		return ""
	}
	return filepath.Base(matches[0])
}

// zipDurations đọc thời lượng từng tiểu mục từ chapters.json trong gói zip,
// khoá "chương/tiểu mục" (đánh số từ 1).
func zipDurations(zipPath string) map[string]int {
	out, _ := zipInfo(zipPath)
	return out
}

// zipInfo đọc trong một lần mở gói zip: thời lượng từng tiểu mục (xem
// zipDurations) và giọng đọc (manifest.json → voice_id).
func zipInfo(zipPath string) (map[string]int, string) {
	out := map[string]int{}
	voice := ""
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return out, voice
	}
	defer func() { _ = zr.Close() }()
	if len(zr.File) > maxZipEntries {
		return out, voice // gói lạ: quá nhiều mục, bỏ qua
	}
	// Chỉ đọc mục đầu tiên mỗi tên: gói độc có thể chứa hàng nghìn mục trùng tên
	// cùng trỏ vào một luồng nén → giải nén lặp gần như vô hạn.
	seenManifest, seenChapters := false, false
	for _, f := range zr.File {
		if f.Name == "manifest.json" && !seenManifest {
			seenManifest = true
			if data, err := readZipJSON(f); err == nil {
				var man struct {
					VoiceID string `json:"voice_id"`
				}
				if json.Unmarshal(data, &man) == nil {
					voice = strings.TrimSpace(man.VoiceID)
				}
			}
			continue
		}
		if f.Name != "chapters.json" || seenChapters {
			continue
		}
		seenChapters = true
		data, err := readZipJSON(f)
		if err != nil {
			continue // quá lớn / hỏng: bỏ qua thời lượng, không làm hỏng cả thư viện
		}
		var cj struct {
			Chapters []struct {
				Order    int `json:"order"`
				Sections []struct {
					Order       int `json:"order"`
					DurationSec int `json:"duration_sec"`
				} `json:"sections"`
			} `json:"chapters"`
		}
		if err := json.Unmarshal(data, &cj); err != nil {
			continue
		}
		for _, ch := range cj.Chapters {
			for _, s := range ch.Sections {
				out[strconv.Itoa(ch.Order)+"/"+strconv.Itoa(s.Order)] = s.DurationSec
			}
		}
	}
	return out, voice
}

// maxZipJSONBytes — dung lượng tối đa khi đọc chapters.json / manifest.json
// trong gói zip (chống gói "bom nén"). Biến để test đặt nhỏ.
var maxZipJSONBytes int64 = 8 << 20

// errZipJSONTooLarge — entry JSON trong gói zip vượt maxZipJSONBytes.
var errZipJSONTooLarge = errors.New("quá lớn")

// maxZipEntries — gói zip sách có nhiều mục hơn thì coi là gói lạ, không lặp
// qua (một cuốn thật chỉ vài trăm mục: MP3 từng tiểu mục + vài file JSON).
var maxZipEntries = 10000

// maxMetadataBytes — dung lượng tối đa của metadata.json (cùng mức với JSON
// trong gói zip). Biến riêng để test đặt nhỏ độc lập.
var maxMetadataBytes = maxZipJSONBytes

// errJSONTooLarge — file JSON trong thư mục sách vượt maxMetadataBytes.
var errJSONTooLarge = errors.New("file JSON quá lớn")

// errNotRegular — không phải file thường (symlink, thiết bị, FIFO...).
var errNotRegular = errors.New("không phải file thường")

// readJSONFile đọc một file JSON trong thư mục sách (metadata.json), tối đa
// maxMetadataBytes. Chỉ nhận file thường, không đi theo symlink: thư mục sách
// người khác gửi có thể chứa symlink tới /dev/zero hoặc file khổng lồ.
func readJSONFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errNotRegular
	}
	if info.Size() > maxMetadataBytes {
		return nil, errJSONTooLarge
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, maxMetadataBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxMetadataBytes {
		return nil, errJSONTooLarge
	}
	return data, nil
}

// readZipJSON đọc một entry JSON trong gói zip, tối đa maxZipJSONBytes.
func readZipJSON(f *zip.File) ([]byte, error) {
	if f.UncompressedSize64 > uint64(maxZipJSONBytes) {
		return nil, errZipJSONTooLarge
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	data, err := io.ReadAll(io.LimitReader(rc, maxZipJSONBytes+1))
	if int64(len(data)) > maxZipJSONBytes {
		return nil, errZipJSONTooLarge
	}
	return data, err
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// SectionText — chữ một tiểu mục cho chế độ xem lời: Text để hiện (original_text,
// trống thì lấy lời đọc), Script là lời đã đọc thật (reading_script: có tên tiểu
// mục ở đầu, số viết thành chữ) để ước lượng thời điểm từng câu.
type SectionText struct {
	Text   string `json:"text"`
	Script string `json:"script"`
}

// Texts trả chữ của từng tiểu mục theo đúng thứ tự Detail.Tracks (chữ chạy theo
// ở màn nghe). Không có gói zip / tiểu mục không có chữ → phần tử rỗng.
func (l *Library) Texts(slug string) ([]SectionText, error) {
	d, err := l.Get(slug)
	if err != nil {
		return nil, err
	}
	out := make([]SectionText, len(d.Tracks))
	if d.Zip == "" {
		return out, nil
	}
	texts := zipTexts(filepath.Join(l.root, filepath.FromSlash(d.Zip)))
	dir, _ := l.Dir(slug)
	data, err := readJSONFile(filepath.Join(dir, "metadata.json"))
	if err != nil {
		return out, nil
	}
	var m metadata
	if json.Unmarshal(data, &m) != nil {
		return out, nil
	}
	i := 0
	for ci, ch := range m.Chapters {
		for si := range ch.Sections {
			if i < len(out) {
				out[i] = texts[strconv.Itoa(ci+1)+"/"+strconv.Itoa(si+1)]
			}
			i++
		}
	}
	return out, nil
}

// zipTexts đọc chữ từng tiểu mục trong chapters.json, khoá "chương/tiểu mục".
func zipTexts(zipPath string) map[string]SectionText {
	out := map[string]SectionText{}
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return out
	}
	defer func() { _ = zr.Close() }()
	if len(zr.File) > maxZipEntries {
		return out
	}
	for _, f := range zr.File {
		if f.Name != "chapters.json" {
			continue
		}
		data, err := readZipJSON(f) // chỉ mục đầu tiên (xem zipInfo)
		if err != nil {
			return out
		}
		var cj struct {
			Chapters []struct {
				Order    int `json:"order"`
				Sections []struct {
					Order         int    `json:"order"`
					OriginalText  string `json:"original_text"`
					ReadingScript string `json:"reading_script"`
				} `json:"sections"`
			} `json:"chapters"`
		}
		if json.Unmarshal(data, &cj) != nil {
			return out
		}
		for _, ch := range cj.Chapters {
			for _, s := range ch.Sections {
				st := SectionText{Text: strings.TrimSpace(s.OriginalText), Script: strings.TrimSpace(s.ReadingScript)}
				if st.Text == "" {
					st.Text = st.Script
				}
				out[strconv.Itoa(ch.Order)+"/"+strconv.Itoa(s.Order)] = st
			}
		}
		break
	}
	return out
}
