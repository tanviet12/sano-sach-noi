# Sano — Tạo sách nói bằng AI từ file Word

[![vbsec: đã quét bảo mật, đạt](https://img.shields.io/badge/vbsec-%C4%91%C3%A3%20qu%C3%A9t%20b%E1%BA%A3o%20m%E1%BA%ADt%20%C2%B7%20%C4%91%E1%BA%A1t-2ea44f)](https://github.com/tanviet12/vbsec) [![Giấy phép MIT](https://img.shields.io/badge/gi%E1%BA%A5y%20ph%C3%A9p-MIT-blue)](./LICENSE)

Biến tài liệu của chính bạn thành sách nói. Sano giúp bạn **tự làm sách nói bằng AI** từ file Word: giọng đọc tiếng Việt chạy ngay trên máy, có mục lục chương, xuất file M4B nghe trên điện thoại. Miễn phí, mã nguồn mở.

Sano là phần mềm cài trên máy tính (Windows, macOS, Linux): đọc file Word, trích mục lục theo Heading, chuẩn hóa lời đọc, đọc thành giọng nói bằng [VieNeu-TTS](https://github.com/pnnbao97/VieNeu-TTS) chạy ngay trên máy, rồi lưu vào thư viện để nghe. Muốn nghe trên điện thoại hay trên xe thì xuất một file M4B (có mục lục chương + bìa).

**Sano làm được gì**

- **Tạo sách nói** từ tài liệu Word của chính bạn
- **Nghe trên máy tính** ngay trong phần mềm, nhớ chỗ nghe dở
- **Nghe trên điện thoại**: xuất một file M4B, nghe bằng app BookPlayer (miễn phí, có cho iPhone và Android)
- **Nghe khi lái xe ô tô**: BookPlayer chạy trên CarPlay và Android Auto, chọn sách, chọn chương ngay trên màn hình xe
- **25 giọng đọc AI tiếng Việt**: nam, nữ, giọng Bắc, giọng Nam
- **Không cần API key, không tốn tiền token**: mô hình AI tải về một lần rồi chạy ngay trên máy, không cần tài khoản ChatGPT hay dịch vụ AI nào

> Repo không kèm sách hay audio nào. Xem mục [Bản quyền và trách nhiệm](#bản-quyền-và-trách-nhiệm) trước khi dùng.

## Stack

- Phần mềm: [Wails v2](https://wails.io) — Go + Vue 3, TypeScript, Tailwind, component kiểu shadcn-vue, icon lucide
- Tạo sách: Go (`internal/bookmaker`), bìa tự vẽ (`internal/cover`), xuất M4B (`internal/m4b`), CLI `cmd/sano-docx2tts`
- Đọc giọng: VieNeu-TTS v3 đọc trước (không realtime), ffmpeg ghi MP3 / M4B. Phần mềm tự cài bộ đọc lần mở đầu

**Brand**: đỏ `#c60505`. Xem [`BRAND.md`](./BRAND.md).

## Chạy dev

Cần Go 1.26+, Node 20+, Wails CLI (chi tiết từng hệ điều hành ở [`desktop/README.md`](desktop/README.md)).

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
make desktop-dev      # cửa sổ app + Vite :5390
make desktop-build    # ra desktop/build/bin/
make desktop-test     # go vet + go test + typecheck giao diện của desktop/
make test             # go test module gốc (tạo sách, bìa, M4B, script đọc giọng)
```

Chỉ sửa giao diện, không cần Go: `cd desktop/frontend && npm install && npm run dev` rồi mở http://localhost:5390 (dữ liệu giả).

Tạo sách bằng dòng lệnh, không cần mở phần mềm:

```bash
go run ./cmd/sano-docx2tts -gen-sample-docx /tmp/s.docx                       # docx mẫu
go run ./cmd/sano-docx2tts --input /tmp/s.docx --output-dir /tmp/s/ --tts-mode stub   # thử nhanh, audio im lặng
go run ./cmd/sano-docx2tts -h                                                 # mọi cờ (giọng, M4B, gói zip...)
```

## Tài liệu

**Trang hướng dẫn: [tanviet12.github.io/sano-sach-noi](https://tanviet12.github.io/sano-sach-noi)** (cài đặt, tạo sách, nghe thử, câu hỏi thường gặp). Mã nguồn trang ở `docs/` (VitePress): `make docs-dev` để xem khi sửa, `make docs-build` để kiểm link gãy.

- Làm mượt tài liệu bằng ChatGPT, Gemini, Claude (bảng, hình thành lời văn): [`docs/lam-muot-tai-lieu.md`](docs/lam-muot-tai-lieu.md)
- Đọc giọng bằng VieNeu-TTS (dòng lệnh): [`docs/tts-build-guide.md`](docs/tts-build-guide.md)
- Nghe trên điện thoại, trên xe (xuất một file M4B): [`docs/nghe-tren-dien-thoai.md`](docs/nghe-tren-dien-thoai.md)
- Gói zip sao lưu / chuyển máy: [`docs/book-zip-format.md`](docs/book-zip-format.md)
- Giao diện: [`docs/design-system.md`](docs/design-system.md)

## Tải bản cài

Vào trang **Releases** của repo, chọn bản mới nhất:

| Máy | File |
|---|---|
| Windows 10/11 | `Sano-<version>-windows-amd64-setup.exe` (bộ cài, không cần quyền admin) hoặc `...-portable.zip` |
| macOS (Apple Silicon + Intel) | `Sano-<version>-macos-universal.dmg` |
| Linux x86_64 | `Sano-<version>-linux-amd64.AppImage` (cần WebKitGTK 4.1) |

App chưa ký số nên lần đầu mở: macOS → Cài đặt hệ thống → Quyền riêng tư & Bảo mật → **Vẫn mở**; Windows → **Thông tin thêm** → **Vẫn chạy**. Kiểm file bằng `SHA256SUMS` đi kèm. Mọi bản cài build trên GitHub Actions từ thẻ phiên bản — cách phát hành xem [`desktop/README.md`](desktop/README.md#phát-hành).

## Cấu trúc

```
sano-sach-noi/
├── desktop/             # phần mềm (Wails, Go module riêng) — giao diện ở desktop/frontend/
├── cmd/sano-docx2tts/   # CLI tạo sách từ file Word
├── internal/
│   ├── bookmaker/       # đọc docx, chuẩn hóa lời đọc, gọi bộ đọc, ghi MP3 + metadata, gói zip
│   ├── cover/           # vẽ bìa mặc định
│   └── m4b/             # xuất một file M4B có mục lục chương
├── scripts/
│   ├── tts/             # script Python gọi VieNeu-TTS + phiên bản ghim (nhúng vào app)
│   └── release/         # build + đóng gói bản cài 3 hệ điều hành
└── docs/                # hướng dẫn, điều khoản sử dụng, design system
```

## Kế hoạch tiếp theo

Bản hiện tại là phần mềm máy tính: tạo sách, nghe ngay trong phần mềm, nghe trên điện thoại bằng file M4B.

Giai đoạn sau: máy chủ nghe sách riêng, dành cho gia đình và nhóm nhỏ — nghe qua web và app điện thoại, nhớ vị trí nghe giữa các máy.

## Bản quyền và trách nhiệm

Sano là công cụ chuyển văn bản thành giọng đọc, dùng cho **tài liệu của chính bạn** hoặc tài liệu bạn **có quyền sử dụng**: bài viết, giáo trình, ghi chép, tài liệu nội bộ, sách đã hết thời hạn bảo hộ, sách được tác giả cho phép.

- **Bạn tự chịu trách nhiệm** về bản quyền của nội dung mình đưa vào Sano và của sách nói tạo ra.
- **Không tạo sách nói từ tác phẩm còn bản quyền**, kể cả để nghe riêng, trừ khi được tác giả hoặc chủ sở hữu cho phép. Sách nói là tác phẩm phái sinh, cũng không được phát tán khi chưa có sự đồng ý của chủ sở hữu.
- Sano **không có và sẽ không có** tính năng gỡ khoá bảo vệ. File Word hoặc PDF có mật khẩu hay khoá hạn chế sẽ bị từ chối.
- Sano **không có máy chủ hay thư viện chung**: sách bạn tạo chỉ nằm trên máy bạn. Đừng dùng sách nói tạo ra để dựng thư viện mở cho người lạ.
- Phần mềm desktop yêu cầu đồng ý [Điều khoản sử dụng](docs/dieu-khoan-su-dung.md) trước khi dùng.
- Tác giả và người đóng góp Sano không chịu trách nhiệm về cách người dùng sử dụng phần mềm. Phần mềm cung cấp "nguyên trạng" theo [giấy phép MIT](./LICENSE).

## Đơn vị tài trợ

Sano miễn phí và mã nguồn mở nhờ sự tài trợ của:

<table>
  <tr>
    <td align="center" width="50%">
      <a href="https://sepay.vn?utm_source=github&utm_medium=readme&utm_campaign=sano"><img src="desktop/frontend/src/assets/sponsors/sepay.svg" alt="SePay" height="44"></a><br>
      <b><a href="https://sepay.vn?utm_source=github&utm_medium=readme&utm_campaign=sano">SePay</a></b><br>
      Nền tảng Open Banking: tự động xác nhận thanh toán chuyển khoản, kết nối API với các ngân hàng Việt Nam
    </td>
    <td align="center" width="50%">
      <a href="https://123host.vn?utm_source=github&utm_medium=readme&utm_campaign=sano"><img src="desktop/frontend/src/assets/sponsors/123host.svg" alt="123HOST" height="44"></a><br>
      <b><a href="https://123host.vn?utm_source=github&utm_medium=readme&utm_campaign=sano">123HOST</a></b><br>
      Hosting, VPS, máy chủ và tên miền cho doanh nghiệp, nhà phát triển Việt Nam
    </td>
  </tr>
</table>

## Tác giả

Sano do **[Bùi Tấn Việt](https://www.facebook.com/buitanviet)** viết.

Tôi học liên tục và muốn tranh thủ nghe lại tài liệu của mình lúc lái xe, lúc rảnh tay. Ngoài thị trường có nhiều app sách nói, nhưng không có cái nào đọc được tài liệu riêng của mình. Tôi tự làm một công cụ để dùng, thấy hữu ích nên mở mã nguồn cho ai cần.

Hiện tôi là CEO của:

- **[SePay](https://sepay.vn)** — nền tảng Open Banking, tự động xác nhận thanh toán chuyển khoản cho doanh nghiệp
- **[123HOST](https://123host.vn)** — dịch vụ hosting, VPS, tên miền

Dự án mã nguồn mở khác của tôi:

- **[vbsec](https://github.com/tanviet12/vbsec)** — quét bảo mật mã nguồn bằng AI, phát hiện hơn 20 loại lỗ hổng phổ biến. Sano được kiểm tra bằng vbsec.
- **[Chat Quality Agent](https://github.com/tanviet12/chat-quality-agent)** — dùng AI chấm chất lượng chăm sóc khách hàng qua Zalo OA, Facebook Messenger ([hướng dẫn](https://tanviet12.github.io/chat-quality-agent/))

## Cảm ơn

Mọi giọng đọc của Sano do **[VieNeu-TTS](https://github.com/pnnbao97/VieNeu-TTS)** tạo — dự án mã nguồn mở chuyển văn bản thành giọng nói tiếng Việt của [Pham Nguyen Ngoc Bao](https://github.com/pnnbao97), giấy phép Apache-2.0 ([vieneu.io](https://www.vieneu.io)). Sano dùng mô hình [VieNeu-TTS v3 Turbo](https://huggingface.co/pnnbao-ump/VieNeu-TTS-v3-Turbo) chạy ngay trên máy, không chỉnh sửa mã hay mô hình. Cảm ơn tác giả đã chia sẻ.

Các thành phần mã nguồn mở khác và giấy phép của chúng: [`desktop/licenses/THIRD-PARTY-NOTICES.md`](desktop/licenses/THIRD-PARTY-NOTICES.md).

## Giấy phép

[MIT](./LICENSE)
