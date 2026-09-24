package bookmaker

import (
	"strings"
	"testing"
	"time"
)

func TestListMarkerLines(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"gạch đầu dòng", "- Ý thứ nhất\n- Ý thứ hai", "Ý thứ nhất\nÝ thứ hai"},
		{"chấm tròn", "• Chuẩn bị tài liệu", "Chuẩn bị tài liệu"},
		{"dấu cộng", "+ Kiểm tra lại", "Kiểm tra lại"},
		{"chữ cái", "a) Đọc kỹ đề\nb) Lập dàn ý", "Đọc kỹ đề\nLập dàn ý"},
		{"số có dấu chấm", "1. Đọc kỹ đề\n2. Lập dàn ý", "Thứ nhất, Đọc kỹ đề\nThứ hai, Lập dàn ý"},
		{"số có ngoặc", "4) Viết bản nháp", "Thứ tư, Viết bản nháp"},
		{"số lớn", "11. Tổng kết", "Thứ mười một, Tổng kết"},
		{"đã có Thứ nhất thì không thêm", "1. Thứ nhất, đọc kỹ đề", "Thứ nhất, đọc kỹ đề"},
		{"đã có Đầu tiên thì không thêm", "1. Đầu tiên là chuẩn bị", "Đầu tiên là chuẩn bị"},
		{"dòng kết thúc bằng hai chấm giữ nguyên, ý sau vẫn xuống dòng", "Chuẩn bị như sau:\n\n- Bước một\n\n- Bước hai", "Chuẩn bị như sau:\n\nBước một\n\nBước hai"},
		{"gạch nối giữa câu không đụng", "Khách quen - khách mới", "Khách quen - khách mới"},
		{"số âm không phải danh sách", "-5 độ vào buổi sáng", "-5 độ vào buổi sáng"},
		{"năm đầu dòng không phải danh sách", "2026. Một năm mới", "2026. Một năm mới"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := listMarkerLines(tc.in); got != tc.want {
				t.Errorf("listMarkerLines(%q) = %q, muốn %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestExpandAmpersands(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Nhóm XY&Z họp", "Nhóm XY và Z họp"},
		{"Trà & bánh", "Trà và bánh"},
		{"A&B&C", "A và B và C"},
		{"Cuối câu có ký hiệu &", "Cuối câu có ký hiệu &"},
	}
	for _, tc := range cases {
		if got := expandAmpersands(tc.in); got != tc.want {
			t.Errorf("expandAmpersands(%q) = %q, muốn %q", tc.in, got, tc.want)
		}
	}
}

