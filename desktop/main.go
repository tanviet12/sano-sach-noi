// Sano desktop — phần mềm tạo sách nói từ file Word, chạy trên máy người dùng.
// Vỏ Wails v2: Go ở đây, giao diện Vue ở frontend/.
package main

import (
	"context"
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

// frontend/dist do `wails build` / `npm run build` sinh ra. Trong git chỉ có
// file giữ chỗ gitkeep để `go build` / `go vet` chạy được trên bản clone mới.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	startAfterUpdate(os.Args[1:])
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "Sano",
		Width:     1100,
		Height:    720,
		MinWidth:  960,
		MinHeight: 640,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: mediaMiddleware(app.lib),
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeClose,
		OnShutdown: func(context.Context) {
			app.CancelRender()
			app.CancelM4B()
			app.CancelSetup() // lần sau mở cài tiếp từ bước dở
			app.CancelUpdate()
			app.applyOnShutdown() // đã chọn "Khởi động lại sau" → thay bản mới lúc thoát
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
		Mac: &mac.Options{
			// Thanh tiêu đề trong suốt: nút đóng/thu nhỏ của macOS nằm trên thanh
			// tiêu đề tự vẽ của giao diện (giống wireframe).
			TitleBar:   mac.TitleBarHiddenInset(),
			Appearance: mac.DefaultAppearance,
			About: &mac.AboutInfo{
				Title:   "Sano",
				Message: "Biến tài liệu của chính bạn thành sách nói, chạy trên máy bạn. Mã nguồn mở.",
			},
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
