package main

// Phần "Cài bộ đọc" của App: tìm bộ đọc (app cài / SANO_TTS_PYTHON / cài tay),
// cài lần mở đầu có tiến độ + huỷ, gỡ bộ đọc. Logic cài nằm ở internal/setup.

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sano/desktop/internal/setup"
	"sano/desktop/internal/tts"
	ttsscripts "sano/scripts/tts"
)

// Sự kiện cài bộ đọc đẩy lên giao diện.
const (
	eventSetupProgress = "setup:progress"
	eventSetupFinished = "setup:finished"
)

type setupJob struct {
	cancel context.CancelFunc
	inst   *setup.Installer
}

// layout — vị trí bộ đọc app cài (<dữ liệu app>/tts). Lỗi khi không xác định
// được thư mục dữ liệu (không có HOME, SANO_DATA_DIR sai).
func layout() (tts.Layout, error) {
	home, _ := os.UserHomeDir()
	dir, err := tts.DataDir(os.Getenv, runtime.GOOS, home)
	if err != nil {
		return tts.Layout{GOOS: runtime.GOOS}, err
	}
	return tts.NewLayout(dir, runtime.GOOS), nil
}

// ttsRuntime — bộ đọc sẽ dùng (xem tts.Resolve về thứ tự ưu tiên).
func ttsRuntime() tts.Runtime {
	home, _ := os.UserHomeDir()
	l, _ := layout()
	o := tts.ResolveOptions{
		Getenv:  os.Getenv,
		Home:    home,
		GOOS:    runtime.GOOS,
		Layout:  l,
		Scripts: ttsscripts.Files,
		Dev:     devBuild,
	}
	if devBuild { // chỉ bản dev mới dò scripts/tts theo thư mục hiện tại / file chạy
		o.Starts = searchRoots()
	}
	return tts.Resolve(o)
}

// findFFmpeg — ffmpeg của máy hoặc bản app đã tải.
func findFFmpeg() string {
	l, _ := layout()
	extra := ""
	if l.Root != "" {
		extra = l.FFmpeg()
	}
	return tts.FindFFmpeg(os.Getenv, exec.LookPath, runtime.GOOS, extra)
}

// SetupInfo — máy của người dùng + dung lượng/thời gian cần để cài bộ đọc.
func (a *App) SetupInfo() setup.Info {
	l, _ := layout()
	return setup.MachineInfo(l, runtime.GOOS, runtime.GOARCH)
}

// SetupStatus — trạng thái lượt cài hiện tại / gần nhất (chưa cài lần nào → các
// dòng chưa chạy).
func (a *App) SetupStatus() setup.Status {
	a.mu.Lock()
	job := a.setup
	a.mu.Unlock()
	if job == nil {
		return setup.InitialStatus()
	}
	return job.inst.Status()
}

// errUninstalling — từ chối dùng / cài bộ đọc trong lúc đang gỡ.
var errUninstalling = errors.New("đang gỡ bộ đọc — đợi gỡ xong rồi thử lại")

// uninstallingNow báo đang gỡ bộ đọc.
func (a *App) uninstallingNow() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.uninstalling
}

// beginTTSUse đánh dấu một lượt dùng bộ đọc ngắn (nghe thử, câu mẫu, hỏi
// giọng) để không gỡ bộ đọc giữa chừng. Gọi hàm trả về khi dùng xong.
func (a *App) beginTTSUse() (func(), error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.uninstalling {
		return nil, errUninstalling
	}
	a.ttsUsers++
	return func() {
		a.mu.Lock()
		a.ttsUsers--
		a.mu.Unlock()
	}, nil
}

// beginTTSCheck đánh dấu một lượt kiểm tra bộ đọc (CheckTTS) để không gỡ bộ
// đọc giữa chừng. Khác beginTTSUse: không chặn cài bộ đọc.
func (a *App) beginTTSCheck() (func(), error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.uninstalling {
		return nil, errUninstalling
	}
	a.ttsChecks++
	return func() {
		a.mu.Lock()
		a.ttsChecks--
		a.mu.Unlock()
	}, nil
}

// setupBusyLocked — lý do không được bắt đầu cài bộ đọc lúc này. Gọi khi đang
// giữ a.mu: cài (uv sync) viết lại venv, không được chạy song song với render
// hay nghe thử / câu mẫu đang dùng bộ đọc.
func (a *App) setupBusyLocked() error {
	switch {
	case a.uninstalling:
		return errUninstalling
	case a.job != nil && a.job.status.Running:
		return errors.New("đang tạo sách — cài bộ đọc sau khi render xong")
	case a.ttsUsers > 0:
		return errors.New("bộ đọc đang được dùng (nghe thử, đọc câu mẫu) — thử lại sau giây lát")
	}
	return nil
}

// beginUninstall kiểm dưới khoá (ngay trước khi xoá, sau hộp xác nhận) rằng
// bộ đọc không bận rồi giữ cờ đang gỡ. Gọi hàm trả về khi gỡ xong.
func (a *App) beginUninstall() (func(), error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch {
	case a.uninstalling:
		return nil, errUninstalling
	case a.job != nil && a.job.status.Running:
		return nil, errors.New("đang tạo sách — gỡ bộ đọc sau khi render xong")
	case a.setup != nil && a.setup.inst.Status().Running:
		return nil, errors.New("đang cài bộ đọc — huỷ cài trước khi gỡ")
	case a.m4b != nil && a.m4b.status.Running:
		return nil, errors.New("đang xuất M4B — đợi xong hoặc huỷ trước khi gỡ bộ đọc")
	case a.ttsUsers > 0:
		return nil, errors.New("bộ đọc đang được dùng (nghe thử, đọc câu mẫu) — thử lại sau giây lát")
	case a.ttsChecks > 0:
		return nil, errors.New("đang kiểm tra bộ đọc — thử lại sau giây lát")
	}
	a.uninstalling = true
	return func() {
		a.mu.Lock()
		a.uninstalling = false
		a.mu.Unlock()
	}, nil
}

