package bookmaker

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ttsMode — chế độ sinh audio.
const (
	TTSModeVieNeu = "vieneu" // gọi VieNeu-TTS thật (cần venv + model)
	TTSModeStub   = "stub"   // sinh MP3 im lặng ngắn (smoke/CI, không cần model)
)

// TTSConfig — cấu hình bước pre-render TTS.
type TTSConfig struct {
	Mode      string // TTSModeVieNeu | TTSModeStub
	Python    string // đường dẫn python venv VieNeu
	Script    string // đường dẫn audio_gen_batch.py
	ScriptDir string // thư mục chứa script (cho PYTHONPATH import audio_gen)
	Voice     string // tên voice preset (vd "Trúc Ly")
	FFmpeg    string // binary ffmpeg
	Bitrate   string // bitrate MP3 CBR (vd "128k")
	StubSec   int    // độ dài MP3 stub (giây)

	// Env — biến môi trường thêm khi chạy python bộ đọc (vd HF_HOME của bộ đọc
	// phần mềm desktop tự cài). nil = chỉ dùng môi trường hiện tại.
	Env []string

	// KeepTxt — giữ lại file <stem>.txt đã nạp cho VieNeu trong output-dir
	// (mặc định xóa sau khi convert sang mp3). Bật để soát lời đọc 1:1 với audio
	// khi phát hiện chỗ đọc sai.
	KeepTxt bool

	// OnlyStems — nếu != nil và Mode == vieneu: chỉ render giọng thật cho các
	// stem trong tập này, các tiểu mục còn lại sinh stub im lặng. Dùng để
	// pre-render thử 1 đoạn cho user nghe trước khi render full sách.
	OnlyStems map[string]bool

	// PreviewChars — nếu > 0: cắt văn bản đọc của mỗi tiểu mục render giọng thật
	// còn ~PreviewChars ký tự (cắt mềm tại cuối câu/khoảng trắng). Audio test sẽ
	// ngắn hơn lời gốc; chapters.json vẫn giữ reading_script đầy đủ.
	PreviewChars int

	Logf func(string, ...any)
}

// ttsJob — 1 đơn vị render: tên gốc file (không đuôi) + văn bản đọc.
type ttsJob struct {
	Stem string
	Text string
}

// preflight kiểm tra công cụ cần thiết trước khi render thật.
func (c *TTSConfig) preflight() error {
	if _, err := exec.LookPath(c.FFmpeg); err != nil {
		return fmt.Errorf("không tìm thấy ffmpeg %q (cần convert WAV→MP3): %w", c.FFmpeg, err)
	}
	if c.Mode == TTSModeStub {
		return nil
	}
	if !fileExistsTTS(c.Python) {
		return fmt.Errorf("không tìm thấy python venv VieNeu %q (dùng --tts-mode stub để bỏ qua TTS)", c.Python)
	}
	if !fileExistsTTS(c.Script) {
		return fmt.Errorf("không tìm thấy script TTS %q", c.Script)
	}
	return nil
}

// Render sinh MP3 cho mọi job vào workDir/<stem>.mp3.
func (c *TTSConfig) Render(workDir string, jobs []ttsJob) error {
	return c.RenderContext(context.Background(), workDir, jobs, nil)
}

// RenderContext như Render nhưng hủy được qua ctx (dừng tiến trình bộ đọc) và
// gọi onDone(stem) ngay khi từng tiểu mục có MP3.
func (c *TTSConfig) RenderContext(ctx context.Context, workDir string, jobs []ttsJob, onDone func(stem string)) error {
	if len(jobs) == 0 {
		return fmt.Errorf("không có tiểu mục nào để render")
	}
	if onDone == nil {
		onDone = func(string) {}
	}
	if c.Logf == nil {
		c.Logf = func(string, ...any) {}
	}
	if c.Mode == TTSModeStub {
		return c.renderStub(ctx, workDir, jobs, onDone)
	}

	// Mode vieneu: nếu giới hạn OnlyStems → render thật phần được chọn,
	// phần còn lại sinh stub im lặng (đủ file hợp lệ cho zip).
	real, silent := splitJobsByStems(jobs, c.OnlyStems)
	if len(c.OnlyStems) > 0 && len(real) == 0 {
		return fmt.Errorf("--tts-only không khớp tiểu mục nào (kiểm tra lại stem, vd \"ch03-sec01\")")
	}
	if len(silent) > 0 {
		c.Logf("TTS: %d tiểu mục render giọng thật, %d tiểu mục stub im lặng", len(real), len(silent))
		if err := c.renderStub(ctx, workDir, silent, onDone); err != nil {
			return err
		}
	}
	return c.renderVieNeu(ctx, workDir, real, onDone)
}

