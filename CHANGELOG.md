# Changelog

## Chưa phát hành

### Tính năng
- **Rời màn nghe vẫn nghe tiếp:** thanh nghe nhỏ ở đáy cửa sổ (phát/dừng, tua −15s/+30s, dừng hẳn); bấm vào tên sách để mở lại màn nghe. Nghe mẫu giọng khi đang tạo sách thì sách tự tạm dừng
- **Sách ghi tên giọng đọc** trên thẻ trong Thư viện và màn nghe (cả sách đã tạo trước đây); tìm sách được theo tên giọng
- **Chọn giọng theo miền:** Miền Bắc / Trung / Nam / Tất cả kèm lọc giọng Nam, Nữ; mở đúng miền đã chọn lần trước, giọng đã dùng có nhãn "Dùng lần trước"

### Sửa lỗi
- Mục lục ở màn nghe tự cuộn tới tiểu mục đang phát, không phải kéo tìm
- Chữ **P** đứng riêng ("chữ P đầu tiên", "P thứ hai là giá") và "4Ps" đọc là "pê" thay vì "phê"

## v0.1.6 (25/09/2026)

### Cải thiện
- **Ngắt nghỉ ở dấu phẩy đều hơn:** câu dài nhiều dấu phẩy trước đây đôi khi bị đọc một mạch; nay bộ đọc cắt câu ở dấu phẩy và nghỉ khoảng 0,3 giây. Sách dài thêm khoảng 1–2%, tốc độ tạo sách không đổi

### Lưu ý khi nâng cấp
- Từ **0.1.2 – 0.1.5**: bấm **Cập nhật ngay** trong app. Từ **0.1.0 / 0.1.1**: tải bản 0.1.6 và cài đè một lần. Sách đã tạo không tự đọc lại; muốn nghe cách ngắt mới thì tạo lại sách

## v0.1.5 (25/09/2026)

### Bảo mật
- **Vá thêm thư viện Python của bộ đọc** trên macOS và Windows: anyio 4.11 → 4.14.2 (lỗi có thể giả mạo chứng chỉ TLS), pygments 2.19 → 2.21. Mô hình giọng đọc giữ nguyên
  - Máy đã cài bộ đọc: mở Sano sẽ thấy **Cập nhật bộ đọc** (khoảng 1 phút). Bỏ qua vẫn tạo sách được; nút cập nhật có trong **Cài đặt → Bộ đọc**
- Nâng công cụ dựng giao diện (vite 8, vue-tsc 3); giao diện không đổi

### Lưu ý khi nâng cấp
- Từ **0.1.2 – 0.1.4**: bấm **Cập nhật ngay** trong app. Từ **0.1.0 / 0.1.1**: tải bản 0.1.5 và cài đè một lần

## v0.1.4 (25/09/2026)

### Sửa lỗi
- **Linux: phát được sách và nghe mẫu giọng.** Trước đây trình phát báo "Không phát được: NotSupportedError" vì WebKitGTK không đọc được file âm thanh qua đường dẫn nội bộ của app; nay Sano đọc file trước rồi mới phát. macOS và Windows không đổi

### Lưu ý khi nâng cấp
- Từ **0.1.2 / 0.1.3**: bấm **Cập nhật ngay** trong app. Từ **0.1.0 / 0.1.1**: tải bản 0.1.4 và cài đè một lần

## v0.1.3 (25/09/2026)

### Bảo mật
- **Vá thư viện Python của bộ đọc**: bỏ phần giao diện web của VieNeu-TTS (gradio cùng fastapi, starlette, python-multipart, pillow… kéo theo) — Sano không dùng; nâng bản đã vá cho filelock, idna, msgpack, protobuf, requests, urllib3. Bộ đọc nhẹ hơn: 92 → 49 thư viện, thư mục bộ đọc còn khoảng 1,1 GB
  - Máy đã cài bộ đọc: mở Sano sẽ thấy **Cập nhật bộ đọc** (khoảng 1 phút, chỉ tải lại vài thư viện, mô hình giọng đọc giữ nguyên). Bỏ qua vẫn tạo sách được; nút cập nhật có trong **Cài đặt → Bộ đọc**
- Mỗi bản phát hành được quét bằng VirusTotal, kết quả ghi trong ghi chú phát hành
- Trang [Chính sách ký số](https://tanviet12.github.io/sano-sach-noi/chinh-sach-ky-so): file nào được ký, ký ở đâu, ai duyệt; đang xin ký số Windows miễn phí qua SignPath Foundation

### Sửa lỗi
- Bấm **Huỷ** khi đang tải bản mới dừng ngay
- Bản build macOS trên CI không còn hỏng vì lỗi tạm "Resource busy" khi đóng gói .dmg

### Lưu ý khi nâng cấp
- Từ **0.1.2**: bấm **Cập nhật ngay** trong app. Từ **0.1.0 / 0.1.1**: tải bản 0.1.3 và cài đè một lần

## v0.1.2 (25/09/2026)

### Tính năng
- **Tự cập nhật ngay trong app**: có bản mới, bấm **Cập nhật ngay** là Sano tự tải đúng bản cho máy (có tiến độ, huỷ được), kiểm chữ ký rồi thay bản và tự mở lại. Sách, tiến độ nghe và bộ đọc giữ nguyên
  - macOS: thay Sano.app trong thư mục Applications; Windows: chạy bộ cài im lặng (hoặc thay Sano.exe nếu dùng bản chạy ngay); Linux: thay file AppImage
  - Đang render thì hẹn **Tự cập nhật khi render xong**; chọn **Khởi động lại sau** thì thay khi thoát Sano
  - Chạy thẳng từ file .dmg hoặc thư mục không ghi được thì vẫn mở trang tải như trước

### Bảo mật
- Mỗi bản phát hành có chữ ký ed25519 (`SHA256SUMS.sig`) ký trên `SHA256SUMS`. App kiểm chữ ký bằng khoá công khai nhúng sẵn **trước khi** tải file cài, rồi kiểm SHA256 của file tải về; sai là từ chối, xoá file, báo rõ. Chỉ tải từ trang phát hành của dự án, giới hạn dung lượng

### Lưu ý khi nâng cấp
- Từ **0.1.0 / 0.1.1**: hai bản này chưa có phần tự cập nhật, cần tải bản 0.1.2 và cài đè **một lần**. Từ 0.1.2 trở đi cập nhật ngay trong app

## v0.1.1 (25/09/2026)

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
