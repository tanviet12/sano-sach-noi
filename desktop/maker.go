package main

// Phần "Tạo sách" của App: nạp file thật, nghe thử, render nền có tiến độ + hủy.
// Toàn bộ logic làm sách nằm ở sano/internal/bookmaker; ở đây chỉ nối vào
// giao diện (tìm công cụ, thư mục thư viện, sự kiện Wails).

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sano/desktop/internal/library"
	"sano/desktop/internal/tts"
	"sano/internal/bookmaker"
)

// Sự kiện Wails đẩy lên giao diện.
const (
	eventRenderProgress = "render:progress"
	eventRenderFinished = "render:finished"
)

// Nghe thử: mỗi đoạn đọc ~previewChars ký tự đầu (≈ 30–40 giây nghe), đủ để
// nghe giọng + cách đọc mà không phải chờ lâu.
const (
	previewChars = 500
	mp3Bitrate   = "128k"
)

// BookSettings — lựa chọn của người dùng ở các bước Tạo sách.
type BookSettings struct {
	Path               string                           `json:"path"`
	Title              string                           `json:"title"`
	Author             string                           `json:"author"`
	Category           string                           `json:"category"`
	Series             string                           `json:"series"` // tên bộ sách; trống = sách lẻ
	Volume             int                              `json:"volume"` // số tập; <= 0 = tập kế tiếp
	Voice              string                           `json:"voice"`
	IntroText          string                           `json:"introText"`
	KeepHeadingNumbers bool                             `json:"keepHeadingNumbers"`
	DropStems          []string                         `json:"dropStems"`
	CoverPath          string                           `json:"coverPath"`
	ReadingEdits       map[string]bookmaker.ReadingEdit `json:"readingEdits"`
	// RightsConfirmedAt — lúc người dùng tick "có quyền dùng tài liệu này" (RFC 3339); bắt buộc để render.
	RightsConfirmedAt string `json:"rightsConfirmedAt"`
}

// Clip — đoạn nghe thử kèm URL phát trong app.
type Clip struct {
	bookmaker.PreviewClip
	URL string `json:"url"`
}

// RenderStatus — trạng thái lượt render hiện tại / gần nhất.
type RenderStatus struct {
	Running   bool               `json:"running"`
	Done      bool               `json:"done"`
	Cancelled bool               `json:"cancelled"`
	Error     string             `json:"error"`
	Title     string             `json:"title"`
	Slug      string             `json:"slug"` // slug trong thư viện khi đã xong
	Progress  bookmaker.Progress `json:"progress"`
}

type renderJob struct {
	cancel context.CancelFunc
	status RenderStatus
}

// toolPaths — python + script bộ đọc + ffmpeg (+ biến môi trường của bộ đọc app
// cài); thiếu gì báo lỗi rõ.
type toolPaths struct {
	python, script, ffmpeg string
	env                    []string
}

// tools — như findTools nhưng từ chối khi đang cài bộ đọc (venv đang thay đổi).
func (a *App) tools() (toolPaths, error) {
	if a.Installing() {
		return toolPaths{}, errors.New("đang cài bộ đọc — đợi cài xong rồi thử lại")
	}
	return findTools()
}

func findTools() (toolPaths, error) {
	rt := ttsRuntime()
	t := toolPaths{python: rt.Python, ffmpeg: findFFmpeg(), env: rt.Env}
	if rt.ScriptsDir != "" {
		t.script = filepath.Join(rt.ScriptsDir, "audio_gen_batch.py")
	}
	switch {
	case t.python == "" || !fileExists(t.python):
		return t, fmt.Errorf("chưa cài bộ đọc VieNeu-TTS (không thấy %s)", orDash(t.python))
	case t.script == "" || !fileExists(t.script):
		return t, fmt.Errorf("không thấy script đọc giọng scripts/tts/audio_gen_batch.py — đặt biến %s", tts.EnvScripts)
	case t.ffmpeg == "":
		return t, fmt.Errorf("không tìm thấy ffmpeg. %s", tts.FFmpegHint(runtime.GOOS))
	}
	return t, nil
}

func (t toolPaths) ttsConfig(voice string) bookmaker.TTSConfig {
	if strings.TrimSpace(voice) == "" {
		voice = bookmaker.DefaultVoice
	}
	return bookmaker.TTSConfig{
		Mode:      bookmaker.TTSModeVieNeu,
		Python:    t.python,
		Script:    t.script,
		ScriptDir: filepath.Dir(t.script),
		Voice:     voice,
		FFmpeg:    t.ffmpeg,
		Env:       t.env,
		Bitrate:   mp3Bitrate,
		KeepTxt:   true, // giữ lời đọc .txt cạnh MP3 để soát khi có chỗ đọc sai
	}
}

