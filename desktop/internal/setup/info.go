package setup

import (
	"os"
	"path/filepath"
	"runtime"

	"sano/desktop/internal/tts"
)

// minRAM — dưới mức này đọc giọng rất chậm hoặc hết bộ nhớ (mô hình ~600 MB).
const minRAM = 4 << 30

// Info — thông tin hiện trên màn cài trước khi bấm Cài.
type Info struct {
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	CPU           string `json:"cpu"`
	Cores         int    `json:"cores"`
	RAMBytes      int64  `json:"ramBytes"`
	FreeBytes     int64  `json:"freeBytes"`
	NeedBytes     int64  `json:"needBytes"`     // dung lượng bộ đọc sau khi cài
	RequiredFree  int64  `json:"requiredFree"`  // cần trống lúc cài
	DownloadBytes int64  `json:"downloadBytes"` // tổng tải về
	DataDir       string `json:"dataDir"`
	UsedBytes     int64  `json:"usedBytes"` // thư mục bộ đọc app đang chiếm
	Enough        bool   `json:"enough"`
	Supported     bool   `json:"supported"`
	Note          string `json:"note"`
}

// MachineInfo đo máy + thư mục bộ đọc. Không lỗi: thiếu gì thì để trống.
func MachineInfo(l tts.Layout, goos, goarch string) Info {
	inf := Info{
		OS:            osName(goos),
		Arch:          goarch,
		CPU:           cpuName(goos),
		Cores:         runtime.NumCPU(),
		RAMBytes:      int64(ramBytes(goos)),
		NeedBytes:     InstalledSize,
		RequiredFree:  RequiredBytes,
		DownloadBytes: DownloadBytes,
		DataDir:       l.Root,
		UsedBytes:     DirSize(l.Root),
		Supported:     officiallySupported(goos, goarch),
	}
	if free, err := freeBytes(existingParent(l.Root)); err == nil {
		inf.FreeBytes = int64(free)
	}
	need := RequiredBytes - inf.UsedBytes
	if need < minFreeBytes {
		need = minFreeBytes
	}
	inf.Enough = inf.FreeBytes == 0 || inf.FreeBytes >= need
	switch {
	case !inf.Enough:
		inf.Note = "Ổ đĩa còn trống " + humanGB(inf.FreeBytes) + ", cần khoảng " + humanGB(need)
	case inf.RAMBytes > 0 && inf.RAMBytes < minRAM:
		inf.Note = "Máy dưới 4 GB RAM — đọc giọng sẽ chậm"
	case !inf.Supported:
		inf.Note = "Bộ đọc chưa được thử trên máy " + inf.OS + " " + goarch + " — vẫn cài thử được"
	}
	return inf
}

func osName(goos string) string {
	switch goos {
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	}
	return goos
}

// existingParent — thư mục gần nhất đã tồn tại (để đo đĩa trống trước khi tạo).
func existingParent(p string) string {
	for {
		if _, err := os.Stat(p); err == nil {
			return p
		}
		parent := filepath.Dir(p)
		if parent == p {
			return p
		}
		p = parent
	}
}
