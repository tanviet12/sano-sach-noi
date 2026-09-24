package setup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sano/desktop/internal/tts"
)

// ErrNotRemovable — thư mục không phải thư mục bộ đọc do Sano tạo (hoặc chạm
// vào dữ liệu cần giữ) → không xoá.
var ErrNotRemovable = errors.New("không xoá: thư mục này không phải bộ đọc do Sano cài")

// userData — dữ liệu của người dùng mà gỡ bộ đọc tuyệt đối không được đụng: sách
// đã tạo, VieNeu cài tay, cache Hugging Face dùng chung.
func userData(home string) []string {
	if home == "" {
		return nil
	}
	return []string{
		filepath.Join(home, "Sano"),
		filepath.Join(home, tts.VenvDir),
		filepath.Join(home, "VieNeu-TTS"),
		filepath.Join(home, ".cache", "huggingface"),
	}
}

// CheckRemovable kiểm chặt trước khi xoá thư mục bộ đọc root:
//   - đường dẫn tuyệt đối, đã chuẩn hoá, tên cuối đúng "tts", không phải liên kết;
//   - có file đánh dấu .sano-tts do Sano ghi;
//   - không chứa HOME hay dữ liệu người dùng, cũng không nằm trong dữ liệu người
//     dùng (~/Sano, ~/VieNeu-TTS*, ~/.cache/huggingface).
func CheckRemovable(root, home string) error {
	if root == "" || !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return fmt.Errorf("%w (đường dẫn không hợp lệ: %q)", ErrNotRemovable, root)
	}
	if filepath.Base(root) != tts.RuntimeDirName || filepath.Dir(root) == root {
		return fmt.Errorf("%w (%s)", ErrNotRemovable, root)
	}
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if info.Mode()&fs.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("%w (%s là liên kết hoặc không phải thư mục)", ErrNotRemovable, root)
	}
	mark, err := os.ReadFile(filepath.Join(root, tts.MarkerFile))
	if err != nil || !strings.HasPrefix(string(mark), "sano-tts") {
		return fmt.Errorf("%w (thiếu file đánh dấu %s)", ErrNotRemovable, tts.MarkerFile)
	}
	if home != "" && within(filepath.Clean(home), root) {
		return fmt.Errorf("%w (chứa thư mục người dùng %s)", ErrNotRemovable, home)
	}
	for _, p := range userData(home) {
		if within(p, root) || within(root, p) {
			return fmt.Errorf("%w (chạm vào %s)", ErrNotRemovable, p)
		}
	}
	return nil
}

// within báo child nằm trong (hoặc bằng) parent.
func within(child, parent string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// Uninstall xoá thư mục bộ đọc app đã cài sau khi CheckRemovable qua. Trả số byte
// đã giải phóng. Thư mục dữ liệu app (cha) cũng xoá nếu đã rỗng.
func Uninstall(l tts.Layout, home string) (int64, error) {
	if err := CheckRemovable(l.Root, home); err != nil {
		return 0, err
	}
	size := DirSize(l.Root)
	if err := os.RemoveAll(l.Root); err != nil {
		return 0, fmt.Errorf("xoá bộ đọc: %w", err)
	}
	_ = os.Remove(filepath.Dir(l.Root)) // chỉ xoá được khi rỗng
	return size, nil
}

// DirSize — tổng dung lượng file thường trong các thư mục (bỏ qua liên kết).
func DirSize(paths ...string) int64 {
	var total int64
	for _, root := range paths {
		_ = filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.Type().IsRegular() {
				if info, err := d.Info(); err == nil {
					total += info.Size()
				}
			}
			return nil
		})
	}
	return total
}

// humanMB / humanGB — dung lượng kiểu Việt Nam (dấu phẩy thập phân).
func humanMB(b int64) string {
	if b >= 1<<30 {
		return humanGB(b)
	}
	return strconv.FormatInt((b+(1<<19))>>20, 10) + " MB"
}

func humanGB(b int64) string {
	s := strconv.FormatFloat(float64(b)/float64(1<<30), 'f', 1, 64)
	return strings.Replace(s, ".", ",", 1) + " GB"
}
