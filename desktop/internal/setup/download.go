package setup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ErrChecksum — file tải về không khớp SHA256 ghim.
var ErrChecksum = errors.New("file tải về không khớp mã SHA256 ghim")

// ErrTooLarge — file tải về lớn hơn giới hạn cho phép.
var ErrTooLarge = errors.New("file tải về lớn bất thường")

// ErrNetwork — không tải được (mất mạng, máy chủ lỗi, bị ngắt giữa chừng).
var ErrNetwork = errors.New("lỗi mạng")

// progressFunc nhận số byte đã tải + tổng (tổng = -1 nếu máy chủ không báo).
type progressFunc func(done, total int64)

// progressEvery — khoảng cách tối thiểu giữa hai lần báo tiến độ tải.
const progressEvery = 200 * time.Millisecond

// download tải url về dst, kiểm SHA256 (wantSHA rỗng = chỉ trả mã băm để kiểm
// sau). Ghi vào dst.part rồi đổi tên: lỗi/huỷ giữa chừng thì xoá file dở, không
// bao giờ để lại dst sai nội dung. Quá maxBytes thì dừng (chống nguồn bị giả mạo
// đổ dữ liệu vô hạn làm đầy ổ đĩa). Trả SHA256 thực của file.
func download(ctx context.Context, client *http.Client, url, dst, wantSHA string, maxBytes int64, onProgress progressFunc) (string, error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Sano-desktop")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: không tải được từ %s (%v)", ErrNetwork, hostOf(url), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %s trả mã %d", ErrNetwork, hostOf(url), resp.StatusCode)
	}

	part := dst + ".part"
	f, err := os.Create(part)
	if err != nil {
		return "", err
	}
	ok := false
	defer func() {
		if !ok {
			_ = f.Close()
			_ = os.Remove(part)
		}
	}()

	h := sha256.New()
	total := resp.ContentLength
	if total > maxBytes {
		return "", fmt.Errorf("%w (%s báo %d MB, tối đa %d MB)", ErrTooLarge, hostOf(url), total>>20, maxBytes>>20)
	}
	var done int64
	last := time.Time{}
	buf := make([]byte, 256<<10)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := f.Write(buf[:n]); err != nil {
				return "", err
			}
			h.Write(buf[:n])
			done += int64(n)
			if done > maxBytes {
				return "", fmt.Errorf("%w (%s vượt %d MB)", ErrTooLarge, hostOf(url), maxBytes>>20)
			}
			if onProgress != nil && time.Since(last) >= progressEvery {
				last = time.Now()
				onProgress(done, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			return "", fmt.Errorf("%w: tải từ %s bị ngắt (%v)", ErrNetwork, hostOf(url), rerr)
		}
	}
	if onProgress != nil {
		onProgress(done, total)
	}
	got := hex.EncodeToString(h.Sum(nil))
	if wantSHA != "" && !strings.EqualFold(got, wantSHA) {
		return got, fmt.Errorf("%w (%s: có %s, cần %s)", ErrChecksum, filepath.Base(dst), got[:12], strings.ToLower(wantSHA)[:12])
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(part, dst); err != nil {
		return "", err
	}
	ok = true
	return got, nil
}

func hostOf(url string) string {
	s := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	if i := strings.IndexByte(s, '/'); i > 0 {
		return s[:i]
	}
	return s
}
