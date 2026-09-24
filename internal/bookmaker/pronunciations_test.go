package bookmaker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPronunciationDict_Loads(t *testing.T) {
	d := mustDefaultPronunciationDict()
	for _, k := range []string{"KPI", "CEO", "SWOT", "R&D", "P&L", "M&A", "B2B", "MBA", "F&B", "VIP", "PR", "NATO"} {
		if !d.has(k) {
			t.Errorf("từ điển mặc định thiếu %q", k)
		}
	}
}

func TestPronunciationDict_Apply(t *testing.T) {
	d := mustDefaultPronunciationDict()
	cases := []struct {
		name, in, want string
	}{
		{"viết tắt đứng riêng", "Mục tiêu KPI rõ ràng.", "Mục tiêu ca pê i rõ ràng."},
		{"có dấu &", "Bộ phận R&D và P&L.", "Bộ phận a en đi và pê en eo."},
		{"có chữ số", "Mô hình B2B.", "Mô hình bi tu bi."},
		{"hai lần liền nhau", "KPI KPI", "ca pê i ca pê i"},
		{"sát dấu câu", "(CEO), VIP.", "(xi i âu), vi ai pi."},
		{"phân biệt hoa thường", "làm pr cho vui", "làm pr cho vui"},
		{"không khớp trong từ dài hơn", "Chương trình PRO và KPIs.", "Chương trình PRO và KPIs."},
		{"dòng in hoa toàn bộ giữ nguyên", "AI LÀ NGƯỜI QUYẾT ĐỊNH", "AI LÀ NGƯỜI QUYẾT ĐỊNH"},
		{"giữ xuống dòng", "Dùng AI.\n\nĐo KPI.", "Dùng ây ai.\n\nĐo ca pê i."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := d.apply(tc.in); got != tc.want {
				t.Errorf("apply(%q) = %q, muốn %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestLoadPronunciationDict_UserFileOverrides(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "doc.tsv")
	content := "# chú thích\n\nKPI\tchỉ tiêu chính\nXYZ\tích i dét\nVIP\t\n"
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := loadPronunciationDict(f)
	if err != nil {
		t.Fatalf("loadPronunciationDict: %v", err)
	}
	got := d.apply("Đo KPI, đọc XYZ, gặp VIP và CEO.")
	want := "Đo chỉ tiêu chính, đọc ích i dét, gặp VIP và xi i âu."
	if got != want {
		t.Errorf("apply = %q, muốn %q", got, want)
	}
}

func TestParsePronunciations_Errors(t *testing.T) {
	cases := map[string]string{
		"thiếu TAB":        "KPI ca pê i\n",
		"khóa có dấu chấm": "T.P\tthành phố\n",
		"khóa có khoảng":   "A B\ta bê\n",
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parsePronunciations(data, "x.tsv"); err == nil {
				t.Errorf("phải báo lỗi với %q", data)
			} else if !strings.Contains(err.Error(), "dòng 1") {
				t.Errorf("lỗi nên chỉ rõ dòng: %v", err)
			}
		})
	}
}

func TestNormalizeReadingScript_UsesDictionary(t *testing.T) {
	got := normalizeReadingScript("Vai trò của CEO trong mô hình B2B.")
	want := "Vai trò của xi i âu trong mô hình bi tu bi."
	if got != want {
		t.Errorf("normalizeReadingScript = %q, muốn %q", got, want)
	}
}
