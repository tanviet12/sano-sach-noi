//go:build !darwin && !windows && !linux

package main

import "errors"

func detectInstall() installInfo {
	return installInfo{err: errors.New("hệ điều hành này chưa tự cập nhật được — tải bản mới ở trang phát hành")}
}

func applyUpdate(installInfo, string, string, bool) error {
	return errors.New("hệ điều hành này chưa tự cập nhật được")
}

func cleanupAfterUpdate() {}
