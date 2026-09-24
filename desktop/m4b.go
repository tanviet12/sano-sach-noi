package main

// Xuất M4B: một file cho cả cuốn (mục lục chương, bìa) để chép sang điện thoại
// nghe bằng app sách nói bất kỳ. Chạy nền, tiến độ qua sự kiện Wails, hủy
// được; mỗi lúc một cuốn, chạy song song với render được.

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sano/desktop/internal/tts"
	"sano/internal/m4b"
)

// Sự kiện Wails của xuất M4B.
const (
	eventM4BProgress = "m4b:progress"
	eventM4BFinished = "m4b:finished"
)

// EnvM4BOut — đặt nơi lưu M4B thay cho hộp lưu file (dùng khi phát triển /
// kiểm thử tự động: hộp thoại của hệ điều hành không điều khiển được). Giá trị
// là đường dẫn .m4b hoặc một thư mục (lưu <Tên sách>.m4b trong đó). Khi đặt
// biến này, xuất xong không tự mở thư mục.
const EnvM4BOut = "SANO_M4B_OUT"

// M4BStatus — trạng thái lượt xuất M4B hiện tại / gần nhất.
type M4BStatus struct {
	Running     bool         `json:"running"`
	Done        bool         `json:"done"`
	Cancelled   bool         `json:"cancelled"`
	Error       string       `json:"error"`
	Slug        string       `json:"slug"`
	Title       string       `json:"title"`
	Path        string       `json:"path"` // file .m4b (đã lưu khi Done)
	Size        int64        `json:"size"`
	DurationSec float64      `json:"durationSec"`
	Chapters    int          `json:"chapters"`
	Progress    m4b.Progress `json:"progress"`
}

type m4bJob struct {
	cancel context.CancelFunc
	status M4BStatus
}

// ExportM4B hỏi nơi lưu (mặc định <Tên sách>.m4b trong thư mục Tải về) rồi
// xuất cuốn slug trong nền. Người dùng huỷ hộp lưu → (nil, nil).
func (a *App) ExportM4B(slug string) (*M4BStatus, error) {
	if a.exportingM4B() {
		return nil, errors.New("đang xuất M4B một cuốn khác — đợi xong hoặc huỷ trước")
	}
	d, err := a.lib.Get(slug)
	if err != nil {
		return nil, err
	}
	dir, err := a.lib.Dir(slug)
	if err != nil {
		return nil, err
	}
	book, err := m4b.FromDir(dir)
	if err != nil {
		return nil, err
	}
	ffmpeg := findFFmpeg()
	if ffmpeg == "" {
		return nil, fmt.Errorf("không tìm thấy ffmpeg. %s", tts.FFmpegHint(runtime.GOOS))
	}
	out, reveal, err := a.m4bTarget(d.Title)
	if err != nil || out == "" {
		return nil, err
	}

	a.mu.Lock()
	if a.uninstalling { // ffmpeg có thể nằm trong thư mục bộ đọc đang bị xoá
		a.mu.Unlock()
		return nil, errUninstalling
	}
	if a.m4b != nil && a.m4b.status.Running {
		a.mu.Unlock()
		return nil, errors.New("đang xuất M4B một cuốn khác — đợi xong hoặc huỷ trước")
	}
	ctx, cancel := context.WithCancel(a.context())
	job := &m4bJob{cancel: cancel, status: M4BStatus{
		Running: true, Slug: slug, Title: d.Title, Path: out,
		Progress: m4b.Progress{Phase: m4b.PhaseEncode, Tracks: len(book.Tracks)},
	}}
	a.m4b = job
	st := job.status
	a.mu.Unlock()

	go a.runM4B(ctx, job, book, ffmpeg, reveal)
	return &st, nil
}

