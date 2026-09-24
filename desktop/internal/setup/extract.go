package setup

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ulikunitz/xz"
)

// Giới hạn giải nén — chặn file nén "bom" (vài MB nén bung ra hàng chục GB)
// trước khi kịp kiểm tree hash / chạy thử.
const (
	maxTreeEntries    = 10_000    // số mục trong tarball mã nguồn
	maxTreeBytes      = 256 << 20 // tổng dung lượng sau giải nén tarball mã nguồn
	maxExtractedBytes = 1 << 30   // một file lấy ra (uv, ffmpeg)
)

// errTooBig — file nén bung ra vượt giới hạn.
var errTooBig = errors.New("file nén bung ra lớn bất thường")

// errUnsafePath — đường dẫn trong file nén thoát ra ngoài thư mục đích.
var errUnsafePath = errors.New("file nén chứa đường dẫn không an toàn")

// safeRel kiểm đường dẫn trong file nén (dạng /) và trả dạng hệ điều hành.
// Chặn đường dẫn tuyệt đối, "..", ký tự ổ đĩa Windows.
func safeRel(name string) (string, error) {
	name = strings.TrimPrefix(name, "./")
	if name == "" {
		return "", nil
	}
	clean := path.Clean(name)
	if path.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, ":") || strings.Contains(clean, `\`) {
		return "", fmt.Errorf("%w: %q", errUnsafePath, name)
	}
	return filepath.FromSlash(clean), nil
}

// stripTop bỏ thư mục gốc của tarball GitHub ("<repo>-<commit>/...").
func stripTop(name string) string {
	name = strings.TrimPrefix(name, "./")
	if i := strings.IndexByte(name, '/'); i >= 0 {
		return name[i+1:]
	}
	return ""
}

// treeEntry — một dòng của "tree hash".
type treeEntry struct {
	path string // dạng /
	line string
}

// treeHash — sha256 của các dòng "<sha256 nội dung>  <đường dẫn>" (file thường)
// hoặc "symlink:<đích>  <đường dẫn>" (liên kết), sắp theo đường dẫn. Không phụ
// thuộc cách nén lại tarball (thứ tự, thời gian, quyền file).
func treeHash(entries []treeEntry) string {
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	h := sha256.New()
	for _, e := range entries {
		io.WriteString(h, e.line)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// extractTarGz giải nén tarball GitHub vào dst (bỏ thư mục gốc), trả tree hash.
// Liên kết (symlink) chỉ ghi vào tree hash, không tạo ra đĩa (mã VieNeu không cần,
// và tránh liên kết trỏ ra ngoài). Kiểm ctx giữa các file để huỷ được.
func extractTarGz(ctx context.Context, archive, dst string) (string, error) {
	f, err := os.Open(archive)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("đọc file nén: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	var entries []treeEntry
	var count int
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("đọc file nén: %w", err)
		}
		name := stripTop(hdr.Name)
		rel, err := safeRel(name)
		if err != nil {
			return "", err
		}
		if rel == "" || hdr.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		if count++; count > maxTreeEntries {
			return "", fmt.Errorf("%w: quá %d mục", errTooBig, maxTreeEntries)
		}
		target := filepath.Join(dst, rel)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			sum, n, err := writeFile(target, tr, hdr.FileInfo().Mode(), maxTreeBytes-total)
			if err != nil {
				return "", err
			}
			total += n
			entries = append(entries, treeEntry{path: path.Clean(name), line: sum + "  " + path.Clean(name) + "\n"})
		case tar.TypeSymlink:
			entries = append(entries, treeEntry{path: path.Clean(name), line: "symlink:" + hdr.Linkname + "  " + path.Clean(name) + "\n"})
		}
	}
	if len(entries) == 0 {
		return "", errors.New("file nén rỗng")
	}
	return treeHash(entries), nil
}

// writeFile ghi r vào target (tạo thư mục cha), trả sha256 nội dung + số byte.
// Quá max byte thì dừng, xoá file dở, trả errTooBig.
func writeFile(target string, r io.Reader, mode os.FileMode, max int64) (string, int64, error) {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", 0, err
	}
	perm := os.FileMode(0o644)
	if mode&0o111 != 0 {
		perm = 0o755
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return "", 0, err
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(out, h), io.LimitReader(r, max+1))
	if err == nil && n > max {
		err = errTooBig
	}
	if err != nil {
		out.Close()
		_ = os.Remove(target)
		return "", 0, err
	}
	if err := out.Close(); err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

// extractOneFromTarGz lấy đúng một file (so theo tên cuối, vd "uv") ra dst.
func extractOneFromTarGz(archive, base, dst string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("đọc file nén: %w", err)
	}
	defer gz.Close()
	return extractOneFromTar(tar.NewReader(gz), base, dst)
}

// extractOneFromTarXz lấy đúng một file (so theo tên cuối, vd "ffmpeg") ra dst
// từ file .tar.xz (bản ffmpeg Linux của BtbN).
func extractOneFromTarXz(archive, base, dst string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	xr, err := xz.NewReader(f)
	if err != nil {
		return fmt.Errorf("đọc file nén: %w", err)
	}
	return extractOneFromTar(tar.NewReader(xr), base, dst)
}

// extractOneFromTar lấy file thường đầu tiên có tên cuối là base ra dst.
func extractOneFromTar(tr *tar.Reader, base, dst string) error {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("không thấy %s trong file nén", base)
		}
		if err != nil {
			return fmt.Errorf("đọc file nén: %w", err)
		}
		if hdr.Typeflag == tar.TypeReg && path.Base(hdr.Name) == base {
			_, _, err := writeFile(dst, tr, 0o755, maxExtractedBytes)
			return err
		}
	}
}

// extractOneFromZip lấy đúng một file (so theo tên cuối) ra dst.
func extractOneFromZip(archive, base, dst string) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return fmt.Errorf("đọc file nén: %w", err)
	}
	defer zr.Close()
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() || path.Base(zf.Name) != base {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		_, _, err = writeFile(dst, rc, 0o755, maxExtractedBytes)
		rc.Close()
		return err
	}
	return fmt.Errorf("không thấy %s trong file nén", base)
}

// gunzipTo giải nén file .gz (một file) ra dst.
func gunzipTo(archive, dst string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("đọc file nén: %w", err)
	}
	defer gz.Close()
	_, _, err = writeFile(dst, gz, 0o755, maxExtractedBytes)
	return err
}
