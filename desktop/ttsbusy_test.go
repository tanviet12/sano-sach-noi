package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"sano/desktop/internal/library"
)

// Đang gỡ bộ đọc: kiểm tra bộ đọc không được tạo lại thư mục bộ đọc (trước đây
// CheckTTS giải nén script vào <dữ liệu>/tts giữa lúc RemoveAll).
func TestCheckTTS_WhileUninstalling_DoesNotRecreateDir(t *testing.T) {
	data := t.TempDir()
	t.Setenv("SANO_DATA_DIR", data)
	a := &App{lib: library.New(t.TempDir())}
	done, err := a.beginUninstall()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	st := a.CheckTTS()
	if st.Ready {
		t.Error("đang gỡ thì không được báo sẵn sàng")
	}
	if _, err := os.Stat(filepath.Join(data, "tts")); !os.IsNotExist(err) {
		t.Errorf("CheckTTS khi đang gỡ đã tạo lại thư mục bộ đọc (err=%v)", err)
	}
}

// Kiểm tra bộ đọc đang chạy thì không gỡ được; nhưng không chặn cài.
func TestTTSCheck_BlocksUninstallNotSetup(t *testing.T) {
	a := &App{lib: library.New(t.TempDir())}
	release, err := a.beginTTSCheck()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.beginUninstall(); err == nil {
		t.Error("đang kiểm tra bộ đọc mà gỡ được")
	}
	a.mu.Lock()
	if err := a.setupBusyLocked(); err != nil {
		t.Errorf("kiểm tra bộ đọc không được chặn cài: %v", err)
	}
	a.mu.Unlock()
	release()
	done, err := a.beginUninstall()
	if err != nil {
		t.Fatalf("hết kiểm tra phải gỡ được: %v", err)
	}
	done()
}

// Cài bộ đọc không bắt đầu khi đang render hoặc bộ đọc đang được dùng (kiểm dưới khoá).
func TestStartSetup_RefusesWhileTTSInUse(t *testing.T) {
	t.Setenv("SANO_DATA_DIR", t.TempDir())
	a := &App{lib: library.New(t.TempDir())}
	release, err := a.beginTTSUse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.StartSetup(); err == nil || !strings.Contains(err.Error(), "đang được dùng") {
		t.Errorf("đang nghe thử mà bắt đầu cài được: %v", err)
	}
	release()
	a.mu.Lock()
	a.job = &renderJob{status: RenderStatus{Running: true}}
	if err := a.setupBusyLocked(); err == nil {
		t.Error("đang render mà cài được")
	}
	if err := a.renderBusyLocked(); err == nil {
		t.Error("đang render mà render cuốn khác được")
	}
	a.job = nil
	a.mu.Unlock()
}

// Chạy song song kiểm tra / dùng bộ đọc / gỡ: không có data race, gỡ không bao
// giờ chen vào giữa một lượt đang chạy.
func TestTTSBusy_ConcurrentNoRace(t *testing.T) {
	a := &App{lib: library.New(t.TempDir())}
	var wg sync.WaitGroup
	var overlap atomic.Bool
	for range 50 {
		wg.Add(3)
		go func() {
			defer wg.Done()
			if rel, err := a.beginTTSCheck(); err == nil {
				if a.uninstallingNow() {
					overlap.Store(true)
				}
				rel()
			}
		}()
		go func() {
			defer wg.Done()
			if rel, err := a.beginTTSUse(); err == nil {
				if a.uninstallingNow() {
					overlap.Store(true)
				}
				rel()
			}
		}()
		go func() {
			defer wg.Done()
			if done, err := a.beginUninstall(); err == nil {
				done()
			}
		}()
	}
	wg.Wait()
	if overlap.Load() {
		t.Error("gỡ bộ đọc chen vào giữa lượt kiểm tra / dùng bộ đọc")
	}
}
