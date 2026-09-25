package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func releaseServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestCheckRelease_CoBanMoi(t *testing.T) {
	srv := releaseServer(t, 200, `{"tag_name":"v0.2.0","published_at":"2026-10-20T08:00:00Z","body":"## Có gì mới\n- Nghe thử **mọi** đoạn\n* Xuất M4B nhanh hơn\nđoạn văn không phải gạch đầu dòng\n- "}`)
	info, _, err := checkRelease(context.Background(), srv.Client(), srv.URL, "0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Available || info.Version != "0.2.0" {
		t.Errorf("muốn có bản 0.2.0, được %+v", info)
	}
	if info.URL != releasesPage+"/tag/v0.2.0" {
		t.Errorf("URL = %q", info.URL)
	}
	if len(info.Notes) != 2 || info.Notes[0] != "Nghe thử mọi đoạn" || info.Notes[1] != "Xuất M4B nhanh hơn" {
		t.Errorf("Notes = %q", info.Notes)
	}
}

func TestCheckRelease_KhongCoBanMoi(t *testing.T) {
	for _, cur := range []string{"0.2.0", "0.3.1", "v0.2.0", "dev"} {
		srv := releaseServer(t, 200, `{"tag_name":"v0.2.0","body":""}`)
		info, _, err := checkRelease(context.Background(), srv.Client(), srv.URL, cur)
		if err != nil {
			t.Fatal(err)
		}
		if info.Available {
			t.Errorf("đang dùng %s: không được báo có bản mới", cur)
		}
	}
}

func TestCheckRelease_Loi(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
	}{
		"repo chưa công khai": {404, `{"message":"Not Found"}`},
		"tag lạ":              {200, `{"tag_name":"latest"}`},
		"bản thử nghiệm":      {200, `{"tag_name":"v0.2.0","prerelease":true}`},
		"JSON hỏng":           {200, `{`},
	}
	for name, c := range cases {
		srv := releaseServer(t, c.status, c.body)
		if _, _, err := checkRelease(context.Background(), srv.Client(), srv.URL, "0.1.0"); err == nil {
			t.Errorf("%s: phải trả lỗi", name)
		}
	}
}

func TestReleaseNotes_GioiHan(t *testing.T) {
	body := strings.Repeat("- dòng\n", 20) + "- " + strings.Repeat("x", 500)
	notes := releaseNotes(body)
	if len(notes) != maxReleaseNotes {
		t.Errorf("số dòng = %d, muốn %d", len(notes), maxReleaseNotes)
	}
	long := releaseNotes("- " + strings.Repeat("ạ", 500))
	if n := len([]rune(long[0])); n != maxReleaseNoteLen+1 {
		t.Errorf("dòng dài không bị cắt: %d ký tự", n)
	}
}

func TestParseSemver(t *testing.T) {
	if v, ok := parseSemver("v1.2.3"); !ok || v != [3]int{1, 2, 3} {
		t.Errorf("v1.2.3 → %v %v", v, ok)
	}
	for _, s := range []string{"dev", "1.2", "v1.2.3-rc1", ""} {
		if _, ok := parseSemver(s); ok {
			t.Errorf("%q không phải semver", s)
		}
	}
	if !newer([3]int{0, 10, 0}, [3]int{0, 9, 9}) || newer([3]int{1, 0, 0}, [3]int{1, 0, 0}) {
		t.Error("so phiên bản sai")
	}
}
