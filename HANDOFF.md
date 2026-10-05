# HANDOFF — Sano (sách nói tiếng Việt)

Cập nhật: 05/10/2026 · Phiên bản đang phát hành: **0.1.23** (Latest, công khai 05/10, CI xanh 3 hệ + smoke) · `main` = bản phát hành

Phiên mới: đọc file này + `README.md` + `CHANGELOG.md` là đủ nắm trạng thái.

## 1. Dự án là gì

Phần mềm desktop (Wails: Go + Vue 3 + shadcn-vue + Tailwind) biến file Word, EPUB, PDF thành sách nói tiếng Việt.
Giọng đọc VieNeu-TTS v3 Turbo chạy ngay trên máy, không cần API key. Chạy Windows, macOS, Linux.

| Thư mục | Nội dung |
|---|---|
| `internal/bookmaker/` | Lõi làm sách: đọc docx / epub (`epub.go`) / pdf (`pdf.go` dựng sách, `pdfium.go` đọc bằng PDFium WASM), chuẩn hoá lời đọc (tiếng Việt), từ điển cách đọc, gọi bộ đọc, đóng gói zip, sửa sách (`rework.go`), bìa (`coverfit.go`) |
| `desktop/` | App Wails. `maker.go` tạo sách, `rework.go` sửa sách, `dict.go` từ điển, `thumb.go` bìa thu nhỏ, `media.go` phục vụ mp3/bìa |
| `desktop/internal/library/` | Thư viện `~/Sano/Sach`: đọc/ghi metadata, gói zip, nhập gói, sửa sách, từ điển |
| `desktop/internal/setup/`, `tts/` | Cài bộ đọc VieNeu (uv + Python + mô hình), kiểm tra, gỡ |
| `desktop/frontend/src/` | Giao diện. `wireframes/` là wireframe đã duyệt (D1–D12), mở bằng `?wireframe=<tên>` ở bản dev |
| `scripts/tts/` | Script Python gọi VieNeu (`audio_gen_batch.py`), khoá phiên bản |
| `scripts/release/` | Build, đóng gói, ký. Phát hành = push tag `vX.Y.Z` → CI tạo **release nháp** → công khai tay |
| `docs/` | Trang hướng dẫn (VitePress, GitHub Pages) |

## 2. Chạy, test, phát hành

- Dev: `make desktop-dev` (wails dev, Vite :5390, trình duyệt gọi được Go ở :34115). Thử không đụng thư viện thật: `SANO_HOME=/thu/muc/tam wails dev`.
- Test: `go vet ./... && go test ./... -race` ở gốc **và** trong `desktop/`; frontend `cd desktop/frontend && npm run build`.
- QA giao diện: `agent-browser` CLI (xem quy tắc chung). Lưu ý: `agent-browser fill` không kích sự kiện input của Vue → đặt `value` + `dispatchEvent(new Event('input'))` bằng `eval`.
- Build RC tại máy: `scripts/release/build.sh 0.1.18-rc.1 darwin/arm64` → `desktop/build/bin/Sano.app` (wails dev dùng chung thư mục này, `-clean` xoá nó).
- Phát hành: cập nhật `VERSION`, `CHANGELOG.md` (ngắn, hộp cập nhật của bản cũ hiện đoạn này), link tải trong `README.md` → commit `chore: phát hành vX.Y.Z` → push `main` → `git tag -a vX.Y.Z` + push tag → chờ CI → `gh release edit vX.Y.Z --draft=false --latest`.
- Phiên Claude trên web (cloud): đẩy được nhánh và `main`, **không đẩy được tag** (proxy ngắt kết nối), không công khai được release (công cụ GitHub chỉ đọc). Anh tự đẩy tag từ `~/Claude-Code/sano-sach-noi-public` (thư mục `sano-sach-noi` cạnh đó là repo cũ `-archive`). Công khai trên điện thoại: mở bằng trình duyệt, **Edit bản nháp "Sano X.Y.Z" do CI tạo** rồi Publish — đừng tạo release mới từ tag (05/10 đã lỡ tạo một bản rỗng, không có file, phải xoá).
- VirusTotal luôn báo 2 phần mềm nhầm bản Windows (đã có từ 0.1.16, file Go chưa ký), không chặn phát hành.

