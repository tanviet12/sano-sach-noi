package bookmaker

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

// ErrProtectedFile: file có mật khẩu hoặc khoá bảo vệ. Sano không mở loại file này
// và không có tính năng gỡ khoá — người dùng chỉ nạp tài liệu mình có quyền sử dụng.
var ErrProtectedFile = errors.New("file có mật khẩu hoặc khoá bảo vệ, Sano không mở. Chỉ dùng tài liệu bạn có quyền sử dụng")

// ErrLegacyDoc: file .doc định dạng Word cũ (không phải .docx).
var ErrLegacyDoc = errors.New("file là định dạng Word cũ (.doc). Mở bằng Word và lưu lại dưới dạng .docx rồi nạp lại")

// cfbMagic — chữ ký file Compound File Binary. File .docx có mật khẩu mở được Word
// bọc trong CFB (kèm luồng EncryptionInfo); .doc định dạng cũ cũng là CFB.
var cfbMagic = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}

// encryptionInfoUTF16 — tên luồng "EncryptionInfo" trong thư mục CFB (UTF-16LE).
var encryptionInfoUTF16 = func() []byte {
	var b []byte
	for _, r := range "EncryptionInfo" {
		b = append(b, byte(r), 0)
	}
	return b
}()

// checkNotProtected từ chối file Word có khoá bảo vệ (và báo rõ khi là .doc cũ)
// trước khi mở như zip OOXML.
func checkNotProtected(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("mở file %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	head := make([]byte, len(cfbMagic))
	if _, err := io.ReadFull(f, head); err != nil || !bytes.Equal(head, cfbMagic) {
		return nil // không phải CFB → để bước mở zip tự báo nếu hỏng
	}
	data, err := io.ReadAll(io.LimitReader(f, 64<<20))
	if err != nil {
		return fmt.Errorf("đọc file %q: %w", path, err)
	}
	if bytes.Contains(data, encryptionInfoUTF16) {
		return ErrProtectedFile
	}
	return ErrLegacyDoc
}
