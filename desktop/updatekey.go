package main

// Khoá công khai ed25519 (32 byte, base64) kiểm chữ ký SHA256SUMS.sig của bản
// phát hành. Khoá bí mật chỉ nằm trong secret SANO_UPDATE_SIGNING_KEY của repo
// (CI ký lúc tạo release), không bao giờ vào mã nguồn. Nhiều khoá cách nhau
// dấu phẩy (đổi khoá: thêm khoá mới, giữ khoá cũ vài bản rồi bỏ).
//
// Tạo khoá + ghi dòng này: scripts/release/update-key.sh. Rỗng → app không tự
// cập nhật, chỉ mở trang tải.
//
// Là biến (không phải hằng) để bản thử cục bộ gắn khoá thử qua
// -ldflags "-X main.updatePublicKeys=..." — người build vốn đã kiểm soát bản
// build của mình, không mở thêm đường nào cho kẻ khác.
var updatePublicKeys = ""
