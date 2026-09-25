<p align="center"><img src="docs/public/logo.svg" alt="Sano" width="72" height="72"></p>

<h1 align="center">Sano — Tạo sách nói bằng AI từ file Word</h1>

<p align="center">Biến tài liệu của chính bạn thành sách nói. Giọng đọc AI tiếng Việt chạy ngay trên máy tính.<br>Miễn phí · Mã nguồn mở · Không cần API key · Windows, macOS, Linux</p>

<p align="center">
  <a href="https://github.com/tanviet12/sano-sach-noi/releases/latest"><img src="https://img.shields.io/github/v/release/tanviet12/sano-sach-noi?label=b%E1%BA%A3n%20m%E1%BB%9Bi%20nh%E1%BA%A5t&color=c60505" alt="Bản mới nhất"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/gi%E1%BA%A5y%20ph%C3%A9p-MIT-blue" alt="Giấy phép MIT"></a>
  <a href="https://github.com/tanviet12/vbsec"><img src="https://img.shields.io/badge/vbsec-%C4%91%C3%A3%20qu%C3%A9t%20b%E1%BA%A3o%20m%E1%BA%ADt%20%C2%B7%20%C4%91%E1%BA%A1t-2ea44f" alt="vbsec: đã quét bảo mật, đạt"></a>
</p>

<p align="center">
  <b><a href="https://github.com/tanviet12/sano-sach-noi/releases/latest">Tải Sano</a></b> ·
  <b><a href="https://tanviet12.github.io/sano-sach-noi/demo">Nghe thử</a></b> ·
  <b><a href="https://tanviet12.github.io/sano-sach-noi/">Trang chủ</a></b> ·
  <b><a href="https://tanviet12.github.io/sano-sach-noi/cai-dat">Hướng dẫn</a></b>
</p>

<p align="center">
  <a href="https://tanviet12.github.io/sano-sach-noi/video/sano-gioi-thieu.mp4"><img src="docs/images/readme/sano-demo.gif" alt="Sano: nạp file Word, nghe thử và sửa lời đọc, tạo sách, nghe trên máy hoặc điện thoại" width="860"></a><br>
  <sub>Bấm vào ảnh để xem video 40 giây có tiếng (giọng Hải Đăng)</sub>
</p>

## Sano làm được gì

- **Tạo sách nói** từ tài liệu Word của chính bạn: tự đọc mục lục theo Heading, tự chuẩn hoá số, chữ viết tắt, ký hiệu thành lời đọc tự nhiên
- **Nghe trên máy tính** ngay trong phần mềm, nhớ chỗ nghe dở
- **Nghe trên điện thoại**: xuất một file M4B có mục lục chương và bìa, nghe bằng app BookPlayer (miễn phí, có cho iPhone và Android)
- **Nghe khi lái xe ô tô**: BookPlayer chạy trên CarPlay và Android Auto, chọn sách, chọn chương ngay trên màn hình xe
- **25 giọng đọc AI tiếng Việt**: nam, nữ, giọng Bắc, Trung, Nam
- **Không cần API key, không tốn tiền token**: mô hình AI tải về một lần rồi chạy ngay trên máy, không cần tài khoản ChatGPT hay dịch vụ AI nào; tài liệu không gửi lên mạng

## Nghe thử

