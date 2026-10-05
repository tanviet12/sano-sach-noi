# File PDF mẫu cho test (pdf_test.go)

Tạo bằng Chromium (Playwright `page.pdf`) và pypdf, không chứa nội dung có bản quyền:

- `outline.pdf` — sách tiếng Việt 7 trang: trang bìa, 3 chương (h1), mỗi chương 2 mục (h2), có bookmark, đầu trang + số trang.
- `tcvn3.pdf` — chữ "phông cũ TCVN3" (ký tự Latin-1 thay chữ có dấu), để thử cảnh báo lỗi phông.
- `scanned.pdf` — 1 trang chỉ có ảnh chụp chữ (giống bản scan).
- `password.pdf` — `outline.pdf` khoá mật khẩu mở (AES-256).
- `nocopy.pdf` — `outline.pdf` có khoá chủ sở hữu, cấm trích chữ (chỉ cho in).
