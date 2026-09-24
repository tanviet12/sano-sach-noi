package main

import "syscall"

// moveToSystemTrash đưa thư mục vào Recycle Bin qua PowerShell (không bật cửa sổ).
// Đường dẫn đi qua biến môi trường, không ghép vào script (xem recycleBinCommand).
func moveToSystemTrash(path string) error {
	cmd := recycleBinCommand(path)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return cmd.Run()
}
