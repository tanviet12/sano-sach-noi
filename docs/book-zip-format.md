# Format gói sách nói chuẩn `book-<slug>.zip`

**Mục đích:** Định nghĩa duy nhất (source of truth) cấu trúc gói zip một sách nói.

Gói zip dùng để **sao lưu hoặc chuyển sách sang máy khác**: một file chứa đủ audio, bìa, lời gốc và lời đọc của cả cuốn. Sinh ra bởi:

- Phần mềm: nút **Xuất gói zip** ở trình phát hoặc khi render xong (lưu `book-<slug>.zip` trong thư mục sách).
- CLI: `cmd/sano-docx2tts --output-zip` (hoặc `--repack-dir` để đóng gói lại từ thư mục đã render).

Định dạng này **giữ ổn định** để máy chủ nghe sách ở giai đoạn sau nhận thẳng gói zip. Đổi spec = tăng `version` và giữ đọc được bản cũ.

**Input format:** `cmd/sano-docx2tts` chỉ nhận `.docx`. File `.pdf`: chạy `scripts/pdf_to_docx.py` trước (trích outline/bookmark PDF → `.docx` tạm có Heading 1/2 chuẩn) rồi tạo sách như thường.

## 1. Cấu trúc thư mục

```
book-<slug>.zip
├── manifest.json
├── chapters.json
├── cover.jpg            (hoặc cover.png)
└── audio/
    ├── ch01/
    │   ├── sec01.mp3
    │   └── sec02.mp3
    └── ch02/
        └── sec01.mp3
```

- `manifest.json` — metadata sách.
- `chapters.json` — mục lục nhiều cấp + nội dung từng tiểu mục.
- `cover.jpg` / `cover.png` — ảnh bìa (tên khớp `manifest.cover_filename`).
- `audio/` — file audio, gom theo chương.

## 2. manifest.json

```json
{
  "title": "Cẩm nang quản trị bản thân",
  "slug": "cam-nang-quan-tri-ban-than",
  "author": "Nguyễn Văn A",
  "description": "Tổng hợp bài giảng về quản trị bản thân.",
  "category_slug": "quan-tri",
  "category": "Quản trị",
  "tags": ["mindset", "ky-luat"],
  "cover_filename": "cover.jpg",
  "voice_id": "Trúc Ly",
  "version": 1,
  "build_timestamp": "2026-06-07T10:30:00+07:00"
}
```

| Field | Kiểu | Bắt buộc | Mô tả |
|---|---|---|---|
| `title` | string | **bắt buộc** | Tiêu đề sách |
| `slug` | string | **bắt buộc** | Định danh URL-safe kebab-case, khớp `<slug>` trong tên file zip |
| `author` | string | tùy chọn | Tác giả |
| `description` | string | tùy chọn | Mô tả ngắn |
| `category_slug` | string | tùy chọn | Slug danh mục (map sang `categories`); không khớp → bỏ qua, không fail. Phần mềm tạo sách suy ra từ `category` (gấp dấu, kebab-case: "Kỹ năng" → `ky-nang`) |
| `category` | string | tùy chọn | Tên danh mục hiển thị người dùng đặt khi tạo sách, vd `Kỹ năng`. Nơi nhận gói zip ưu tiên `category_slug` |
| `tags` | string[] | tùy chọn | Danh sách slug tag |
| `cover_filename` | string | tùy chọn | Tên file ảnh bìa trong zip, vd `cover.jpg` |
| `voice_id` | string | tùy chọn | Preset giọng VieNeu-TTS dùng khi render, vd `Trúc Ly` |
| `version` | int | **bắt buộc** | Phiên bản schema = `1` |
| `build_timestamp` | string | tùy chọn | Thời điểm build, định dạng RFC3339 |

## 3. chapters.json

Mục lục nhiều cấp: sách → chương (chapter) → tiểu mục (section). Section là đơn vị phát.

```json
{
  "chapters": [
    {
      "order": 1,
      "title": "Chương 1: Nền tảng",
      "sections": [
        {
          "order": 1,
          "title": "1.1 Tư duy gốc",
          "audio_filename": "audio/ch01/sec01.mp3",
          "duration_sec": 123,
          "original_text": "Lời gốc của sách, giữ nguyên không sửa...",
          "reading_script": "Bản đã chuẩn hóa để đọc TTS...",
          "image_description": "Slide minh họa sơ đồ tư duy gốc (tùy chọn)"
        }
      ]
    }
  ]
}
```

