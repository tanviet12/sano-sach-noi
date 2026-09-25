# Hướng dẫn build audiobook bằng VieNeu-TTS

Quy trình đọc giọng của Sano: text tiếng Việt → WAV (VieNeu-TTS) → MP3 → thư mục sách (`metadata.json` + MP3 + bìa) mà phần mềm mở được, xuất M4B được. Audio **không realtime** — đọc sẵn một lần lúc tạo sách.

## Model & môi trường

- **Model:** VieNeu-TTS **v3 Turbo** (on-device, tiếng Việt + song ngữ, output **48 kHz**), chạy **ONNX trên CPU** — không cần torch, không cần GPU, không cần llama-cpp. HuggingFace `pnnbao-ump/VieNeu-TTS-v3-Turbo` (thư mục `onnx_update/` + `config.json`, `denoiser.onnx`) + codec `OpenMOSS-Team/MOSS-Audio-Tokenizer-Nano-ONNX`. Cả hai Apache-2.0, **không khoá** → không cần tài khoản Hugging Face / `HF_TOKEN`. Tải ~580 MB lần đầu, sau đó chạy **offline**.
- **Vị trí:** VieNeu-TTS clone ở **`~/VieNeu-TTS-v3/`** (mã nguồn upstream, không sửa) — `sano-docx2tts` mặc định tìm python ở đây. Tách khỏi `~/VieNeu-TTS/` nếu máy còn bản v2 cũ: script v3 chạy bằng venv v2 sẽ lỗi. Script đọc giọng nằm trong repo này ở `scripts/tts/`.
- **Python:** venv `~/VieNeu-TTS-v3/.venv` (Python 3.12, tạo bằng `uv sync`). Package chính `vieneu` (`Vieneu(mode="v3turbo")`).

### Ghim phiên bản

Mọi phiên bản nằm ở **`scripts/tts/versions.env`** (một nguồn duy nhất, script + CI cùng đọc):

| Thành phần | Ghim |
|---|---|
| VieNeu-TTS | commit `VIENEU_COMMIT` (bản 3.8.3) của `https://github.com/pnnbao97/VieNeu-TTS` |
| Thư viện Python | `uv.lock` của đúng commit đó (`uv sync --frozen`); bản đọc được: `scripts/tts/requirements.txt` |
| Model HF | `HF_BACKBONE_REVISION`, `HF_CODEC_REVISION` + danh sách file — `scripts/tts/models.py` tải đúng revision và nạp offline |
| uv / Python | `UV_VERSION`, `PYTHON_VERSION` (3.12) |

### Cài (macOS / Linux / Windows)

