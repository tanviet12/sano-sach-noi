package setup

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ulikunitz/xz"

	ttsscripts "sano/scripts/tts"

	"sano/desktop/internal/tts"
)

func sha(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func TestDownload_KiemSHA256(t *testing.T) {
	body := bytes.Repeat([]byte("sano"), 100_000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.Write(body) }))
	defer srv.Close()
	dir := t.TempDir()

	dst := filepath.Join(dir, "a.bin")
	var last int64
	got, err := download(context.Background(), srv.Client(), srv.URL, dst, sha(body), 1<<20, func(d, _ int64) { last = d })
	if err != nil || got != sha(body) {
		t.Fatalf("tải đúng mã băm phải qua: %v", err)
	}
	if last != int64(len(body)) {
		t.Errorf("tiến độ cuối = %d, muốn %d", last, len(body))
	}

	bad := filepath.Join(dir, "b.bin")
	_, err = download(context.Background(), srv.Client(), srv.URL, bad, strings.Repeat("0", 64), 1<<20, nil)
	if !errors.Is(err, ErrChecksum) {
		t.Fatalf("lệch mã băm phải báo ErrChecksum, được %v", err)
	}
	for _, p := range []string{bad, bad + ".part"} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("lệch mã băm không được để lại %s", filepath.Base(p))
		}
	}
}

func TestDownload_HuyGiuaChung(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100000000")
		w.Write([]byte("mot it"))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer srv.Close()
	dst := filepath.Join(t.TempDir(), "c.bin")
	ctx, cancel := context.WithCancel(context.Background())
	_, err := download(ctx, srv.Client(), srv.URL, dst, "", 1<<30, func(d, _ int64) {
		if d > 0 {
			cancel()
		}
	})
	if err == nil {
		t.Fatal("huỷ phải trả lỗi")
	}
	if _, err := os.Stat(dst + ".part"); !os.IsNotExist(err) {
		t.Error("huỷ phải xoá file .part")
	}
}

type tarFile struct {
	name, body, link string
	dir              bool
}

