package main

import (
	"context"
	"errors"
	"fmt"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sano/desktop/internal/library"
)

// Nhập sách từ gói zip (wireframe D6). Kiểm tra an toàn nằm ở library/importzip.go.

const eventImportProgress = "import:progress"

// ImportProgress — tiến độ giải nén gửi lên giao diện.
type ImportProgress struct {
	Done  int `json:"done"`
	Total int `json:"total"`
}

// ChooseBookZip mở hộp chọn gói sách; huỷ → chuỗi rỗng.
func (a *App) ChooseBookZip() (string, error) {
	if a.ctx == nil {
		return "", errors.New("ứng dụng chưa khởi động xong")
	}
	path, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title:   "Chọn gói sách để nhập",
		Filters: []wruntime.FileFilter{{DisplayName: "Gói sách Sano (*.zip)", Pattern: "*.zip"}},
	})
	if err != nil {
		return "", fmt.Errorf("mở hộp chọn file: %w", err)
	}
	return path, nil
}

// PreviewBookZip kiểm gói và trả thông tin xem trước (không ghi gì).
func (a *App) PreviewBookZip(path string) (*library.ImportPreview, error) {
	return a.lib.PreviewImport(path)
}

// ImportBookZip nhập gói vào thư viện. replaceSlug: mã cuốn trùng mà người dùng đã
// thấy ở hộp xem trước và chọn Thay thế (rỗng = giữ cả hai nếu trùng). Chỉ chuyển
// cuốn cũ vào Thùng rác khi gói hiện tại đúng là trùng cuốn đó — gói bị tráo sau
// lúc xem trước thì dừng. Mỗi lúc chỉ một lượt nhập. Trả mã sách đã nhập.
func (a *App) ImportBookZip(path, replaceSlug string) (string, error) {
	ctx, cancel := context.WithCancel(a.context())
	a.mu.Lock()
	if a.importCancel != nil {
		a.mu.Unlock()
		cancel()
		return "", errors.New("đang nhập một cuốn khác, đợi xong rồi thử lại")
	}
	a.importCancel = cancel
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		a.importCancel = nil
		a.mu.Unlock()
		cancel()
	}()

	replace := replaceSlug != ""
	work, slug, existing, err := a.lib.PrepareImport(ctx, path, replace, func(done, total int) {
		a.emit(eventImportProgress, ImportProgress{Done: done, Total: total})
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return "", errors.New("đã huỷ nhập sách")
		}
		return "", err
	}
	if ctx.Err() != nil {
		a.lib.CancelPrepared(work)
		return "", errors.New("đã huỷ nhập sách")
	}
	if replace && existing != "" {
		if existing != replaceSlug {
			a.lib.CancelPrepared(work)
			return "", errors.New("gói sách đã thay đổi sau lúc xem trước, hãy chọn lại file")
		}
		dir, derr := a.lib.Dir(existing)
		if derr == nil {
			if _, derr = trashBook(dir, a.lib.Root()); derr != nil {
				a.lib.CancelPrepared(work)
				return "", fmt.Errorf("không chuyển được cuốn cũ vào Thùng rác: %w", derr)
			}
		}
	}
	name, err := a.lib.Commit(work, slug)
	if err != nil {
		a.lib.CancelPrepared(work)
		return "", err
	}
	return name, nil
}

// CancelImport huỷ lượt nhập đang chạy (không có thì thôi).
func (a *App) CancelImport() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.importCancel != nil {
		a.importCancel()
	}
}
