package setup

import (
	"fmt"
	"strings"
)

// artifact — một file tải về: URL + SHA256 ghim + cách lấy file chạy ra.
type artifact struct {
	URL    string
	SHA256 string
	Kind   string // "tar.gz" | "zip" | "gz"
	Member string // tên file cần lấy trong file nén ("" với .gz)
}

// uvTargets — tên bản dựng uv theo GOOS/GOARCH (đúng tên file trên GitHub Release).
var uvTargets = map[string]string{
	"darwin/arm64":  "aarch64-apple-darwin",
	"darwin/amd64":  "x86_64-apple-darwin",
	"windows/amd64": "x86_64-pc-windows-msvc",
	"windows/arm64": "aarch64-pc-windows-msvc",
	"linux/amd64":   "x86_64-unknown-linux-gnu",
	"linux/arm64":   "aarch64-unknown-linux-gnu",
}

// uvArtifact — bản uv đúng UV_VERSION cho máy này, SHA256 lấy từ versions.env.
func uvArtifact(pins map[string]string, goos, goarch string) (artifact, error) {
	target, ok := uvTargets[goos+"/"+goarch]
	if !ok {
		return artifact{}, fmt.Errorf("chưa hỗ trợ máy %s/%s", goos, goarch)
	}
	ver := pins["UV_VERSION"]
	sum := pins["UV_SHA256_"+strings.ToUpper(strings.ReplaceAll(target, "-", "_"))]
	if ver == "" || sum == "" {
		return artifact{}, fmt.Errorf("versions.env thiếu UV_VERSION hoặc SHA256 uv cho %s", target)
	}
	a := artifact{
		URL:    "https://github.com/astral-sh/uv/releases/download/" + ver + "/uv-" + target,
		SHA256: sum,
		Kind:   "tar.gz",
		Member: "uv",
	}
	if goos == "windows" {
		a.URL += ".zip"
		a.Kind, a.Member = "zip", "uv.exe"
	} else {
		a.URL += ".tar.gz"
	}
	return a, nil
}

// ffmpegArtifact — bản ffmpeg tĩnh ghim cho máy này. Windows ARM chạy bản x64
// (Windows 11 ARM giả lập được). Không có bản ghim → lỗi (hướng dẫn cài tay).
func ffmpegArtifact(pins map[string]string, goos, goarch string) (artifact, error) {
	arch := goarch
	if goos == "windows" && goarch == "arm64" {
		arch = "amd64"
	}
	key := "FFMPEG_" + strings.ToUpper(goos) + "_" + strings.ToUpper(arch)
	url, sum := pins[key+"_URL"], pins[key+"_SHA256"]
	if url == "" || sum == "" {
		return artifact{}, fmt.Errorf("chưa có bản ffmpeg dựng sẵn cho %s/%s", goos, goarch)
	}
	a := artifact{URL: url, SHA256: sum}
	switch {
	case strings.HasSuffix(url, ".zip"):
		a.Kind, a.Member = "zip", "ffmpeg"
		if goos == "windows" {
			a.Member = "ffmpeg.exe"
		}
	case strings.HasSuffix(url, ".tar.xz"):
		a.Kind, a.Member = "tar.xz", "ffmpeg"
	case strings.HasSuffix(url, ".gz"):
		a.Kind = "gz"
	default:
		return artifact{}, fmt.Errorf("không rõ định dạng file ffmpeg: %s", url)
	}
	return a, nil
}

// vieneuTarball — URL tarball GitHub của commit VieNeu ghim + tree hash cần khớp.
func vieneuTarball(pins map[string]string) (url, commit, treeSHA string, err error) {
	repo, commit, treeSHA := strings.TrimSuffix(pins["VIENEU_REPO"], "/"), pins["VIENEU_COMMIT"], pins["VIENEU_TREE_SHA256"]
	if repo == "" || commit == "" || len(treeSHA) != 64 {
		return "", "", "", fmt.Errorf("versions.env thiếu VIENEU_REPO / VIENEU_COMMIT / VIENEU_TREE_SHA256")
	}
	return repo + "/archive/" + commit + ".tar.gz", commit, treeSHA, nil
}

// officiallySupported — hệ máy VieNeu khai báo chạy được (required-environments
// trong pyproject.toml của VieNeu). Máy khác vẫn thử cài, nhưng báo trước.
func officiallySupported(goos, goarch string) bool {
	switch goos + "/" + goarch {
	case "darwin/arm64", "windows/amd64", "linux/amd64":
		return true
	}
	return false
}
