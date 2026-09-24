package main

import (
	"reflect"
	"testing"
)

// macOS không bao giờ "open" thẳng thư mục (thư mục X.app sẽ bị chạy): luôn -R.
func TestOpenCommand_Darwin_AlwaysReveal(t *testing.T) {
	for _, reveal := range []bool{false, true} {
		got := openCommand("darwin", "/Users/a/Sano/Sach/Tro-Choi.app", reveal)
		want := openCmd{Name: "open", Args: []string{"-R", "/Users/a/Sano/Sach/Tro-Choi.app"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("reveal=%v: got %+v, want %+v", reveal, got, want)
		}
	}
}

// Windows: đường dẫn luôn trong ngoặc kép, dấu phẩy không tách thành tham số khác.
func TestOpenCommand_Windows_QuotesPath(t *testing.T) {
	p := `C:\Users\a\Sano\Sach\book,calc.exe`
	if got := openCommand("windows", p, false).WinCmdLine; got != `explorer.exe "C:\Users\a\Sano\Sach\book,calc.exe"` {
		t.Errorf("mở thư mục: %s", got)
	}
	if got := openCommand("windows", p+`\book-x.zip`, true).WinCmdLine; got != `explorer.exe /select,"C:\Users\a\Sano\Sach\book,calc.exe\book-x.zip"` {
		t.Errorf("chọn sẵn file: %s", got)
	}
}

func TestOpenCommand_Linux(t *testing.T) {
	if got := openCommand("linux", "/home/a/Sano/Sach/x", false); got.Name != "xdg-open" || got.Args[0] != "/home/a/Sano/Sach/x" {
		t.Errorf("mở thư mục: %+v", got)
	}
	if got := openCommand("linux", "/home/a/Sano/Sach/x/book-x.zip", true); got.Args[0] != "/home/a/Sano/Sach/x" {
		t.Errorf("reveal phải mở thư mục chứa file: %+v", got)
	}
}
