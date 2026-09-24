package m4b

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestEscapeMeta(t *testing.T) {
	cases := map[string]string{
		"Bình thường":      "Bình thường",
		"a=b":              `a\=b`,
		"a;b":              `a\;b`,
		"#1 Mở đầu":        `\#1 Mở đầu`,
		`C:\sach`:          `C:\\sach`,
		"dòng 1\ndòng 2":   "dòng 1\\\ndòng 2",
		"dòng 1\r\ndòng 2": "dòng 1\\\ndòng 2",
		`=;#\` + "\n":      `\=\;\#\\` + "\\\n",
	}
	for in, want := range cases {
		if got := escapeMeta(in); got != want {
			t.Errorf("escapeMeta(%q) = %q, muốn %q", in, got, want)
		}
	}
}

func TestFFMetadata(t *testing.T) {
	got := FFMetadata(Meta{Title: "Sách = hay; #1", Author: "Tác giả"}, []Chapter{
		{Title: "Lời mở đầu", StartMs: 0, EndMs: 1500},
		{Title: "Phần 1; a=b", StartMs: 1500, EndMs: 4200},
	})
	want := `;FFMETADATA1
title=Sách \= hay\; \#1
album=Sách \= hay\; \#1
artist=Tác giả
album_artist=Tác giả
genre=Audiobook
media_type=2
comment=Tạo bằng Sano

[CHAPTER]
TIMEBASE=1/1000
START=0
END=1500
title=Lời mở đầu

[CHAPTER]
TIMEBASE=1/1000
START=1500
END=4200
title=Phần 1\; a\=b
`
	if got != want {
		t.Fatalf("FFMetadata sai:\n%s\n--- muốn ---\n%s", got, want)
	}
	// Không có tác giả → bỏ hẳn dòng artist, không ghi "artist=".
	if s := FFMetadata(Meta{Title: "X"}, nil); strings.Contains(s, "artist") {
		t.Fatalf("không có tác giả mà vẫn ghi artist:\n%s", s)
	}
}

func TestChapterTitle(t *testing.T) {
	cases := []struct {
		chapter, section string
		n                int
		want             string
	}{
		{"Chương 1", "Chương 1", 1, "Chương 1"},
		{"Chương 1", "chương  1", 1, "Chương 1"},
		{"Chương 1", "Mục 1.1", 2, "Mục 1.1"},
		{"Chương 1", "Mục riêng", 1, "Mục riêng"},
		{"Chương 1", "", 3, "Chương 1"},
		{"Chương\n2", "Phần\na", 2, "Phần a"},
	}
	for _, c := range cases {
		if got := ChapterTitle(c.chapter, c.section, c.n); got != c.want {
			t.Errorf("ChapterTitle(%q, %q, %d) = %q, muốn %q", c.chapter, c.section, c.n, got, c.want)
		}
	}
}

func TestFileName(t *testing.T) {
	cases := map[string]string{
		"Kỹ năng giao tiếp":  "Kỹ năng giao tiếp.m4b",
		`A/B: C*? "D" <E>|F`: "A B C D E F.m4b",
		"  ...  ":            "Sách nói.m4b",
		"Tên.\n":             "Tên.m4b",
	}
	for in, want := range cases {
		if got := FileName(in); got != want {
			t.Errorf("FileName(%q) = %q, muốn %q", in, got, want)
		}
	}
}

func TestBuildChapters(t *testing.T) {
	tracks := []Track{{Title: "A"}, {Title: "Rỗng"}, {Title: ""}, {Title: "C"}}
	chs, total := buildChapters(tracks, []int64{44100, 0, 22050, 88200})
	want := []Chapter{{"A", 0, 1000}, {"Phần 2", 1000, 1500}, {"C", 1500, 3500}}
	if total != 3500 || len(chs) != len(want) {
		t.Fatalf("chapters = %+v, total = %d", chs, total)
	}
	for i := range want {
		if chs[i] != want[i] {
			t.Errorf("chapter %d = %+v, muốn %+v", i, chs[i], want[i])
		}
	}
}

// ── Kiểm với ffmpeg thật (bỏ qua nếu máy không có) ──────────────────────

func needFFmpeg(t *testing.T) string {
	t.Helper()
	p, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("không có ffmpeg")
	}
	return p
}