// Installing báo đang cài bộ đọc.
func (a *App) Installing() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.setup != nil && a.setup.inst.Status().Running
}

// StartSetup bắt đầu cài bộ đọc trong nền. Tiến độ qua sự kiện setup:progress,
// kết thúc qua setup:finished. Đang cài thì trả trạng thái hiện tại.
func (a *App) StartSetup() (*setup.Status, error) {
	if a.uninstallingNow() {
		return nil, errUninstalling
	}
	if a.Rendering() {
		return nil, errors.New("đang tạo sách — cài bộ đọc sau khi render xong")
	}
	l, err := layout()
	if err != nil {
		return nil, err
	}
	pins, err := ttsscripts.Pins()
	if err != nil {
		return nil, fmt.Errorf("đọc phiên bản ghim: %w", err)
	}

	a.mu.Lock()
	if a.setup != nil && a.setup.inst.Status().Running {
		st := a.setup.inst.Status()
		a.mu.Unlock()
		return &st, nil
	}
	if err := a.setupBusyLocked(); err != nil {
		a.mu.Unlock()
		return nil, err
	}
	ctx, cancel := context.WithCancel(a.context())
	inst := setup.New(setup.Config{
		Layout:     l,
		Pins:       pins,
		Scripts:    ttsscripts.Files,
		GOOS:       runtime.GOOS,
		GOARCH:     runtime.GOARCH,
		FindFFmpeg: findFFmpeg,
		OnProgress: func(st setup.Status) { a.emit(eventSetupProgress, st) },
	})
	a.setup = &setupJob{cancel: cancel, inst: inst}
	a.mu.Unlock()
	log.Printf("bắt đầu cài bộ đọc vào %s", l.Root)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("setup panic: %v\n%s", r, debug.Stack())
			}
			cancel()
			a.mu.Lock()
			a.voices = nil // bộ đọc đổi → hỏi lại danh sách giọng
			a.mu.Unlock()
			a.emit(eventSetupFinished, inst.Status())
		}()
		if err := inst.Run(ctx); err != nil {
			log.Printf("cài bộ đọc: %v", err)
		}
	}()
	st := inst.Status()
	st.Running = true
	return &st, nil
}

// CancelSetup dừng lượt cài đang chạy (lần sau cài tiếp từ bước dở).
func (a *App) CancelSetup() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.setup != nil && a.setup.inst.Status().Running {
		log.Printf("huỷ cài bộ đọc")
		a.setup.cancel()
	}
}

// UninstallResult — kết quả gỡ bộ đọc.
type UninstallResult struct {
	Cancelled  bool   `json:"cancelled"`
	FreedBytes int64  `json:"freedBytes"`
	Dir        string `json:"dir"`
}

// UninstallTTS hỏi xác nhận (hộp thoại của hệ điều hành) rồi xoá thư mục bộ đọc
// app đã cài. Chỉ xoá <dữ liệu app>/tts có file đánh dấu của Sano; không đụng
// sách (~/Sano/Sach), VieNeu cài tay (~/VieNeu-TTS*) hay cache Hugging Face.
func (a *App) UninstallTTS() (*UninstallResult, error) {
	if a.ctx == nil {
		return nil, errors.New("ứng dụng chưa khởi động xong")
	}
	if a.Rendering() {
		return nil, errors.New("đang tạo sách — gỡ bộ đọc sau khi render xong")
	}
	if a.Installing() {
		return nil, errors.New("đang cài bộ đọc — huỷ cài trước khi gỡ")
	}
	l, err := layout()
	if err != nil {
		return nil, err
	}
	home, _ := os.UserHomeDir()
	if err := setup.CheckRemovable(l.Root, home); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("chưa có bộ đọc nào do Sano cài để gỡ")
		}
		return nil, err
	}
	size := setup.DirSize(l.Root)
	const yes = "Gỡ bộ đọc"
	res, err := wruntime.MessageDialog(a.ctx, wruntime.MessageDialogOptions{
		Type:  wruntime.QuestionDialog,
		Title: "Gỡ bộ đọc?",
		Message: fmt.Sprintf("Xoá bộ đọc VieNeu-TTS, mô hình giọng đọc và Python do Sano cài (giải phóng %s).\n\n"+
			"Sách đã tạo vẫn giữ nguyên. Muốn tạo sách tiếp thì cài lại bộ đọc.", humanSize(size)),
		Buttons:       []string{yes, "Huỷ"},
		DefaultButton: "Huỷ",
		CancelButton:  "Huỷ",
	})
	if err != nil {
		return nil, err
	}
	if res != yes && res != "Yes" && res != "Ok" {
		return &UninstallResult{Cancelled: true}, nil
	}
	// Hộp thoại có thể mở lâu: kiểm lại dưới khoá ngay trước khi xoá.
	done, err := a.beginUninstall()
	if err != nil {
		return nil, err
	}
	defer done()
	freed, err := setup.Uninstall(l, home)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.voices = nil
	a.setup = nil
	a.mu.Unlock()
	return &UninstallResult{FreedBytes: freed, Dir: l.Root}, nil
}

// humanSize — dung lượng dễ đọc cho hộp thoại (dấu phẩy thập phân).
func humanSize(b int64) string {
	if b >= 1<<30 {
		return strings.Replace(fmt.Sprintf("%.1f GB", float64(b)/float64(1<<30)), ".", ",", 1)
	}
	return fmt.Sprintf("%d MB", (b+(1<<19))>>20)
}
