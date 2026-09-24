package main

// Điều khoản sử dụng: lưu lần người dùng đồng ý (phiên bản + thời điểm) trong thư mục
// dữ liệu của app. Nội dung điều khoản nằm ở docs/dieu-khoan-su-dung.md (giao diện nhúng),
// điều khoản đổi phiên bản thì giao diện hỏi lại.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sano/desktop/internal/tts"
)

// TermsStatus — lần đồng ý gần nhất; AcceptedVersion = 0 là chưa đồng ý.
type TermsStatus struct {
	AcceptedVersion int    `json:"acceptedVersion"`
	AcceptedAt      string `json:"acceptedAt"` // RFC 3339
}

const termsFileName = "terms.json"

func termsPath() (string, error) {
	home, _ := os.UserHomeDir()
	dir, err := tts.DataDir(os.Getenv, runtime.GOOS, home)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, termsFileName), nil
}

func readTerms(path string) TermsStatus {
	var st TermsStatus
	b, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(b, &st) != nil || st.AcceptedVersion < 0 {
		return TermsStatus{}
	}
	return st
}

func writeTerms(path string, version int, now time.Time) (TermsStatus, error) {
	if version < 1 {
		return TermsStatus{}, errors.New("phiên bản điều khoản không hợp lệ")
	}
	st := TermsStatus{AcceptedVersion: version, AcceptedAt: now.Format(time.RFC3339)}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return TermsStatus{}, err
	}
	b, _ := json.MarshalIndent(st, "", "  ")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return TermsStatus{}, err
	}
	if err := os.Rename(tmp, path); err != nil {
		return TermsStatus{}, err
	}
	return st, nil
}

// TermsStatus trả lần đồng ý gần nhất (chưa có → phiên bản 0).
func (a *App) TermsStatus() TermsStatus {
	p, err := termsPath()
	if err != nil {
		return TermsStatus{}
	}
	return readTerms(p)
}

// AcceptTerms ghi nhận người dùng đồng ý điều khoản phiên bản version.
func (a *App) AcceptTerms(version int) (TermsStatus, error) {
	p, err := termsPath()
	if err != nil {
		return TermsStatus{}, err
	}
	return writeTerms(p, version, time.Now())
}

// QuitApp đóng phần mềm (nút "Thoát Sano" ở màn điều khoản).
func (a *App) QuitApp() {
	if a.ctx != nil {
		wruntime.Quit(a.ctx)
	}
}