> Người dùng phần mềm desktop **không cần làm tay**: app tự cài đúng các bước dưới đây (uv, Python, VieNeu đúng commit, `uv sync --frozen`, mô hình, ffmpeg) vào thư mục dữ liệu của app, kiểm SHA256 từng thứ — xem [`desktop/README.md`](../desktop/README.md#cài-bộ-đọc-lần-mở-đầu). Cách cài tay bên dưới dành cho máy dev / CLI `sano-docx2tts`.

```bash
# 1. uv: https://docs.astral.sh/uv/  (Windows: lệnh bên dưới chạy được trong Git Bash)
source scripts/tts/versions.env
git clone "$VIENEU_REPO" ~/VieNeu-TTS-v3
git -C ~/VieNeu-TTS-v3 checkout "$VIENEU_COMMIT"
(cd ~/VieNeu-TTS-v3 && uv sync --frozen --no-dev --python "$PYTHON_VERSION")

# 2. Tải model ghim (~580 MB, 1 lần, không cần HF_TOKEN)
PY=~/VieNeu-TTS-v3/.venv/bin/python       # Windows: ~/VieNeu-TTS-v3/.venv/Scripts/python.exe
"$PY" scripts/tts/models.py fetch        # tải xong tự kiểm SHA256 theo scripts/tts/models.sha256
"$PY" scripts/tts/models.py check        # kiểm tra đủ file
"$PY" scripts/tts/models.py verify       # kiểm lại SHA256 từng file
```

- Cả 3 hệ đều cài bằng wheel dựng sẵn — **không build gì từ mã nguồn** (trừ chính gói `vieneu`, vài giây). Trên GitHub Actions 2 lõi: Linux ~7s, Windows ~22s (khi uv đã tải thư viện về).
- `ffmpeg` cần riêng cho bước WAV → MP3.
- Kết quả chạy thử (Windows / Linux, macOS bật tay): workflow `.github/workflows/tts-smoke.yml`.

### Giọng có sẵn (25 giọng, ⭐ = upstream chọn nổi bật)

In danh sách: `PYTHONPATH=scripts/tts "$PY" scripts/tts/audio_gen_batch.py --list-voices`. Mặc định **Hải Đăng** (nam Bắc, tự nhiên).

| Nhóm | Giọng |
|---|---|
| Nữ Bắc | ⭐Trúc Ly, ⭐Mai Anh, ⭐Ngọc Huyền, Ngọc Linh, Đoan Trang, Quỳnh Anh |
| Nữ Nam | ⭐Thùy Dung, Thục Đoan, Mỹ Duyên, Kim Thanh |
| Nữ Trung | ⭐Ngọc Trân |
| Nam Bắc | ⭐Adam bựa, ⭐Hải Đăng (mặc định), ⭐Thiện Minh, ⭐Thiền Tâm Đức, Minh Đức, Phạm Tuyên, Xuân Vĩnh, Thanh Bình, Quốc Tuấn |
| Nam Nam | Thái Sơn, Minh Triết, Đức Trí, Adam |
| Nam Trung | ⭐Quang Sơn |

`--voice` nhận đúng tên (có dấu), bí danh upstream, hoặc một phần tên không phân biệt hoa thường. Không thấy giọng → báo lỗi kèm danh sách, không tự đổi giọng khác.

## Script

`scripts/tts/audio_gen.py` (1 file) và `scripts/tts/audio_gen_batch.py` (nhiều file, nạp model 1 lần — `sano-docx2tts` gọi script này). Mỗi file đưa **nguyên văn bản** vào `tts.infer(text, voice=...)`: v3 Turbo tự chia câu, tự chống lỗi "nói thêm" và tự nghỉ theo ranh giới (đoạn 0,70s > câu 0,50s > trong câu 0,30s) — xem [`docs/vieneu-tts-patch.md`](vieneu-tts-patch.md).

Output: `<stem>_full.wav` (48 kHz) đặt ở **CWD**. File lỗi → script thoát mã 1.

## Đọc thử bằng tay (không qua phần mềm)

Dùng khi muốn thử giọng hoặc kiểm bộ đọc. Làm cả cuốn thì dùng phần mềm hoặc CLI `sano-docx2tts` ở mục dưới.

### 1. Soạn text — mỗi chương 1 file `.txt`

Giữ nguyên lời gốc. Dòng trắng = ranh giới đoạn. Heading `#` bị bỏ tự động. Text mẫu tự viết (không vướng bản quyền) có sẵn ở `docs/demo-books/<slug>/chapN.txt`.

```
chap1.txt  chap2.txt  chap3.txt
```

### 2. Build WAV (batch, load model 1 lần)

```bash
cd /duong-dan/thu-muc-sach            # WAV xuất ra thư mục hiện tại
PYTHONPATH=/duong-dan/sano-sach-noi/scripts/tts \
  ~/VieNeu-TTS-v3/.venv/bin/python /duong-dan/sano-sach-noi/scripts/tts/audio_gen_batch.py \
  chap1.txt chap2.txt chap3.txt
# Mặc định giọng "Hải Đăng". Đổi: --voice "Thái Sơn"
```

Tốc độ tham khảo (thời gian đọc / thời lượng audio, đã có model): Mac Apple Silicon ~0,17x; máy 2 lõi (GitHub Actions) Windows/Linux ~0,6–0,7x. Nạp model 1–4s.

> **Lưu ý zsh:** truyền nhiều file `.txt` qua biến shell phải dùng **mảng** (`files+=(...)` rồi `"${files[@]}"`). zsh KHÔNG word-split biến unquoted như bash → `$files` thành 1 arg duy nhất → "File không tồn tại".

### 3. Convert WAV → MP3

Sano đọc thời lượng bằng `github.com/tcolgate/mp3` (đếm frame) → **phải là MP3 hợp lệ CBR**. Dùng libmp3lame mono CBR:

```bash
for n in 1 2 3; do
  ffmpeg -y -i chap${n}_full.wav -codec:a libmp3lame -b:a 128k -ar 44100 -ac 1 chap${n}.mp3
done
```

## Tạo cả cuốn bằng dòng lệnh — `cmd/sano-docx2tts`

Cùng lõi với phần mềm (`internal/bookmaker`), làm tất cả từ **1 file `.docx`**:

```
docx → trích mục lục nhiều cấp → chuẩn hóa cách đọc → sinh mô tả ảnh slide
     → pre-render TTS (VieNeu) → MP3 + metadata.json → (tùy chọn) gói zip, file M4B
```

### Quy ước docx (bắt buộc để trích đúng)

| Style trong Word | Vai trò |
|---|---|
| **Title** | Tiêu đề sách (có thể ghi đè bằng `--title`) |
| **Cấp heading nhỏ nhất** có mặt | Chương (thường Heading 1; nếu docx chỉ có Heading 2/3 thì Heading 2 đóng vai chương) |
| Các cấp heading **lớn hơn** | Tiểu mục trong chương (đơn vị phát) |
| Đoạn thường | Nội dung tiểu mục (giữ `original_text`) |
| **Caption ảnh/bảng** (style chứa `caption`: `ImageCaption`, `CaptionedFigure`, `Caption`…) | **KHÔNG đọc** — nhãn "Hình N — ..." vô nghĩa với người nghe; loại khỏi `reading_script` **và** `original_text`. Ảnh kèm trong đoạn vẫn được trích. |
| Ảnh nhúng | Trích ra `images/`, sinh `image_description` |

> Cấp heading nhỏ nhất xuất hiện trong tài liệu được coi là chương (docx chuẩn Heading 1/2 → chương = Heading 1; docx export Heading 2/3 → chương = Heading 2). Đoạn body nằm trực tiếp dưới 1 chương (chưa có tiểu mục) → tự tạo 1 tiểu mục ngầm mang tên chương. Số trang (dòng chỉ chứa số) bị bỏ tự động.

### Chạy

```bash
go build -o bin/sano-docx2tts ./cmd/sano-docx2tts

bin/sano-docx2tts \
  --input /duong-dan/cam-nang.docx \
  --output-dir /tmp/cam-nang/ \
  --title "Cẩm Nang ..." --author "..." --category "Kỹ năng" \
  --voice "Trúc Ly" --tts-mode vieneu \
  --verbose
```

Cờ chính: `--voice` (mặc định Hải Đăng) · `--tts-mode vieneu|stub` (stub = MP3 im lặng, thử nhanh không cần model) · `--output-zip <file>` (đóng gói thêm zip chuẩn — [`book-zip-format.md`](book-zip-format.md)) · `--repack-dir <thư mục>` (đóng gói lại zip từ thư mục đã render, không render lại) · `--cover` / `--cover-first-image` (mặc định tự vẽ bìa theo tên sách). Sinh docx mẫu: `--gen-sample-docx /tmp/s.docx`. Xem hết: `bin/sano-docx2tts -h`.
Bộ đọc: `--tts-python` (mặc định `~/VieNeu-TTS-v3/.venv/bin/python`, Windows `.venv\Scripts\python.exe`) · `--tts-script` (mặc định: `bin/../scripts/tts/audio_gen_batch.py` cạnh file chạy; không có thì dùng bản script nhúng sẵn trong chương trình, giải nén vào thư mục cache của máy. Không tìm theo thư mục hiện tại để tránh chạy nhầm script lạ — đang sửa script trong repo thì truyền `--tts-script scripts/tts/audio_gen_batch.py`).

Muốn phần mềm thấy sách vừa tạo: đặt `--output-dir` là một thư mục con trong `~/Sano/Sach/`.

### Đầu ra

```
output-dir/
├── metadata.json        # chapters nhiều cấp (xem book-zip-format.md)
├── cover.png            # bìa tự vẽ, hoặc ảnh từ --cover / --cover-first-image
├── images/              # ảnh slide trích ra
└── chNN-secMM.mp3       # 1 MP3 / tiểu mục (CBR 128k mono)
```

`reading_script` đã chuẩn hóa: số La Mã→chữ ("Chương I"→"Chương một"), viết tắt mở rộng (`vd.`→ví dụ, `TP.HCM`→Thành phố Hồ Chí Minh, `v.v.`→vân vân); **caption ảnh (style chứa "caption") bị loại — không đọc nhãn "Hình N — ..."**. `image_description` **KHÔNG** ghép vào lời đọc (giữ trong metadata). `original_text` giữ nguyên lời gốc (đã loại caption ảnh).

### Xuất M4B (nghe trên điện thoại, trên xe)

Một file `.m4b` cho cả cuốn: AAC mono 64 kbps (~29 MB mỗi giờ nghe), mốc mục lục theo từng tiểu mục (chương chỉ có một tiểu mục trùng tên thì lấy tên chương), thẻ tên sách / tác giả / thể loại Audiobook, bìa nhúng (bìa của sách, không có thì tự vẽ bìa vuông 1400×1400). Chỉ cần ffmpeg.

```bash
# Render xong thì xuất luôn M4B:
bin/sano-docx2tts --input cam-nang.docx --output-dir /tmp/cam-nang/ --m4b ~/Downloads/cam-nang.m4b

# Xuất từ thư mục sách đã render sẵn (không render lại, không cần docx):
bin/sano-docx2tts --m4b-from-dir /tmp/cam-nang/ --m4b ~/Downloads/cam-nang.m4b
bin/sano-docx2tts --m4b-from-dir ~/Sano/Sach/cam-nang/     # lưu <thư mục>/<Tên sách>.m4b
```

Cờ: `--m4b <file>` · `--m4b-from-dir <thư mục>` · `--m4b-bitrate` (mặc định `64k`) · `--ffmpeg`. Cách chép sang điện thoại: [`nghe-tren-dien-thoai.md`](nghe-tren-dien-thoai.md).

Kiểm file: `ffprobe -v error -show_chapters -show_format out.m4b` (mốc chương, thẻ, thời lượng), `ffprobe -show_streams` có stream ảnh `attached_pic`.

### Thử nhanh không cần model

```bash
go run ./cmd/sano-docx2tts -gen-sample-docx /tmp/s.docx
go run ./cmd/sano-docx2tts --input /tmp/s.docx --output-dir /tmp/s/ --tts-mode stub --output-zip /tmp/book-s.zip
go test ./internal/...
```

### Mô tả ảnh (chưa bật)

`image_description` hiện là placeholder (tham chiếu vị trí + tên tệp). Cấu trúc đã sẵn (ảnh đã trích ra đĩa) để sau này gọi mô hình đọc ảnh sinh mô tả thật.
