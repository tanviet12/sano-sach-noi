package bookmaker

import (
	"bytes"
	"strings"
	"testing"
)

const testDocXML = `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>
<w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>Chương một</w:t></w:r></w:p>
<w:p><w:pPr><w:rPr><w:b/></w:rPr></w:pPr><w:r><w:rPr><w:sz w:val="26"/></w:rPr><w:t>Đoạn thân bài bình thường, dấu đoạn in đậm không tính.</w:t></w:r></w:p>
<w:p><w:r><w:rPr><w:b/><w:sz w:val="26"/></w:rPr><w:t>Tiêu đề gõ tay in đậm</w:t></w:r></w:p>
<w:p><w:r><w:rPr><w:sz w:val="36"/></w:rPr><w:t>Tiêu đề chữ to</w:t></w:r></w:p>
<w:p><w:r><w:rPr><w:b/><w:sz w:val="26"/></w:rPr><w:t>Nhãn in đậm:</w:t></w:r><w:r><w:rPr><w:sz w:val="26"/></w:rPr><w:t> phần giải thích thường.</w:t></w:r></w:p>
<w:p><w:r><w:rPr><w:b w:val="0"/><w:sz w:val="26"/></w:rPr><w:t>Không in đậm vì val bằng 0</w:t></w:r></w:p>
<w:tbl><w:tr><w:tc><w:p><w:r><w:rPr><w:sz w:val="26"/></w:rPr><w:t>Ô một của bảng thứ nhất.</w:t></w:r></w:p></w:tc></w:tr></w:tbl>
<w:tbl><w:tr><w:tc><w:p><w:r><w:rPr><w:sz w:val="26"/></w:rPr><w:t>Ô của bảng thứ hai.</w:t></w:r></w:p></w:tc></w:tr></w:tbl>
</w:body></w:document>`

func TestParseDocumentXML_StatsAndFakeHeadings(t *testing.T) {
	paras, tables, err := parseDocumentXML([]byte(testDocXML))
	if err != nil {
		t.Fatalf("parseDocumentXML: %v", err)
	}
	if tables != 2 {
		t.Errorf("số bảng = %d, muốn 2", tables)
	}
	book := buildBook(paras, nil, nil)
	want := []string{"Tiêu đề gõ tay in đậm", "Tiêu đề chữ to"}
	got := book.Stats.FakeHeadings
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("FakeHeadings = %q, muốn %q", got, want)
	}
}

func TestUnknownAcronyms(t *testing.T) {
	d := mustDefaultPronunciationDict()
	texts := []string{
		"Nhóm XYZ họp với ca pê i. XYZ báo cáo QRS.",
		"Chương II bàn về thế kỷ XXI và bi tu bi.",
		"BẢN TIN XYZ ĐẶC BIỆT HÔM NAY", // dòng in hoa toàn bộ: bỏ qua
		"Dùng KPI và AB&C.",
	}
	got := unknownAcronyms(texts, d)
	want := []acronymCount{{"XYZ", 2}, {"AB&C", 1}, {"QRS", 1}}
	if len(got) != len(want) {
		t.Fatalf("unknownAcronyms = %+v, muốn %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %+v, muốn %+v", i, got[i], want[i])
		}
	}
}

func TestPrintLoadWarnings(t *testing.T) {
	var buf bytes.Buffer
	st := DocStats{Images: 3, Tables: 1, FakeHeadings: []string{"A", "B", "C", "D"}}
	printLoadWarnings(&buf, st, []acronymCount{{"XYZ", 2}, {"QRS", 1}})
	out := buf.String()
	for _, s := range []string{"3 hình", "1 bảng", "4 đoạn in đậm", "«A», «B», «C»", "2 từ viết hoa liền", "(3 lần)", "XYZ ×2, QRS ×1", "--pronunciations"} {
		if !strings.Contains(out, s) {
			t.Errorf("cảnh báo thiếu %q:\n%s", s, out)
		}
	}
	if strings.Contains(out, "«D»") {
		t.Errorf("chỉ in tối đa %d ví dụ:\n%s", warnExamples, out)
	}
}
