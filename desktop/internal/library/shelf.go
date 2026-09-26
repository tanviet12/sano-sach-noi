package library

// Kệ sách: bộ sách nhiều tập, đổi tên / xoá danh mục và bộ sách hàng loạt,
// thứ tự "Tự sắp xếp". Bộ sách + số tập ghi trong metadata.json của từng cuốn
// (như danh mục); thứ tự tự sắp xếp là một file nhỏ ở gốc ~/Sano.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MaxSeriesLen — độ dài tối đa (ký tự) của tên bộ sách.
const MaxSeriesLen = 80

// MaxVolume — số tập lớn nhất trong một bộ.
const MaxVolume = 999

// ErrVolumeTaken — số tập đã có cuốn khác trong cùng bộ.
var ErrVolumeTaken = errors.New("số tập này đã có cuốn khác trong bộ")

// NormalizeSeries bỏ khoảng trắng thừa và cắt còn tối đa MaxSeriesLen ký tự.
func NormalizeSeries(s string) string {
	return clip(strings.Join(strings.Fields(s), " "), MaxSeriesLen)
}

// SeriesKey — khoá so sánh tên bộ (không phân biệt hoa thường), dùng cả trong
// file thứ tự.
func SeriesKey(name string) string { return strings.ToLower(NormalizeSeries(name)) }

// PlaceInSeries — placeInSeries cho cuốn sắp tạo (slug rỗng) hoặc đang sửa.
func (l *Library) PlaceInSeries(name string, volume int, slug string) (string, int, error) {
	return l.placeInSeries(name, volume, slug)
}

// placeInSeries chuẩn hoá tên bộ (gộp với cách viết đã có) và số tập cho cuốn
// slug. Không có bộ → ("", 0). Số tập <= 0 → tập kế tiếp của bộ. Trùng số tập
// với cuốn khác → ErrVolumeTaken.
func (l *Library) placeInSeries(name string, volume int, slug string) (string, int, error) {
	name = NormalizeSeries(name)
	if name == "" {
		return "", 0, nil
	}
	books, err := l.List()
	if err != nil {
		return "", 0, err
	}
	key := SeriesKey(name)
	top := 0
	for _, b := range books {
		if b.Slug == slug || SeriesKey(b.Series) != key {
			continue
		}
		name = b.Series // cách viết đã có
		top = max(top, b.Volume)
		if volume > 0 && b.Volume == volume {
			return "", 0, fmt.Errorf("%w (tập %d là %q)", ErrVolumeTaken, volume, b.Title)
		}
	}
	if volume <= 0 {
		volume = top + 1
	}
	if volume > MaxVolume {
		return "", 0, fmt.Errorf("số tập tối đa là %d", MaxVolume)
	}
	return name, volume, nil
}

