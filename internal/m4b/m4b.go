// Package m4b xuất một cuốn sách nói (danh sách MP3 từng tiểu mục) thành một
// file .m4b: AAC trong MP4, có mục lục chương, thẻ tên sách/tác giả và ảnh bìa
// nhúng — nghe được trên Apple Books, các app sách nói Android và hiện tên
// chương trên màn hình xe (CarPlay/Android Auto).
//
// Cách làm (chỉ cần ffmpeg):
//
//  1. Giải mã lần lượt từng MP3 ra PCM (s16le mono 44,1 kHz) và dồn vào một
//     tiến trình ffmpeg mã hoá AAC. Đếm số mẫu của từng file nên mốc chương
//     khớp tuyệt đối với âm thanh (không lệch dồn qua hàng trăm tiểu mục).
//  2. Ghép AAC + file ffmetadata (thẻ + mục lục) + ảnh bìa thành MP4 brand
//     "M4B " (-c copy, -movflags +faststart), ghi file tạm rồi đổi tên.
package m4b

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tcolgate/mp3"

	"sano/internal/cover"
)

// Tham số âm thanh: MP3 của bookmaker là mono 44,1 kHz; AAC-LC 64 kbps mono
// đủ cho giọng đọc (~29 MB mỗi giờ nghe).
const (
	sampleRate     = 44100
	bytesPerSample = 2 // s16le mono
	DefaultBitrate = "64k"
	coverSize      = 1400 // bìa vuông tự vẽ khi sách không có bìa
)

// Các giai đoạn báo trong Progress.Phase.
const (
	PhaseEncode = "encode" // giải mã MP3 + mã hoá AAC
	PhaseMux    = "mux"    // ghép mục lục, thẻ, ảnh bìa
	PhaseDone   = "done"
)

// Progress — tiến độ xuất.
type Progress struct {
	Phase    string  `json:"phase"`
	Track    int     `json:"track"`  // tiểu mục đang xử lý (đếm từ 1)
	Tracks   int     `json:"tracks"` // tổng số tiểu mục
	DoneSec  float64 `json:"doneSec"`
	TotalSec float64 `json:"totalSec"` // ước lượng từ các frame MP3
	Percent  int     `json:"percent"`
}

// Options — công cụ + callback.
type Options struct {
	FFmpeg   string // đường dẫn ffmpeg; "" = "ffmpeg" trong PATH
	Bitrate  string // "" = DefaultBitrate
	TempDir  string // thư mục làm việc tạm; "" = os.TempDir()
	Progress func(Progress)
	Logf     func(string, ...any)
}

// Result — file đã xuất.
type Result struct {
	Path        string  `json:"path"`
	DurationSec float64 `json:"durationSec"`
	Size        int64   `json:"size"`
	Chapters    int     `json:"chapters"`
}

// Export ghi b thành file .m4b ở out. Hủy được qua ctx: dừng ffmpeg, xoá hết
// file tạm, không để lại file dở ở out.
func Export(ctx context.Context, b Book, out string, opt Options) (*Result, error) {
	if opt.FFmpeg == "" {
		opt.FFmpeg = "ffmpeg"
	}
	if opt.Bitrate == "" {
		opt.Bitrate = DefaultBitrate
	}
	if opt.Logf == nil {
		opt.Logf = func(string, ...any) {}
	}
	if len(b.Tracks) == 0 {
		return nil, errors.New("sách không có tiểu mục nào để xuất")
	}
	for _, t := range b.Tracks {
		if !fileExists(t.File) {
			return nil, fmt.Errorf("thiếu file âm thanh %q", t.File)
		}
	}
	if _, err := exec.LookPath(opt.FFmpeg); err != nil {
		return nil, fmt.Errorf("không tìm thấy ffmpeg %q: %w", opt.FFmpeg, err)
	}
	if !strings.EqualFold(filepath.Ext(out), ".m4b") {
		out += ".m4b"
	}
	outDir := filepath.Dir(out)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, fmt.Errorf("tạo thư mục lưu: %w", err)
	}

	work, err := os.MkdirTemp(opt.TempDir, "sano-m4b-")
	if err != nil {
		return nil, fmt.Errorf("tạo thư mục tạm: %w", err)
	}
	defer func() { _ = os.RemoveAll(work) }()
	// File tạm nằm cạnh file đích (cùng ổ đĩa) để đổi tên là xong.
	part := filepath.Join(outDir, "."+filepath.Base(out)+".part")
	defer func() { _ = os.Remove(part) }()

	rep := newReporter(opt.Progress, len(b.Tracks), estimateSec(b.Tracks))
	audio := filepath.Join(work, "audio.m4a")
	counts, err := encode(ctx, opt, b.Tracks, audio, rep)
	if err != nil {
		return nil, err
	}

	chapters, totalMs := buildChapters(b.Tracks, counts)
	metaPath := filepath.Join(work, "chapters.txt")
	if err := os.WriteFile(metaPath, []byte(FFMetadata(b.Meta, chapters)), 0o644); err != nil {
		return nil, fmt.Errorf("ghi mục lục: %w", err)
	}
	coverPath, coverCodec, err := prepareCover(b, work)
	if err != nil {
		return nil, err
	}
	// Bìa: ép bộ tách ảnh (không dò dạng theo nội dung — file giả dạng playlist
	// HLS không thành bộ tách khác) và giao thức file:, giống decode().
	coverAbs, err := filepath.Abs(coverPath)
	if err != nil {
		return nil, err
	}

	rep.phase(PhaseMux)
	args := []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-i", audio,
		"-f", "ffmetadata", "-i", metaPath,
		"-f", "image2", "-pattern_type", "none", "-i", "file:" + coverAbs,
		"-map", "0:a", "-map", "2:v",
		"-map_metadata", "1", "-map_chapters", "1",
		"-c:a", "copy", "-c:v", coverCodec,
		"-disposition:v:0", "attached_pic",
		"-movflags", "+faststart",
		"-brand", "M4B ",
		"-f", "mp4", part,
	}
	if err := run(ctx, opt.FFmpeg, args); err != nil {
		return nil, fmt.Errorf("ghép file M4B: %w", err)
	}
	if err := os.Rename(part, out); err != nil {
		return nil, fmt.Errorf("lưu file M4B: %w", err)
	}
	res := &Result{Path: out, DurationSec: float64(totalMs) / 1000, Chapters: len(chapters)}
	if info, err := os.Stat(out); err == nil {
		res.Size = info.Size()
	}
	rep.phase(PhaseDone)
	opt.Logf("✓ M4B: %s (%d mốc chương, %.0f giây)", out, len(chapters), res.DurationSec)
	return res, nil
}

