package bookmaker

import "testing"

func TestSplitJobsByStems_EmptyOnly_AllReal(t *testing.T) {
	jobs := []ttsJob{{Stem: "ch01-sec01"}, {Stem: "ch01-sec02"}}
	real, silent := splitJobsByStems(jobs, nil)
	if len(real) != 2 {
		t.Errorf("only rỗng phải render thật toàn bộ, got real=%d", len(real))
	}
	if len(silent) != 0 {
		t.Errorf("only rỗng không được có stub, got silent=%d", len(silent))
	}
}

func TestSplitJobsByStems_OneStem_SplitsRealAndSilent(t *testing.T) {
	jobs := []ttsJob{
		{Stem: "ch01-sec01"},
		{Stem: "ch03-sec01"},
		{Stem: "ch03-sec02"},
	}
	only := map[string]bool{"ch03-sec01": true}
	real, silent := splitJobsByStems(jobs, only)
	if len(real) != 1 || real[0].Stem != "ch03-sec01" {
		t.Errorf("real phải đúng 1 stem ch03-sec01, got %+v", real)
	}
	if len(silent) != 2 {
		t.Errorf("silent phải gồm 2 tiểu mục còn lại, got %d", len(silent))
	}
}

func TestSplitJobsByStems_NoMatch_RealEmpty(t *testing.T) {
	jobs := []ttsJob{{Stem: "ch01-sec01"}}
	only := map[string]bool{"ch99-sec99": true}
	real, silent := splitJobsByStems(jobs, only)
	if len(real) != 0 {
		t.Errorf("stem không khớp → real rỗng, got %d", len(real))
	}
	if len(silent) != 1 {
		t.Errorf("toàn bộ job thành stub, got silent=%d", len(silent))
	}
}

func TestTruncatePreview(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{name: "n=0 giữ nguyên", s: "Một hai ba bốn năm.", n: 0, want: "Một hai ba bốn năm."},
		{name: "ngắn hơn n giữ nguyên", s: "Ngắn gọn.", n: 100, want: "Ngắn gọn."},
		{name: "cắt tại cuối câu", s: "Câu một. Câu hai dài hơn nhiều.", n: 12, want: "Câu một."},
		{name: "cắt tại khoảng trắng khi không có dấu câu", s: "một hai ba bốn năm sáu", n: 10, want: "một hai"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncatePreview(tt.s, tt.n)
			if got != tt.want {
				t.Errorf("truncatePreview(%q, %d) = %q, muốn %q", tt.s, tt.n, got, tt.want)
			}
		})
	}
}

func TestParseStemSet(t *testing.T) {
	tests := []struct {
		name string
		csv  string
		want []string // các key kỳ vọng; nil nghĩa là kết quả phải nil
	}{
		{name: "rỗng trả nil", csv: "", want: nil},
		{name: "chỉ khoảng trắng trả nil", csv: "  ,  ", want: nil},
		{name: "một stem", csv: "ch03-sec01", want: []string{"ch03-sec01"}},
		{name: "nhiều stem có khoảng trắng", csv: " ch03-sec01 , ch04-sec02 ", want: []string{"ch03-sec01", "ch04-sec02"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseStemSet(tt.csv)
			if tt.want == nil {
				if got != nil {
					t.Errorf("muốn nil, got %+v", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("số lượng stem khác: got %d muốn %d", len(got), len(tt.want))
			}
			for _, k := range tt.want {
				if !got[k] {
					t.Errorf("thiếu stem %q trong kết quả %+v", k, got)
				}
			}
		})
	}
}
