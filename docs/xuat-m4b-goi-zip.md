---
title: Xuất M4B và gói zip
description: 'Làm sách nói bằng AI với Sano rồi xuất thành một file M4B có mục lục chương và bìa để nghe trên điện thoại, trên xe; hoặc lấy gói zip để sao lưu, chuyển máy.'
---

# Xuất M4B và gói zip

Sano có hai cách mang sách ra khỏi phần mềm: **file M4B** để nghe ở nơi khác, và **gói zip** để sao lưu.

## Xuất file M4B

M4B là định dạng sách nói chuẩn: **một file duy nhất** chứa cả cuốn, có mục lục chương, tên sách, tác giả và ảnh bìa. App sách nói trên điện thoại và màn hình xe (CarPlay, Android Auto) đọc được.

1. Mở cuốn sách trong Thư viện → bấm **Xuất M4B** (hoặc **Xuất file M4B** ngay khi render xong).
2. Chọn nơi lưu. Mặc định là thư mục Tải về, tên file là tên sách.
3. Sano xuất chạy nền, hiện phần trăm, bấm **Huỷ** được (không để lại file dở).
4. Xong, Sano mở thư mục và chọn sẵn file.

<img class="app-shot" src="./images/app/xuat-m4b.jpg" alt="Trình phát với nút Xuất M4B, Xuất gói zip; dòng báo đã lưu file M4B kèm số mốc chương" width="1600" height="955">

Dung lượng khoảng **29 MB cho mỗi giờ nghe** (AAC 64 kbps mono, đủ rõ cho giọng đọc). Mỗi lúc xuất một cuốn.

Tiếp theo: [chép file M4B sang điện thoại](./nghe-tren-dien-thoai).

::: tip File M4B mẫu
Muốn thử xuất M4B mà chưa tạo cuốn nào? Thư viện có sẵn 3 cuốn mẫu ngay lần đầu mở, cuốn ngắn nhất là "Giới thiệu Sano" (3 phút). Mở cuốn đó, bấm **Xuất M4B** rồi chép sang điện thoại để xem.
:::

## Gói zip

Mỗi cuốn sách có sẵn một file `book-<tên-sách>.zip` trong thư mục của nó: gồm audio từng mục, bìa và thông tin sách. Bấm **Xuất gói zip** (trong trình phát hoặc ngay khi render xong), Sano mở trình quản lý file và chọn sẵn file zip đó để bạn chép đi sao lưu.

### Nhập gói zip (từ 0.1.8)

Nhận gói zip từ người khác, hoặc lấy lại bản sao lưu: vào **Thư viện**, bấm **Nhập sách** rồi chọn file `.zip` (hoặc kéo thả file vào cửa sổ). Sano cho xem trước tên sách, giọng đọc, số chương, thời lượng, rồi mới nhập. Sách nhập vào giữ đủ mục lục, bìa, tên giọng và chữ chạy theo.

- **Trùng sách đã có:** chọn **Giữ cả hai** (cuốn mới mang tên "… (2)") hoặc **Thay thế** (cuốn cũ vào Thùng rác, lấy lại được).
- **An toàn:** Sano chỉ đọc đúng các file của gói sách (thông tin sách, mục lục, bìa ảnh, file mp3), kiểm từng file đúng định dạng và giới hạn dung lượng, rồi đóng gói lại. Gói lạ, bị sửa hay có file nằm ngoài chỗ quy định bị từ chối, không có gì được thêm vào thư viện.
- Chỉ nhập sách bạn có quyền nghe: sách tự làm, sách được tác giả cho phép chia sẻ, hoặc tác phẩm đã hết bản quyền.

Người nhận không dùng Sano thì gửi file M4B: nghe được bằng app sách nói bất kỳ. Muốn chuyển cả thư viện sang máy khác, chép nguyên thư mục `~/Sano/Sach` sang cùng vị trí trên máy mới. Cấu trúc gói zip ghi ở [Định dạng gói zip](./book-zip-format).

## Dùng dòng lệnh

Công cụ dòng lệnh `sano-docx2tts` (trong mã nguồn) xuất M4B từ một thư mục sách đã tạo:

```bash
sano-docx2tts --m4b-from-dir ~/Sano/Sach/<tên-sách> --m4b ~/Downloads/sach.m4b
```

Muốn file nhỏ hơn thì thêm `--m4b-bitrate 48k`. Chi tiết ở [Đọc giọng bằng dòng lệnh](./tts-build-guide).
