package bookmaker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writeSample ghi docx mẫu (docxgen) vào thư mục tạm.
func writeSample(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "sach-mau.docx")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteSampleDocx(f); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return p
}

func needFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("máy không có ffmpeg")
	}
}

func stubOpts(t *testing.T, docx string) Options {
	t.Helper()
	return Options{
		InputDocx: docx,
		OutputDir: t.TempDir(),
		TTS:       TTSConfig{Mode: TTSModeStub, FFmpeg: "ffmpeg", Bitrate: "64k", StubSec: 1, Voice: DefaultVoice},
	}
}

func TestInspect_SampleDocx_ReturnsOutlineAndWarnings(t *testing.T) {
	o, err := Inspect(writeSample(t), InspectOptions{})
	if err != nil {
		t.Fatalf("Inspect lỗi: %v", err)
	}
	if len(o.Chapters) == 0 || o.Sections == 0 || o.Chars == 0 {
		t.Fatalf("mục lục rỗng: %+v", o)
	}
	first := o.Chapters[0].Sections[0]
	if first.Stem != "ch01-sec01" {
		t.Errorf("stem đầu = %q, muốn ch01-sec01", first.Stem)
	}
	if o.Warnings.Images != 1 {
		t.Errorf("docx mẫu có 1 hình, got %d", o.Warnings.Images)
	}
	if o.Warnings.UnknownAcronyms == nil || o.Warnings.FakeHeadings == nil {
		t.Error("danh sách cảnh báo phải là mảng rỗng, không nil (giao diện đọc .length)")
	}
	if o.SampleSentence == "" {
		t.Error("thiếu câu nghe mẫu")
	}
}

func TestInspect_TOCSection_MarkedNotCounted(t *testing.T) {
	sec := Section{Title: "Mục lục", Text: "Mở đầu 3\nThân bài 9\nKết luận 15\nPhụ lục 20"}
	if tocReason(sec) == "" {
		t.Fatal("tiểu mục mục lục phải được nhận ra")
	}
	if tocReason(Section{Title: "Mở đầu", Text: "Một câu bình thường."}) != "" {
		t.Error("tiểu mục thường không được coi là mục lục")
	}
}

func TestApplyReadingEdit(t *testing.T) {
	tests := []struct {
		name, reading string
		e             ReadingEdit
		want          string
	}{
		{"không sửa", "Câu một. Câu hai.", ReadingEdit{}, "Câu một. Câu hai."},
		{"thay phần đầu, giữ phần sau", "Câu một. Câu hai.", ReadingEdit{From: "Câu một.", To: "Câu thứ nhất."}, "Câu thứ nhất. Câu hai."},
		{"lời đọc đã đổi thì giữ nguyên", "Câu khác. Câu hai.", ReadingEdit{From: "Câu một.", To: "X."}, "Câu khác. Câu hai."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := applyReadingEdit(tt.reading, tt.e); got != tt.want {
				t.Errorf("got %q, muốn %q", got, tt.want)
			}
		})
	}
}

func TestParseVoices(t *testing.T) {
	out := "🎤 Loading\n\n📋 Giọng có sẵn:\n   • ⭐ Trúc Ly — Nữ · Bắc · Phong cách tự nhiên\n   • Minh Đức — Nam · Bắc · Phong cách tin tức\n"
	v := parseVoices(out)
	if len(v) != 2 {
		t.Fatalf("muốn 2 giọng, got %+v", v)
	}
	if v[0].Name != "Trúc Ly" || !v[0].Featured || v[0].Desc != "Nữ · Bắc · Phong cách tự nhiên" {
		t.Errorf("giọng 1 sai: %+v", v[0])
	}
	if v[1].Name != "Minh Đức" || v[1].Featured {
		t.Errorf("giọng 2 sai: %+v", v[1])
	}
}

func TestRunContext_Stub_ReportsProgressAndZips(t *testing.T) {
	needFFmpeg(t)
	opts := stubOpts(t, writeSample(t))
	opts.OutputZip = filepath.Join(opts.OutputDir, "book.zip")
	opts.IntroText = "Bạn đang nghe sách nói."
	opts.DropStems = map[string]bool{"ch01-sec02": true}
	var got []Progress
	opts.Progress = func(p Progress) { got = append(got, p) }
	n, err := RunContext(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunContext lỗi: %v", err)
	}
	if len(got) < 3 {
		t.Fatalf("thiếu sự kiện tiến độ: %+v", got)
	}
	last := got[len(got)-1]
	if last.Phase != PhaseDone || last.Done != n || last.Total != n || last.DoneChars != last.TotalChars {
		t.Errorf("tiến độ cuối sai (n=%d): %+v", n, last)
	}
	if _, err := os.Stat(opts.OutputZip); err != nil {
		t.Errorf("chưa có zip: %v", err)
	}
}

func TestRunContext_Cancelled_ReturnsContextError(t *testing.T) {
	needFFmpeg(t)
	opts := stubOpts(t, writeSample(t))
	ctx, cancel := context.WithCancel(context.Background())
	opts.Progress = func(p Progress) {
		if p.Done == 1 {
			cancel()
		}
	}
	_, err := RunContext(ctx, opts)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("muốn context.Canceled, got %v", err)
	}
}

func TestPreview_Stub_RendersRequestedStemsWithEdit(t *testing.T) {
	needFFmpeg(t)
	docx := writeSample(t)
	o, err := Inspect(docx, InspectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	stem := o.Chapters[0].Sections[0].Stem
	opts := stubOpts(t, docx)
	opts.IntroText = "Bạn đang nghe sách nói."
	first, err := Preview(context.Background(), opts, []string{IntroStem, stem, "ch99-sec99"}, 30)
	if err != nil {
		t.Fatalf("Preview lỗi: %v", err)
	}
	if len(first) != 2 || first[0].Stem != IntroStem || first[1].Stem != stem {
		t.Fatalf("muốn 2 đoạn intro + %s, got %+v", stem, first)
	}
	if len([]rune(first[1].Text)) > 30 || first[1].Full {
		t.Errorf("đoạn nghe thử phải cắt ≤30 ký tự: %q", first[1].Text)
	}
	if _, err := os.Stat(first[1].File); err != nil {
		t.Errorf("chưa có MP3 nghe thử: %v", err)
	}

	// Sửa lời đọc phần đã nghe → lượt sau dùng lời mới, phần sau giữ nguyên.
	opts.ReadingEdits = map[string]ReadingEdit{stem: {From: first[1].Text, To: "Lời đọc đã sửa."}}
	again, err := Preview(context.Background(), opts, []string{stem}, 30)
	if err != nil {
		t.Fatal(err)
	}
	if again[0].Text != "Lời đọc đã sửa." || again[0].Full {
		t.Errorf("nghe thử sau khi sửa phải đúng phần đã sửa: %q (full=%v)", again[0].Text, again[0].Full)
	}
	// Bản render thật: phần đầu thay bằng lời mới, phần sau giữ nguyên.
	p, err := opts.prepare()
	if err != nil {
		t.Fatal(err)
	}
	for _, j := range p.jobs {
		if p.origStems[j.Stem] == stem && (!strings.HasPrefix(j.Text, "Lời đọc đã sửa.") || len(j.Text) <= len("Lời đọc đã sửa.")) {
			t.Errorf("lời đọc bản cuối sai: %q", j.Text)
		}
	}
}
