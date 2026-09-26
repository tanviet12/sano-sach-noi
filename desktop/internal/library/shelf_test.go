package library

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func metaJSON(title, category, series string, volume int) string {
	m := map[string]any{"title": title, "chapters": []any{}}
	if category != "" {
		m["category"] = category
	}
	if series != "" {
		m["series"] = series
		m["series_volume"] = volume
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func getBook(t *testing.T, lib *Library, slug string) Book {
	t.Helper()
	d, err := lib.Get(slug)
	if err != nil {
		t.Fatal(err)
	}
	return d.Book
}

func manifestOf(t *testing.T, lib *Library, slug string) map[string]any {
	t.Helper()
	_, c := readZip(t, filepath.Join(lib.BooksRoot(), slug, "book-"+slug+".zip"))
	var m map[string]any
	if err := json.Unmarshal([]byte(c["manifest.json"]), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestUpdateInfoSeries(t *testing.T) {
	lib := New(t.TempDir())
	makeFullBook(t, lib, "tap-1", metaJSON("Tập 1", "", "Kinh doanh", 1))
	makeFullBook(t, lib, "tap-2", metaJSON("Tập 2", "", "", 0))

	// số tập 0 → tự lấy tập kế tiếp, tên bộ gộp theo cách viết đã có
	d, err := lib.UpdateInfo("tap-2", Info{Title: "Tập 2", Series: "  kinh   DOANH "})
	if err != nil {
		t.Fatal(err)
	}
	if d.Series != "Kinh doanh" || d.Volume != 2 {
		t.Fatalf("got %q tập %d", d.Series, d.Volume)
	}
	if m := manifestOf(t, lib, "tap-2"); m["series"] != "Kinh doanh" || m["series_volume"] != 2.0 {
		t.Fatalf("manifest = %v", m)
	}

	// trùng số tập → lỗi, không ghi gì
	if _, err := lib.UpdateInfo("tap-2", Info{Title: "Tập 2", Series: "Kinh doanh", Volume: 1}); !errors.Is(err, ErrVolumeTaken) {
		t.Fatalf("err = %v, muốn ErrVolumeTaken", err)
	}
	if b := getBook(t, lib, "tap-2"); b.Volume != 2 {
		t.Fatalf("volume đổi dù lỗi: %d", b.Volume)
	}

	// bỏ bộ → xoá cả hai trường
	if _, err := lib.UpdateInfo("tap-2", Info{Title: "Tập 2"}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(lib.BooksRoot(), "tap-2", "metadata.json"))
	var meta map[string]any
	_ = json.Unmarshal(raw, &meta)
	if _, ok := meta["series"]; ok {
		t.Fatalf("còn series: %s", raw)
	}
	if _, ok := meta["series_volume"]; ok {
		t.Fatalf("còn series_volume: %s", raw)
	}
}

func TestRenameAndDeleteCategory(t *testing.T) {
	lib := New(t.TempDir())
	makeFullBook(t, lib, "a", metaJSON("A", "Sức khoẻ", "", 0))
	makeFullBook(t, lib, "b", metaJSON("B", "sức khoẻ", "", 0))
	makeFullBook(t, lib, "c", metaJSON("C", "Kỹ năng", "", 0))

	n, err := lib.RenameCategory("Sức khoẻ", "kỹ NĂNG") // trùng danh mục khác → gộp
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	for _, s := range []string{"a", "b", "c"} {
		if b := getBook(t, lib, s); b.Category != "Kỹ năng" {
			t.Fatalf("%s: %q", s, b.Category)
		}
	}
	if m := manifestOf(t, lib, "a"); m["category"] != "Kỹ năng" || m["category_slug"] != "ky-nang" {
		t.Fatalf("manifest = %v", m)
	}

	// đổi hoa thường của chính nó vẫn được
	if _, err := lib.RenameCategory("Kỹ năng", "KỸ NĂNG"); err != nil {
		t.Fatal(err)
	}
	if b := getBook(t, lib, "c"); b.Category != "KỸ NĂNG" {
		t.Fatalf("got %q", b.Category)
	}

	if _, err := lib.RenameCategory("KỸ NĂNG", "   "); err == nil {
		t.Fatal("tên rỗng phải lỗi")
	}
	n, err = lib.DeleteCategory("kỹ năng")
	if err != nil || n != 3 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if b := getBook(t, lib, "a"); b.Category != "" || b.Title != "A" {
		t.Fatalf("got %+v", b)
	}
}

func TestSeriesRenameDelete(t *testing.T) {
	lib := New(t.TempDir())
	makeFullBook(t, lib, "t1", metaJSON("T1", "", "Bộ A", 1))
	makeFullBook(t, lib, "t2", metaJSON("T2", "", "bộ a", 2))
	makeFullBook(t, lib, "x", metaJSON("X", "", "Bộ B", 1))

	got, _ := lib.SeriesList()
	if want := []Group{{"Bộ A", 2}, {"Bộ B", 1}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SeriesList = %v", got)
	}
	if _, err := lib.RenameSeries("Bộ A", "bộ b"); err == nil {
		t.Fatal("trùng bộ khác phải lỗi")
	}
	if err := lib.SetOrder([]string{"b:x", "s:bộ a"}); err != nil {
		t.Fatal(err)
	}
	n, err := lib.RenameSeries("Bộ A", "Bộ Mới")
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if b := getBook(t, lib, "t2"); b.Series != "Bộ Mới" || b.Volume != 2 {
		t.Fatalf("got %+v", b)
	}
	if o, _ := lib.Order(); !reflect.DeepEqual(o, []string{"b:x", "s:bộ mới"}) {
		t.Fatalf("order = %v", o)
	}
	n, err = lib.DeleteSeries("bộ mới")
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if b := getBook(t, lib, "t1"); b.Series != "" || b.Volume != 0 {
		t.Fatalf("got %+v", b)
	}
}

func TestOrderFile(t *testing.T) {
	lib := New(t.TempDir())
	if o, err := lib.Order(); err != nil || len(o) != 0 {
		t.Fatalf("o=%v err=%v", o, err)
	}
	if err := lib.SetOrder([]string{"b:a", "x:lạ", "b:a", "s:bộ", "b:", ""}); err != nil {
		t.Fatal(err)
	}
	if o, _ := lib.Order(); !reflect.DeepEqual(o, []string{"b:a", "s:bộ"}) {
		t.Fatalf("order = %v", o)
	}
	// file hỏng → coi như chưa sắp xếp
	if err := os.WriteFile(filepath.Join(lib.Root(), orderFile), []byte("{hỏng"), 0o644); err != nil {
		t.Fatal(err)
	}
	if o, err := lib.Order(); err != nil || len(o) != 0 {
		t.Fatalf("o=%v err=%v", o, err)
	}
	if err := lib.SetOrder(make([]string, maxOrderItems+1)); err == nil {
		t.Fatal("quá nhiều mục phải lỗi")
	}
}
