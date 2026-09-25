package main

import (
	"archive/zip"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

func newKey(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return pub, priv
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func sign(priv ed25519.PrivateKey, sums []byte) []byte {
	return []byte(b64(ed25519.Sign(priv, sums)) + "\n")
}

func sha(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func TestVerifySums(t *testing.T) {
	pub, priv := newKey(t)
	other, otherPriv := newKey(t)
	sums := []byte(sha([]byte("a")) + "  Sano-0.2.0-macos-universal.dmg\n")
	sig := sign(priv, sums)

	if err := verifySums(sums, sig, []ed25519.PublicKey{pub}); err != nil {
		t.Errorf("chữ ký đúng bị từ chối: %v", err)
	}
	// Đổi khoá: khoá mới đứng trước, khoá cũ vẫn nhận.
	if err := verifySums(sums, sig, []ed25519.PublicKey{other, pub}); err != nil {
		t.Errorf("khớp một trong các khoá phải đạt: %v", err)
	}
	bad := map[string]struct {
		sums, sig []byte
		keys      []ed25519.PublicKey
	}{
		"SHA256SUMS bị sửa":   {append([]byte("x"), sums...), sig, []ed25519.PublicKey{pub}},
		"ký bằng khoá khác":   {sums, sign(otherPriv, sums), []ed25519.PublicKey{pub}},
		"chữ ký không base64": {sums, []byte("không phải chữ ký"), []ed25519.PublicKey{pub}},
		"chữ ký cụt":          {sums, []byte(b64([]byte("ngắn"))), []ed25519.PublicKey{pub}},
		"chữ ký rỗng":         {sums, nil, []ed25519.PublicKey{pub}},
		"không có khoá":       {sums, sig, nil},
	}
	for name, c := range bad {
		if err := verifySums(c.sums, c.sig, c.keys); !errors.Is(err, errBadSignature) {
			t.Errorf("%s: muốn errBadSignature, được %v", name, err)
		}
	}
}

func TestParsePublicKeys(t *testing.T) {
	a, _ := newKey(t)
	b, _ := newKey(t)
	keys, err := parsePublicKeys(" " + b64(a) + " , " + b64(b) + ",")
	if err != nil || len(keys) != 2 || !keys[1].Equal(b) {
		t.Errorf("hai khoá → %d khoá, lỗi %v", len(keys), err)
	}
	if _, err := parsePublicKeys(""); !errors.Is(err, errNoKey) {
		t.Errorf("rỗng → %v", err)
	}
	for _, s := range []string{"khong-phai-base64", b64([]byte("32 byte? không"))} {
		if _, err := parsePublicKeys(s); err == nil {
			t.Errorf("%q phải lỗi", s)
		}
	}
}

func TestSumFor(t *testing.T) {
	h1, h2 := sha([]byte("1")), sha([]byte("2"))
	sums := []byte(h1 + "  Sano-0.2.0-linux-amd64.AppImage\n" + h2 + " *Sano-0.2.0-windows-amd64-setup.exe\n\n")
	if got, err := sumFor(sums, "Sano-0.2.0-linux-amd64.AppImage"); err != nil || got != h1 {
		t.Errorf("AppImage → %q %v", got, err)
	}
	if got, err := sumFor(sums, "Sano-0.2.0-windows-amd64-setup.exe"); err != nil || got != h2 {
		t.Errorf("dấu * (chế độ nhị phân) → %q %v", got, err)
	}
	cases := map[string][]byte{
		"không có tên":      sums,
		"hai dòng cùng tên": []byte(h1 + "  Sano-0.2.0-macos-universal.dmg\n" + h2 + "  Sano-0.2.0-macos-universal.dmg\n"),
		"mã ngắn":           []byte("abc  Sano-0.2.0-macos-universal.dmg\n"),
		"mã chữ hoa":        []byte(strings.ToUpper(h1) + "  Sano-0.2.0-macos-universal.dmg\n"),
		"mã không hex":      []byte(strings.Repeat("z", 64) + "  Sano-0.2.0-macos-universal.dmg\n"),
		"dòng thiếu tên":    []byte(h1 + "\n"),
	}
	for name, s := range cases {
		if _, err := sumFor(s, "Sano-0.2.0-macos-universal.dmg"); err == nil {
			t.Errorf("%s: phải lỗi", name)
		}
	}
}

func asset(name string, size int64) releaseAsset {
	return releaseAsset{Name: name, Size: size, URL: releasesPage + "/download/v0.2.0/" + name}
}

func testRelease() *latestRelease {
	return &latestRelease{Version: "0.2.0", Assets: []releaseAsset{
		asset("Sano-0.2.0-macos-universal.dmg", 12<<20),
		asset("Sano-0.2.0-windows-amd64-setup.exe", 9<<20),
		asset("Sano-0.2.0-windows-amd64-portable.zip", 8<<20),
		asset("Sano-0.2.0-linux-amd64.AppImage", 10<<20),
		asset(sumsName, 400),
		asset(sigName, 89),
	}}
}

func withKey(t *testing.T, pub ed25519.PublicKey) {
	t.Helper()
	old := updatePublicKeys
	updatePublicKeys = b64(pub)
	t.Cleanup(func() { updatePublicKeys = old })
}

func TestPlanUpdate_ChonDungFile(t *testing.T) {
	pub, _ := newKey(t)
	withKey(t, pub)
	for suffix, size := range map[string]int64{
		"macos-universal.dmg":        12 << 20,
		"windows-amd64-setup.exe":    9 << 20,
		"windows-amd64-portable.zip": 8 << 20,
		"linux-amd64.AppImage":       10 << 20,
	} {
		p, err := planUpdate(testRelease(), installInfo{kind: "x", suffix: suffix})
		if err != nil {
			t.Fatalf("%s: %v", suffix, err)
		}
		if p.asset.Name != "Sano-0.2.0-"+suffix || p.asset.Size != size || p.sums.Name != sumsName || p.sig.Name != sigName {
			t.Errorf("%s: chọn sai %+v", suffix, p)
		}
	}
}

func TestPlanUpdate_TuChoi(t *testing.T) {
	pub, _ := newKey(t)
	mac := installInfo{kind: "dmg", suffix: "macos-universal.dmg"}

	withKey(t, pub)
	updatePublicKeys = ""
	if _, err := planUpdate(testRelease(), mac); !errors.Is(err, errNoKey) {
		t.Errorf("chưa gắn khoá → %v", err)
	}
	updatePublicKeys = b64(pub)

	drop := func(name string) *latestRelease {
		r := testRelease()
		var keep []releaseAsset
		for _, a := range r.Assets {
			if a.Name != name {
				keep = append(keep, a)
			}
		}
		r.Assets = keep
		return r
	}
	edit := func(name string, f func(*releaseAsset)) *latestRelease {
		r := testRelease()
		for i := range r.Assets {
			if r.Assets[i].Name == name {
				f(&r.Assets[i])
			}
		}
		return r
	}
	dmg := "Sano-0.2.0-macos-universal.dmg"
	cases := map[string]struct {
		rel  *latestRelease
		inst installInfo
	}{
		"chưa kiểm tra bản mới": {nil, mac},
		"kiểu cài không hỗ trợ": {testRelease(), installInfo{err: errors.New("chạy từ .dmg")}},
		"thiếu file cho máy":    {drop(dmg), mac},
		"thiếu chữ ký":          {drop(sigName), mac},
		"thiếu SHA256SUMS":      {drop(sumsName), mac},
		"file quá lớn":          {edit(dmg, func(a *releaseAsset) { a.Size = maxInstallerBytes + 1 }), mac},
		"tải từ chỗ lạ":         {edit(dmg, func(a *releaseAsset) { a.URL = "https://example.com/" + dmg }), mac},
		"chữ ký ở chỗ lạ":       {edit(sigName, func(a *releaseAsset) { a.URL = "https://example.com/" + sigName }), mac},
		"file của bản khác":     {edit(dmg, func(a *releaseAsset) { a.URL = releasesPage + "/download/v0.1.0/" + dmg }), mac},
		"http thường":           {edit(dmg, func(a *releaseAsset) { a.URL = strings.Replace(a.URL, "https://", "http://", 1) }), mac},
	}
	for name, c := range cases {
		if _, err := planUpdate(c.rel, c.inst); err == nil {
			t.Errorf("%s: phải từ chối", name)
		}
	}
	dup := testRelease()
	dup.Assets = append(dup.Assets, asset(dmg, 1))
	if _, err := planUpdate(dup, mac); err == nil {
		t.Error("hai file cùng tên: phải từ chối")
	}
}

func TestUpdateAPI_ChiNhanMayChuCucBo(t *testing.T) {
	for v, want := range map[string]string{
		"":                             latestReleaseAPI,
		"http://127.0.0.1:8765/latest": "http://127.0.0.1:8765/latest",
		"http://localhost:1/x":         "http://localhost:1/x",
		"https://evil.example.com/x":   latestReleaseAPI,
		"http://10.0.0.1/x":            latestReleaseAPI,
		"file:///etc/passwd":           latestReleaseAPI,
	} {
		t.Setenv(envUpdateAPI, v)
		if got := updateAPI(); got != want {
			t.Errorf("%s=%q → %q, muốn %q", envUpdateAPI, v, got, want)
		}
	}
	t.Setenv(envUpdateAPI, "http://127.0.0.1:8765/latest")
	if !allowedDownload("http://127.0.0.1:8765/f/Sano.dmg", "0.2.0") || allowedDownload("http://127.0.0.1:9999/f", "0.2.0") ||
		allowedDownload(releasesPage+"/download/v0.2.0/x", "0.2.0") {
		t.Error("chạy thử cục bộ: chỉ nhận đúng máy chủ thử")
	}
}

// fakeRelease dựng máy chủ phát hành thử: file cài, SHA256SUMS, chữ ký.
type fakeRelease struct {
	srv       *httptest.Server
	files     map[string][]byte
	installer atomic.Int32 // số lần file cài bị tải
}

func newFakeRelease(t *testing.T, files map[string][]byte) *fakeRelease {
	f := &fakeRelease{files: files}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Base(r.URL.Path)
		b, ok := f.files[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(name, "Sano-") {
			f.installer.Add(1)
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		w.Write(b)
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeRelease) plan(name string) updatePlan {
	at := func(n string) releaseAsset {
		return releaseAsset{Name: n, URL: f.srv.URL + "/dl/" + n, Size: int64(len(f.files[n]))}
	}
	return updatePlan{version: "0.2.0", asset: at(name), sums: at(sumsName), sig: at(sigName)}
}

func TestRunUpdateDownload(t *testing.T) {
	pub, priv := newKey(t)
	_, otherPriv := newKey(t)
	withKey(t, pub)
	name := "Sano-0.2.0-linux-amd64.AppImage"
	body := []byte(strings.Repeat("bản mới ", 50000))
	sums := []byte(sha([]byte("khác")) + "  Sano-0.2.0-macos-universal.dmg\n" + sha(body) + "  " + name + "\n")

	t.Run("đúng hết", func(t *testing.T) {
		f := newFakeRelease(t, map[string][]byte{name: body, sumsName: sums, sigName: sign(priv, sums)})
		var last int64
		file, err := runUpdateDownload(context.Background(), f.srv.Client(), f.plan(name), t.TempDir(), func(done, total int64) { last = done })
		if err != nil {
			t.Fatal(err)
		}
		got, _ := os.ReadFile(file)
		if string(got) != string(body) || last != int64(len(body)) {
			t.Errorf("file tải về sai (%d byte, tiến độ cuối %d)", len(got), last)
		}
	})

	t.Run("chữ ký sai: không tải file cài", func(t *testing.T) {
		f := newFakeRelease(t, map[string][]byte{name: body, sumsName: sums, sigName: sign(otherPriv, sums)})
		dir := filepath.Join(t.TempDir(), "cap-nhat")
		_, err := runUpdateDownload(context.Background(), f.srv.Client(), f.plan(name), dir, nil)
		if !errors.Is(err, errBadSignature) {
			t.Fatalf("muốn errBadSignature, được %v", err)
		}
		if n := f.installer.Load(); n != 0 {
			t.Errorf("chữ ký sai mà vẫn tải file cài %d lần", n)
		}
	})

	t.Run("file cài bị tráo", func(t *testing.T) {
		evil := append([]byte{}, body...)
		evil[100] ^= 1
		f := newFakeRelease(t, map[string][]byte{name: evil, sumsName: sums, sigName: sign(priv, sums)})
		dir := t.TempDir()
		_, err := runUpdateDownload(context.Background(), f.srv.Client(), f.plan(name), dir, nil)
		if !errors.Is(err, errBadChecksum) {
			t.Fatalf("muốn errBadChecksum, được %v", err)
		}
		if left, _ := os.ReadDir(dir); len(left) != 0 {
			t.Errorf("file tải hỏng còn lại: %v", left)
		}
	})

	t.Run("SHA256SUMS đã ký không có file này", func(t *testing.T) {
		other := []byte(sha(body) + "  Sano-0.1.0-linux-amd64.AppImage\n")
		f := newFakeRelease(t, map[string][]byte{name: body, sumsName: other, sigName: sign(priv, other)})
		if _, err := runUpdateDownload(context.Background(), f.srv.Client(), f.plan(name), t.TempDir(), nil); err == nil {
			t.Fatal("phải từ chối: bản ký cũ không chứa tên file bản mới (chống hạ cấp)")
		}
	})

	t.Run("SHA256SUMS quá lớn", func(t *testing.T) {
		big := []byte(strings.Repeat("#", maxSumsBytes+1))
		f := newFakeRelease(t, map[string][]byte{name: body, sumsName: big, sigName: sign(priv, big)})
		if _, err := runUpdateDownload(context.Background(), f.srv.Client(), f.plan(name), t.TempDir(), nil); err == nil {
			t.Fatal("phải từ chối SHA256SUMS lớn bất thường")
		}
	})

	t.Run("huỷ giữa chừng", func(t *testing.T) {
		f := newFakeRelease(t, map[string][]byte{name: body, sumsName: sums, sigName: sign(priv, sums)})
		ctx, cancel := context.WithCancel(context.Background())
		dir := t.TempDir()
		_, err := runUpdateDownload(ctx, f.srv.Client(), f.plan(name), dir, func(done, total int64) { cancel() })
		if err == nil {
			t.Fatal("huỷ mà vẫn xong")
		}
		if left, _ := os.ReadDir(dir); len(left) != 0 {
			t.Errorf("file tải dở còn lại: %v", left)
		}
	})
}

func TestDownloadVerified_GioiHanDungLuong(t *testing.T) {
	body := []byte(strings.Repeat("x", 5000))
	// Máy chủ không báo Content-Length: vẫn phải dừng ở giới hạn.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.(http.Flusher).Flush()
		w.Write(body)
	}))
	defer srv.Close()
	dest := filepath.Join(t.TempDir(), "f")
	err := downloadVerified(context.Background(), srv.Client(), srv.URL+"/f", dest, sha(body), 1000, nil)
	if err == nil || !strings.Contains(err.Error(), "giới hạn") {
		t.Fatalf("muốn lỗi vượt giới hạn, được %v", err)
	}
	if _, err := os.Stat(dest + ".dang-tai"); !os.IsNotExist(err) {
		t.Error("file tải dở còn lại")
	}
	if err := downloadVerified(context.Background(), srv.Client(), srv.URL+"/f", dest, sha(body), 5000, nil); err != nil {
		t.Errorf("vừa đúng giới hạn phải đạt: %v", err)
	}
}

func TestCheckRelease_BanThuLenBanChinhThuc(t *testing.T) {
	srv := releaseServer(t, 200, `{"tag_name":"v0.1.2","assets":[{"name":"SHA256SUMS","browser_download_url":"u","size":3}]}`)
	for cur, want := range map[string]bool{"0.1.2-rc.1": true, "0.1.1": true, "0.1.2": false, "0.1.3-rc.1": false, "0.0.0-dev.abc1234": true} {
		info, rel, err := checkRelease(context.Background(), srv.Client(), srv.URL, cur)
		if err != nil {
			t.Fatal(err)
		}
		if info.Available != want {
			t.Errorf("đang dùng %s: có bản mới = %v, muốn %v", cur, info.Available, want)
		}
		if rel == nil || rel.Version != "0.1.2" || len(rel.Assets) != 1 || rel.Assets[0].Name != sumsName {
			t.Errorf("thông tin file đính kèm sai: %+v", rel)
		}
	}
}

func TestExtractExe(t *testing.T) {
	dir := t.TempDir()
	mk := func(name string, entries map[string]string) string {
		p := filepath.Join(dir, name)
		f, _ := os.Create(p)
		zw := zip.NewWriter(f)
		for n, c := range entries {
			w, _ := zw.Create(n)
			w.Write([]byte(c))
		}
		zw.Close()
		f.Close()
		return p
	}
	ok := mk("ok.zip", map[string]string{"Sano.exe": "MZ bản mới", "khac.txt": "x"})
	dest := filepath.Join(dir, "Sano.exe.moi")
	if err := extractExe(ok, dest); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(dest); string(b) != "MZ bản mới" {
		t.Errorf("nội dung = %q", b)
	}
	for name, entries := range map[string]map[string]string{
		"thieu.zip":  {"Khac.exe": "x"},
		"thumuc.zip": {"sub/Sano.exe": "x"},
	} {
		if err := extractExe(mk(name, entries), filepath.Join(dir, "out")); err == nil {
			t.Errorf("%s: phải lỗi", name)
		}
	}
	if err := extractExe(filepath.Join(dir, "khong-co.zip"), dest); err == nil {
		t.Error("file không tồn tại: phải lỗi")
	}
}

func TestOldPID(t *testing.T) {
	me := os.Getpid()
	for args, want := range map[string]int{
		"--sano-doi 123":              123,
		"-x --sano-doi 42 --y":        42,
		"--sano-doi":                  0,
		"--sano-doi abc":              0,
		"--sano-doi -5":               0,
		"":                            0,
		fmt.Sprint("--sano-doi ", me): 0, // không tự đợi chính mình
	} {
		if got := oldPID(strings.Fields(args)); got != want {
			t.Errorf("%q → %d, muốn %d", args, got, want)
		}
	}
}

func TestProcessAlive(t *testing.T) {
	if !processAlive(os.Getpid()) {
		t.Error("tiến trình đang chạy phải còn sống")
	}
	if processAlive(1 << 30) {
		t.Error("pid không tồn tại phải báo đã thoát")
	}
}
