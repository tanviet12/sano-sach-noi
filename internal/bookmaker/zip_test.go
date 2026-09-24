package bookmaker

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocxSlugify_VietnameseTitle_KebabCase(t *testing.T) {
	cases := map[string]string{
		"Cẩm nang thử nghiệm":     "cam-nang-thu-nghiem",
		"  Đường tới thành công ": "duong-toi-thanh-cong",
		"A/B & C":                 "a-b-c",
		"":                        "audiobook",
	}
	for in, want := range cases {
		if got := docxSlugify(in); got != want {
			t.Errorf("docxSlugify(%q) = %q, muốn %q", in, got, want)
		}
	}
}

func TestCoverInZip_Extension_Normalized(t *testing.T) {
	cases := map[string]string{
		"cover.jpg":   "cover.jpg",
		"anh-bia.PNG": "cover.png",
		"x.gif":       "cover.png", // đuôi lạ → png
		"":            "",
	}
	for in, want := range cases {
		if got := coverInZip(in); got != want {
			t.Errorf("coverInZip(%q) = %q, muốn %q", in, got, want)
		}
	}
}

func TestValidateZipContent_Rules(t *testing.T) {
	ok := zipManifest{Title: "T", Slug: "t-book", Version: zipFormatVersion}
	chs := zipChapters{Chapters: []zipChapter{{Order: 1, Title: "C", Sections: []zipSection{
		{Order: 1, Title: "S", AudioFilename: "audio/ch01/sec01.mp3"},
	}}}}
	if err := validateZipContent(ok, chs); err != nil {
		t.Fatalf("hợp lệ nhưng báo lỗi: %v", err)
	}

	tests := []struct {
		name string
		m    zipManifest
		c    zipChapters
	}{
		{"thiếu title", zipManifest{Slug: "t", Version: 1}, chs},
		{"slug sai", zipManifest{Title: "T", Slug: "Bad Slug", Version: 1}, chs},
		{"version sai", zipManifest{Title: "T", Slug: "t", Version: 99}, chs},
		{"chapters rỗng", ok, zipChapters{}},
		{"thiếu audio_filename", ok, zipChapters{Chapters: []zipChapter{{Sections: []zipSection{{Title: "S"}}}}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateZipContent(tc.m, tc.c); err == nil {
				t.Errorf("mong đợi lỗi cho case %q nhưng nil", tc.name)
			}
		})
	}
}

func TestRepackBookZip_FromRenderedDir(t *testing.T) {
	dir := t.TempDir()
	meta := outMeta{
		Title:    "Kỹ năng giao tiếp Phần 1",
		Author:   "Nguyễn Văn A",
		Category: "Kỹ năng",
		Chapters: []outChapter{{
			Title: "Chương một",
			Sections: []outSection{{
				Title:         "Mở đầu",
				File:          "ch01-sec01.mp3",
				OriginalText:  "Mô hình 4P.",     // lời gốc giữ nguyên
				ReadingScript: "Mô hình bốn Pê.", // đã sửa khớp audio mới
			}},
		}},
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatalf("marshal meta: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), data, 0o644); err != nil {
		t.Fatalf("ghi metadata.json: %v", err)
	}
	// mp3 giả: chỉ cần tồn tại để packageBookZip nhúng; duration đọc lỗi → 0, không crash.
	if err := os.WriteFile(filepath.Join(dir, "ch01-sec01.mp3"), []byte("fake-mp3"), 0o644); err != nil {
		t.Fatalf("ghi mp3 giả: %v", err)
	}

	zipPath := filepath.Join(dir, "out.zip")
	slug, err := RepackZip(dir, zipPath, "Thanh Bình", nil)
	if err != nil {
		t.Fatalf("RepackZip: %v", err)
	}
	if want := "ky-nang-giao-tiep-phan-1"; slug != want {
		t.Errorf("slug = %q, muốn %q", slug, want)
	}

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("mở zip: %v", err)
	}
	defer func() { _ = zr.Close() }()

	contents := map[string]string{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("mở entry %q: %v", f.Name, err)
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("đọc entry %q: %v", f.Name, err)
		}
		contents[f.Name] = string(b)
	}

	for _, want := range []string{"manifest.json", "chapters.json", "audio/ch01/sec01.mp3"} {
		if _, ok := contents[want]; !ok {
			t.Errorf("zip thiếu entry %q", want)
		}
	}
	if !strings.Contains(contents["manifest.json"], `"voice_id": "Thanh Bình"`) {
		t.Errorf("manifest.json thiếu voice_id 'Thanh Bình': %s", contents["manifest.json"])
	}
	if !strings.Contains(contents["manifest.json"], `"category": "Kỹ năng"`) || !strings.Contains(contents["manifest.json"], `"category_slug": "ky-nang"`) {
		t.Errorf("manifest.json thiếu category/category_slug: %s", contents["manifest.json"])
	}
	if !strings.Contains(contents["chapters.json"], "bốn Pê") {
		t.Errorf("chapters.json thiếu reading_script đã sửa 'bốn Pê'")
	}
	if !strings.Contains(contents["chapters.json"], "Mô hình 4P.") {
		t.Errorf("chapters.json phải giữ original_text gốc 'Mô hình 4P.'")
	}
}

func TestCategorySlug(t *testing.T) {
	tests := map[string]string{
		"Kỹ năng":      "ky-nang",
		"  Sức khoẻ  ": "suc-khoe",
		"quan-tri":     "quan-tri", // CLI cũ truyền slug → giữ nguyên
		"":             "",
		"   ":          "",
		"!!!":          "",
		"Đầu tư 101":   "dau-tu-101",
	}
	for in, want := range tests {
		if got := CategorySlug(in); got != want {
			t.Errorf("CategorySlug(%q) = %q, muốn %q", in, got, want)
		}
	}
}
