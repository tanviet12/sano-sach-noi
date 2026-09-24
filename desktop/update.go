package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Kiểm tra bản mới: hỏi GitHub Releases bản phát hành mới nhất, không gửi dữ
// liệu gì của người dùng. Chỉ báo có bản mới + mở trang tải; người dùng tự tải
// bản cài (tự tải và thay cần chữ ký số, làm ở bản sau).
const (
	releasesPage      = "https://github.com/tanviet12/sano-sach-noi/releases"
	latestReleaseAPI  = "https://api.github.com/repos/tanviet12/sano-sach-noi/releases/latest"
	updateTimeout     = 10 * time.Second
	maxReleaseBytes   = 1 << 20
	maxReleaseNotes   = 8   // số dòng "có gì mới" hiện trong hộp thoại
	maxReleaseNoteLen = 200 // ký tự mỗi dòng
)

// UpdateInfo — kết quả kiểm tra bản mới cho giao diện.
type UpdateInfo struct {
	Available bool     `json:"available"`
	Version   string   `json:"version"`   // "0.2.0" (không có "v")
	Published string   `json:"published"` // RFC3339, rỗng nếu không rõ
	Notes     []string `json:"notes"`     // các dòng gạch đầu dòng trong ghi chú phát hành
	URL       string   `json:"url"`       // trang tải bản đó (luôn trên github.com/tanviet12/sano-sach-noi)
}

var semverRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

// CheckUpdate hỏi GitHub bản phát hành mới nhất. Lỗi mạng hay repo chưa công
// khai (404) trả lỗi để giao diện báo "không kiểm tra được"; bản dev không so.
func (a *App) CheckUpdate() (UpdateInfo, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return checkRelease(ctx, &http.Client{Timeout: updateTimeout}, latestReleaseAPI, a.Version())
}

func checkRelease(ctx context.Context, client *http.Client, apiURL, current string) (UpdateInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return UpdateInfo{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Sano-desktop")
	resp, err := client.Do(req)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("không kết nối được GitHub: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return UpdateInfo{}, fmt.Errorf("GitHub trả mã %d", resp.StatusCode)
	}
	var rel struct {
		TagName     string `json:"tag_name"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		Draft       bool   `json:"draft"`
		Prerelease  bool   `json:"prerelease"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxReleaseBytes)).Decode(&rel); err != nil {
		return UpdateInfo{}, fmt.Errorf("đọc thông tin bản phát hành: %w", err)
	}
	latest, ok := parseSemver(rel.TagName)
	if !ok || rel.Draft || rel.Prerelease {
		return UpdateInfo{}, errors.New("bản phát hành mới nhất không có số phiên bản hợp lệ")
	}
	info := UpdateInfo{
		Version:   strings.TrimPrefix(rel.TagName, "v"),
		Published: rel.PublishedAt,
		Notes:     releaseNotes(rel.Body),
		URL:       releasesPage + "/tag/" + rel.TagName, // tag đã khớp semverRe
	}
	if cur, ok := parseSemver(current); ok {
		info.Available = newer(latest, cur)
	}
	return info, nil
}

func parseSemver(s string) ([3]int, bool) {
	m := semverRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return [3]int{}, false
	}
	var v [3]int
	for i := range v {
		n, err := strconv.Atoi(m[i+1])
		if err != nil {
			return [3]int{}, false
		}
		v[i] = n
	}
	return v, true
}

func newer(a, b [3]int) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

// releaseNotes lấy các dòng gạch đầu dòng ("- ", "* ") trong ghi chú phát hành,
// bỏ dấu markdown đơn giản, giới hạn số dòng và độ dài.
func releaseNotes(body string) []string {
	var out []string
	for _, ln := range strings.Split(body, "\n") {
		ln = strings.TrimSpace(ln)
		if !strings.HasPrefix(ln, "- ") && !strings.HasPrefix(ln, "* ") {
			continue
		}
		ln = strings.TrimSpace(strings.NewReplacer("**", "", "`", "").Replace(ln[2:]))
		if ln == "" {
			continue
		}
		if r := []rune(ln); len(r) > maxReleaseNoteLen {
			ln = string(r[:maxReleaseNoteLen]) + "…"
		}
		out = append(out, ln)
		if len(out) == maxReleaseNotes {
			break
		}
	}
	return out
}
