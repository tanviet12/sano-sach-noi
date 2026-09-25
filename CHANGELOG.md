# Changelog

## Chưa phát hành

### Tính năng
- **Sách mẫu có sẵn trong thư viện**: lần đầu mở app, thư viện có cuốn "Kỹ năng mềm cho người trẻ" (giọng Hải Đăng, 4 phút) để nghe thử ngay. Xoá đi thì không tự thêm lại

### Thay đổi
- Giọng mặc định đổi từ Thiện Minh sang **Hải Đăng** (nam, miền Bắc, tự nhiên)
- Bước Nghe thử không còn bắt buộc nghe đủ 2 đoạn: chỉ cần tick xác nhận quyền dùng tài liệu là render được cả cuốn

## v0.1.0 (24/09/2026)

Bản đầu tiên. Tạo sách nói từ file Word của chính bạn, giọng đọc tiếng Việt chạy ngay trên máy, nghe trong phần mềm hoặc xuất M4B nghe trên điện thoại.

### Tính năng
- **Phần mềm máy tính** cho Windows, macOS, Linux (Wails); công cụ dòng lệnh `sano-docx2tts` cho ai muốn tự động hoá
- **Tự cài bộ đọc lần mở đầu**: báo trước dung lượng và thời gian, tự tải uv → Python → VieNeu-TTS → mô hình → ffmpeg (nếu thiếu), tiến độ từng bước, huỷ và cài tiếp được; mọi thứ ghim phiên bản và kiểm SHA256; gỡ bộ đọc trong Cài đặt
- **Tạo sách 6 bước**: nạp file Word (đọc mục lục theo Heading, cảnh báo bảng, hình, chữ viết tắt lạ; chặn file có mật khẩu), chọn phần sẽ đọc, chọn giọng (mặc định Thiện Minh), lời mở đầu, **nghe thử bắt buộc** ít nhất 2 đoạn và sửa được lời đọc từng đoạn, render chạy nền huỷ được. **File Word mẫu** (đặt sẵn Heading 1/2, có lời hướng dẫn chuẩn bị file): tải về làm theo, hoặc nạp thẳng để thử
- **Chuẩn hoá lời đọc**: đọc đúng số, ngày tháng, chữ viết tắt, khoảng số ("4-6 tháng" → "4 đến 6 tháng"), dấu "/" theo ngữ cảnh ("300 triệu/năm" → "300 triệu một năm", "2-3 giờ/ngày" → "2 đến 3 giờ một ngày", "có/không" → "có hoặc không")
- **Thư viện**: tìm không dấu, lọc theo danh mục, sắp xếp, "Đang nghe" để nghe tiếp; trình phát nhớ vị trí, tua, đổi tốc độ, mục lục chương
- **Xuất M4B** (một file có mục lục chương + bìa, nghe trên Apple Books, BookPlayer, app sách nói Android, màn hình xe) và **gói zip** để sao lưu hoặc chuyển máy
- **Điều khoản sử dụng** hiện một lần khi mở; xác nhận quyền dùng tài liệu cho từng cuốn
- **Kiểm tra bản mới** qua GitHub Releases khi mở (chỉ lấy số phiên bản, tắt được trong Cài đặt); có bản mới thì mở trang tải để cài đè, sách và bộ đọc giữ nguyên
- **Trang tài liệu**: hướng dẫn cài đặt, tạo sách, làm mượt tài liệu, nghe trên điện thoại, câu hỏi thường gặp

### Bảo mật
- Quét bằng [vbsec](https://github.com/tanviet12/vbsec) 4 lượt, lượt cuối **đạt** (không có lỗi nghiêm trọng, cao hay trung bình)
- ffmpeg tải về (khi máy chưa có): Windows gyan.dev 9.0.2 và Linux BtbN 9.0.1 (nguồn ffmpeg.org giới thiệu), macOS Martin Riedl 9.0.2; đều là bản GPL, ghim SHA256
- Mọi thứ tải về có giới hạn dung lượng; file nén giới hạn số mục và dung lượng khi giải nén; thư mục sách của người khác gửi không được đi theo symlink ra ngoài; giải mã MP3 và đọc ảnh bìa khi xuất M4B ép đúng định dạng
- Build bằng Go 1.26.8, govulncheck sạch