// truncatePreview cắt s còn ~n ký tự (rune), ưu tiên dừng ở cuối câu gần nhất,
// nếu không có thì cắt tại khoảng trắng cuối để không đứt từ. n <= 0 hoặc s ngắn
// hơn n → trả nguyên s.
func truncatePreview(s string, n int) string {
	if n <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	cut := string(r[:n])
	if i := strings.LastIndexAny(cut, ".!?"); i > 0 {
		return strings.TrimSpace(cut[:i+1])
	}
	if i := strings.LastIndexByte(cut, ' '); i > 0 {
		return strings.TrimSpace(cut[:i])
	}
	return strings.TrimSpace(cut)
}

// splitJobsByStems tách jobs thành nhóm render thật (stem ∈ only) và nhóm stub
// (còn lại). only rỗng → tất cả vào real (giữ hành vi render full).
func splitJobsByStems(jobs []ttsJob, only map[string]bool) (real, silent []ttsJob) {
	if len(only) == 0 {
		return jobs, nil
	}
	for _, j := range jobs {
		if only[j.Stem] {
			real = append(real, j)
		} else {
			silent = append(silent, j)
		}
	}
	return real, silent
}

// wavDoneRe khớp dòng audio_gen_batch.py in ra khi đọc xong 1 file: "✨ <stem>_full.wav".
var wavDoneRe = regexp.MustCompile(`✨\s+(\S+)_full\.wav`)

// renderVieNeu gọi audio_gen_batch.py 1 lần (load model 1 lần) cho mọi job.
// Đọc log của script theo từng dòng: file nào xong thì convert ngay
// <stem>_full.wav → <stem>.mp3 và báo onDone, không đợi cả lượt.
func (c *TTSConfig) renderVieNeu(ctx context.Context, workDir string, jobs []ttsJob, onDone func(string)) error {
	var txtPaths []string
	for _, j := range jobs {
		text := j.Text
		if c.PreviewChars > 0 {
			text = truncatePreview(text, c.PreviewChars)
			c.Logf("  ✂ %s: cắt còn ~%d ký tự (preview)", j.Stem, len([]rune(text)))
		}
		txt := filepath.Join(workDir, j.Stem+".txt")
		if err := os.WriteFile(txt, []byte(text), 0o644); err != nil {
			return fmt.Errorf("ghi txt %q: %w", txt, err)
		}
		txtPaths = append(txtPaths, txt)
	}

	args := append([]string{c.Script, "--voice", c.Voice}, txtPaths...)
	cmd := exec.CommandContext(ctx, c.Python, args...)
	cmd.Dir = workDir // wav xuất ra CWD = workDir
	// PYTHONUNBUFFERED: log in ra từng dòng ngay → biết file nào xong để báo tiến độ.
	cmd.Env = append(append(os.Environ(), c.Env...), "PYTHONPATH="+c.ScriptDir, "PYTHONUNBUFFERED=1")
	hideWindow(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("chạy VieNeu-TTS: %w", err)
	}
	cmd.Stderr = cmd.Stdout
	cmd.WaitDelay = 3 * time.Second
	c.Logf("TTS: %s %s (voice=%q, %d tiểu mục)", c.Python, c.Script, c.Voice, len(jobs))
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("chạy VieNeu-TTS thất bại: %w", err)
	}

	pending := make(map[string]bool, len(jobs))
	for _, j := range jobs {
		pending[j.Stem] = true
	}
	var tail tailBuffer
	var convErr error
	sc := bufio.NewScanner(stdout)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		tail.add(line)
		m := wavDoneRe.FindStringSubmatch(line)
		if m == nil || !pending[m[1]] || convErr != nil {
			continue
		}
		if convErr = c.finishWav(workDir, m[1]); convErr == nil {
			delete(pending, m[1])
			onDone(m[1])
		}
	}
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if waitErr != nil {
		return fmt.Errorf("chạy VieNeu-TTS thất bại: %w\n%s", waitErr, truncateBytes([]byte(tail.String()), 2000))
	}
	if convErr != nil {
		return convErr
	}
	// Phòng khi log không khớp mẫu (script đổi cách in): convert phần còn lại theo thứ tự.
	for _, j := range jobs {
		if !pending[j.Stem] {
			continue
		}
		if err := c.finishWav(workDir, j.Stem); err != nil {
			return err
		}
		onDone(j.Stem)
	}
	return nil
}

