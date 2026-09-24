package bookmaker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PreviewClip — một đoạn nghe thử đã render.
type PreviewClip struct {
	Stem        string `json:"stem"`  // stem gốc (IntroStem cho lời mở đầu)
	Title       string `json:"title"` // tiêu đề tiểu mục
	Text        string `json:"text"`  // đúng phần lời đọc đã đưa vào bộ đọc
	Full        bool   `json:"full"`  // true = đã đọc hết tiểu mục (không cắt)
	File        string `json:"file"`  // đường dẫn MP3 tuyệt đối
	DurationSec int    `json:"durationSec"`
}

// Preview render nghe thử các tiểu mục theo stem gốc (IntroStem = lời mở đầu),
// mỗi đoạn cắt còn ~chars ký tự (0 = đọc hết), vào opts.OutputDir. Áp đúng các
// tuỳ chọn của lượt render thật (chuẩn hóa, bỏ tiểu mục, lời đọc đã sửa) để
// nghe thử khớp bản cuối. Stem không có trong sách bị bỏ qua.
func Preview(ctx context.Context, opts Options, stems []string, chars int) ([]PreviewClip, error) {
	if len(stems) == 0 {
		return nil, fmt.Errorf("chưa chọn đoạn nào để nghe thử")
	}
	if opts.Logf == nil {
		opts.Logf = func(string, ...any) {}
	}
	p, err := opts.prepare()
	if err != nil {
		return nil, err
	}
	byOrig := make(map[string]ttsJob, len(p.jobs))
	for _, j := range p.jobs {
		byOrig[p.origStems[j.Stem]] = j
	}
	var jobs []ttsJob
	var clips []PreviewClip
	for _, s := range stems {
		j, ok := byOrig[s]
		if !ok {
			continue
		}
		text := truncatePreview(j.Text, chars)
		// Đoạn đã sửa lời đọc: nghe đúng phần đã sửa, để lần sửa sau vẫn thay
		// cùng phần đầu gốc (ReadingEdit.From) mà không lặp phần phía sau.
		if e, ok := opts.ReadingEdits[s]; ok {
			if to := strings.TrimSpace(e.To); to != "" && strings.HasPrefix(j.Text, to) {
				text = to
			}
		}
		// Tên file riêng cho nghe thử: không đè MP3 của lượt render thật nếu cùng thư mục.
		stem := "nghe-thu-" + s
		jobs = append(jobs, ttsJob{Stem: stem, Text: text})
		clips = append(clips, PreviewClip{
			Stem:  s,
			Title: p.titles[j.Stem],
			Text:  text,
			Full:  text == strings.TrimSpace(j.Text),
			File:  filepath.Join(opts.OutputDir, stem+".mp3"),
		})
	}
	if len(jobs) == 0 {
		return nil, fmt.Errorf("không tìm thấy đoạn cần nghe thử trong sách")
	}
	tts := opts.TTS
	tts.OnlyStems, tts.PreviewChars, tts.KeepTxt = nil, 0, true
	if err := tts.preflight(); err != nil {
		return nil, err
	}
	if err := tts.RenderContext(ctx, opts.OutputDir, jobs, nil); err != nil {
		return nil, err
	}
	for i := range clips {
		clips[i].DurationSec, _ = docxMP3DurationSec(clips[i].File)
	}
	return clips, nil
}

// Speak đọc một câu bất kỳ (vd câu nghe mẫu giọng) ra workDir/<name>.mp3, trả
// đường dẫn file. Chuẩn hóa văn nói bằng norm (nil = mặc định).
func Speak(ctx context.Context, tts TTSConfig, norm *Normalizer, text, workDir, name string) (string, error) {
	if norm == nil {
		norm = defaultNormalizer
	}
	text = norm.script(strings.TrimSpace(text))
	if text == "" {
		return "", fmt.Errorf("chưa có câu để đọc")
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return "", fmt.Errorf("tạo thư mục nghe thử: %w", err)
	}
	tts.OnlyStems, tts.PreviewChars = nil, 0
	if err := tts.preflight(); err != nil {
		return "", err
	}
	if err := tts.RenderContext(ctx, workDir, []ttsJob{{Stem: name, Text: text}}, nil); err != nil {
		return "", err
	}
	return filepath.Join(workDir, name+".mp3"), nil
}

// MP3DurationSec đọc thời lượng MP3 (giây), không cần ffprobe.
func MP3DurationSec(path string) (int, error) {
	return docxMP3DurationSec(path)
}

// Slugify gấp dấu tiếng Việt + kebab-case, giống slug trong manifest zip.
func Slugify(s string) string {
	return docxSlugify(s)
}