// makeBookDir tạo thư mục sách giả: MP3 sóng sin độ dài secs[i] + metadata.json.
func makeBookDir(t *testing.T, ffmpeg string, secs []float64) string {
	t.Helper()
	dir := t.TempDir()
	type sec struct {
		Title string `json:"title"`
		File  string `json:"file"`
	}
	type ch struct {
		Title    string `json:"title"`
		Sections []sec  `json:"sections"`
	}
	meta := struct {
		Title    string `json:"title"`
		Author   string `json:"author"`
		Chapters []ch   `json:"chapters"`
	}{Title: "Sách thử = M4B", Author: "Tác Giả"}
	for i, s := range secs {
		file := fmt.Sprintf("ch%02d-sec01.mp3", i+1)
		out, err := exec.Command(ffmpeg, "-loglevel", "error", "-y", "-f", "lavfi",
			"-i", fmt.Sprintf("sine=f=%d:r=44100:d=%g", 300+100*i, s),
			"-ac", "1", "-c:a", "libmp3lame", "-b:a", "64k", filepath.Join(dir, file)).CombinedOutput()
		if err != nil {
			t.Skipf("ffmpeg không tạo được MP3 thử (thiếu libmp3lame?): %v %s", err, out)
		}
		title := fmt.Sprintf("Chương %d", i+1)
		meta.Chapters = append(meta.Chapters, ch{Title: title, Sections: []sec{{Title: title, File: file}}})
	}
	data, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestExportThat(t *testing.T) {
	ffmpeg := needFFmpeg(t)
	secs := []float64{1.5, 2.25, 1}
	dir := makeBookDir(t, ffmpeg, secs)
	b, err := FromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	outDir := t.TempDir()
	tmp := t.TempDir()
	var last Progress
	calls := 0
	res, err := Export(context.Background(), b, filepath.Join(outDir, "sach"), Options{
		FFmpeg: ffmpeg, TempDir: tmp,
		Progress: func(p Progress) { last = p; calls++ },
	})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(res.Path) != ".m4b" || res.Chapters != 3 || res.Size == 0 {
		t.Fatalf("kết quả lạ: %+v", res)
	}
	if last.Phase != PhaseDone || last.Percent != 100 || calls < 4 {
		t.Errorf("tiến độ cuối = %+v (%d lần gọi)", last, calls)
	}
	if d := res.DurationSec; d < 4.7 || d > 4.8 {
		t.Errorf("thời lượng = %.3f, muốn ≈ 4.75", d)
	}
	entries, _ := os.ReadDir(outDir)
	if len(entries) != 1 {
		t.Errorf("thư mục lưu có file thừa: %v", entries)
	}
	if left, _ := os.ReadDir(tmp); len(left) != 0 {
		t.Errorf("còn file tạm: %v", left)
	}
	head := make([]byte, 12)
	f, _ := os.Open(res.Path)
	_, _ = f.Read(head)
	_ = f.Close()
	if string(head[4:12]) != "ftypM4B " {
		t.Errorf("brand = %q, muốn ftypM4B", head[4:12])
	}

	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return
	}
	out, err := exec.Command(ffprobe, "-v", "error", "-of", "json", "-show_chapters", "-show_format",
		"-show_streams", res.Path).Output()
	if err != nil {
		t.Fatal(err)
	}
	var probe struct {
		Chapters []struct {
			StartTime string            `json:"start_time"`
			Tags      map[string]string `json:"tags"`
		} `json:"chapters"`
		Streams []struct {
			CodecType   string         `json:"codec_type"`
			CodecName   string         `json:"codec_name"`
			Disposition map[string]int `json:"disposition"`
		} `json:"streams"`
		Format struct {
			Duration string            `json:"duration"`
			Tags     map[string]string `json:"tags"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &probe); err != nil {
		t.Fatal(err)
	}
	wantStart := []float64{0, 1.5, 3.75}
	if len(probe.Chapters) != 3 {
		t.Fatalf("ffprobe thấy %d chương", len(probe.Chapters))
	}
	for i, c := range probe.Chapters {
		st, _ := strconv.ParseFloat(c.StartTime, 64)
		if d := st - wantStart[i]; d < -0.03 || d > 0.03 {
			t.Errorf("chương %d bắt đầu %.3f, muốn %.3f", i+1, st, wantStart[i])
		}
		if c.Tags["title"] != fmt.Sprintf("Chương %d", i+1) {
			t.Errorf("chương %d tên %q", i+1, c.Tags["title"])
		}
	}
	var hasAAC, hasCover bool
	for _, s := range probe.Streams {
		hasAAC = hasAAC || (s.CodecType == "audio" && s.CodecName == "aac")
		hasCover = hasCover || (s.CodecType == "video" && s.Disposition["attached_pic"] == 1)
	}
	if !hasAAC || !hasCover {
		t.Errorf("thiếu stream: aac=%v bìa=%v", hasAAC, hasCover)
	}
	tags := probe.Format.Tags
	for k, v := range map[string]string{
		"title": "Sách thử = M4B", "album": "Sách thử = M4B", "artist": "Tác Giả",
		"album_artist": "Tác Giả", "genre": "Audiobook", "comment": DefaultComment,
	} {
		if tags[k] != v {
			t.Errorf("thẻ %s = %q, muốn %q", k, tags[k], v)
		}
	}
}

func TestExportHuyDonFileTam(t *testing.T) {
	ffmpeg := needFFmpeg(t)
	dir := makeBookDir(t, ffmpeg, []float64{3, 3, 3, 3})
	b, err := FromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	outDir := t.TempDir()
	tmp := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := filepath.Join(outDir, "sach.m4b")
	_, err = Export(ctx, b, out, Options{
		FFmpeg: ffmpeg, TempDir: tmp,
		Progress: func(p Progress) {
			if p.Track >= 2 {
				cancel() // hủy giữa chừng, khi đang ở tiểu mục thứ 2
			}
		},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, muốn context.Canceled", err)
	}
	if left, _ := os.ReadDir(outDir); len(left) != 0 {
		t.Errorf("còn file ở thư mục lưu sau khi hủy: %v", left)
	}
	if left, _ := os.ReadDir(tmp); len(left) != 0 {
		t.Errorf("còn file tạm sau khi hủy: %v", left)
	}
}

func TestExportLoiRo(t *testing.T) {
	if _, err := Export(context.Background(), Book{}, "x.m4b", Options{}); err == nil {
		t.Fatal("sách rỗng phải báo lỗi")
	}
	b := Book{Meta: Meta{Title: "X"}, Tracks: []Track{{Title: "A", File: "/khong/co.mp3"}}}
	if _, err := Export(context.Background(), b, "x.m4b", Options{}); err == nil || !strings.Contains(err.Error(), "thiếu file") {
		t.Fatalf("err = %v", err)
	}
}

func TestFromDirLoi(t *testing.T) {
	if _, err := FromDir(t.TempDir()); err == nil {
		t.Fatal("thư mục không có metadata.json phải báo lỗi")
	}
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(`{"title":"X","chapters":[{"title":"A","sections":[{"title":"A","file":"../x.mp3"}]}]}`), 0o644)
	if _, err := FromDir(dir); err == nil {
		t.Fatal("file ngoài thư mục sách phải bị từ chối")
	}
	_ = os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(`{"title":"X","chapters":[{"title":"A","sections":[{"title":"A","file":".."}]}]}`), 0o644)
	if _, err := FromDir(dir); err == nil {
		t.Fatal(`tên file ".." phải bị từ chối`)
	}
}

// Bìa người dùng chọn: ép bộ tách image2 + giao thức file:. Tên có "%" không bị
// hiểu là mẫu dãy ảnh; file giả dạng playlist HLS bị từ chối thay vì mở mạng.
func TestExportUserCover(t *testing.T) {
	ffmpeg := needFFmpeg(t)
	dir := makeBookDir(t, ffmpeg, []float64{1})
	for _, name := range []string{"bia 100%d.png", "bia.jpg", "bia.webp"} {
		t.Run(name, func(t *testing.T) {
			cover := filepath.Join(t.TempDir(), name)
			mk := exec.Command(ffmpeg, "-loglevel", "error", "-y", "-f", "lavfi", "-i", "color=red:s=64x64",
				"-frames:v", "1", "-f", "image2", "-update", "1", cover)
			if out, err := mk.CombinedOutput(); err != nil {
				t.Skipf("ffmpeg không tạo được %s: %s", name, out)
			}
			b, err := FromDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			b.Cover = cover
			if _, err := Export(context.Background(), b, filepath.Join(t.TempDir(), "sach"), Options{
				FFmpeg: ffmpeg, TempDir: t.TempDir(),
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("playlist giả dạng ảnh", func(t *testing.T) {
		cover := filepath.Join(t.TempDir(), "bia.jpg")
		if err := os.WriteFile(cover, []byte("#EXTM3U\n#EXTINF:1,\nhttp://127.0.0.1:9/x.ts\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		b, err := FromDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		b.Cover = cover
		_, err = Export(context.Background(), b, filepath.Join(t.TempDir(), "sach"), Options{
			FFmpeg: ffmpeg, TempDir: t.TempDir(),
		})
		if err == nil {
			t.Fatal("muốn lỗi khi bìa không phải ảnh")
		}
		if strings.Contains(strings.ToLower(err.Error()), "hls") {
			t.Fatalf("ffmpeg dò nội dung bìa thành playlist HLS: %v", err)
		}
	})
}