// options dựng bookmaker.Options từ lựa chọn của người dùng.
func (s BookSettings) options(t toolPaths, outDir string) (bookmaker.Options, error) {
	if !strings.EqualFold(filepath.Ext(s.Path), ".docx") {
		return bookmaker.Options{}, ErrNotDocx
	}
	norm, err := bookmaker.NewNormalizer("", s.KeepHeadingNumbers)
	if err != nil {
		return bookmaker.Options{}, err
	}
	drop := map[string]bool{}
	for _, st := range s.DropStems {
		drop[st] = true
	}
	return bookmaker.Options{
		InputDocx:         s.Path,
		OutputDir:         outDir,
		Title:             strings.TrimSpace(s.Title),
		Author:            strings.TrimSpace(s.Author),
		IntroText:         strings.TrimSpace(s.IntroText),
		Visibility:        "private",
		TTS:               t.ttsConfig(s.Voice),
		Norm:              norm,
		DropStems:         drop,
		CoverPath:         strings.TrimSpace(s.CoverPath),
		ReadingEdits:      s.ReadingEdits,
		RightsConfirmedAt: strings.TrimSpace(s.RightsConfirmedAt),
	}, nil
}

// InspectDocx nạp file Word thật: mục lục, số ký tự, cảnh báo lúc nạp.
func (a *App) InspectDocx(path string, keepHeadingNumbers bool) (*bookmaker.Outline, error) {
	if _, err := describeDocx(path); err != nil {
		return nil, err
	}
	return bookmaker.Inspect(path, bookmaker.InspectOptions{KeepHeadingNumbers: keepHeadingNumbers})
}

// Voices trả danh sách giọng của bộ đọc (hỏi bộ đọc lần đầu rồi nhớ lại).
func (a *App) Voices() ([]bookmaker.Voice, error) {
	a.mu.Lock()
	cached := a.voices
	a.mu.Unlock()
	if len(cached) > 0 {
		return cached, nil
	}
	release, err := a.beginTTSUse()
	if err != nil {
		return nil, err
	}
	defer release()
	t, err := a.tools()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(a.context(), 60*time.Second)
	defer cancel()
	v, err := bookmaker.ListVoices(ctx, t.python, t.script, t.env)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.voices = v
	a.mu.Unlock()
	return v, nil
}

// PreviewClips render nghe thử các đoạn (stem gốc; "intro" = lời mở đầu) với
// đúng lựa chọn hiện tại. Mỗi đoạn ~previewChars ký tự đầu.
func (a *App) PreviewClips(s BookSettings, stems []string) ([]Clip, error) {
	if a.Rendering() {
		return nil, errors.New("đang render cả cuốn — nghe thử lại sau khi render xong")
	}
	release, err := a.beginTTSUse()
	if err != nil {
		return nil, err
	}
	defer release()
	t, err := a.tools()
	if err != nil {
		return nil, err
	}
	dir, err := a.previewDir()
	if err != nil {
		return nil, err
	}
	opts, err := s.options(t, dir)
	if err != nil {
		return nil, err
	}
	clips, err := bookmaker.Preview(a.context(), opts, stems, previewChars)
	if err != nil {
		return nil, err
	}
	out := make([]Clip, 0, len(clips))
	for _, c := range clips {
		out = append(out, Clip{PreviewClip: c, URL: a.mediaURL(c.File)})
	}
	return out, nil
}

// SpeakSample đọc câu mẫu bằng một giọng, trả URL phát.
func (a *App) SpeakSample(voice, text string) (string, error) {
	release, err := a.beginTTSUse()
	if err != nil {
		return "", err
	}
	defer release()
	t, err := a.tools()
	if err != nil {
		return "", err
	}
	dir, err := a.previewDir()
	if err != nil {
		return "", err
	}
	name := "giong-" + bookmaker.Slugify(voice)
	file, err := bookmaker.Speak(a.context(), t.ttsConfig(voice), nil, text, dir, name)
	if err != nil {
		return "", err
	}
	return a.mediaURL(file), nil
}

// renderBusyLocked — lý do không được bắt đầu render lúc này. Gọi khi đang giữ
// a.mu (kiểm lại dưới khoá: cài bộ đọc có thể vừa bắt đầu sau lần kiểm trước).
func (a *App) renderBusyLocked() error {
	switch {
	case a.uninstalling:
		return errUninstalling
	case a.setup != nil && a.setup.inst.Status().Running:
		return errors.New("đang cài bộ đọc — đợi cài xong rồi thử lại")
	case a.job != nil && a.job.status.Running:
		return errors.New("đang render một cuốn khác")
	}
	return nil
}