func TestExpandArrows(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"giữa câu", "Mưa lớn → đường ngập", "Mưa lớn, dẫn tới đường ngập"},
		{"mũi tên đôi", "Làm nhanh ⇒ sai sót", "Làm nhanh, dẫn tới sai sót"},
		{"=>", "Nhập => xuất", "Nhập, dẫn tới xuất"},
		{"->", "Bước một -> bước hai", "Bước một, dẫn tới bước hai"},
		{"chuỗi", "Giai đoạn 1 → 2 → 3", "Giai đoạn 1, dẫn tới 2, dẫn tới 3"},
		{"sau dấu phẩy", "Hàng về chậm, → khách phàn nàn", "Hàng về chậm, dẫn tới khách phàn nàn"},
		{"đầu dòng là ký hiệu đầu ý", "→ Kết quả tốt hơn", "Kết quả tốt hơn"},
		{"sau câu hỏi là câu trả lời", "Ai làm? → Trưởng nhóm.", "Ai làm? Trưởng nhóm."},
		{"sau dấu hai chấm", "Thời hạn: → cuối tháng", "Thời hạn: cuối tháng"},
		{"trước từ nối", "Ngủ muộn, → dẫn đến mệt mỏi", "Ngủ muộn, dẫn đến mệt mỏi"},
		{"trước từ nối không có dấu phẩy", "Làm kỹ → vì vậy ít sai", "Làm kỹ, vì vậy ít sai"},
		{"cuối dòng", "Dòng một →\nDòng hai", "Dòng một\nDòng hai"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := expandArrows(tc.in); got != tc.want {
				t.Errorf("expandArrows(%q) = %q, muốn %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestExpandSlashes(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"hai cụm từ thành dấu phẩy", "Đọc sách/tạp chí mỗi tối", "Đọc sách, tạp chí mỗi tối"},
		{"có khoảng trắng", "mua rau/ trái cây / thịt cá", "mua rau, trái cây, thịt cá"},
		{"cặp lựa chọn", "Chọn có/không", "Chọn có hoặc không"},
		{"cặp lựa chọn tiếng Anh", "Học online/offline", "Học online hoặc offline"},
		{"hai chữ cái", "Phương án A/B", "Phương án A hoặc B"},
		{"và hoặc", "Dùng thẻ và/hoặc tiền mặt", "Dùng thẻ và hoặc tiền mặt"},
		{"lượng trên đơn vị", "Quỹ lương 300 triệu/năm", "Quỹ lương 300 triệu một năm"},
		{"số trên đơn vị", "Lương 15k/giờ", "Lương 15k một giờ"},
		{"phần trăm trên đơn vị", "Lãi 1%/tháng", "Lãi 1% một tháng"},
		{"khoảng số giờ trên ngày", "Học 2-3 giờ/ngày", "Học 2-3 giờ một ngày"},
		{"số trang trên ngày", "Đọc 5 trang/ngày", "Đọc 5 trang một ngày"},
		{"tốc độ", "Chạy 60 km/giờ", "Chạy 60 km một giờ"},
		{"đơn vị sau số nhưng vế phải không phải đơn vị", "Chia 12 người/nhóm", "Chia 12 người, nhóm"},
		{"nhãn ngày tháng năm không có số", "Ghi theo ngày/tháng/năm", "Ghi theo ngày, tháng, năm"},
		{"ngày tháng giữ cho v3", "Ngày 12/09/2026", "Ngày 12/09/2026"},
		{"phân số giữ cho v3", "Khoảng 1/3 số người", "Khoảng 1/3 số người"},
		{"24/7", "Phục vụ 24/7", "Phục vụ 24 7"},
		{"địa chỉ web giữ nguyên", "Xem https://example.com/huong-dan/buoc-1 nhé", "Xem https://example.com/huong-dan/buoc-1 nhé"},
		{"mã có số", "Mô hình B2B/B2C", "Mô hình B2B, B2C"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := expandSlashes(tc.in); got != tc.want {
				t.Errorf("expandSlashes(%q) = %q, muốn %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestExpandSmallRanges(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Mất khoảng 4-6 tháng", "Mất khoảng 4 đến 6 tháng"},
		{"Học trong 2 - 4 năm.", "Học trong 2 đến 4 năm."},
		{"Nhóm 20–30 tuổi", "Nhóm 20 đến 30 tuổi"},
		{"Tăng 5-10%", "Tăng 5 đến 10%"},
		{"Khoảng 2-3, có khi 4-5.", "Khoảng 2 đến 3, có khi 4 đến 5."},
		{"Ngày 12-09-2026", "Ngày 12-09-2026"},
		{"Gọi 090-123-45", "Gọi 090-123-45"},
		{"Hệ số 1.2-1.5", "Hệ số 1.2-1.5"},
		{"Giai đoạn 2020-2025", "Giai đoạn 2020-2025"},
		{"Ngày 5-10/3", "Ngày 5-10/3"},
	}
	for _, tc := range cases {
		if got := expandSmallRanges(tc.in); got != tc.want {
			t.Errorf("expandSmallRanges(%q) = %q, muốn %q", tc.in, got, tc.want)
		}
	}
}

func TestScript_SpokenRulesTogether(t *testing.T) {
	in := "Các bước chính:\n\n1. Đo KPI mỗi tuần\n\n- Rà soát R&D/marketing → điều chỉnh sau 4-6 tháng"
	want := "Các bước chính:\n\nThứ nhất, Đo ca pê i mỗi tuần\n\nRà soát a en đi, marketing, dẫn tới điều chỉnh sau 4 đến 6 tháng"
	if got := normalizeReadingScript(in); got != want {
		t.Errorf("normalizeReadingScript =\n%q\nmuốn\n%q", got, want)
	}
}

func TestScript_StripsInvisibleChars(t *testing.T) {
	in := "\uFEFF\uFEFFDòng có\u200B ký tự ẩn"
	if got := normalizeReadingScript(in); got != "Dòng có ký tự ẩn" {
		t.Errorf("phải bỏ BOM / zero-width: %q", got)
	}
}

// Cả chuỗi dấu "/" rồi khoảng số: "2-3 giờ/ngày" không được thành "2 đến 3 giờ, ngày".
func TestSlashThenRange_GioTrenNgay(t *testing.T) {
	got := expandSmallRanges(expandSlashes("Dành 2-3 giờ/ngày để đọc."))
	if want := "Dành 2 đến 3 giờ một ngày để đọc."; got != want {
		t.Errorf("= %q, muốn %q", got, want)
	}
}

// Dòng rất dài toàn mũi tên (file cố tình tạo) phải xử lý trong thời gian
// tuyến tính — trước đây mỗi mũi tên ghi lại cả dòng, 800 KB mất ~18 giây.
func TestExpandArrowsLongLineLinear(t *testing.T) {
	in := strings.Repeat("a,→", 300_000)
	start := time.Now()
	got := expandArrows(in)
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("expandArrows 900 KB mất %v, muốn < 2s", d)
	}
	if !strings.HasPrefix(got, "a, dẫn tới a, dẫn tới") {
		t.Errorf("đầu kết quả = %q", got[:40])
	}
}