func makeTarGz(t *testing.T, files []tarFile) string {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, f := range files {
		h := &tar.Header{Name: f.name, Mode: 0o644}
		switch {
		case f.dir:
			h.Typeflag = tar.TypeDir
		case f.link != "":
			h.Typeflag, h.Linkname = tar.TypeSymlink, f.link
		default:
			h.Typeflag, h.Size = tar.TypeReg, int64(len(f.body))
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if h.Typeflag == tar.TypeReg {
			tw.Write([]byte(f.body))
		}
	}
	tw.Close()
	gz.Close()
	p := filepath.Join(t.TempDir(), "x.tar.gz")
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestExtractTarGz_TreeHashKhongPhuThuocThuTu(t *testing.T) {
	a := makeTarGz(t, []tarFile{{name: "repo-abc/", dir: true}, {name: "repo-abc/a.txt", body: "A"}, {name: "repo-abc/src/b.py", body: "B"}, {name: "repo-abc/l", link: "a.txt"}})
	b := makeTarGz(t, []tarFile{{name: "repo-abc/l", link: "a.txt"}, {name: "repo-abc/src/b.py", body: "B"}, {name: "repo-abc/a.txt", body: "A"}})
	d1, d2 := t.TempDir(), t.TempDir()
	h1, err := extractTarGz(context.Background(), a, d1)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := extractTarGz(context.Background(), b, d2)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Errorf("tree hash phải giống nhau khi chỉ đổi thứ tự: %s vs %s", h1, h2)
	}
	if got, _ := os.ReadFile(filepath.Join(d1, "src", "b.py")); string(got) != "B" {
		t.Errorf("giải nén sai: %q", got)
	}
	if _, err := os.Lstat(filepath.Join(d1, "l")); !os.IsNotExist(err) {
		t.Error("không được tạo liên kết ra đĩa")
	}
	c := makeTarGz(t, []tarFile{{name: "repo-abc/a.txt", body: "khác"}, {name: "repo-abc/src/b.py", body: "B"}, {name: "repo-abc/l", link: "a.txt"}})
	h3, _ := extractTarGz(context.Background(), c, t.TempDir())
	if h3 == h1 {
		t.Error("đổi nội dung phải đổi tree hash")
	}
}

func TestExtractTarGz_ChanDuongDanNguyHiem(t *testing.T) {
	for _, name := range []string{"repo/../../thoat.txt", "repo//etc/passwd", `repo/C:\x`} {
		p := makeTarGz(t, []tarFile{{name: name, body: "x"}})
		dst := t.TempDir()
		_, err := extractTarGz(context.Background(), p, dst)
		if name == "repo//etc/passwd" {
			// "//etc/passwd" sau khi bỏ thư mục gốc thành "/etc/passwd" → phải chặn
			if !errors.Is(err, errUnsafePath) {
				t.Errorf("%q: muốn errUnsafePath, được %v", name, err)
			}
			continue
		}
		if !errors.Is(err, errUnsafePath) {
			t.Errorf("%q: muốn errUnsafePath, được %v", name, err)
		}
	}
}

func TestExtractOne(t *testing.T) {
	tgz := makeTarGz(t, []tarFile{{name: "uv-x/uvx", body: "no"}, {name: "uv-x/uv", body: "UV"}})
	dst := filepath.Join(t.TempDir(), "uv")
	if err := extractOneFromTarGz(tgz, "uv", dst); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(dst); string(b) != "UV" {
		t.Errorf("lấy sai file: %q", b)
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("ffmpeg")
	w.Write([]byte("FF"))
	zw.Close()
	zp := filepath.Join(t.TempDir(), "f.zip")
	os.WriteFile(zp, buf.Bytes(), 0o644)
	dz := filepath.Join(t.TempDir(), "ffmpeg")
	if err := extractOneFromZip(zp, "ffmpeg", dz); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(dz); string(b) != "FF" {
		t.Errorf("zip lấy sai: %q", b)
	}
	if err := extractOneFromZip(zp, "khong-co", dz); err == nil {
		t.Error("thiếu file trong zip phải báo lỗi")
	}
}

func TestArtifacts_TheoMay(t *testing.T) {
	pins, err := ttsscripts.Pins()
	if err != nil {
		t.Fatal(err)
	}
	for plat := range uvTargets {
		goos, goarch, _ := strings.Cut(plat, "/")
		a, err := uvArtifact(pins, goos, goarch)
		if err != nil {
			t.Errorf("uv %s: %v", plat, err)
			continue
		}
		if !strings.HasPrefix(a.URL, "https://github.com/astral-sh/uv/releases/download/"+pins["UV_VERSION"]+"/") || len(a.SHA256) != 64 {
			t.Errorf("uv %s: %+v", plat, a)
		}
		if (goos == "windows") != (a.Kind == "zip") {
			t.Errorf("uv %s: sai định dạng %s", plat, a.Kind)
		}
	}
	for _, plat := range []string{"darwin/arm64", "darwin/amd64", "windows/amd64", "windows/arm64", "linux/amd64", "linux/arm64"} {
		goos, goarch, _ := strings.Cut(plat, "/")
		a, err := ffmpegArtifact(pins, goos, goarch)
		if err != nil || len(a.SHA256) != 64 {
			t.Errorf("ffmpeg %s: %+v %v", plat, a, err)
		}
		want := map[string]string{"darwin": "zip/ffmpeg", "windows": "zip/ffmpeg.exe", "linux": "tar.xz/ffmpeg"}[goos]
		if got := a.Kind + "/" + a.Member; got != want {
			t.Errorf("ffmpeg %s: định dạng %s, cần %s", plat, got, want)
		}
	}
	if _, err := ffmpegArtifact(pins, "freebsd", "amd64"); err == nil {
		t.Error("máy không có bản ghim phải báo lỗi")
	}
	if _, err := uvArtifact(pins, "plan9", "386"); err == nil {
		t.Error("máy lạ phải báo lỗi")
	}
	url, commit, tree, err := vieneuTarball(pins)
	if err != nil || !strings.HasSuffix(url, "/archive/"+commit+".tar.gz") || len(tree) != 64 {
		t.Errorf("tarball VieNeu: %q %q %v", url, tree, err)
	}
}

func newRoot(t *testing.T) (tts.Layout, string) {
	t.Helper()
	home := t.TempDir()
	data := filepath.Join(home, "Library", "Application Support", "Sano")
	l := tts.NewLayout(data, "darwin")
	if err := l.EnsureRoot(); err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(filepath.Join(l.Root, "hf"), 0o755)
	os.WriteFile(filepath.Join(l.Root, "hf", "m.bin"), bytes.Repeat([]byte{1}, 4096), 0o644)
	return l, home
}

func TestUninstall_XoaDungThuMuc(t *testing.T) {
	l, home := newRoot(t)
	books := filepath.Join(home, "Sano", "Sach", "cuon-1")
	os.MkdirAll(books, 0o755)
	os.WriteFile(filepath.Join(books, "a.mp3"), []byte("x"), 0o644)

	freed, err := Uninstall(l, home)
	if err != nil {
		t.Fatal(err)
	}
	if freed < 4096 {
		t.Errorf("dung lượng giải phóng = %d", freed)
	}
	if _, err := os.Stat(l.Root); !os.IsNotExist(err) {
		t.Error("thư mục bộ đọc vẫn còn")
	}
	if _, err := os.Stat(filepath.Join(books, "a.mp3")); err != nil {
		t.Error("gỡ bộ đọc đã đụng vào sách!")
	}
}

func TestCheckRemovable_TuChoi(t *testing.T) {
	l, home := newRoot(t)

	// Thiếu file đánh dấu
	os.Remove(l.Marker())
	if err := CheckRemovable(l.Root, home); !errors.Is(err, ErrNotRemovable) {
		t.Errorf("thiếu đánh dấu phải từ chối: %v", err)
	}
	os.WriteFile(l.Marker(), []byte(tts.MarkerContent), 0o644)
	if err := CheckRemovable(l.Root, home); err != nil {
		t.Fatalf("thư mục hợp lệ phải qua: %v", err)
	}

	cases := map[string]string{
		"tên không phải tts":  filepath.Join(home, "khac"),
		"đường dẫn tương đối": "tts",
		"trong ~/Sano":        filepath.Join(home, "Sano", "tts"),
		"trong VieNeu":        filepath.Join(home, "VieNeu-TTS-v3", "tts"),
		"chứa HOME":           filepath.Join(filepath.Dir(home), "tts"),
	}
	for name, root := range cases {
		os.MkdirAll(root, 0o755)
		os.WriteFile(filepath.Join(root, tts.MarkerFile), []byte(tts.MarkerContent), 0o644)
		if name == "chứa HOME" {
			// HOME nằm trong root khi root = <cha của HOME>/tts? Không — dựng HOME giả nằm trong root.
			fakeHome := filepath.Join(root, "nguoi-dung")
			os.MkdirAll(fakeHome, 0o755)
			if err := CheckRemovable(root, fakeHome); !errors.Is(err, ErrNotRemovable) {
				t.Errorf("%s: phải từ chối, được %v", name, err)
			}
			continue
		}
		if err := CheckRemovable(root, home); !errors.Is(err, ErrNotRemovable) {
			t.Errorf("%s: phải từ chối, được %v", name, err)
		}
	}

	// Liên kết trỏ tới thư mục khác
	link := filepath.Join(t.TempDir(), "tts")
	if err := os.Symlink(l.Root, link); err == nil {
		if err := CheckRemovable(link, home); !errors.Is(err, ErrNotRemovable) {
			t.Errorf("liên kết phải từ chối: %v", err)
		}
	}
}

func TestInstaller_BoQuaBuocDaXong(t *testing.T) {
	// ffmpeg có sẵn → bỏ qua; kiểm trạng thái + sự kiện.
	l, _ := newRoot(t)
	var events int
	in := New(Config{Layout: l, FindFFmpeg: func() string { return "/usr/bin/ffmpeg" }, OnProgress: func(Status) { events++ }})
	if err := in.stepFFmpeg(context.Background()); err != nil {
		t.Fatal(err)
	}
	st := in.Status()
	for _, s := range st.Steps {
		if s.Key == StepFFmpeg && (s.State != StateSkipped || s.Pct != 100) {
			t.Errorf("ffmpeg phải bỏ qua: %+v", s)
		}
	}
	if events == 0 {
		t.Error("phải đẩy sự kiện tiến độ")
	}
}

func TestHintFor(t *testing.T) {
	if !strings.Contains(hintFor(ErrChecksum), "Thử lại") {
		t.Error("gợi ý cho lỗi mã băm")
	}
	if !strings.Contains(hintFor(fmt.Errorf("%w: dial tcp", ErrNetwork)), "mạng") {
		t.Error("gợi ý cho lỗi mạng")
	}
	if !strings.Contains(hintFor(errors.New("uv: error: Failed to fetch: https://pypi.org")), "mạng") {
		t.Error("gợi ý cho lỗi mạng của uv")
	}
	if strings.Contains(hintFor(errors.New("ffmpeg tải về không có bộ mã MP3")), "mạng") {
		t.Error("lỗi không liên quan mạng không được gợi ý kiểm tra mạng")
	}
}

func TestHuman(t *testing.T) {
	if got := humanMB(ModelBytes); got != "580 MB" {
		t.Errorf("humanMB = %q", got)
	}
	if got := humanGB(3 << 29); got != "1,5 GB" {
		t.Errorf("humanGB = %q", got)
	}
}

func TestDownload_QuaGioiHan(t *testing.T) {
	body := bytes.Repeat([]byte("x"), 4096)
	for _, khaiBao := range []bool{true, false} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !khaiBao {
				w.Header().Set("Transfer-Encoding", "chunked")
				w.(http.Flusher).Flush()
			}
			w.Write(body)
		}))
		dst := filepath.Join(t.TempDir(), "lon.bin")
		_, err := download(context.Background(), srv.Client(), srv.URL, dst, "", 1024, nil)
		srv.Close()
		if !errors.Is(err, ErrTooLarge) {
			t.Errorf("khai báo Content-Length=%v: muốn ErrTooLarge, được %v", khaiBao, err)
		}
		for _, p := range []string{dst, dst + ".part"} {
			if _, err := os.Stat(p); !os.IsNotExist(err) {
				t.Errorf("không được để lại %s", filepath.Base(p))
			}
		}
	}
}

