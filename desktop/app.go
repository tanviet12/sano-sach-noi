package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sano/desktop/internal/library"
	"sano/desktop/internal/setup"
	"sano/desktop/internal/tts"
	"sano/internal/bookmaker"
	ttsscripts "sano/scripts/tts"
)

// App là phần Go gắn (bind) vào giao diện: mỗi method public gọi được từ JS
// qua wailsjs/go/main/App.
type App struct {
	ctx context.Context
	lib *library.Library

	mu     sync.Mutex // bảo vệ job, voices, setup, m4b, importCancel, uninstalling, ttsUsers, ttsChecks
	job    *renderJob
	voices []bookmaker.Voice
	setup  *setupJob
	m4b    *m4bJob
	// importCancel — huỷ lượt nhập sách đang chạy (nil = không có lượt nào).
	importCancel context.CancelFunc
	// uninstalling — đang xoá bộ đọc; ttsUsers — số lượt dùng bộ đọc ngắn đang
	// chạy (nghe thử, câu mẫu, hỏi giọng). Hai bên loại trừ nhau (xem setup.go).
	uninstalling bool
	ttsUsers     int
	// ttsChecks — số lượt kiểm tra bộ đọc đang chạy (CheckTTS). Kiểm tra giải nén
	// script vào thư mục bộ đọc nên chặn gỡ, nhưng không chặn cài (xem setup.go).
	ttsChecks int

	updateState // tự cập nhật (selfupdate.go), khoá riêng updMu
}

// NewApp tạo App với thư viện ~/Sano; ctx gắn ở startup.
func NewApp() *App {
	lib, err := library.Default()
	if err != nil {
		lib = library.New(filepath.Join(os.TempDir(), library.DirName))
	}
	return &App{lib: lib}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := seedSampleBook(a.lib); err != nil {
		log.Printf("sách mẫu: %v", err)
	}
}

// searchRoots — nơi bắt đầu tìm VERSION và scripts/tts: thư mục hiện tại và
// thư mục chứa file chạy.
func searchRoots() []string {
	roots := []string{}
	if wd, err := os.Getwd(); err == nil {
		roots = append(roots, wd)
	}
	if exe, err := os.Executable(); err == nil {
		roots = append(roots, filepath.Dir(exe))
	}
	return roots
}

// Version trả phiên bản app (ldflags → file VERSION → "dev").
func (a *App) Version() string {
	return resolveVersion(version, searchRoots()...)
}

// CheckTTS kiểm tra bộ đọc VieNeu-TTS: python (bộ đọc app cài → SANO_TTS_PYTHON
// → ~/VieNeu-TTS-v3), script đọc giọng nhúng trong app, mô hình ghim, ffmpeg.
//
// Đang gỡ bộ đọc thì không kiểm (kiểm tra ghi script vào thư mục bộ đọc, sẽ tạo
// lại thư mục đang bị xoá).
func (a *App) CheckTTS() tts.Status {
	var dataDir string
	if l, err := layout(); err == nil {
		dataDir = l.Root
	}
	release, err := a.beginTTSCheck()
	if err != nil {
		return tts.Status{DataDir: dataDir, Message: "Đang gỡ bộ đọc", Detail: err.Error()}
	}
	defer release()
	st := tts.Check(a.context(), ttsRuntime(), tts.ExecRunner)
	st.DataDir = dataDir
	st.FFmpeg = findFFmpeg()
	if st.Ready && st.FFmpeg == "" {
		st.Ready = false
		st.Message = "Thiếu ffmpeg"
		st.Detail = "Sano cần ffmpeg để ghi file MP3. " + tts.FFmpegHint(runtime.GOOS)
	}
	// Vẫn Ready: bộ đọc cũ chạy được (không chặn người đang offline), giao diện
	// mời cập nhật khi mở app và trong Cài đặt.
	if st.Ready && st.Source == tts.SourceApp && ttsNeedsResync() {
		st.Update = true
		st.Message = "Bộ đọc cần cập nhật thư viện"
		st.Detail = "Bản Sano này vá lỗi bảo mật trong thư viện Python của bộ đọc. Chỉ tải lại vài thư viện, mô hình giọng đọc giữ nguyên."
	}
	return st
}

// ttsNeedsResync — bộ đọc app cài có thư viện Python chưa khớp bản Sano này.
func ttsNeedsResync() bool {
	l, err := layout()
	if err != nil {
		return false
	}
	pins, err := ttsscripts.Pins()
	if err != nil {
		return false
	}
	return setup.NeedsResync(l, pins, ttsscripts.Files)
}

// DocxFile mô tả file Word người dùng chọn.
type DocxFile struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// ErrNotDocx — file không phải .docx.
var ErrNotDocx = errors.New("chỉ nhận file Word .docx")

// ErrRightsNotConfirmed — chưa tick xác nhận có quyền dùng tài liệu (bước Nghe thử).
var ErrRightsNotConfirmed = errors.New("hãy xác nhận bạn có quyền dùng tài liệu này trước khi render")

// ChooseDocx mở hộp chọn file của hệ điều hành, chỉ lọc .docx.
// Người dùng bấm huỷ → trả (nil, nil).
func (a *App) ChooseDocx() (*DocxFile, error) {
	if a.ctx == nil {
		return nil, errors.New("ứng dụng chưa khởi động xong")
	}
	path, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title:   "Chọn file Word",
		Filters: []wruntime.FileFilter{{DisplayName: "File Word (*.docx)", Pattern: "*.docx"}},
	})
	if err != nil {
		return nil, fmt.Errorf("mở hộp chọn file: %w", err)
	}
	if path == "" {
		return nil, nil
	}
	return describeDocx(path)
}

// DescribeDocx đọc thông tin file .docx được kéo thả vào cửa sổ.
func (a *App) DescribeDocx(path string) (*DocxFile, error) {
	return describeDocx(path)
}

func describeDocx(path string) (*DocxFile, error) {
	if !strings.EqualFold(filepath.Ext(path), ".docx") {
		return nil, ErrNotDocx
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("đọc file: %w", err)
	}
	if info.IsDir() {
		return nil, ErrNotDocx
	}
	return &DocxFile{Path: path, Name: info.Name(), Size: info.Size()}, nil
}