// encode giải mã từng MP3 ra PCM, dồn vào một ffmpeg mã hoá AAC. Trả số mẫu
// PCM của từng tiểu mục.
func encode(ctx context.Context, opt Options, tracks []Track, audio string, rep *reporter) ([]int64, error) {
	enc := exec.CommandContext(ctx, opt.FFmpeg,
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-f", "s16le", "-ar", fmt.Sprint(sampleRate), "-ac", "1", "-i", "pipe:0",
		"-c:a", "aac", "-b:a", opt.Bitrate,
		"-f", "mp4", audio,
	)
	hideWindow(enc)
	var encErr tailBuf
	enc.Stderr = &encErr
	stdin, err := enc.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("chạy ffmpeg: %w", err)
	}
	enc.WaitDelay = 3 * time.Second
	if err := enc.Start(); err != nil {
		return nil, fmt.Errorf("chạy ffmpeg: %w", err)
	}
	fail := func(err error) ([]int64, error) {
		_ = stdin.Close()
		_ = enc.Wait()
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if s := encErr.String(); s != "" {
			return nil, fmt.Errorf("%w\n%s", err, s)
		}
		return nil, err
	}

	counts := make([]int64, len(tracks))
	for i, t := range tracks {
		if ctx.Err() != nil {
			return fail(ctx.Err())
		}
		rep.track(i + 1)
		cw := &countWriter{w: stdin, onWrite: rep.addBytes}
		if err := decode(ctx, opt.FFmpeg, t.File, cw); err != nil {
			return fail(fmt.Errorf("đọc %q: %w", filepath.Base(t.File), err))
		}
		counts[i] = cw.n / bytesPerSample
	}
	if err := stdin.Close(); err != nil {
		return fail(fmt.Errorf("mã hoá AAC: %w", err))
	}
	if err := enc.Wait(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("mã hoá AAC: %w\n%s", err, encErr.String())
	}
	return counts, nil
}

// decode giải mã một MP3 ra PCM s16le mono 44,1 kHz vào w. File có thể nằm
// trong thư mục sách của người khác gửi: ép bộ tách mp3 (không để ffmpeg dò định
// dạng theo nội dung, chạm tới bộ tách khác) và giao thức file: (tên file dạng
// "pipe:0" hay "http:..." không thành giao thức khác).
func decode(ctx context.Context, ffmpeg, file string, w io.Writer) error {
	abs, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, ffmpeg,
		"-nostdin", "-hide_banner", "-loglevel", "error",
		"-f", "mp3", "-i", "file:"+abs, "-map", "0:a:0", "-vn", "-sn", "-dn",
		"-ac", "1", "-ar", fmt.Sprint(sampleRate), "-f", "s16le", "pipe:1",
	)
	hideWindow(cmd)
	var errBuf tailBuf
	cmd.Stdout = w
	cmd.Stderr = &errBuf
	cmd.WaitDelay = 3 * time.Second
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if s := errBuf.String(); s != "" {
			return fmt.Errorf("%w: %s", err, s)
		}
		return err
	}
	return nil
}