// Group — một danh mục hoặc bộ sách kèm số cuốn.
type Group struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// SeriesList trả các bộ sách đang có (mỗi tên một lần), xếp theo tên.
func (l *Library) SeriesList() ([]Group, error) {
	books, err := l.List()
	if err != nil {
		return nil, err
	}
	idx := map[string]int{}
	out := []Group{}
	for _, b := range books {
		if b.Series == "" {
			continue
		}
		k := SeriesKey(b.Series)
		if i, ok := idx[k]; ok {
			out[i].Count++
			continue
		}
		idx[k] = len(out)
		out = append(out, Group{Name: b.Series, Count: 1})
	}
	sort.SliceStable(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out, nil
}

// RenameCategory đổi tên danh mục old cho mọi cuốn. Tên mới trùng một danh mục
// khác đã có (không phân biệt hoa thường) thì hai danh mục gộp làm một. Trả số
// cuốn đã đổi.
func (l *Library) RenameCategory(old, name string) (int, error) {
	name = NormalizeCategory(name)
	if name == "" {
		return 0, errors.New("tên danh mục không được để trống")
	}
	books, err := l.List()
	if err != nil {
		return 0, err
	}
	var others []string
	for _, b := range books {
		if b.Category != "" && !strings.EqualFold(b.Category, NormalizeCategory(old)) {
			others = append(others, b.Category)
		}
	}
	name = MergeCategory(name, others)
	return l.eachBook(books, func(b Book) bool { return strings.EqualFold(b.Category, NormalizeCategory(old)) }, func(f *fields) { f.category = name })
}

// DeleteCategory gỡ danh mục khỏi mọi cuốn (sách về "Chưa phân loại").
func (l *Library) DeleteCategory(name string) (int, error) {
	books, err := l.List()
	if err != nil {
		return 0, err
	}
	return l.eachBook(books, func(b Book) bool { return b.Category != "" && strings.EqualFold(b.Category, NormalizeCategory(name)) }, func(f *fields) { f.category = "" })
}

// RenameSeries đổi tên bộ sách cho mọi tập, giữ số tập. Tên mới trùng một bộ
// khác thì báo lỗi (gộp hai bộ dễ trùng số tập — người dùng tự chuyển từng cuốn).
func (l *Library) RenameSeries(old, name string) (int, error) {
	name = NormalizeSeries(name)
	if name == "" {
		return 0, errors.New("tên bộ sách không được để trống")
	}
	books, err := l.List()
	if err != nil {
		return 0, err
	}
	oldKey, newKey := SeriesKey(old), SeriesKey(name)
	for _, b := range books {
		if k := SeriesKey(b.Series); b.Series != "" && k == newKey && k != oldKey {
			return 0, fmt.Errorf("đã có bộ sách %q", b.Series)
		}
	}
	n, err := l.eachBook(books, func(b Book) bool { return b.Series != "" && SeriesKey(b.Series) == oldKey }, func(f *fields) { f.series = name })
	if err == nil && oldKey != newKey {
		err = l.renameOrderKey(seriesOrderKey(old), seriesOrderKey(name))
	}
	return n, err
}

// DeleteSeries tách các tập thành sách lẻ (sách giữ nguyên).
func (l *Library) DeleteSeries(name string) (int, error) {
	books, err := l.List()
	if err != nil {
		return 0, err
	}
	key := SeriesKey(name)
	return l.eachBook(books, func(b Book) bool { return b.Series != "" && SeriesKey(b.Series) == key }, func(f *fields) { f.series, f.volume = "", 0 })
}

// eachBook ghi lại thông tin mọi cuốn khớp match sau khi sửa bằng edit.
func (l *Library) eachBook(books []Book, match func(Book) bool, edit func(*fields)) (int, error) {
	n := 0
	for _, b := range books {
		if !match(b) {
			continue
		}
		f := fieldsOf(b)
		edit(&f)
		if err := l.writeFields(b.Slug, f); err != nil {
			return n, fmt.Errorf("sửa %q: %w", b.Title, err)
		}
		n++
	}
	return n, nil
}

// --- Thứ tự "Tự sắp xếp" ---

// orderFile — file thứ tự ở gốc ~/Sano (ẩn, cạnh thư mục Sach).
const orderFile = ".thu-tu.json"

// maxOrderItems — số mục tối đa trong file thứ tự (chặn file hỏng / quá lớn).
const maxOrderItems = 20000

// Khoá trong file thứ tự: "b:<slug>" cho sách lẻ, "s:<tên bộ thường>" cho bộ sách.
func seriesOrderKey(name string) string { return "s:" + SeriesKey(name) }

type orderDoc struct {
	Version int      `json:"version"`
	Items   []string `json:"items"`
}

// Order trả thứ tự tự sắp xếp đã lưu (danh sách khoá). Chưa có file → rỗng.
func (l *Library) Order() ([]string, error) {
	data, err := readJSONFile(filepath.Join(l.root, orderFile))
	if errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("đọc thứ tự sách: %w", err)
	}
	var d orderDoc
	if err := json.Unmarshal(data, &d); err != nil {
		return []string{}, nil // file hỏng: coi như chưa sắp xếp
	}
	return cleanOrder(d.Items), nil
}

// SetOrder lưu thứ tự tự sắp xếp.
func (l *Library) SetOrder(items []string) error {
	if len(items) > maxOrderItems {
		return fmt.Errorf("quá nhiều mục (%d)", len(items))
	}
	if err := os.MkdirAll(l.root, 0o755); err != nil {
		return err
	}
	out, err := marshalIndent(orderDoc{Version: 1, Items: cleanOrder(items)})
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(l.root, orderFile), out)
}

// cleanOrder giữ khoá hợp lệ ("b:" / "s:" + tối đa 200 ký tự), bỏ trùng.
func cleanOrder(items []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, it := range items {
		if len(out) >= maxOrderItems {
			break
		}
		if (!strings.HasPrefix(it, "b:") && !strings.HasPrefix(it, "s:")) || len(it) < 3 || len([]rune(it)) > 202 || seen[it] {
			continue
		}
		seen[it] = true
		out = append(out, it)
	}
	return out
}

// renameOrderKey đổi khoá bộ sách trong file thứ tự (khi đổi tên bộ).
func (l *Library) renameOrderKey(from, to string) error {
	items, err := l.Order()
	if err != nil || len(items) == 0 {
		return err
	}
	changed := false
	for i, it := range items {
		if it == from {
			items[i] = to
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return l.SetOrder(items)
}
