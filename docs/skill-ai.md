---
title: Skill làm sách nói cho AI
description: 'Skill miễn phí cho Claude và ChatGPT: biến tài liệu Word, PDF thành bản đọc sách nói nghe cuốn hút, giữ đủ kiến thức, trả về file Word nạp thẳng vào Sano.'
---

# Skill làm sách nói cho AI

Tài liệu viết để **đọc bằng mắt** thường khó nghe: bảng biểu, hình, gạch đầu dòng, chữ viết tắt, câu dài. Skill **sano-sach-noi** dạy Claude hoặc ChatGPT biên tập lại thành bản để **nghe bằng tai**, rồi trả về file Word có sẵn chương, mục. Bạn nạp file đó vào Sano là đọc thành sách nói.

Nạp skill **một lần**. Mỗi lần làm sách chỉ cần đính kèm file Word hoặc PDF và gõ một câu, không phải dán prompt dài.

<div class="sano-actions">
  <a class="sano-btn brand" href="/skill/sano-sach-noi.zip" download>Tải skill sano-sach-noi.zip</a>
  <a class="sano-btn outline" href="/skill/sano-huong-dan-ai.txt" download>Tải sano-huong-dan-ai.txt (Dự án ChatGPT)</a>
</div>

Miễn phí, mã nguồn mở. Mã nguồn skill ở thư mục [`skills/`](https://github.com/tanviet12/sano-sach-noi/tree/main/skills) trên GitHub.

## Skill làm được gì

| Việc | Khi nào dùng | AI làm gì |
|---|---|---|
| **Viết lại thành văn sách nói** (cấp độ 3) | Sách kiến thức, tài liệu nội bộ, bài giảng muốn nghe cuốn như nghe kể chuyện | Mỗi chương mở bằng một câu chuyện, rồi mới vào lý thuyết. Câu ngắn, viết như nói. Chương dài 5 đến 10 phút nghe, cuối chương có **Ba ý cần nhớ** |
| **Làm mượt, giữ nguyên ý** (cấp độ 2) | Muốn giữ nguyên giọng văn tác giả | Đổi bảng, hình, danh sách, chữ viết tắt, ký hiệu thành lời. Không thêm, không bớt ý |
| **Soát lại** | Sau khi viết lại, muốn chắc không mất ý | Đối chiếu bản viết lại với bản gốc, bổ sung chỗ thiếu, sửa chỗ tự thêm |

Hai điều skill luôn giữ: **không làm mất kiến thức** (đủ khái niệm, định nghĩa, điều kiện, ngoại lệ) và **không bịa** (không thêm số liệu, tên người, trích dẫn mà tài liệu không có). Trước khi viết, AI lập danh sách ý của tài liệu; viết xong tự soát lại theo danh sách đó.

So sánh ba cách đọc và cách chọn: xem [Ba cách đọc](./lam-muot-tai-lieu).

## Nạp skill (làm một lần)

### Claude

Mọi gói Claude, kể cả gói miễn phí, đều nạp được skill.

1. Tải `sano-sach-noi.zip` ở trên (không cần giải nén).
2. Mở [claude.ai](https://claude.ai) → **Customize** → **Skills** → nút **+** → **Create skill** → **Upload a skill** → chọn file zip.

Dùng **Claude Code**: chép thư mục `sano-sach-noi/` vào `~/.claude/skills/`.

### ChatGPT

- **Gói có mục Skills** (Business, Enterprise, Edu): **Skills** → **Tạo** → **Tải lên** → chọn `sano-sach-noi.zip`.
- **Các gói khác, dùng Dự án (Project):**
  1. Tải `sano-huong-dan-ai.txt` ở trên.
  2. Tạo một Dự án tên "Sano – sách nói", thêm file vào phần **Tệp** của dự án.
  3. Dán câu này vào ô **Hướng dẫn** (Instructions):

     > Mỗi khi tôi gửi tài liệu để làm sách nói, làm đúng theo file sano-huong-dan-ai.txt đã đính kèm. Mặc định làm việc A, cấp 3 viết lại thành văn sách nói. Tôi ghi "cấp 2" thì làm việc B, làm mượt.

### Gemini

Gemini chưa nạp được skill. Trong Sano, ở màn nhờ AI chọn Gemini rồi bấm **Sao chép prompt**, mỗi lần làm sách dán prompt một lần.

### Nạp ngay trong phần mềm

Ở bước **Tạo sách nói** → **Cách đọc**, chọn cấp 2 hoặc 3 → bấm **Nạp skill cho Claude** (hoặc ChatGPT). Sano cho tải file skill, mở trang Skills của AI và có sẵn các câu gõ mẫu để sao chép.

<img class="app-shot" src="./images/app/nap-skill.jpg" alt="Hộp Nạp skill cho Claude: tải file skill, mở Skills của Claude, và các câu gõ mẫu có nút Sao chép" width="1600" height="955">

## Mỗi lần làm sách

Mở cuộc trò chuyện mới, đính kèm file Word hoặc PDF rồi gõ một câu:

| Việc | Claude, ChatGPT có mục Skills | ChatGPT dùng Dự án |
|---|---|---|
| Viết lại thành văn sách nói | Làm file sách nói dùng skill sano-sach-noi (cấp độ 3) | Làm file sách nói theo file sano-huong-dan-ai.txt (cấp độ 3) |
| Làm mượt, giữ nguyên ý | Làm file sách nói dùng skill sano-sach-noi (cấp độ 2) | Làm file sách nói theo file sano-huong-dan-ai.txt (cấp độ 2) |
| Soát lại (gửi kèm bản gốc và bản viết lại) | Soát lại file sách nói dùng skill sano-sach-noi | Soát lại file sách nói theo file sano-huong-dan-ai.txt |

Tài liệu dài, AI làm lần lượt từng chương, hết mỗi phần thì dừng chờ bạn gõ **tiếp**. Xong, AI gộp lại một file Word. Tải file đó về, mở Sano → **Tạo sách nói** → nạp file → chọn giọng → nghe thử → tạo sách. Chi tiết: [Tạo sách đầu tiên](./tao-sach-dau-tien).

::: tip Nghe thử trước khi tạo cả cuốn
Bản AI viết lại vẫn nên nghe thử vài đoạn trong Sano trước khi tạo cả cuốn. Chỗ nào đọc chưa ổn, sửa ngay trong bước nghe thử.
:::

## Câu hỏi thường gặp

**Skill có mất phí không?** Không. Skill miễn phí. Bạn chỉ cần tài khoản Claude hoặc ChatGPT (gói miễn phí dùng được, tài liệu dài có thể chạm giới hạn lượt dùng của gói).

**Tài liệu của tôi có bị gửi đi đâu không?** Skill chỉ là hướng dẫn cho AI. Tài liệu bạn đính kèm được gửi tới Claude hoặc ChatGPT như mọi lần bạn chat với AI đó. Tài liệu nhạy cảm thì dùng cách đọc **nguyên văn** của Sano: không qua AI, chạy hoàn toàn trên máy bạn.

**AI trả nội dung trong khung chat, không có file Word?** Một số gói không tạo được file. Khi đó AI trả nội dung trong một khối mã, dòng đầu là tên sách. Sao chép khối đó, rồi ở bước nạp file của Sano bấm **Dán văn bản AI trả về**.

**Skill có bản mới thì sao?** Tải lại file ở trang này và nạp đè lên skill cũ.