## 3d. Phiên 05/10 (Claude web) — 0.1.23: nạp EPUB + PDF

Đã phát hành **0.1.23** (gộp bản nháp 0.1.22 chỉ có EPUB, đã xoá nháp; tag `v0.1.22` còn, bỏ qua). PR đã gộp:
tanviet12/sano-sach-noi#33 (EPUB), tanviet12/sano-sach-noi#34 (PDF). Issue #28 xong cả EPUB lẫn PDF có chữ (chưa OCR).

- **EPUB** (`internal/bookmaker/epub.go`): đọc theo spine OPF, h1–h6 → Heading; trang không có h lấy tên từ nav/NCX
  (cả neo `#id`); bỏ bìa / trang chỉ ảnh / nav / chú thích (`epub:type` footnote, noteref) / ruby; h trùng `dc:title`
  → tên sách; DRM (`encryption.xml` mã hoá nội dung, `rights.xml`, `license.lcpl`) → `ErrProtectedFile`. Đã chạy
  45 sách mẫu IDPF epub3-samples. Còn lẻ: phần đầu sách gom thành chương "Nội dung" (bỏ tick ở Mục lục), EPUB dàn
  trang cố định dính chữ giữa dòng, thơ có số dòng đọc cả số.
- **PDF** (`pdfium.go` + `pdf.go`): PDFium WASM qua go-pdfium + wazero (Go thuần, không CGO, không Python; module
  không gắn thư mục). Bản dịch máy cache ở `UserCacheDir/Sano/pdfium` (~21 MB; lần đầu ~4 giây, sau ~0,1 giây).
  Chương/mục từ bookmark, không có thì theo cỡ chữ (cỡ lớn nhất chỉ 1 đoạn ở 2 trang đầu → tên sách). Bỏ đầu/chân
  trang theo vị trí (2 dòng trên/dưới cùng, trong 12% mép, chữ nhỏ, tách khỏi thân bài, lặp ≥ 30% trang) + số trang.
  Cảnh báo `LoadWarnings.notes` (`severe` tô đỏ): thiếu bookmark, lỗi phông (≥1% / ≥10% nặng, đếm ký tự Latin-1 mà
  TCVN3/VNI dùng), trang chỉ ảnh, nhiều cột. Từ chối: không có chữ, lỗi phông ≥ 60%, mật khẩu, cấm trích chữ (cho
  trích hỗ trợ tiếp cận thì đọc). Giới hạn 256 MB / 5000 trang, cache 3 file, PDFium lỗi thì tự lấy phiên mới.
  Fixture test `internal/bookmaker/testdata/pdf/` (tạo bằng Chromium + pypdf). Bản cài nặng thêm ~4 MB/hệ (~9 MB mac universal).
- App: `bookmaker.IsSourceFile` (.docx/.epub/.pdf) dùng ở hộp chọn file, kéo thả, CLI; bước 2 đổi tên "Nạp file".
- **Chưa thử với PDF / EPUB tiếng Việt thật** (cloud chặn tải sách). Anh hứa gửi PDF test → chạy lại, chỉnh ngưỡng
  lọc đầu trang / đoán tiêu đề nếu cần.
