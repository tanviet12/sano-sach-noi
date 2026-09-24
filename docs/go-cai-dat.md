---
title: Gỡ cài đặt
description: 'Gỡ Sano, phần mềm làm sách nói bằng AI, khỏi Windows, macOS, Linux: gỡ bộ đọc và mô hình giọng đọc để giải phóng dung lượng, giữ hoặc xoá sách nói đã tạo.'
---

# Gỡ cài đặt

Sano có ba phần nằm ở ba chỗ khác nhau: **phần mềm**, **bộ đọc** (giọng đọc AI và mô hình, khoảng 1,5 GB) và **sách bạn đã tạo**. Gỡ phần nào cũng được, phần khác giữ nguyên.

## 1. Gỡ bộ đọc (giải phóng dung lượng)

Trong Sano: **Cài đặt** → **Bộ đọc** → **Gỡ**, xác nhận **Gỡ bộ đọc**.

Sano xoá thư mục bộ đọc trong thư mục dữ liệu của app: VieNeu-TTS, mô hình giọng đọc, Python và ffmpeg do Sano tải về. Sano **không đụng** tới sách của bạn (`~/Sano`), bộ đọc bạn tự cài tay, hay bộ nhớ đệm Hugging Face dùng chung.

Muốn dùng lại, mở Sano và bấm **Cài bộ đọc**. Không gỡ được khi đang render hoặc đang cài.

## 2. Gỡ phần mềm

**Windows:** **Cài đặt** → **Ứng dụng** → **Ứng dụng đã cài đặt** → Sano → **Gỡ cài đặt**. Bản `portable.zip` thì xoá file `Sano.exe`.

**macOS:** kéo **Sano** từ thư mục Applications vào Thùng rác.

**Linux:** xoá file `Sano-*.AppImage`.

## 3. Xoá dữ liệu còn lại (không bắt buộc)

Gỡ phần mềm chưa xoá dữ liệu. Muốn xoá sạch, xoá thêm thư mục dữ liệu của app (chứa bộ đọc nếu chưa gỡ ở bước 1, và lần đồng ý điều khoản):

| Máy | Thư mục dữ liệu |
|---|---|
| macOS | `~/Library/Application Support/Sano` |
| Windows | `%LOCALAPPDATA%\Sano` |
| Linux | `~/.local/share/sano` (hoặc `$XDG_DATA_HOME/sano`) |

::: danger Sách của bạn
Sách nói đã tạo nằm ở `~/Sano/Sach`. Chỉ xoá thư mục `~/Sano` khi chắc chắn không cần sách nữa, hoặc đã sao lưu. Thư mục này không tự mất khi gỡ phần mềm.
:::