// finishWav convert <stem>_full.wav → <stem>.mp3 rồi dọn file trung gian.
func (c *TTSConfig) finishWav(workDir, stem string) error {
	wav := filepath.Join(workDir, stem+"_full.wav")
	if !fileExistsTTS(wav) {
		return fmt.Errorf("không thấy WAV đầu ra %q (TTS lỗi cho tiểu mục %q)", wav, stem)
	}
	mp3 := filepath.Join(workDir, stem+".mp3")
	if err := c.convertToMP3(wav, mp3); err != nil {
		return err
	}
	// Dọn file trung gian (wav + chunks). File <stem>.txt giữ lại nếu KeepTxt
	// để người làm sách soát lời đọc đã nạp cho TTS 1:1 với audio khi có chỗ phát sai.
	if !c.KeepTxt {
		_ = os.Remove(filepath.Join(workDir, stem+".txt"))
	}
	_ = os.Remove(wav)
	_ = os.RemoveAll(filepath.Join(workDir, "chunks_"+stem))
	c.Logf("  ✓ %s.mp3", stem)
	return nil
}

// tailBuffer giữ ~40 dòng log cuối của bộ đọc để in kèm khi lỗi.
type tailBuffer struct{ lines []string }

func (t *tailBuffer) add(s string) {
	t.lines = append(t.lines, s)
	if len(t.lines) > 40 {
		t.lines = t.lines[len(t.lines)-40:]
	}
}

func (t *tailBuffer) String() string { return strings.Join(t.lines, "\n") }

// renderStub sinh MP3 im lặng ngắn cho mỗi job — đủ cho smoke pipeline (MP3 CBR
// hợp lệ, đọc được thời lượng), không cần model.
func (c *TTSConfig) renderStub(ctx context.Context, workDir string, jobs []ttsJob, onDone func(string)) error {
	for _, j := range jobs {
		if err := ctx.Err(); err != nil {
			return err
		}
		mp3 := filepath.Join(workDir, j.Stem+".mp3")
		// anullsrc → MP3 mono CBR, độ dài StubSec giây.
		args := []string{
			"-y", "-f", "lavfi",
			"-i", "anullsrc=r=44100:cl=mono",
			"-t", strconv.Itoa(c.StubSec),
			"-codec:a", "libmp3lame", "-b:a", c.Bitrate, "-ac", "1",
			mp3,
		}
		cmd := exec.CommandContext(ctx, c.FFmpeg, args...)
		hideWindow(cmd)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("stub TTS ffmpeg %q: %w\n%s", j.Stem, err, truncateBytes(out, 800))
		}
		c.Logf("  ✓ %s.mp3 (stub %ds)", j.Stem, c.StubSec)
		onDone(j.Stem)
	}
	return nil
}

// convertToMP3 chuyển WAV → MP3 CBR mono qua ffmpeg (đếm frame được bằng
// tcolgate/mp3 để tính thời lượng).
func (c *TTSConfig) convertToMP3(wav, mp3 string) error {
	args := []string{
		"-y", "-i", wav,
		"-codec:a", "libmp3lame", "-b:a", c.Bitrate, "-ar", "44100", "-ac", "1",
		mp3,
	}
	cmd := exec.Command(c.FFmpeg, args...)
	hideWindow(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("convert %q → MP3: %w\n%s", wav, err, truncateBytes(out, 800))
	}
	return nil
}

func fileExistsTTS(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func truncateBytes(b []byte, n int) []byte {
	if len(b) <= n {
		return b
	}
	return b[len(b)-n:] // giữ phần cuối (thường chứa lỗi)
}
