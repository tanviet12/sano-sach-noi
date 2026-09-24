// Package bookmaker là lõi làm sách nói từ file Word, dùng chung cho lệnh
// cmd/sano-docx2tts và phần mềm desktop:
//
//   - nạp .docx → mục lục chương/tiểu mục + cảnh báo lúc nạp (docx.go, warnings.go)
//   - chuẩn hóa văn nói, không dùng AI (normalize.go, spoken.go, headings.go,
//     pronunciations.go, toc.go)
//   - đọc giọng bằng VieNeu-TTS qua scripts/tts/audio_gen_batch.py rồi ffmpeg ra
//     MP3 (tts.go)
//   - đóng gói zip chuẩn docs/book-zip-format.md (zip.go)
//
// Run chạy trọn một lượt như lệnh CLI.
package bookmaker
