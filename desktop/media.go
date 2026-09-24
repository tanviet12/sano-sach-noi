package main

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"sano/desktop/internal/library"
)

// mediaPrefix — đường dẫn giao diện dùng để phát MP3 / hiện bìa nằm trong
// ~/Sano (thư viện + file nghe thử). Phục vụ qua middleware của asset server
// Wails nên chạy được cả bản build lẫn `wails dev` (trình duyệt :34115).
const mediaPrefix = "/sano-media/"

// mediaTypes — chỉ phục vụ các loại file này, không lộ file khác trong ~/Sano.
var mediaTypes = map[string]string{
	".mp3":  "audio/mpeg",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
}

// mediaMiddleware chặn request /sano-media/<đường dẫn tương đối ~/Sano>, còn
// lại chuyển tiếp cho asset server (giao diện nhúng hoặc Vite khi dev).
func mediaMiddleware(lib *library.Library) assetserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, mediaPrefix) {
				next.ServeHTTP(w, r)
				return
			}
			serveMedia(lib, w, r)
		})
	}
}

func serveMedia(lib *library.Library, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rel := strings.TrimPrefix(r.URL.Path, mediaPrefix)
	ctype, ok := mediaTypes[strings.ToLower(filepath.Ext(rel))]
	if !ok {
		http.NotFound(w, r)
		return
	}
	full, err := lib.Resolve(rel)
	if err != nil || !fileExists(full) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, full) // hỗ trợ Range → tua được trong thẻ <audio>
}
