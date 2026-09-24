package main

// Phần Thư viện của App: đọc các cuốn đã tạo trong ~/Sano/Sach, mở thư mục,
// chọn ảnh bìa.

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sano/desktop/internal/library"
)

// BookView — một cuốn trong thư viện, kèm URL bìa để hiện.
type BookView struct {
	library.Book
	CoverURL string `json:"coverUrl"`
}

// TrackView — một tiểu mục phát được, kèm URL MP3.
type TrackView struct {
	library.Track
	URL string `json:"url"`
}

// BookDetail — cuốn sách + danh sách tiểu mục cho trình phát.
type BookDetail struct {
	BookView
	Dir    string      `json:"dir"`
	Tracks []TrackView `json:"tracks"`
}

// LibraryInfo — thư viện + nơi lưu (hiện dưới tiêu đề Thư viện).
type LibraryInfo struct {
	Dir   string     `json:"dir"`
	Books []BookView `json:"books"`
}

// Library liệt kê các cuốn đã tạo, mới nhất trước.
func (a *App) Library() (*LibraryInfo, error) {
	books, err := a.lib.List()
	if err != nil {
		return nil, err
	}
	out := &LibraryInfo{Dir: a.lib.BooksRoot(), Books: make([]BookView, 0, len(books))}
	for _, b := range books {
		out.Books = append(out.Books, a.bookView(b))
	}
	return out, nil
}

// LibrarySize — tổng dung lượng các cuốn đã tạo (byte).
func (a *App) LibrarySize() (int64, error) {
	return a.lib.Size()
}

// Book đọc một cuốn để phát.
func (a *App) Book(slug string) (*BookDetail, error) {
	d, err := a.lib.Get(slug)
	if err != nil {
		return nil, err
	}
	dir, _ := a.lib.Dir(slug)
	out := &BookDetail{BookView: a.bookView(d.Book), Dir: dir, Tracks: make([]TrackView, 0, len(d.Tracks))}
	for _, t := range d.Tracks {
		out.Tracks = append(out.Tracks, TrackView{Track: t, URL: mediaPrefix + t.File})
	}
	return out, nil
}

func (a *App) bookView(b library.Book) BookView {
	v := BookView{Book: b}
	if b.Cover != "" {
		v.CoverURL = mediaPrefix + b.Cover
	}
	return v
}

// OpenBookFolder hiện thư mục của một cuốn trong trình quản lý file của máy
// (macOS: chọn sẵn trong Finder, không "mở" để khỏi chạy nhầm gói ứng dụng).
func (a *App) OpenBookFolder(slug string) error {
	dir, err := a.lib.Dir(slug)
	if err != nil {
		return err
	}
	return openPath(dir, false)
}

// RevealBookZip mở thư mục và chọn sẵn gói zip của cuốn sách.
func (a *App) RevealBookZip(slug string) error {
	d, err := a.lib.Get(slug)
	if err != nil {
		return err
	}
	if d.Zip == "" {
		return errors.New("cuốn này chưa có gói zip")
	}
	full, err := a.lib.Resolve(d.Zip)
	if err != nil {
		return err
	}
	return openPath(full, true)
}

// DeleteBook hỏi xác nhận rồi chuyển cuốn sách vào Thùng rác của máy (lấy lại
// được). Người dùng huỷ → ("", nil). Trả nơi đã chuyển tới.
func (a *App) DeleteBook(slug string) (string, error) {
	if a.ctx == nil {
		return "", errors.New("ứng dụng chưa khởi động xong")
	}
	dir, err := a.lib.Dir(slug)
	if err != nil {
		return "", err
	}
	title := slug
	if d, err := a.lib.Get(slug); err == nil && d.Title != "" {
		title = d.Title
	}
	const yes = "Chuyển vào Thùng rác"
	res, err := wruntime.MessageDialog(a.ctx, wruntime.MessageDialogOptions{
		Type:          wruntime.QuestionDialog,
		Title:         "Xoá sách?",
		Message:       fmt.Sprintf("Chuyển \"%s\" vào Thùng rác? Bạn vẫn lấy lại được từ Thùng rác nếu cần.", title),
		Buttons:       []string{yes, "Huỷ"},
		DefaultButton: "Huỷ",
		CancelButton:  "Huỷ",
	})
	if err != nil {
		return "", err
	}
	if res != yes && res != "Yes" && res != "Ok" {
		return "", nil
	}
	return trashBook(dir, a.lib.Root())
}

