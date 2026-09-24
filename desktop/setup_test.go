package main

import (
	"strings"
	"testing"

	"sano/desktop/internal/library"
)

// Đang gỡ bộ đọc thì nghe thử, đọc câu mẫu, hỏi giọng, tạo sách, cài bộ đọc
// đều bị từ chối (trước đây chỉ kiểm trước hộp thoại xác nhận).
func TestTTSBusy_UninstallBlocksUsers(t *testing.T) {
	a := &App{lib: library.New(t.TempDir())}
	done, err := a.beginUninstall()
	if err != nil {
		t.Fatalf("beginUninstall: %v", err)
	}
	defer done()

	if _, err := a.beginTTSUse(); err == nil || !strings.Contains(err.Error(), "gỡ bộ đọc") {
		t.Errorf("beginTTSUse khi đang gỡ phải lỗi, got %v", err)
	}
	if _, err := a.PreviewClips(BookSettings{}, nil); err == nil || !strings.Contains(err.Error(), "gỡ bộ đọc") {
		t.Errorf("PreviewClips khi đang gỡ phải lỗi, got %v", err)
	}
	if _, err := a.SpeakSample("x", "y"); err == nil || !strings.Contains(err.Error(), "gỡ bộ đọc") {
		t.Errorf("SpeakSample khi đang gỡ phải lỗi, got %v", err)
	}
	if _, err := a.Voices(); err == nil || !strings.Contains(err.Error(), "gỡ bộ đọc") {
		t.Errorf("Voices khi đang gỡ phải lỗi, got %v", err)
	}
	if _, err := a.StartRender(BookSettings{RightsConfirmedAt: "2026-01-01T00:00:00Z"}); err == nil || !strings.Contains(err.Error(), "gỡ bộ đọc") {
		t.Errorf("StartRender khi đang gỡ phải lỗi, got %v", err)
	}
	if _, err := a.StartSetup(); err == nil || !strings.Contains(err.Error(), "gỡ bộ đọc") {
		t.Errorf("StartSetup khi đang gỡ phải lỗi, got %v", err)
	}
	if _, err := a.beginUninstall(); err == nil {
		t.Error("gỡ hai lượt cùng lúc phải lỗi")
	}
}

// Ngược lại: bộ đọc đang được dùng (nghe thử / render / cài) thì không gỡ được,
// kiểm dưới khoá ngay trước khi xoá.
func TestTTSBusy_UsersBlockUninstall(t *testing.T) {
	a := &App{lib: library.New(t.TempDir())}
	release, err := a.beginTTSUse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.beginUninstall(); err == nil {
		t.Error("đang nghe thử mà gỡ được bộ đọc")
	}
	release()
	done, err := a.beginUninstall()
	if err != nil {
		t.Fatalf("hết lượt dùng phải gỡ được: %v", err)
	}
	done()

	a.job = &renderJob{status: RenderStatus{Running: true}}
	if _, err := a.beginUninstall(); err == nil {
		t.Error("đang render mà gỡ được bộ đọc")
	}
	a.job = nil
	a.m4b = &m4bJob{status: M4BStatus{Running: true}}
	if _, err := a.beginUninstall(); err == nil {
		t.Error("đang xuất M4B mà gỡ được bộ đọc")
	}
}
