package setup

// Test cần mạng — mặc định bỏ qua, bật bằng biến môi trường:
//
//	SANO_PRINT_TREE_HASH=1  in tree hash tarball VieNeu (khi đổi VIENEU_COMMIT)
//	SANO_SETUP_INTEGRATION=1 SANO_DATA_DIR=/thư/mục/tạm  cài thật trọn luồng
//	  (thêm SANO_SETUP_FORCE_FFMPEG=1 để thử cả bước tải ffmpeg)

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	ttsscripts "sano/scripts/tts"

	"sano/desktop/internal/tts"
)

func TestPrintTreeHash(t *testing.T) {
	if os.Getenv("SANO_PRINT_TREE_HASH") == "" {
		t.Skip("đặt SANO_PRINT_TREE_HASH=1 để chạy (cần mạng)")
	}
	pins, err := ttsscripts.Pins()
	if err != nil {
		t.Fatal(err)
	}
	pins["VIENEU_TREE_SHA256"] = "0000000000000000000000000000000000000000000000000000000000000000"
	url, _, _, err := vieneuTarball(pins)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	archive := filepath.Join(dir, "v.tar.gz")
	sum, err := download(context.Background(), &http.Client{Timeout: 5 * time.Minute}, url, archive, "", maxVieNeuBytes, nil)
	if err != nil {
		t.Fatal(err)
	}
	tree, err := extractTarGz(context.Background(), archive, filepath.Join(dir, "x"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("tarball sha256 = %s", sum)
	t.Logf("VIENEU_TREE_SHA256=%s", tree)
}

// Gỡ thật bộ đọc đã cài bằng TestInstallThat (cùng SANO_DATA_DIR): đi đúng đường
// App.UninstallTTS sau khi người dùng bấm xác nhận.
func TestUninstallThat(t *testing.T) {
	if os.Getenv("SANO_SETUP_UNINSTALL") == "" {
		t.Skip("đặt SANO_SETUP_UNINSTALL=1 và SANO_DATA_DIR=<thư mục tạm đã cài> để gỡ thật")
	}
	data := os.Getenv(tts.EnvDataDir)
	if data == "" || !filepath.IsAbs(data) {
		t.Fatal("cần SANO_DATA_DIR tuyệt đối")
	}
	home, _ := os.UserHomeDir()
	l := tts.NewLayout(data, runtime.GOOS)
	freed, err := Uninstall(l, home)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(l.Root); !os.IsNotExist(err) {
		t.Fatal("thư mục bộ đọc vẫn còn")
	}
	t.Logf("đã gỡ %s, giải phóng %s", l.Root, humanMB(freed))
}

func TestInstallThat(t *testing.T) {
	if os.Getenv("SANO_SETUP_INTEGRATION") == "" {
		t.Skip("đặt SANO_SETUP_INTEGRATION=1 và SANO_DATA_DIR=<thư mục tạm> để cài thật (cần mạng, vài phút)")
	}
	data := os.Getenv(tts.EnvDataDir)
	if data == "" || !filepath.IsAbs(data) {
		t.Fatal("cần SANO_DATA_DIR là đường dẫn tuyệt đối tới thư mục tạm")
	}
	pins, err := ttsscripts.Pins()
	if err != nil {
		t.Fatal(err)
	}
	l := tts.NewLayout(data, runtime.GOOS)
	find := func() string {
		if os.Getenv("SANO_SETUP_FORCE_FFMPEG") != "" {
			if fileExists(l.FFmpeg()) {
				return l.FFmpeg()
			}
			return ""
		}
		return tts.FindFFmpeg(os.Getenv, func(string) (string, error) { return "", os.ErrNotExist }, runtime.GOOS, l.FFmpeg())
	}
	last := time.Time{}
	in := New(Config{
		Layout: l, Pins: pins, Scripts: ttsscripts.Files, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		FindFFmpeg: find,
		OnProgress: func(st Status) {
			if time.Since(last) < 3*time.Second && st.Running {
				return
			}
			last = time.Now()
			for _, s := range st.Steps {
				if s.State == StateRunning {
					t.Logf("[%5.0fs] %-7s %5.1f%% %s", st.ElapsedSec, s.Key, s.Pct, s.Detail)
				}
			}
		},
	})
	start := time.Now()
	err = in.Run(context.Background())
	st := in.Status()
	for _, s := range st.Steps {
		t.Logf("%-7s %-8s %s", s.Key, s.State, s.Detail)
	}
	if err != nil {
		t.Fatalf("cài lỗi: %v\n%s", err, st.Hint)
	}
	t.Logf("xong sau %s; thư mục bộ đọc %s = %s", time.Since(start).Round(time.Second), l.Root, humanMB(DirSize(l.Root)))
}