func TestWriteFile_QuaGioiHan(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "f")
	if _, _, err := writeFile(dst, strings.NewReader("12345"), 0o644, 4); !errors.Is(err, errTooBig) {
		t.Errorf("muốn errTooBig, được %v", err)
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Error("vượt giới hạn phải xoá file dở")
	}
	if _, n, err := writeFile(dst, strings.NewReader("1234"), 0o644, 4); err != nil || n != 4 {
		t.Errorf("đúng giới hạn phải qua: n=%d err=%v", n, err)
	}
}

func TestExtractTarGz_QuaSoMuc(t *testing.T) {
	files := make([]tarFile, 0, maxTreeEntries+1)
	for i := 0; i <= maxTreeEntries; i++ {
		files = append(files, tarFile{name: fmt.Sprintf("repo/d%d/", i), dir: true})
	}
	p := makeTarGz(t, files)
	if _, err := extractTarGz(context.Background(), p, t.TempDir()); !errors.Is(err, errTooBig) {
		t.Errorf("muốn errTooBig, được %v", err)
	}
}

func TestExtractOneFromTarXz(t *testing.T) {
	var buf bytes.Buffer
	xw, err := xz.NewWriter(&buf)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(xw)
	for _, f := range []struct{ name, body string }{{"ff/bin/ffprobe", "no"}, {"ff/bin/ffmpeg", "FF"}} {
		tw.WriteHeader(&tar.Header{Name: f.name, Mode: 0o755, Typeflag: tar.TypeReg, Size: int64(len(f.body))})
		tw.Write([]byte(f.body))
	}
	tw.Close()
	xw.Close()
	p := filepath.Join(t.TempDir(), "f.tar.xz")
	os.WriteFile(p, buf.Bytes(), 0o644)
	dst := filepath.Join(t.TempDir(), "ffmpeg")
	if err := extractOneFromTarXz(p, "ffmpeg", dst); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(dst); string(b) != "FF" {
		t.Errorf("lấy sai file: %q", b)
	}
	if err := extractOneFromTarXz(p, "khong-co", dst); err == nil {
		t.Error("thiếu file phải báo lỗi")
	}
}