// buildChapters đổi số mẫu từng tiểu mục thành mốc chương (ms). Tiểu mục rỗng
// (0 mẫu) bỏ khỏi mục lục — mốc dài 0 làm vài app nghe hiển thị sai.
func buildChapters(tracks []Track, counts []int64) ([]Chapter, int64) {
	var out []Chapter
	var pos int64
	for i, t := range tracks {
		start := pos
		pos += counts[i]
		if counts[i] == 0 {
			continue
		}
		title := cleanTitle(t.Title)
		if title == "" {
			title = fmt.Sprintf("Phần %d", len(out)+1)
		}
		out = append(out, Chapter{Title: title, StartMs: samplesToMs(start), EndMs: samplesToMs(pos)})
	}
	return out, samplesToMs(pos)
}

func samplesToMs(n int64) int64 { return n * 1000 / sampleRate }

// prepareCover chọn ảnh bìa nhúng: bìa của sách (png/jpg chép nguyên, định dạng
// khác đổi sang JPEG); không có thì vẽ bìa vuông 1400×1400. Trả đường dẫn + bộ
// mã ảnh cho ffmpeg.
func prepareCover(b Book, work string) (string, string, error) {
	if b.Cover != "" && fileExists(b.Cover) {
		switch strings.ToLower(filepath.Ext(b.Cover)) {
		case ".png", ".jpg", ".jpeg":
			return b.Cover, "copy", nil
		default:
			return b.Cover, "mjpeg", nil
		}
	}
	p := filepath.Join(work, "cover.png")
	f, err := os.Create(p)
	if err != nil {
		return "", "", fmt.Errorf("tạo bìa: %w", err)
	}
	werr := cover.WritePNG(f, b.Title, b.Author, coverSize, coverSize)
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return "", "", fmt.Errorf("vẽ bìa: %w", werr)
	}
	return p, "copy", nil
}

func run(ctx context.Context, ffmpeg string, args []string) error {
	cmd := exec.CommandContext(ctx, ffmpeg, args...)
	hideWindow(cmd)
	var errBuf tailBuf
	cmd.Stderr = &errBuf
	cmd.WaitDelay = 3 * time.Second
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%w\n%s", err, errBuf.String())
	}
	return nil
}

// estimateSec cộng thời lượng các MP3 theo frame (đọc header, không giải mã) —
// chỉ để tính phần trăm tiến độ.
func estimateSec(tracks []Track) float64 {
	var total time.Duration
	for _, t := range tracks {
		f, err := os.Open(t.File)
		if err != nil {
			continue
		}
		dec := mp3.NewDecoder(f)
		var frame mp3.Frame
		skipped := 0
		for dec.Decode(&frame, &skipped) == nil {
			total += frame.Duration()
		}
		_ = f.Close()
	}
	return total.Seconds()
}

// reporter gom tiến độ, gọi callback tối đa ~4 lần/giây (và luôn gọi khi đổi
// tiểu mục / giai đoạn).
type reporter struct {
	mu    sync.Mutex
	cb    func(Progress)
	p     Progress
	bytes int64
	last  time.Time
}

func newReporter(cb func(Progress), tracks int, totalSec float64) *reporter {
	return &reporter{cb: cb, p: Progress{Phase: PhaseEncode, Tracks: tracks, TotalSec: totalSec}}
}

func (r *reporter) emit(force bool) {
	if r.cb == nil || (!force && time.Since(r.last) < 250*time.Millisecond) {
		return
	}
	r.last = time.Now()
	r.cb(r.p)
}

func (r *reporter) track(i int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.p.Track = i
	r.emit(true)
}

func (r *reporter) addBytes(n int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bytes += int64(n)
	r.p.DoneSec = float64(r.bytes) / (sampleRate * bytesPerSample)
	if r.p.TotalSec > 0 {
		// Mã hoá chiếm 0–95%, phần còn lại cho bước ghép.
		r.p.Percent = min(95, int(r.p.DoneSec/r.p.TotalSec*95))
	}
	r.emit(false)
}

func (r *reporter) phase(ph string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.p.Phase = ph
	switch ph {
	case PhaseMux:
		r.p.Percent = 97
	case PhaseDone:
		r.p.Percent = 100
		r.p.Track = r.p.Tracks
		r.p.TotalSec = r.p.DoneSec
	}
	r.emit(true)
}

type countWriter struct {
	w       io.Writer
	n       int64
	onWrite func(int)
}

func (c *countWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	if c.onWrite != nil {
		c.onWrite(n)
	}
	return n, err
}

// tailBuf giữ ~2 KB cuối của stderr ffmpeg để in kèm khi lỗi.
type tailBuf struct{ b bytes.Buffer }

func (t *tailBuf) Write(p []byte) (int, error) {
	t.b.Write(p)
	if t.b.Len() > 4096 {
		keep := append([]byte(nil), t.b.Bytes()[t.b.Len()-2048:]...)
		t.b.Reset()
		t.b.Write(keep)
	}
	return len(p), nil
}

func (t *tailBuf) String() string { return strings.TrimSpace(t.b.String()) }
