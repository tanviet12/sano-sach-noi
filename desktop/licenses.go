package main

import (
	_ "embed"
)

// Giấy phép bên thứ ba đi kèm app (ghi công VieNeu-TTS, mô hình, ffmpeg...).
//
//go:embed licenses/THIRD-PARTY-NOTICES.md
var thirdPartyNotices string

// Toàn văn Apache License 2.0 (file LICENSE của VieNeu-TTS, áp cho cả mô hình).
//
//go:embed licenses/Apache-2.0.txt
var apacheLicense string

// ThirdPartyNotices trả danh sách giấy phép bên thứ ba + toàn văn Apache-2.0.
func (a *App) ThirdPartyNotices() string {
	return thirdPartyNotices + "\n\n---\n\n# Apache License 2.0 (VieNeu-TTS, mô hình VieNeu-TTS v3 Turbo, MOSS Audio Tokenizer)\n\n" + apacheLicense
}