- Issue mới: #32 clone giọng (để dành).
- **Bản phát hành kế tiếp (anh Việt chốt 05/10):** hộp cập nhật trong app chỉ hiện các dòng `- ` của ghi chú bản
  mới nhất (tối đa 8, `desktop/update.go` `releaseNotes`). Nên ở mục CHANGELOG của bản sau: tính năng mới của bản đó
  trước, rồi **nhắc lại 2 dòng EPUB, PDF** của 0.1.23, cuối cùng dòng sửa chữ "ghi đủ Word, EPUB, PDF trong app"
  (PR #35, chưa phát hành riêng). Không phát hành bản chỉ sửa chữ, không build lại trùng số 0.1.23.

## 3c. Phiên 30/09 — 0.1.21: tìm trong mục lục + Kết nối AI (MCP)

Đã phát hành 0.1.21. RC cuối trên máy anh: `desktop/build/bin/Sano.app` = 0.1.21-rc.5 (bản rc sau đánh 0.1.22-rc.N).
VirusTotal lần này: bộ cài Windows 3 phần mềm báo nhầm, bản zip 2 (trước là 2; có thêm file Go chưa ký sano-mcp.exe).

- **Tìm trong Mục lục** màn nghe (`PlayerView.vue`, `lib/find.ts` `tocMatches` / `markParts`): bỏ dấu, khớp đầu từ, số khớp trọn, ⌘F.
- **MCP** (kế hoạch + quyết định: memory `sano-mcp-ke-hoach`; hướng dẫn người dùng `docs/ket-noi-ai.md`):
  - Lõi `desktop/mcp.go`: máy chủ MCP trong app (go-sdk v1.8.0), 127.0.0.1:39390 (bận thì cổng bất kỳ), stateless + JSON,
    mã bí mật `~/Sano/.mcp/token` (0600), chặn Origin. `addTool` = kiểm quyền sửa + ghi nhật ký.
  - Cầu nối `desktop/internal/mcpbridge` (stdio → HTTP, gửi tên phần mềm AI qua `X-Sano-Client`): file `sano-mcp` cạnh
    app (Mac trong Sano.app, Windows cạnh Sano.exe) hoặc `Sano mcp` (Linux AppImage). Hẹn: `internal/mcplink`.
  - Màn MCP (D21, `views/McpView.vue`, `mcp_settings.go`): bật/tắt trong máy, quyền Tạo và sửa, nhật ký 7 ngày
    (`~/Sano/.mcp/log.jsonl`), Thêm vào Claude Desktop (sửa `claude_desktop_config.json`, sao lưu `.bak-sano`).
    Kết nối từ xa hiện "Sắp có".
  - Tạo sách (D22, `mcp_create.go` bản nháp, `mcp_queue.go`): tạo trước, cam kết sau. Hàng đợi, khu chờ
    `~/Sano/.cho-cam-ket/<id>/sach` (7 ngày), `McpPledgeDialog.vue` cam kết một lần nhiều cuốn, ghi `rights_confirmed_at`.
  - Sửa sách (M3, `mcp_edit.go`): update_sections, find_replace, update_book_info, set_cover (base64), change_voice,
    set_pronunciation, get_edit_status. Bản cũ `library/snapshot.go` → `~/Sano/.ban-cu/<slug>/<id>` (hardlink mp3/zip/ảnh,
    3 bản, 7 ngày), hoàn tác ở nhật ký (chỉ lần sửa mới nhất mỗi cuốn). **Không có công cụ xoá** (test chặn).
  - Đóng gói (M4): `build.sh` build sano-mcp (mac lipo, Windows trước wails build để NSIS kèm), ký lại app mac;
    NSIS đổi tên sano-mcp.exe đang bị khoá (.cu-*); bản zip + tự cập nhật portable kèm cầu nối. CI Windows kiểm.
  - Điều khoản mục 1 sửa câu cam kết (trước khi lưu vào Thư viện; sách AI ở khu chờ). Giữ Phiên bản 2. Anh nhờ luật sư xem.
- **Còn (anh chốt để giai đoạn sau):** kết nối từ xa (trạm Cloudflare trên sanobook.com, mã ghép 40 ký tự 2 phút, OAuth), chưa làm. Bộ cài Windows
  chưa chạy thử trên máy thật (chỉ CI). Mở app còn sách ở khu chờ thì popup tự bật nhắc (anh chưa phản hồi).
- Test dev không đụng thư viện thật: worktree scratchpad + `SANO_HOME=<tạm>`; cổng 39390 bận (RC anh đang mở) thì
  bản dev lấy cổng khác, cầu nối đọc cổng từ `<SANO_HOME>/.mcp/local.json`.

## 3b. Phiên 29/09 — phát hành 0.1.19 và 0.1.20

Cả hai bản: test Go gốc + desktop và build frontend qua, CI xanh (build 3 hệ, smoke). RC cuối trên máy anh Việt:
`desktop/build/bin/Sano.app` = 0.1.20-rc.2 (bản rc sau nên đánh 0.1.21-rc.N).

**0.1.19**
- Khung dọc 9:16 chừa vùng Reels / TikTok / Story che: video cả cuốn (`bookVideoCard.ts`, `TALL_SAFE`:
  14% đầu, 28% đáy, lề phải 64/360), video ngắn (`shareCard.ts` `drawReel`, `REEL_SAFE`: logo lên đầu,
  bìa nhỏ + sóng âm + câu đọc), ảnh Story (14% đầu, 22% đáy). Khung ngang, vuông giữ nguyên.
- **D19** popup cam kết (`RenderPledgeDialog.vue`, `lib/pledge.ts`, `?wireframe=pledge`): 6 ô tick bắt buộc trước
  khi render (thay ô tick ở thanh dưới bước Nghe thử) và trước khi tạo video ngắn / cả cuốn (một lần mỗi cuốn trong
  một lần mở app, chưa lưu lúc cam kết của video). Popup dùng `flex` + `max-h-full` để không tràn khung (grid thì tràn).
- Điều khoản **phiên bản 2** (`docs/dieu-khoan-su-dung.md`): nội dung vi phạm pháp luật, video/ảnh chia sẻ, mục Bồi hoàn.

**0.1.20**
- **D20** Tìm và thay trong lời đọc cả cuốn (Sửa sách · Nội dung, nút hoặc ⌘F; `lib/edit.ts` `textHits` /
  `replaceInText` / `undoReplace`, `?wireframe=findreplace`): khớp nguyên cụm, phân biệt hoa thường; tick "Thay cả
  trong tiêu đề mục". Thay xong là bản sửa chưa lưu → nút có sẵn "Lưu & đọc lại N mục" chỉ đọc lại các mục đó.
  Tô chỗ khớp bằng lớp div nằm dưới textarea trong suốt (đồng bộ cuộn). Tên chương không thay được bằng cách này.

**Còn / ý tưởng**
- Anh nhờ luật sư xem mục Bồi hoàn (có nhắc nhà tài trợ SePay, 123HOST).
- Chia sẻ ảnh chưa có popup cam kết (anh chưa yêu cầu).
- Cáp quốc tế chậm thì tải cập nhật từ GitHub ~20 KB/giây. Đề xuất (chưa làm): bản sao file cập nhật trên
  sanobook.com (Cloudflare), app ưu tiên tải từ đó, GitHub dự phòng.
- Test dev không đụng thư viện thật: worktree trong scratchpad + symlink node_modules + `SANO_HOME=<tạm>` chứa bản
  sao một cuốn, `wails dev -browser=false`, agent-browser vào :34115; xong kill đúng PID rồi `git worktree remove`.

## 3a. Phiên tối 28/09 — D17, D18, phát hành 0.1.18

**Đã phát hành 0.1.18:** test Go gốc + desktop và build frontend qua; CI xanh (build 3 hệ, VirusTotal, smoke);
đã tải dmg kiểm SHA-256, phiên bản, universal. VirusTotal: Windows 2/75 báo nhầm như cũ, Mac/Linux 0/75.

- **D17** (`videohub`): hàng nút màn nghe; "Tạo video" một hộp, đầu hộp `components/VideoKindTabs.vue` chọn Video ngắn
  (ShareDialog phần video) / Video cả cuốn (BookVideoDialog), cùng cỡ 1060×690, nhớ loại (`lib/video.ts`, `sano.video.kind`).
  Hộp Chia sẻ chỉ còn ảnh. Bánh răng bỏ "Tạo video cả cuốn…".
- **D18 A+B** (`sharecta`): nút "Chia sẻ câu hay · ảnh" (Hoàng hôn) + "Tạo video · Reels, YouTube" (tím xanh), hẹp thì chữ phụ
  xuống dòng trong nút, nút điện thoại chỉ còn icon (đo `actionsEl`). Dưới câu đang đọc (LyricsStage) "Chia sẻ câu này" ·
  "Video 15 giây", chỉ hiện khi rê chuột.
- Phóng to lời đọc là mặc định (`sano.lyricsFrame.v2`), bỏ hẳn "Xem cả lời" (`LyricsPanel.vue` đã xoá), Thu lại = icon nhỏ.
  Lời đọc: câu đang đọc 21px đậm, câu khác 17px. Khung Lời đọc ở màn thường tô màu Hoàng hôn nhạt.
- Chia sẻ ảnh cuộn danh sách tới câu đang đọc (`centerPicked`).
- **Bẫy đã gặp:** `wails dev` ghi đè `desktop/build/bin/Sano.app` và tắt dev thì xoá file chạy → test dev bằng worktree
  riêng trong scratchpad (symlink node_modules) rồi `git worktree remove`. `/Applications/Sano.app` cùng bundle id:
  đang chạy thì `open` RC chỉ đưa app cũ lên trước — phải ⌘Q trước.

## 3. Phiên 28/09 — các tính năng của 0.1.18 (đã phát hành)

RC mới nhất đang chạy trên máy anh Việt: `desktop/build/bin/Sano.app` = **0.1.18-rc.15** (build bằng
`scripts/release/build.sh 0.1.18-rc.N darwin/arm64`, mỗi lần ghi đè cùng chỗ). Đã có `CHANGELOG` 0.1.18 chưa viết.

### Đã làm (theo thứ tự, wireframe đã duyệt ở `desktop/frontend/src/wireframes/`, mở `?wireframe=<tên>`)

- **Sửa lỗi sáng 28/09:** mục lục hiện tên chương (`0eed810`); Thiền Tâm Đức đọc đúng "chánh" (`3dfdc6d`, bảng
  `PHONEME_OVERRIDES` theo giọng trong `scripts/tts/audio_gen.py`). Sách "Chánh niệm" đã đọc lại 21/39 mục.
- **D13 Phóng to lời đọc trong khung** (`lyricsframe`): khung Lời đọc → "Phóng to": hàng trên bìa nhỏ + tên sách,
  lời đọc to đậm (`components/LyricsStage.vue`), nền phủ nhẹ màu bìa, logo "Sano · Sách nói" giữa thanh trên.
  Mục lục: thanh tiến độ cả cuốn, tên tiểu mục 2 dòng, đầu mục lục đứng yên. Nhớ chế độ (`sano.lyricsFrame`).
- **D14 Chia sẻ đoạn hay** (`share`): nút Chia sẻ ở Phóng to / Xem cả lời → `components/ShareDialog.vue`.
  Ảnh có lời (≤4 câu) / video có tiếng đọc (≤8 câu, ≤30 giây, mở ra tự chọn ~15 giây). Mặc định Dọc 9:16 +
  nền Hoàng hôn. Thẻ vẽ bằng canvas `lib/shareCard.ts` (xem trước = file xuất, 1080px), chân thẻ
  "Sano · Tự tạo sách nói · sanobook.com". Video: `desktop/share.go` (ffmpeg, cột sóng âm sáng dần bằng mặt nạ
  trượt). Sao chép ảnh vào clipboard Mac `clipboard_darwin.go` (máy khác dùng API trình duyệt). Lưu / AirDrop.
- **D15 Tạo video cả cuốn** (`bookvideo`): bánh răng màn nghe → "Tạo video cả cuốn…" → `components/BookVideoDialog.vue`.
  Trích đoạn hay nhất (tuỳ chọn) → màn tựa → thẻ chương → câu đang đọc chữ to + câu kế nhạt + tiến độ có vạch
  chương → màn kết sanobook.com. Ngang 16:9 1920×1080 (mặc định) / dọc; cả cuốn hoặc vài chương ("NGHE THỬ",
  thư mục riêng). Kèm thumbnail 1280×720, `phu-de.srt`, `mo-ta-youtube.txt` (mốc chương). Lưu vào
  `Tải về/Sano video/<Tên>`. Dòng thời gian `lib/bookVideo.ts` (`buildTimeline`, dùng chung cho tạo + nghe thử),
  vẽ cảnh `lib/bookVideoCard.ts`, Go `desktop/bookvideo.go` (mỗi đoạn tiếng → WAV đúng số giây rồi nối; khung
  hình JPEG gửi dần; mặt nạ trượt cho sóng âm / tiến độ). Đo: cuốn 29 phút → 64 MB, 3 phút tạo.
- **Nghe thử trong hộp chọn câu** (`lib/previewAudio.ts`): ▶ từng câu nghe tiếp, "Nghe đoạn". **Nghe thử như video**
  ngay trong hộp D15 (`lib/videoPreview.ts`): phát nối liền cả dòng thời gian, ô xem trước vẽ theo giờ, tua được.
- **D16 Mở đầu kiểu Fonos** (`videointro`): "Bạn đang nghe sách nói, tạo bằng Sano" → nhạc hiệu → giới thiệu sách
  (tên, tác giả, dịch giả, NXB) → nội dung; màn kết lời kết + nhạc; chuông sang chương (mặc định TẮT). Bật lời
  giới thiệu thì bỏ tiểu mục mở đầu tự có ("Cuốn sách: …"). Câu đọc bằng bộ đọc + giọng của cuốn,
  nhạc hiệu Sano tạm tự tổng hợp (aevalsrc, không bản quyền, nhỏ hơn 20% = −2 dB) hoặc file người dùng:
  `desktop/videoextras.go` (đệm `~/Sano/.tam/video-them`, mã → file, không nhận đường dẫn từ giao diện).
  Sửa sách → Thông tin & bìa có thêm **Dịch giả, Nhà xuất bản** (`library.Book/Info`, metadata + manifest zip).
- **Sửa giờ câu (quan trọng, ảnh hưởng cả màn nghe):** `lib/lyrics.ts` `alignSentences` ghép câu chữ gốc với câu lời
  đọc theo nội dung — trước đây tiểu mục có dấu ":" (0.1.17 đọc thành ngắt câu) lệch cả tiểu mục một câu
  (6/39 mục cuốn Chánh niệm). `sentenceBounds` cắt trích đoạn ở giữa khoảng lặng thật (hết dính chữ câu kế).

### Việc dở dang — phiên sau làm tiếp

1. **Chờ anh Việt chọn cách đọc tên** (đã gửi mp3): "Sano" / "Sa-nô" (đang dùng) / "Xa Nô"; "sanobook.com" /
   "Sa-nô búc chấm com" (đang dùng). Sửa ở `BookVideoDialog.vue` hằng `BRAND.say`, `END_LINE.say`.
2. **Âm lượng nhạc hiệu:** đang −2 dB (−20% biên độ, nghe nhỏ ~13%). Anh hỏi lại "đã giảm 20% chưa" — nếu muốn
   nghe rõ nhỏ 20% thì −3,2 dB (`volume=0.69` trong `videoextras.go`, đổi mã `sano-v3` để tạo lại). Chờ anh nghe rc.15.
3. **Nhạc hiệu chính thức:** anh Việt sẽ đặt làm / mua; khi có thì nhúng file thay bản tổng hợp tạm.
4. **Tên miền sanobook.com: XONG 28/09.** DNS ở luutruso.vn (4 A + CNAME www + TXT xác minh GitHub), repo gắn
   custom domain, base VitePress `/`, link đổi hết, homepage repo + ghi chú release cũ đã sửa. Link cũ 301 sang tên miền mới.
   Còn: anh tích Enforce HTTPS; Search Console thêm property Domain `sanobook.com` + nộp sitemap; giữ property cũ 3–6 tháng.
5. **Chưa kiểm tự động được (cửa sổ hệ điều hành):** Lưu ảnh / Lưu video / AirDrop / Mở thư mục — nhờ anh bấm thử.
   Sao chép ảnh trên Windows / Linux (API trình duyệt) chưa thử máy thật.
6. ~~Phát hành 0.1.18~~ — XONG 28/09 (mục 3a).
7. ~~Nút "Tạo video" cạnh "Nghe trên điện thoại"~~ — XONG (D17, D18).
8. Ghi chú kỹ thuật: bước vẽ khung hình video chạy ở giao diện (đóng hộp vẫn chạy, tắt app thì dừng lượt);
   trong `Tải về/Sano video/` còn video test của phiên này.

### QA / test phiên này

- Test bằng `agent-browser` trên `wails dev` với thư viện tạm: `SANO_HOME=<scratch>/sanohome wails dev` (đã chép
  cuốn Chánh niệm sang). Tắt dev đúng tiến trình: `kill` con của `pgrep -f "wails dev"` — **không** `pkill` theo tên
  Sano.app (dev và RC cùng đường dẫn, từng lỡ tắt app của anh).
- Âm thanh: đo bằng `ffmpeg silencedetect/volumedetect`, không tin Whisper cho ch/tr; kết luận âm gửi mp3 anh nghe.

## 4. Đã làm trong phiên 27/09 (0.1.16 → 0.1.17)

- Dấu ":" giữa câu đọc thành ngắt câu (VieNeu chỉ nghỉ ~0,25 s ở ":"), giữ 10:30, 3:1, link.
- Phím tắt khi nghe: Space dừng / nghe tiếp, ← → tiểu mục trước / sau (bỏ qua khi gõ chữ, có hộp thoại, ở Tạo sách / Sửa sách).
- **Sửa sách (D11)**: menu ⋯ ở Thư viện "Sửa sách", bánh răng màn nghe "Sửa mục đang nghe". 4 tab:
  Nội dung (sửa chữ từng mục → Lưu & đọc lại, chấm vàng = chưa lưu, bản sửa giữ trong localStorage),
  Thông tin & bìa (đổi tên → vẽ lại bìa tự vẽ + đọc lại lời giới thiệu), Giọng đọc (đổi giọng cả cuốn,
  đọc vào `.doc-lai/` trong thư mục sách, xong hết mới thay, tắt app thì mở lại tự đọc tiếp), Từ điển.
  Hộp "Sửa thông tin" cũ đã bỏ.
- **Từ điển cách đọc (D12)**: bộ chuẩn 65 từ (nhúng) → từ điển chung `~/Sano/.tu-dien.tsv` → từ điển
  của cuốn `tu-dien.tsv` (trong zip là `pronunciations.tsv`, tuỳ chọn). Dùng ở Nạp file (bấm từ viết tắt
  bị cảnh báo), Nghe thử (bôi đen → "Đọc từ này là…", khung từ điển bên phải), Sửa sách (tab Từ điển,
  đổi cách đọc → đánh dấu mục cần đọc lại), Cài đặt (Từ điển chung, ghi đè từ có sẵn).
- Nghe thử: đổi giọng / tên / lời mở đầu không còn xoá chỗ đã sửa; chỗ sửa không khớp nữa thì báo.
- Ô xác nhận quyền dùng tài liệu nằm cố định trên thanh dưới cạnh nút render.
- Bìa: kệ sách dùng bản thu nhỏ 480px (đệm `~/Sano/.tam/bia`), ảnh lỗi thì thử lại rồi hiện bìa tự vẽ;
  ảnh bìa tự chọn lớn hơn 1200×1600 thì thu nhỏ giữ tỉ lệ.

## 5. Việc còn tồn

- MCP chưa nhận file EPUB/PDF (`create_book` chỉ nhận văn bản, cố ý không nhận đường dẫn). Hướng đã bàn 05/10:
  công cụ mới `create_book_from_file` → Sano mở hộp chọn file, người dùng tự chọn, AI không thấy đường dẫn. Để sau.
- Đề xuất (chưa làm, chờ anh gật): workflow `workflow_dispatch` "Công khai bản phát hành" (gắn tag + publish nháp +
  xoá nháp cũ) để phiên Claude cloud tự phát hành khi anh bảo, khỏi đẩy tag / bấm tay.
- PDF scan cần OCR: chưa làm (#28 ghi giai đoạn sau).

- App chưa có nút "Đọc lại cả cuốn" (khi script đọc / bộ đọc đổi, sách cũ không tự đọc lại; hiện phải mượn từ điển
  của cuốn để đánh dấu mục, hoặc chạy tay như mục 3). Tính năng mới → cần wireframe (tab Giọng đọc trong Sửa sách).

- 11 nhánh cũ trên máy + remote (`feat/cach-doc-3-cap`, `feat/thu-vien-bo-sach`, `release/0.1.3`, `release/0.1.4`,
  `fix/linux-audio`…): nhiều nhánh đã gộp vào `main`, cần kiểm rồi mới xoá.
- 3 nhánh Dependabot trên remote (TypeScript 7.0.2, Vite 8.3.0, vue-tsc 3.3.11) chưa gộp.
- Thanh "Đọc từ này là…" ở Nghe thử đang nằm góc trên phải ô chữ, chưa bám sát vị trí từ bôi đen như wireframe D12.
- Từ điển chỉ nhận một từ liền (chữ, số, &), chưa nhận cụm nhiều từ ("Hồ Chí Minh").
- Đổi Từ điển chung không đánh dấu mục cần đọc lại ở sách cũ (phải thêm từ vào từ điển của cuốn).
- Trang hướng dẫn (`docs/`) chưa có bài riêng cho Sửa sách và Từ điển cách đọc.
- Khi app bị tắt cứng (không qua hộp hỏi), tiến trình Python của bộ đọc có thể còn chạy tới hết lượt (có từ trước, cả Tạo sách).

## 6. Ý tưởng chưa làm: giọng đọc tiếng Anh, tiếng Nhật (anh Việt: "chưa cần làm", 27/09)

Nhiều người hỏi. Đã phân tích, **chưa làm**. Khi làm lại thì đọc đoạn này trước.

**Hướng đã thống nhất:** mỗi ngôn ngữ là một **gói bộ đọc cài riêng**. Lần đầu mở Sano hỏi làm sách tiếng nào
(mặc định tiếng Việt → VieNeu như hiện nay); thêm / gỡ tiếng khác ở Cài đặt → Bộ đọc → Thêm ngôn ngữ; khi tạo sách
chọn ngôn ngữ (chỉ hiện tiếng đã cài); mỗi cuốn nhớ ngôn ngữ (`metadata.json` đã có `language: "vi"`).
Bộ đọc gợi ý: **Kokoro-82M** (Apache 2.0, chạy trên máy) — một mô hình dùng chung nhiều tiếng, mỗi tiếng thêm giọng
+ phần riêng (tiếng Nhật cần từ điển Kanji nặng). Số liệu Kokoro chưa kiểm trên máy thật.

**Mức độ dính VieNeu / tiếng Việt trong code:**

| Phần | Việc phải làm | Rủi ro |
|---|---|---|
| Cài bộ đọc (`desktop/internal/setup`, ~3.400 dòng, 1 venv VieNeu, khoá phiên bản, SHA-256 mô hình) | Thành "gói theo ngôn ngữ", mỗi bộ đọc một môi trường Python riêng | **Cao** — 3 hệ điều hành, Mac Intel từng lỗi onnxruntime; hỏng thì người dùng tiếng Việt mất bộ đọc |
| Chuẩn hoá lời đọc (`normalize.go`, `spoken.go`, từ điển) | Bộ chuẩn hoá theo ngôn ngữ | Trung bình |
| Đọc docx (nhận "Mục lục", cảnh báo viết tắt, lời mở đầu tiếng Việt) | Theo ngôn ngữ của sách | Thấp |
| Giọng + giao diện (tab miền Bắc/Trung/Nam, giọng khuyên dùng, "VieNeu-TTS (giọng)" viết cứng) | Theo ngôn ngữ; màn cài lần đầu; chọn ngôn ngữ khi tạo sách (cần wireframe) | Trung bình |
| Chữ chạy theo (`lyrics.ts`) | Chỉnh nhịp ước lượng | Thấp |

Thuận lợi: Go gọi bộ đọc qua script Python theo giao ước đơn giản (file chữ vào → wav ra, in `✨ <stem>_full.wav`).
Kokoro chỉ cần `kokoro_batch.py` cùng giao ước; tạo sách, Sửa sách, M4B, zip, trình phát không đổi.

**Ước lượng (AI làm):** tiếng Anh khoảng 1 ngày viết code, 1–2 ngày thật (chờ CI 15–30 phút/lượt + duyệt wireframe).
Tiếng Nhật thêm 2–3 ngày và không ai trong nhóm kiểm được chất lượng.

**Rủi ro tốc độ AI không giảm được:** không test tay được trên Windows / Mac Intel / Linux (chỉ CI build + smoke);
chất lượng giọng phải người nghe; mỗi bộ đọc thêm là thêm bảo trì lâu dài (khoá phiên bản, vá bảo mật, lỗi từng HĐH).

**Khi làm:** (1) chạy thử Kokoro tiếng Anh trong thư mục tạm, anh nghe + đo dung lượng, tốc độ; (2) wireframe 3 chỗ
(chọn ngôn ngữ lúc cài, Cài đặt → Thêm ngôn ngữ, chọn ngôn ngữ khi tạo sách); (3) không đụng đường cài VieNeu,
gói mới là phần cài thêm tuỳ chọn; (4) cho vài người dùng thật trên Windows + Mac Intel thử trước khi phát hành;
(5) tiếng Anh trước, tiếng Nhật sau; không chạy theo "nhiều model".
