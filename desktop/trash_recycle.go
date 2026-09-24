package main

import (
	"os"
	"os/exec"
)

// trashPathEnv — biến môi trường chuyển đường dẫn thư mục sách cho PowerShell.
const trashPathEnv = "SANO_TRASH_PATH"

// recycleBinScript — script PowerShell cố định đưa thư mục vào Recycle Bin.
// Đường dẫn KHÔNG ghép vào script (tên thư mục do người khác đặt có thể chứa
// dấu nháy PowerShell coi là đóng chuỗi → chạy lệnh lạ); script đọc
// $env:SANO_TRASH_PATH.
const recycleBinScript = "Add-Type -AssemblyName Microsoft.VisualBasic; " +
	"[Microsoft.VisualBasic.FileIO.FileSystem]::DeleteDirectory($env:" + trashPathEnv + ", 'OnlyErrorDialogs', 'SendToRecycleBin')"

// recycleBinCommand dựng lệnh PowerShell chuyển path vào Recycle Bin (dùng trên
// Windows; tách riêng để test được trên mọi hệ điều hành).
func recycleBinCommand(path string) *exec.Cmd {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", recycleBinScript)
	cmd.Env = append(os.Environ(), trashPathEnv+"="+path)
	return cmd
}