5 cuốn sách mẫu tự viết, mỗi cuốn một giọng, tạo hoàn toàn bằng Sano. Bấm để nghe chương 1 ngay trong trình duyệt, hoặc mở **[trang nghe thử](https://tanviet12.github.io/sano-sach-noi/demo)** để nghe đủ 3 chương mỗi cuốn và so 12 giọng đọc cùng một đoạn.

| | Sách | Giọng | Nghe |
|---|---|---|---|
| <img src="docs/public/audio/demo/ky-nang-mem-cho-nguoi-tre.jpg" width="56" alt=""> | Kỹ năng mềm cho người trẻ | Hải Đăng · nam · Bắc | [Chương 1 ▶](https://tanviet12.github.io/sano-sach-noi/audio/demo/ky-nang-mem-cho-nguoi-tre-1.mp3) |
| <img src="docs/public/audio/demo/tam-ly-tich-cuc.jpg" width="56" alt=""> | Tâm lý tích cực | Trúc Ly · nữ · Bắc | [Chương 1 ▶](https://tanviet12.github.io/sano-sach-noi/audio/demo/tam-ly-tich-cuc-1.mp3) |
| <img src="docs/public/audio/demo/khoi-nghiep-tu-so-0.jpg" width="56" alt=""> | Khởi nghiệp từ số 0 | Thái Sơn · nam · Nam | [Chương 1 ▶](https://tanviet12.github.io/sano-sach-noi/audio/demo/khoi-nghiep-tu-so-0-1.mp3) |
| <img src="docs/public/audio/demo/tai-chinh-ca-nhan-co-ban.jpg" width="56" alt=""> | Tài chính cá nhân cơ bản | Thục Đoan · nữ · Nam | [Chương 1 ▶](https://tanviet12.github.io/sano-sach-noi/audio/demo/tai-chinh-ca-nhan-co-ban-1.mp3) |
| <img src="docs/public/audio/demo/lanh-dao-cho-quan-ly-moi.jpg" width="56" alt=""> | Lãnh đạo cho quản lý mới | Ngọc Trân · nữ · Trung | [Chương 1 ▶](https://tanviet12.github.io/sano-sach-noi/audio/demo/lanh-dao-cho-quan-ly-moi-1.mp3) |

Giọng đọc do [VieNeu-TTS](https://github.com/pnnbao97/VieNeu-TTS) tạo (xem mục [Cảm ơn](#cảm-ơn)). Bìa sách do Sano tự vẽ theo tên sách.

## Tạo sách trong 4 bước

<table>
  <tr>
    <td width="50%"><img src="docs/images/app/b1-nap-file.jpg" alt="Bước 1: nạp file Word"><br><b>1. Nạp file Word.</b> Sano đọc mục lục, cảnh báo bảng, hình, chữ viết tắt lạ. Chưa có file thì tải file Word mẫu.</td>
    <td width="50%"><img src="docs/images/app/b5-nghe-thu.jpg" alt="Bước 2: nghe thử và sửa lời đọc"><br><b>2. Chọn giọng, nghe thử.</b> Nghe từng đoạn, sửa lời đọc nếu cần trước khi tạo cả cuốn.</td>
  </tr>
  <tr>
    <td><img src="docs/images/app/b6-render.jpg" alt="Bước 3: tạo sách chạy nền"><br><b>3. Tạo sách.</b> Chạy nền ngay trên máy bạn, vẫn làm việc khác được.</td>
    <td><img src="docs/images/app/nghe.jpg" alt="Bước 4: nghe trong thư viện"><br><b>4. Nghe.</b> Trong thư viện của Sano, hoặc xuất M4B chép sang điện thoại.</td>
  </tr>
</table>

## Nghe trên điện thoại và khi lái xe

<table>
  <tr>
    <td width="22%"><img src="docs/images/nghe-tren-dien-thoai/2-thu-vien.jpg" alt="BookPlayer: thư viện"></td>
    <td width="22%"><img src="docs/images/nghe-tren-dien-thoai/3-trinh-phat.jpg" alt="BookPlayer: đang nghe"></td>
    <td width="56%"><img src="docs/images/nghe-tren-dien-thoai/5-carplay-dang-nghe.jpg" alt="CarPlay: đang nghe trên màn hình xe"><br><img src="docs/images/nghe-tren-dien-thoai/7-carplay-chuong.jpg" alt="CarPlay: chọn chương trên màn hình xe"></td>
  </tr>
</table>

Một file M4B có mục lục chương, tên sách, bìa, khoảng 29 MB cho mỗi giờ nghe. Nghe không cần mạng. Ảnh CarPlay chụp trên xe thật. Hướng dẫn: [nghe trên điện thoại](https://tanviet12.github.io/sano-sach-noi/nghe-tren-dien-thoai) · [nghe khi lái xe ô tô](https://tanviet12.github.io/sano-sach-noi/nghe-khi-lai-xe).

## Tải về

| Máy | Tải | Ghi chú |
|---|---|---|
| **Windows** 10/11 | [Bộ cài .exe](https://github.com/tanviet12/sano-sach-noi/releases/download/v0.1.7/Sano-0.1.7-windows-amd64-setup.exe) · [Bản portable .zip](https://github.com/tanviet12/sano-sach-noi/releases/download/v0.1.7/Sano-0.1.7-windows-amd64-portable.zip) | không cần quyền admin |
| **macOS** 10.13+ | [Sano .dmg](https://github.com/tanviet12/sano-sach-noi/releases/download/v0.1.7/Sano-0.1.7-macos-universal.dmg) | Apple Silicon và Intel |
| **Linux** x86_64 | [Sano .AppImage](https://github.com/tanviet12/sano-sach-noi/releases/download/v0.1.7/Sano-0.1.7-linux-amd64.AppImage) | cần WebKitGTK 4.1 |

Lần mở đầu, Sano tự tải bộ đọc giọng Việt về máy (khoảng 1 GB, chỉ một lần). App chưa ký số nên lần đầu mở: macOS → Cài đặt hệ thống → Quyền riêng tư & Bảo mật → **Vẫn mở**; Windows → **Thông tin thêm** → **Vẫn chạy** ([chi tiết](https://tanviet12.github.io/sano-sach-noi/mo-app-lan-dau)). Kiểm file bằng `SHA256SUMS` trong [bản phát hành](https://github.com/tanviet12/sano-sach-noi/releases/latest); mọi bản cài build trên GitHub Actions từ thẻ phiên bản. Tất cả phiên bản: [Releases](https://github.com/tanviet12/sano-sach-noi/releases).

> Repo không kèm sách nào ngoài 5 cuốn mẫu tự viết. Xem [Bản quyền và trách nhiệm](#bản-quyền-và-trách-nhiệm) trước khi dùng.

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

## Dành cho nhà phát triển

### Stack

- Phần mềm: [Wails v2](https://wails.io) — Go + Vue 3, TypeScript, Tailwind, component kiểu shadcn-vue, icon lucide
- Tạo sách: Go (`internal/bookmaker`), bìa tự vẽ (`internal/cover`), xuất M4B (`internal/m4b`), CLI `cmd/sano-docx2tts`
- Đọc giọng: VieNeu-TTS v3 đọc trước (không realtime), ffmpeg ghi MP3 / M4B. Phần mềm tự cài bộ đọc lần mở đầu

**Brand**: đỏ `#c60505`. Xem [`BRAND.md`](./BRAND.md).

### Chạy dev

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

### Tài liệu

**Trang hướng dẫn: [tanviet12.github.io/sano-sach-noi](https://tanviet12.github.io/sano-sach-noi)** (cài đặt, tạo sách, nghe thử, câu hỏi thường gặp). Mã nguồn trang ở `docs/` (VitePress): `make docs-dev` để xem khi sửa, `make docs-build` để kiểm link gãy.

- Làm mượt tài liệu bằng ChatGPT, Gemini, Claude (bảng, hình thành lời văn): [`docs/lam-muot-tai-lieu.md`](docs/lam-muot-tai-lieu.md)
- Đọc giọng bằng VieNeu-TTS (dòng lệnh): [`docs/tts-build-guide.md`](docs/tts-build-guide.md)
- Nghe trên điện thoại (xuất một file M4B): [`docs/nghe-tren-dien-thoai.md`](docs/nghe-tren-dien-thoai.md) · khi lái xe ô tô: [`docs/nghe-khi-lai-xe.md`](docs/nghe-khi-lai-xe.md)
- Gói zip sao lưu / chuyển máy: [`docs/book-zip-format.md`](docs/book-zip-format.md)
- Giao diện: [`docs/design-system.md`](docs/design-system.md)

### Cấu trúc

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

### Kế hoạch tiếp theo

Bản hiện tại là phần mềm máy tính: tạo sách, nghe ngay trong phần mềm, nghe trên điện thoại bằng file M4B.

Giai đoạn sau: máy chủ nghe sách riêng, dành cho gia đình và nhóm nhỏ — nghe qua web và app điện thoại, nhớ vị trí nghe giữa các máy.

## Giấy phép

[MIT](./LICENSE)
