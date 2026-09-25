package setup

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"

	"sano/desktop/internal/tts"
	ttsscripts "sano/scripts/tts"
)

// Bản Sano đổi pyproject/uv.lock (vd vá bảo mật) → máy đã cài bị coi là cần sync lại.
func TestNeedsResync(t *testing.T) {
	pins, err := ttsscripts.Pins()
	if err != nil {
		t.Fatal(err)
	}
	l := tts.NewLayout(t.TempDir(), runtime.GOOS)
	if NeedsResync(l, pins, ttsscripts.Files) {
		t.Error("chưa cài (không có python) thì không phải cập nhật")
	}
	if err := os.MkdirAll(filepath.Dir(l.Python()), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(l.Python(), []byte("x"), 0o755)
	if !NeedsResync(l, pins, ttsscripts.Files) {
		t.Error("có python mà không có dấu đã sync → phải sync lại")
	}
	// Dấu của bản cũ (trước khi có mã băm pyproject/uv.lock).
	old := pins["VIENEU_COMMIT"] + "\n" + pins["PYTHON_VERSION"] + "\n" + pins["UV_VERSION"] + "\n"
	os.WriteFile(filepath.Join(l.Venv(), syncedMarker), []byte(old), 0o644)
	if !NeedsResync(l, pins, ttsscripts.Files) {
		t.Error("dấu của bản cũ → phải sync lại")
	}
	_, sum, err := vieneuProject(ttsscripts.Files)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(l.Venv(), syncedMarker), []byte(syncedWant(pins, sum)), 0o644)
	if NeedsResync(l, pins, ttsscripts.Files) {
		t.Error("dấu khớp bản này → không cần sync lại")
	}
	if NeedsResync(l, pins, nil) {
		t.Error("không có file nhúng thì không báo cần cập nhật")
	}
}

func TestVieNeuProject(t *testing.T) {
	files, sum, err := vieneuProject(ttsscripts.Files)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || len(files["uv.lock"]) == 0 || len(files["pyproject.toml"]) == 0 || len(sum) != 64 {
		t.Errorf("files=%d sum=%q", len(files), sum)
	}
	// Đổi một byte uv.lock → mã băm khác.
	lock := append([]byte{}, files["uv.lock"]...)
	lock[len(lock)-1] ^= 1
	other := fstest.MapFS{
		"vieneu-project/pyproject.toml": {Data: files["pyproject.toml"]},
		"vieneu-project/uv.lock":        {Data: lock},
	}
	if _, sum2, err := vieneuProject(other); err != nil || sum2 == sum {
		t.Errorf("uv.lock khác mà mã băm giống (%v)", err)
	}
	if _, _, err := vieneuProject(fstest.MapFS{"vieneu-project/pyproject.toml": {Data: []byte("x")}}); err == nil {
		t.Error("thiếu uv.lock phải lỗi")
	}
}