### Chapter

| Field | Kiểu | Bắt buộc | Mô tả |
|---|---|---|---|
| `order` | int | **bắt buộc** | Thứ tự chương, bắt đầu từ 1 |
| `title` | string | **bắt buộc** | Tiêu đề chương |
| `sections` | Section[] | **bắt buộc** | Danh sách tiểu mục (≥ 1) |

### Section

| Field | Kiểu | Bắt buộc | Mô tả |
|---|---|---|---|
| `order` | int | **bắt buộc** | Thứ tự tiểu mục trong chương, từ 1 |
| `title` | string | **bắt buộc** | Tiêu đề tiểu mục |
| `audio_filename` | string | **bắt buộc** | Đường dẫn tương đối trong zip, vd `audio/ch01/sec01.mp3` |
| `duration_sec` | int | **bắt buộc** | Thời lượng audio (giây), phục vụ progress + tổng thời lượng sách |
| `original_text` | string | **bắt buộc** | Lời gốc của sách, KHÔNG sửa |
| `reading_script` | string | **bắt buộc** | Bản chuẩn hóa đã dùng để render TTS |
| `image_description` | string | tùy chọn | Mô tả ảnh/slide cho người nghe không nhìn màn hình |

## 4. Quy ước đặt tên

- Thư mục chương: `chNN` — `NN` là `order` của chương, zero-pad 2 chữ số (`ch01`, `ch02`, ... `ch10`).
- File tiểu mục: `secNN.mp3` — `NN` là `order` của tiểu mục trong chương, zero-pad 2 chữ số (`sec01.mp3`).
- `audio_filename` trong `chapters.json` là đường dẫn tương đối tính từ gốc zip (gồm tiền tố `audio/`).
- File audio bắt buộc **MP3 CBR** (đếm frame được để tính thời lượng, tua giữa file được).
- `slug` khớp regex kebab-case: `^[a-z0-9]+(?:-[a-z0-9]+)*$`.

## 5. Tại sao format này

- **`original_text` + `reading_script` song song** — giữ lời gốc của sách, kịch bản đọc TTS để riêng để soát lại đối chiếu. Không tự viết lại nội dung chính.
- **`image_description`** — nguồn liệu là slide/cẩm nang có hình; người nghe khi lái xe không nhìn được màn hình, mô tả ảnh giúp không mất thông tin.
- **Tách `manifest` (metadata) khỏi `chapters` (nội dung)** — nơi nhận validate nhanh metadata trước khi đọc toàn bộ nội dung nặng.
- **Một file tự đủ** — render TTS chạy trên máy người làm sách; gói zip mang đủ mọi thứ để mở lại ở máy khác mà không cần render lại.

## 6. Validation checklist

`cmd/sano-docx2tts` kiểm trước khi ghi zip; nơi nhận gói zip cũng phải kiểm TẤT CẢ. Fail bất kỳ mục nào → từ chối toàn bộ, không nhận một phần.

- [ ] Zip giải nén được, có `manifest.json` + `chapters.json` ở gốc.
- [ ] `manifest.title` không rỗng.
- [ ] `manifest.slug` không rỗng và khớp regex `^[a-z0-9]+(?:-[a-z0-9]+)*$`.
- [ ] `manifest.version == 1`.
- [ ] `chapters` có ít nhất 1 chương, mỗi chương có ít nhất 1 tiểu mục.
- [ ] Mỗi chapter có `order` (≥1), `title` không rỗng.
- [ ] Mỗi section có `order` (≥1), `title`, `audio_filename`, `duration_sec` (≥0), `original_text`, `reading_script`.
- [ ] **Mọi `audio_filename` tồn tại thực sự trong zip** (không thiếu file audio nào).
- [ ] Nếu có `cover_filename` → file ảnh bìa tồn tại trong zip.
- [ ] Kích thước zip ≤ 500MB; ảnh bìa ≤ 5MB.
- [ ] Không có path traversal trong tên file (`..`, đường dẫn tuyệt đối).

## 7. Tham chiếu

- `internal/bookmaker/zip.go` — sinh + validate gói zip.
- `cmd/sano-docx2tts --output-zip` / `--repack-dir` — sinh gói zip từ dòng lệnh.
- [`tts-build-guide.md`](tts-build-guide.md) — tạo sách bằng dòng lệnh.
