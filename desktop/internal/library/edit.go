package library

// Sửa thông tin hiển thị của một cuốn (tên, tác giả, danh mục, bộ sách): ghi lại
// metadata.json và manifest.json trong gói zip. Không đổi thư mục/slug, không
// đụng MP3 — lời giới thiệu đã đọc giữ nguyên.

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"sano/internal/bookmaker"
)

// MaxCategoryLen — độ dài tối đa (ký tự) của tên danh mục.
const MaxCategoryLen = 40

// maxTitleLen — độ dài tối đa (ký tự) của tên sách / tác giả khi sửa.
const maxTitleLen = 200

// Info — thông tin hiển thị sửa được của một cuốn.
type Info struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Category string `json:"category"`
	Series   string `json:"series"` // tên bộ sách; trống = sách lẻ
	Volume   int    `json:"volume"` // số tập; <= 0 kèm Series = tự lấy số tập kế tiếp
}

// ErrEmptyTitle — tên sách để trống.
var ErrEmptyTitle = errors.New("tên sách không được để trống")

// NormalizeCategory bỏ khoảng trắng thừa (đầu, cuối, giữa) và cắt còn tối đa
// MaxCategoryLen ký tự.
func NormalizeCategory(s string) string {
	return clip(strings.Join(strings.Fields(s), " "), MaxCategoryLen)
}

// MergeCategory chuẩn hoá tên danh mục; trùng (không phân biệt hoa thường) với
// một danh mục đã có thì dùng đúng cách viết đã có.
func MergeCategory(name string, existing []string) string {
	name = NormalizeCategory(name)
	if name == "" {
		return ""
	}
	for _, e := range existing {
		if strings.EqualFold(NormalizeCategory(e), name) {
			return NormalizeCategory(e)
		}
	}
	return name
}

// Categories trả các danh mục đang dùng trong thư viện (mỗi tên một lần),
// bỏ qua cuốn exceptSlug (cuốn đang sửa).
func (l *Library) Categories(exceptSlug string) []string {
	books, err := l.List()
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	out := []string{}
	for _, b := range books {
		if b.Slug == exceptSlug || b.Category == "" {
			continue
		}
		k := strings.ToLower(b.Category)
		if !seen[k] {
			seen[k] = true
			out = append(out, b.Category)
		}
	}
	return out
}

// CanonicalCategory chuẩn hoá tên danh mục theo các danh mục đã có trong thư
// viện (gộp trùng không phân biệt hoa thường).
func (l *Library) CanonicalCategory(name, exceptSlug string) string {
	if NormalizeCategory(name) == "" {
		return ""
	}
	return MergeCategory(name, l.Categories(exceptSlug))
}

// UpdateInfo sửa tên, tác giả, danh mục, bộ sách của một cuốn: cập nhật
// manifest.json trong gói zip (viết lại zip an toàn, giữ nguyên mọi file khác)
// rồi metadata.json. Thứ tự "Mới tạo nhất" giữ nguyên.
func (l *Library) UpdateInfo(slug string, in Info) (*Detail, error) {
	if _, err := l.Dir(slug); err != nil {
		return nil, err
	}
	title := clip(strings.TrimSpace(in.Title), maxTitleLen)
	if title == "" {
		return nil, ErrEmptyTitle
	}
	series, volume, err := l.placeInSeries(in.Series, in.Volume, slug)
	if err != nil {
		return nil, err
	}
	f := fields{
		title:    title,
		author:   clip(strings.TrimSpace(in.Author), maxTitleLen),
		category: l.CanonicalCategory(in.Category, slug),
		series:   series,
		volume:   volume,
	}
	if err := l.writeFields(slug, f); err != nil {
		return nil, err
	}
	return l.Get(slug)
}

// fields — thông tin hiển thị ghi vào metadata.json + manifest.json.
type fields struct {
	title, author, category, series string
	volume                          int
}

func fieldsOf(b Book) fields {
	return fields{title: b.Title, author: b.Author, category: b.Category, series: b.Series, volume: b.Volume}
}

