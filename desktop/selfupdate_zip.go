package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
)

// Dùng cho bản chạy ngay trên Windows (selfupdate_windows.go); để chung cho
// test chạy được trên mọi máy.

// extractExe lấy đúng Sano.exe trong file .zip bản chạy ngay ra dest.
func extractExe(zipFile, dest string) error {
	zr, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("mở file .zip bản mới: %w", err)
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name != "Sano.exe" {
			continue
		}
		if f.UncompressedSize64 > maxInstallerBytes {
			return errors.New("Sano.exe trong file .zip lớn bất thường")
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return err
		}
		n, err := io.Copy(out, io.LimitReader(rc, maxInstallerBytes+1))
		if cerr := out.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return err
		}
		if n > maxInstallerBytes {
			return errors.New("Sano.exe trong file .zip lớn bất thường")
		}
		return nil
	}
	return errors.New("file .zip bản mới không có Sano.exe")
}