// StartRender bắt đầu render cả cuốn trong nền. Tiến độ đẩy lên qua sự kiện
// render:progress, kết thúc qua render:finished. Mỗi lúc chỉ một cuốn.
func (a *App) StartRender(s BookSettings) (*RenderStatus, error) {
	if strings.TrimSpace(s.RightsConfirmedAt) == "" {
		return nil, ErrRightsNotConfirmed
	}
	// Giữ một lượt dùng bộ đọc từ lúc tìm công cụ (tìm script có giải nén vào
	// thư mục bộ đọc) tới khi đăng ký job: gỡ bộ đọc không chen vào giữa được.
	release, err := a.beginTTSUse()
	if err != nil {
		return nil, err
	}
	defer release()
	t, err := a.tools()
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(s.Title)
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(s.Path), filepath.Ext(s.Path))
	}
	slug := library.BookSlug(title)
	category := a.lib.CanonicalCategory(s.Category, "") // gộp với danh mục đã có (không phân biệt hoa thường)
	series, volume, err := a.lib.PlaceInSeries(s.Series, s.Volume, "")
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	if err := a.renderBusyLocked(); err != nil {
		a.mu.Unlock()
		return nil, err
	}
	work, err := a.lib.NewWorkDir(slug)
	if err != nil {
		a.mu.Unlock()
		return nil, err
	}
	opts, err := s.options(t, work)
	if err != nil {
		a.mu.Unlock()
		_ = os.RemoveAll(work)
		return nil, err
	}
	opts.OutputZip = filepath.Join(work, "book-"+slug+".zip")
	opts.Category = category
	opts.Series, opts.SeriesVolume = series, volume
	ctx, cancel := context.WithCancel(a.context())
	job := &renderJob{cancel: cancel, status: RenderStatus{Running: true, Title: title}}
	a.job = job
	st := job.status
	a.mu.Unlock()

	opts.Progress = func(p bookmaker.Progress) {
		a.mu.Lock()
		job.status.Progress = p
		cur := job.status
		a.mu.Unlock()
		a.emit(eventRenderProgress, cur)
	}
	go a.runRender(ctx, job, opts, work, slug)
	return &st, nil
}

// runRender chạy trong goroutine riêng: render xong thì chuyển thư mục tạm vào
// thư viện; lỗi hoặc hủy thì dọn thư mục tạm (chỉ chứa file của lượt này).
func (a *App) runRender(ctx context.Context, job *renderJob, opts bookmaker.Options, work, slug string) {
	var runErr error
	defer func() {
		if r := recover(); r != nil {
			log.Printf("render panic: %v\n%s", r, debug.Stack())
			runErr = fmt.Errorf("lỗi không mong muốn: %v", r)
		}
		a.finishRender(ctx, job, work, slug, runErr)
	}()
	_, runErr = bookmaker.RunContext(ctx, opts)
}

func (a *App) finishRender(ctx context.Context, job *renderJob, work, slug string, runErr error) {
	final := ""
	if runErr == nil {
		final, runErr = a.lib.Commit(work, slug)
	}
	if runErr != nil {
		_ = os.RemoveAll(work)
	}
	a.mu.Lock()
	job.status.Running = false
	switch {
	case errors.Is(ctx.Err(), context.Canceled):
		job.status.Cancelled = true
	case runErr != nil:
		job.status.Error = runErr.Error()
	default:
		job.status.Done = true
		job.status.Slug = final
	}
	cur := job.status
	a.mu.Unlock()
	job.cancel()
	a.emit(eventRenderFinished, cur)
}

// CancelRender dừng lượt render đang chạy (không lưu gì vào thư viện).
func (a *App) CancelRender() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.job != nil && a.job.status.Running {
		a.job.cancel()
	}
}

// RenderStatus trả trạng thái lượt render hiện tại / gần nhất (nil nếu chưa render lần nào).
func (a *App) RenderStatus() *RenderStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.job == nil {
		return nil
	}
	st := a.job.status
	return &st
}

// Rendering báo có đang render không.
func (a *App) Rendering() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.job != nil && a.job.status.Running
}

// previewDir — thư mục tạm cho file nghe thử (~/Sano/.tam/nghe-thu).
func (a *App) previewDir() (string, error) {
	dir := filepath.Join(a.lib.Root(), ".tam", "nghe-thu")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("tạo thư mục nghe thử: %w", err)
	}
	return dir, nil
}

// mediaURL đổi file trong ~/Sano thành URL phát trong app (kèm tham số chống cache).
func (a *App) mediaURL(file string) string {
	rel, err := a.lib.Rel(file)
	if err != nil {
		return ""
	}
	return mediaPrefix + rel + "?v=" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

func (a *App) emit(event string, data any) {
	if a.ctx == nil {
		return
	}
	wruntime.EventsEmit(a.ctx, event, data)
}

func (a *App) context() context.Context {
	if a.ctx == nil {
		return context.Background()
	}
	return a.ctx
}

// beforeClose: đang render thì hỏi có dừng không (render dài hàng giờ, đóng
// nhầm là mất công). Trả true = giữ cửa sổ.
func (a *App) beforeClose(ctx context.Context) bool {
	if !a.Rendering() {
		return false
	}
	const stop, keep = "Dừng render và thoát", "Tiếp tục render"
	choice, err := wruntime.MessageDialog(ctx, wruntime.MessageDialogOptions{
		Type:          wruntime.QuestionDialog,
		Title:         "Sano đang render sách",
		Message:       "Thoát bây giờ sẽ dừng render và bỏ phần đã làm của cuốn này.",
		Buttons:       []string{stop, keep},
		DefaultButton: keep,
		CancelButton:  keep,
	})
	if err != nil || choice != stop {
		return true
	}
	a.CancelRender()
	return false
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