// writeFields ghi f vào manifest.json trong gói zip rồi metadata.json, giữ giờ
// thư mục (thứ tự "Mới tạo nhất").
func (l *Library) writeFields(slug string, f fields) error {
	dir, err := l.Dir(slug)
	if err != nil {
		return err
	}
	mtime := dirModTime(dir)

	metaPath := filepath.Join(dir, "metadata.json")
	raw, err := readJSONFile(metaPath)
	if err != nil {
		return fmt.Errorf("đọc metadata.json: %w", err)
	}
	var meta map[string]json.RawMessage
	if err := json.Unmarshal(raw, &meta); err != nil {
		return fmt.Errorf("đọc metadata.json: %w", err)
	}
	apply := func(m map[string]json.RawMessage) {
		setJSON(m, "title", f.title)
		setJSON(m, "author", f.author)
		setJSON(m, "category", f.category)
		setJSON(m, "series", f.series)
		setInt(m, "series_volume", f.volume)
	}
	apply(meta)

	if z := findZip(dir); z != "" {
		err := rewriteZipEntry(filepath.Join(dir, z), "manifest.json", func(data []byte) ([]byte, error) {
			var man map[string]json.RawMessage
			if err := json.Unmarshal(data, &man); err != nil {
				return nil, fmt.Errorf("đọc manifest.json trong gói zip: %w", err)
			}
			apply(man)
			setJSON(man, "category_slug", bookmaker.CategorySlug(f.category))
			return marshalIndent(man)
		})
		if err != nil {
			return err
		}
	}

	out, err := marshalIndent(meta)
	if err != nil {
		return err
	}
	if err := writeFileAtomic(metaPath, out); err != nil {
		return fmt.Errorf("ghi metadata.json: %w", err)
	}
	if !mtime.IsZero() {
		_ = os.Chtimes(dir, mtime, mtime) // ghi file tạm làm đổi giờ thư mục → trả lại để giữ thứ tự
	}
	return nil
}

// setInt gán số v cho khoá k; v <= 0 thì xoá khoá.
func setInt(m map[string]json.RawMessage, k string, v int) {
	if v <= 0 {
		delete(m, k)
		return
	}
	b, _ := json.Marshal(v)
	m[k] = b
}

// setJSON gán chuỗi v cho khoá k; v rỗng thì xoá khoá (như omitempty).
func setJSON(m map[string]json.RawMessage, k, v string) {
	if v == "" {
		delete(m, k)
		return
	}
	b, _ := json.Marshal(v) // chuỗi luôn marshal được
	m[k] = b
}

func marshalIndent(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// rewriteZipEntry viết lại gói zip với entry name đã sửa qua edit; mọi entry
// khác chép nguyên (không nén lại). Ghi ra file tạm cùng thư mục rồi đổi tên,
// lỗi giữa chừng thì zip cũ còn nguyên. Zip không có entry name → không làm gì.
func rewriteZipEntry(zipPath, name string, edit func([]byte) ([]byte, error)) (err error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("mở gói zip: %w", err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = zr.Close()
		}
	}()

	if len(zr.File) > maxZipEntries {
		return fmt.Errorf("gói zip có quá nhiều mục (%d)", len(zr.File))
	}
	var target *zip.File
	for _, f := range zr.File {
		if f.Name == name {
			target = f
			break
		}
	}
	if target == nil {
		return nil
	}
	old, err := readZipJSON(target)
	if err != nil {
		return fmt.Errorf("đọc %s trong gói zip: %w", name, err)
	}
	data, err := edit(old)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(zipPath), ".sua-*.zip")
	if err != nil {
		return fmt.Errorf("tạo file tạm: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	zw := zip.NewWriter(tmp)
	for _, f := range zr.File {
		if f != target {
			if err = zw.Copy(f); err != nil {
				return fmt.Errorf("chép %s: %w", f.Name, err)
			}
			continue
		}
		var w io.Writer
		w, err = zw.CreateHeader(&zip.FileHeader{Name: f.Name, Method: f.Method, Modified: time.Now()})
		if err != nil {
			return fmt.Errorf("ghi %s: %w", name, err)
		}
		if _, err = w.Write(data); err != nil {
			return fmt.Errorf("ghi %s: %w", name, err)
		}
	}
	if err = zw.Close(); err != nil {
		return fmt.Errorf("đóng gói zip: %w", err)
	}
	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("ghi gói zip: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("ghi gói zip: %w", err)
	}
	if info, statErr := os.Stat(zipPath); statErr == nil {
		_ = os.Chmod(tmpName, info.Mode().Perm())
	}
	closed = true
	if err = zr.Close(); err != nil { // Windows không đổi tên đè lên file đang mở
		return fmt.Errorf("đóng gói zip cũ: %w", err)
	}
	if err = os.Rename(tmpName, zipPath); err != nil {
		return fmt.Errorf("thay gói zip: %w", err)
	}
	return nil
}

// writeFileAtomic ghi file tạm cùng thư mục rồi đổi tên, giữ quyền file cũ.
func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".sua-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	perm := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	}
	_ = os.Chmod(name, perm)
	if err := os.Rename(name, path); err != nil {
		_ = os.Remove(name)
		return err
	}
	return nil
}

func clip(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return strings.TrimSpace(string([]rune(s)[:n]))
}

func dirModTime(dir string) time.Time {
	info, err := os.Stat(dir)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
