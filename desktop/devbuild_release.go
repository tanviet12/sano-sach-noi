//go:build !dev

package main

// devBuild — false ở bản phát hành (`wails build` gắn tag "production") và khi
// chạy test: không dò script đọc giọng theo thư mục hiện tại.
const devBuild = false
