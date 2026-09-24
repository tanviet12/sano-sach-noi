package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"sano/desktop/internal/library"
	"sano/internal/bookmaker"
)

func TestMediaMiddleware(t *testing.T) {
	root := t.TempDir()
	lib := library.New(root)
	dir := filepath.Join(root, "Sach", "sach-thu")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"ch01-sec01.mp3": "ID3xxxx", "ch01-sec01.txt": "lời đọc", "metadata.json": "{}"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })
	h := mediaMiddleware(lib)(next)

	tests := []struct {
		path string
		want int
	}{
		{"/sano-media/Sach/sach-thu/ch01-sec01.mp3", http.StatusOK},
		{"/sano-media/Sach/sach-thu/ch01-sec01.txt", http.StatusNotFound}, // chỉ phục vụ mp3 + ảnh
		{"/sano-media/Sach/sach-thu/khong-co.mp3", http.StatusNotFound},
		{"/sano-media/../../etc/x.mp3", http.StatusNotFound},
		{"/index.html", http.StatusTeapot}, // không phải media → chuyển tiếp
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if rec.Code != tt.want {
			t.Errorf("%s: status %d, muốn %d", tt.path, rec.Code, tt.want)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sano-media/Sach/sach-thu/ch01-sec01.mp3", nil)
	req.Header.Set("Range", "bytes=0-2")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusPartialContent || rec.Header().Get("Content-Type") != "audio/mpeg" || rec.Body.String() != "ID3" {
		t.Errorf("Range phải trả 206 audio/mpeg \"ID3\", got %d %q %q", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
}

func TestBookSettingsOptions(t *testing.T) {
	s := BookSettings{
		Path: "/tmp/sach.docx", Title: "  Sách  ", Voice: "", IntroText: " Xin chào. ",
		DropStems: []string{"ch01-sec01"}, ReadingEdits: map[string]bookmaker.ReadingEdit{"ch02-sec01": {From: "a", To: "b"}},
	}
	o, err := s.options(toolPaths{python: "py", script: "/x/audio_gen_batch.py", ffmpeg: "ff"}, "/out")
	if err != nil {
		t.Fatal(err)
	}
	if o.Title != "Sách" || o.IntroText != "Xin chào." || !o.DropStems["ch01-sec01"] || o.OutputDir != "/out" {
		t.Errorf("options sai: %+v", o)
	}
	if o.TTS.Voice != bookmaker.DefaultVoice || !o.TTS.KeepTxt || o.TTS.ScriptDir != "/x" || o.TTS.Mode != bookmaker.TTSModeVieNeu {
		t.Errorf("TTS sai: %+v", o.TTS)
	}
	if _, err := (BookSettings{Path: "/tmp/a.pdf"}).options(toolPaths{}, "/out"); err == nil {
		t.Error("file không phải .docx phải lỗi")
	}
}
