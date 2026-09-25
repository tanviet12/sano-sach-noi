package main

import (
	"os"
	"strconv"
	"time"
)

// waitFlag — bản mới vừa thay được mở kèm "--sano-doi <pid bản cũ>": đợi bản cũ
// thoát hẳn rồi mới mở cửa sổ (tránh hai bản cùng dùng thư mục dữ liệu WebView).
const (
	waitFlag    = "--sano-doi"
	waitOldMax  = 20 * time.Second
	waitOldStep = 200 * time.Millisecond
)

// oldPID đọc pid sau waitFlag trong tham số dòng lệnh (0 nếu không có / sai).
func oldPID(args []string) int {
	for i, a := range args {
		if a == waitFlag && i+1 < len(args) {
			if pid, err := strconv.Atoi(args[i+1]); err == nil && pid > 0 && pid != os.Getpid() {
				return pid
			}
		}
	}
	return 0
}

// startAfterUpdate chạy đầu main: đợi bản cũ thoát (tối đa waitOldMax), dọn
// phần còn lại của lần thay bản và file cài đã tải.
func startAfterUpdate(args []string) {
	if pid := oldPID(args); pid != 0 {
		for deadline := time.Now().Add(waitOldMax); time.Now().Before(deadline) && processAlive(pid); {
			time.Sleep(waitOldStep)
		}
	}
	cleanupAfterUpdate()
	if dir, err := updateDir(); err == nil {
		go os.RemoveAll(dir) // Windows: bộ cài có thể còn đang chạy từ đây → xoá hỏng, lần sau xoá tiếp
	}
}