func (a *App) runM4B(ctx context.Context, job *m4bJob, book m4b.Book, ffmpeg string, reveal bool) {
	var (
		res    *m4b.Result
		runErr error
	)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("xuất M4B panic: %v\n%s", r, debug.Stack())
			runErr = fmt.Errorf("lỗi không mong muốn: %v", r)
		}
		a.finishM4B(ctx, job, res, runErr, reveal)
	}()
	res, runErr = m4b.Export(ctx, book, job.status.Path, m4b.Options{
		FFmpeg: ffmpeg,
		Progress: func(p m4b.Progress) {
			a.mu.Lock()
			job.status.Progress = p
			cur := job.status
			a.mu.Unlock()
			a.emit(eventM4BProgress, cur)
		},
	})
}

func (a *App) finishM4B(ctx context.Context, job *m4bJob, res *m4b.Result, runErr error, reveal bool) {
	a.mu.Lock()
	job.status.Running = false
	switch {
	case errors.Is(ctx.Err(), context.Canceled):
		job.status.Cancelled = true
	case runErr != nil:
		job.status.Error = runErr.Error()
	default:
		job.status.Done = true
		job.status.Path = res.Path
		job.status.Size = res.Size
		job.status.DurationSec = res.DurationSec
		job.status.Chapters = res.Chapters
	}
	cur := job.status
	a.mu.Unlock()
	job.cancel()
	if cur.Done && reveal {
		if err := openPath(cur.Path, true); err != nil {
			log.Printf("mở thư mục chứa M4B: %v", err)
		}
	}
	a.emit(eventM4BFinished, cur)
}

// CancelM4B dừng lượt xuất đang chạy (không để lại file dở).
func (a *App) CancelM4B() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.m4b != nil && a.m4b.status.Running {
		a.m4b.cancel()
	}
}

// M4BStatus trả trạng thái lượt xuất hiện tại / gần nhất (nil nếu chưa xuất lần nào).
func (a *App) M4BStatus() *M4BStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.m4b == nil {
		return nil
	}
	st := a.m4b.status
	return &st
}

// RevealM4B mở thư mục và chọn sẵn file M4B vừa xuất.
func (a *App) RevealM4B() error {
	st := a.M4BStatus()
	if st == nil || !st.Done {
		return errors.New("chưa có file M4B nào vừa xuất")
	}
	if !fileExists(st.Path) {
		return fmt.Errorf("không thấy file %s (đã bị chuyển hoặc xoá?)", st.Path)
	}
	return openPath(st.Path, true)
}

func (a *App) exportingM4B() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.m4b != nil && a.m4b.status.Running
}

// m4bTarget chọn nơi lưu: biến SANO_M4B_OUT (dev/test) hoặc hộp lưu file của
// hệ điều hành. Trả reveal = có mở thư mục khi xong không. Huỷ → ("", false, nil).
func (a *App) m4bTarget(title string) (string, bool, error) {
	name := m4b.FileName(title)
	if v := strings.TrimSpace(os.Getenv(EnvM4BOut)); v != "" {
		if strings.EqualFold(filepath.Ext(v), ".m4b") {
			return v, false, nil
		}
		return filepath.Join(v, name), false, nil
	}
	if a.ctx == nil {
		return "", false, errors.New("ứng dụng chưa khởi động xong")
	}
	path, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:                "Lưu file M4B",
		DefaultDirectory:     downloadsDir(),
		DefaultFilename:      name,
		Filters:              []wruntime.FileFilter{{DisplayName: "Sách nói M4B (*.m4b)", Pattern: "*.m4b"}},
		CanCreateDirectories: true,
	})
	if err != nil {
		return "", false, fmt.Errorf("mở hộp lưu file: %w", err)
	}
	if path == "" {
		return "", false, nil
	}
	if !strings.EqualFold(filepath.Ext(path), ".m4b") {
		path += ".m4b"
	}
	return path, true, nil
}

// downloadsDir — thư mục Tải về (Downloads) của người dùng; không có thì thư mục nhà.
func downloadsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	if d := filepath.Join(home, "Downloads"); isDir(d) {
		return d
	}
	return home
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
