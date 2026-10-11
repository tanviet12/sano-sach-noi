package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"sano/desktop/internal/library"
)

func playbackTestApp(t *testing.T, meta string) (*App, string) {
	t.Helper()
	lib := library.New(t.TempDir())
	dir := filepath.Join(lib.BooksRoot(), "sach-thu")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	return &App{lib: lib}, dir
}

func TestBook_RejectsInvalidPlaybackFiles(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "ngoai-thu-vien.mp3")
	if err := os.WriteFile(outside, []byte("ID3"), 0o644); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, file, reason       string
		omit, directory, symlink bool
	}{
		{name: "missing field", omit: true, reason: "thiếu hoặc rỗng trường file"},
		{name: "empty field", reason: "thiếu hoặc rỗng trường file"},
		{name: "blank field", file: " \t", reason: "thiếu hoặc rỗng trường file"},
		{name: "missing MP3", file: "khong-co.mp3", reason: "không tìm thấy file âm thanh"},
		{name: "directory", file: "thu-muc.mp3", directory: true, reason: "không phải file thường"},
		{name: "parent traversal", file: "../ngoai.mp3", reason: "tên file không hợp lệ"},
		{name: "normalized traversal", file: "audio/../sample.mp3", reason: "tên file không hợp lệ"},
		{name: "subdirectory", file: "audio/sample.mp3", reason: "tên file không hợp lệ"},
		{name: "backslash", file: `audio\sample.mp3`, reason: "tên file không hợp lệ"},
		{name: "Windows traversal", file: `..\ngoai.mp3`, reason: "tên file không hợp lệ"},
		{name: "absolute path", file: outside, reason: "tên file không hợp lệ"},
		{name: "Windows absolute path", file: `C:\Users\test\sample.mp3`, reason: "tên file không hợp lệ"},
		{name: "dot", file: ".", reason: "tên file không hợp lệ"},
		{name: "parent", file: "..", reason: "tên file không hợp lệ"},
		{name: "unsupported extension", file: "sample.wav", reason: "file âm thanh phải có đuôi .mp3"},
		{name: "external symlink", file: "link.mp3", symlink: true, reason: "đường dẫn file không hợp lệ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			section := map[string]string{"title": "Mở đầu"}
			if !tt.omit {
				section["file"] = tt.file
			}
			data, err := json.Marshal(section)
			if err != nil {
				t.Fatal(err)
			}
			meta := `{"title":"Sách thử","chapters":[{"title":"Chương 1","sections":[` + string(data) + `]}]}`
			a, dir := playbackTestApp(t, meta)
			if tt.directory {
				if err := os.Mkdir(filepath.Join(dir, tt.file), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if tt.symlink {
				if err := os.Symlink(outside, filepath.Join(dir, tt.file)); err != nil {
					t.Fatal(err)
				}
			}

			d, err := a.Book("sach-thu")
			if err == nil || d != nil {
				t.Errorf("phải từ chối phát, got detail=%+v err=%v", d, err)
			} else {
				for _, want := range []string{"Chương 1", "Mở đầu", tt.reason} {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("lỗi %q phải chứa %q", err, want)
					}
				}
				if tt.file != "" && (tt.directory || tt.symlink || tt.name == "missing MP3" || tt.name == "unsupported extension") {
					if !strings.Contains(err.Error(), tt.file) {
						t.Errorf("lỗi phải chỉ rõ tên file tương đối: %v", err)
					}
				}
				home, _ := os.UserHomeDir()
				for _, private := range []string{a.lib.Root(), outside, home} {
					if private != "" && strings.Contains(err.Error(), private) {
						t.Errorf("lỗi làm lộ đường dẫn tuyệt đối: %v", err)
					}
				}
				if tt.reason == "tên file không hợp lệ" && strings.Contains(err.Error(), fmt.Sprintf("%q", tt.file)) {
					t.Errorf("lỗi không được lặp lại đường dẫn không hợp lệ: %v", err)
				}
			}

			// Chặn phát không được làm sách biến mất hay chặn màn sửa sách.
			if d, err := a.lib.Get("sach-thu"); err != nil || len(d.Tracks) != 1 {
				t.Fatalf("vẫn phải đọc được metadata: %+v %v", d, err)
			}
			if shelf, err := a.Library(); err != nil || len(shelf.Books) != 1 || shelf.Books[0].Slug != "sach-thu" {
				t.Fatalf("sách lỗi vẫn phải hiện trong thư viện: %+v %v", shelf, err)
			}
			if edit, err := a.EditBook("sach-thu"); err != nil || len(edit.Sections) != 1 || edit.Sections[0].Title != "Mở đầu" {
				t.Fatalf("sách lỗi vẫn phải mở được để sửa: %+v %v", edit, err)
			}
		})
	}
}

func TestBook_ValidTracksPreserveOrderAndMediaURLs(t *testing.T) {
	meta := `{"title":"Sách thử","chapters":[
	 {"title":"Chương 1","sections":[{"title":"Mở đầu","file":"z.mp3"},{"title":"Phần hai","file":"a.MP3"}]},
	 {"title":"Chương 2","sections":[{"title":"Kết","file":"kết.mp3"}]}]}`
	a, dir := playbackTestApp(t, meta)
	files := []string{"z.mp3", "a.MP3", "kết.mp3"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("ID3"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	d, err := a.Book("sach-thu")
	if err != nil {
		t.Fatal(err)
	}
	if d.Title != "Sách thử" || d.Chapters != 2 || d.Sections != 3 || len(d.Tracks) != 3 || d.Dir != dir {
		t.Fatalf("thông tin sách sai: %+v", d)
	}
	original, err := a.lib.Get("sach-thu")
	if err != nil {
		t.Fatal(err)
	}
	for i, track := range d.Tracks {
		if !reflect.DeepEqual(track.Track, original.Tracks[i]) {
			t.Errorf("tiểu mục %d bị đổi: %+v", i, track)
		}
		if track.File != "Sach/sach-thu/"+files[i] || track.ChapterStart != (i != 1) {
			t.Errorf("thứ tự / đầu chương sai: %+v", track)
		}
		info, err := os.Stat(filepath.Join(dir, files[i]))
		if err != nil {
			t.Fatal(err)
		}
		wantURL := mediaPrefix + track.File + fmt.Sprintf("?v=%x", info.ModTime().UnixNano())
		if track.URL != wantURL {
			t.Errorf("URL phải giữ nguyên: got %q, muốn %q", track.URL, wantURL)
		}
		data, err := json.Marshal(track)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			t.Fatal(err)
		}
		if len(fields) != 6 {
			t.Errorf("không được thêm trường nội bộ vào JSON track: %s", data)
		}
	}
}