// UpdateBookInfo sửa tên, tác giả, danh mục của một cuốn (metadata.json + gói
// zip). Không đổi thư mục, không đọc lại lời giới thiệu.
func (a *App) UpdateBookInfo(slug string, info library.Info) (*BookView, error) {
	d, err := a.lib.UpdateInfo(slug, info)
	if err != nil {
		return nil, err
	}
	v := a.bookView(d.Book)
	return &v, nil
}

// OpenLibraryFolder hiện thư mục ~/Sano/Sach trong trình quản lý file.
func (a *App) OpenLibraryFolder() error {
	if err := os.MkdirAll(a.lib.BooksRoot(), 0o755); err != nil {
		return err
	}
	return openPath(a.lib.BooksRoot(), false)
}

// CoverFile — ảnh bìa người dùng chọn, kèm data URL để xem trước.
type CoverFile struct {
	Path    string `json:"path"`
	DataURL string `json:"dataUrl"`
}

// maxCoverBytes — bìa lớn hơn thì từ chối (ảnh bìa thường < 2 MB).
const maxCoverBytes = 10 << 20

var coverTypes = map[string]string{".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png", ".webp": "image/webp"}

// ChooseCover mở hộp chọn ảnh bìa (jpg/png/webp). Huỷ → (nil, nil).
func (a *App) ChooseCover() (*CoverFile, error) {
	if a.ctx == nil {
		return nil, errors.New("ứng dụng chưa khởi động xong")
	}
	path, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title:   "Chọn ảnh bìa",
		Filters: []wruntime.FileFilter{{DisplayName: "Ảnh (*.jpg, *.png, *.webp)", Pattern: "*.jpg;*.jpeg;*.png;*.webp"}},
	})
	if err != nil {
		return nil, fmt.Errorf("mở hộp chọn ảnh: %w", err)
	}
	if path == "" {
		return nil, nil
	}
	return describeCover(path)
}

func describeCover(path string) (*CoverFile, error) {
	ctype, ok := coverTypes[strings.ToLower(filepath.Ext(path))]
	if !ok {
		return nil, errors.New("ảnh bìa phải là jpg, png hoặc webp")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("đọc ảnh bìa: %w", err)
	}
	if info.Size() > maxCoverBytes {
		return nil, errors.New("ảnh bìa quá lớn (tối đa 10 MB)")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("đọc ảnh bìa: %w", err)
	}
	return &CoverFile{Path: path, DataURL: "data:" + ctype + ";base64," + base64.StdEncoding.EncodeToString(data)}, nil
}

// openCmd — lệnh mở trình quản lý file. WinCmdLine (chỉ Windows) là dòng lệnh
// nguyên văn, thay cho cách Go tự ghép tham số.
type openCmd struct {
	Name       string
	Args       []string
	WinCmdLine string
}

// openCommand dựng lệnh mở thư mục (hoặc chọn sẵn file nếu reveal) theo hệ
// điều hành goos. Tách riêng để test được trên mọi máy.
//
//   - macOS: luôn "open -R" (hiện và chọn sẵn trong Finder). "open <thư mục>"
//     sẽ CHẠY thư mục nếu nó là gói ứng dụng (X.app, hoặc thư mục gắn cờ gói),
//     còn "-R" chỉ hiện, không bao giờ chạy.
//   - Windows: explorer tách tham số theo dấu phẩy và không cần ngoặc kép cho
//     đường dẫn không có dấu cách → tự ghép dòng lệnh, luôn đặt đường dẫn trong
//     ngoặc kép (tên file Windows không chứa được ").
//   - Linux: xdg-open thư mục (reveal thì mở thư mục chứa file).
func openCommand(goos, path string, reveal bool) openCmd {
	switch goos {
	case "darwin":
		return openCmd{Name: "open", Args: []string{"-R", path}}
	case "windows":
		if reveal {
			return openCmd{Name: "explorer", Args: []string{"/select,", path}, WinCmdLine: `explorer.exe /select,"` + path + `"`}
		}
		return openCmd{Name: "explorer", Args: []string{path}, WinCmdLine: `explorer.exe "` + path + `"`}
	default:
		if reveal {
			path = filepath.Dir(path)
		}
		return openCmd{Name: "xdg-open", Args: []string{path}}
	}
}

// openPath mở thư mục (hoặc chọn sẵn file nếu reveal) bằng trình quản lý file.
func openPath(path string, reveal bool) error {
	oc := openCommand(runtime.GOOS, path, reveal)
	cmd := exec.Command(oc.Name, oc.Args...)
	setRawCmdLine(cmd, oc.WinCmdLine)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("mở thư mục: %w", err)
	}
	go func() { // dọn tiến trình con, không chờ
		defer func() { _ = recover() }()
		_ = cmd.Wait()
	}()
	return nil
}
