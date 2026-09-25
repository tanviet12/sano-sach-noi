# Sano desktop — phần mềm tạo sách

Phần mềm cài trên máy (Windows, macOS, Linux) để biến tài liệu của chính bạn (file Word) thành sách nói. Viết bằng [Wails v2](https://wails.io): phần Go ở thư mục này, giao diện Vue ở `frontend/`.

Chạy thật: Tạo sách 6 bước (nạp file Word → mục lục + cảnh báo → chọn giọng, nghe mẫu → lời mở đầu → nghe thử bắt buộc ≥ 2 đoạn, sửa được lời đọc → render nền có tiến độ, huỷ được), Thư viện đọc từ `~/Sano/Sach/`, Trình phát phát MP3 thật (nhớ vị trí, đổi tốc độ), Xuất M4B (một file có mục lục chương + bìa, chạy nền, huỷ được — xem [Nghe trên điện thoại](../docs/nghe-tren-dien-thoai.md)), cài bộ đọc lần mở đầu + gỡ bộ đọc. Còn là vỏ: phần còn lại của Cài đặt, hộp cập nhật.

## Cần có

- Go 1.26+, Node 20+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest` (kiểm tra máy bằng `wails doctor`)
- macOS: Xcode Command Line Tools. Linux: `libgtk-3-dev` + `libwebkit2gtk-4.1-dev` (build với `-tags webkit2_41`). Windows: WebView2 (có sẵn trên Windows 11)
- ffmpeg (ghi MP3): tìm trong PATH, rồi Homebrew/MacPorts (`/opt/homebrew/bin`, `/usr/local/bin`, `/opt/local/bin`), apt/snap; hoặc đặt biến `SANO_FFMPEG`
- Bộ đọc VieNeu-TTS v3: **app tự cài lần mở đầu** (xem mục "Cài bộ đọc" bên dưới). Máy dev đã cài tay `~/VieNeu-TTS-v3/.venv` theo [`docs/tts-build-guide.md`](../docs/tts-build-guide.md) thì app dùng luôn bản đó.

## Cài bộ đọc (lần mở đầu)

Chưa có bộ đọc → app mở màn cài: báo trước dung lượng (~1,5 GB) và thời gian, bấm **Cài bộ đọc** là tự làm hết, không cần Terminal. Mỗi bước có thanh tiến độ, huỷ được (lần sau bấm **Cài tiếp**, bỏ qua bước đã xong, mô hình tải tiếp phần dở), lỗi thì báo rõ + gợi ý + nút **Thử lại**. Logic ở `internal/setup/`.

| Bước | Nguồn | Kiểm |
|---|---|---|
| uv | GitHub Release `astral-sh/uv` đúng `UV_VERSION` | SHA256 ghim `UV_SHA256_*` trong `scripts/tts/versions.env` |
| Python | uv tải (`uv python install`, không ghi ra `~/.local/bin`, không ghi registry Windows) | uv tự kiểm SHA256 bản Python |
| Mã VieNeu-TTS | tarball GitHub của `VIENEU_COMMIT` (không cần git) | "tree hash" nội dung `VIENEU_TREE_SHA256` |
| Thư viện | `uv sync --frozen --no-dev` theo `scripts/tts/vieneu-project/uv.lock` (uv.lock của VieNeu + ghi đè của Sano: bỏ gradio, bản vá bảo mật; chép đè sau khi kiểm tree hash). Máy đã cài mà khoá đổi → Cài đặt mời "Cập nhật bộ đọc" | uv kiểm hash từng gói trong `uv.lock` |
| Mô hình (~580 MB) | Hugging Face, đúng revision ghim (`models.py fetch`, không cần tài khoản) | SHA256 từng file trong `scripts/tts/models.sha256` |
| ffmpeg | chỉ tải khi máy chưa có — bản dựng tĩnh GPLv3 (`FFMPEG_*` trong `versions.env`) | SHA256 ghim + có bộ mã `libmp3lame` |
| Kiểm tra cuối | `models.py check` + đọc thử một câu ra WAV | |

Thư mục dữ liệu app (`SANO_DATA_DIR` đè được): macOS `~/Library/Application Support/Sano`, Windows `%LOCALAPPDATA%\Sano`, Linux `$XDG_DATA_HOME/sano` (mặc định `~/.local/share/sano`). Bộ đọc nằm gọn trong `<thư mục dữ liệu>/tts/` (uv, Python, mã VieNeu + venv, mô hình qua `HF_HOME` riêng, ffmpeg nếu tải, `install.log`), có file đánh dấu `.sano-tts`. **Gỡ bộ đọc** (Cài đặt → Bộ đọc → Gỡ) hỏi xác nhận rồi xoá đúng thư mục đó — không đụng sách `~/Sano/Sach`, `~/VieNeu-TTS*` hay `~/.cache/huggingface`.

Script đọc giọng (`scripts/tts/*.py` + `versions.env` + `models.sha256`) được nhúng vào app (`scripts/tts/embed.go`) và giải nén vào `<tts>/scripts` — bản `.app` chép đi đâu cũng chạy. Bộ đọc dùng theo thứ tự: `SANO_TTS_PYTHON` → bộ đọc app cài → `~/VieNeu-TTS-v3`.

Thử cài thật mà không ảnh hưởng máy (cần mạng, ~1–5 phút):

```bash
cd desktop
SANO_SETUP_INTEGRATION=1 SANO_DATA_DIR=/tmp/sano-thu go test -count=1 ./internal/setup -run TestInstallThat -v   # thêm SANO_SETUP_FORCE_FFMPEG=1 để thử cả bước tải ffmpeg
SANO_SETUP_UNINSTALL=1  SANO_DATA_DIR=/tmp/sano-thu go test -count=1 ./internal/setup -run TestUninstallThat -v
```

## Chạy

Từ gốc repo:

```bash
make desktop-dev     # cửa sổ app + Vite :5390; mở http://localhost:34115 bằng trình duyệt để gọi được Go
make desktop-build   # ra desktop/build/bin/Sano.app (macOS), Sano.exe (Windows)
make desktop-test    # go vet + go test + typecheck giao diện
```

Thử luồng thật khi dev mà không cần hộp chọn file: mở `http://localhost:34115/?docx=/đường/dẫn/tuyệt/đối/file.docx` (chỉ có ở bản dev). Tạo docx mẫu: `go run ./cmd/sano-docx2tts -gen-sample-docx /tmp/s.docx`.

Chỉ sửa giao diện, không cần Go: `cd desktop/frontend && npm install && npm run dev` rồi mở http://localhost:5390 (dùng dữ liệu giả thay cho phần Go). Mở thẳng một màn: `?screen=setup|library|create|player|settings`, `&step=1..6`, `?update=1`.

## Phát hành

Bản cài chỉ build trên GitHub Actions (`.github/workflows/desktop-release.yml`), không build tay:

- **Gắn thẻ** `vX.Y.Z` (hoặc `vX.Y.Z-beta.N`) rồi đẩy thẻ lên → build 3 hệ điều hành, tạo `SHA256SUMS`, attestation nguồn gốc build (khi repo công khai), tạo GitHub Release **nháp** đính kèm mọi file. Nội dung release lấy từ mục `## X.Y.Z` trong `CHANGELOG.md`. Kiểm lại rồi tự bấm Publish.
- **Run workflow** (tab Actions) → build thử 3 hệ điều hành, tải file ở mục Artifacts, không tạo release.
- Pull request / push vào `main` có đụng mã desktop → build + đóng gói Linux và Windows để bắt lỗi sớm.

| File | Cho máy |
|---|---|
| `Sano-<version>-windows-amd64-setup.exe` | Windows 10/11 64-bit, bộ cài NSIS cài theo người dùng (`%LOCALAPPDATA%\Programs\Sano`), không đòi quyền admin |
| `Sano-<version>-windows-amd64-portable.zip` | Windows, giải nén là chạy (cần WebView2, Windows 11 có sẵn) |
| `Sano-<version>-macos-universal.dmg` | macOS 10.13+, cả Apple Silicon lẫn Intel |
| `Sano-<version>-linux-amd64.AppImage` | Linux x86_64 có GTK3 + WebKitGTK 4.1 (Ubuntu 22.04+, Debian 12+, Fedora 36+). Thiếu thì `sudo apt install libwebkit2gtk-4.1-0` |

Các bước đóng gói là script trong `scripts/release/`, chạy được cả trên máy lẫn CI:

```bash
cd desktop/frontend && npm ci && cd ../..
scripts/release/build.sh 0.1.0                # wails build + gắn phiên bản (-X main.version)
scripts/release/package-macos.sh 0.1.0        # hoặc package-windows.sh / package-linux.sh
scripts/release/checksums.sh                  # SHA256SUMS → desktop/build/bin/release/
```

**App chưa ký số** (chưa mua chứng chỉ). macOS chỉ ký ad-hoc, không notarize: lần đầu mở báo không xác minh được nhà phát triển → Cài đặt hệ thống → Quyền riêng tư & Bảo mật → **Vẫn mở**. Windows SmartScreen → **Thông tin thêm** → **Vẫn chạy**. Linux: `chmod +x Sano-*.AppImage` rồi chạy. Không nén UPX (dễ bị antivirus báo nhầm).

Kiểm file tải về: `sha256sum -c SHA256SUMS --ignore-missing` (macOS: `shasum -a 256 -c ...`), nguồn gốc build: `gh attestation verify <file> --repo <owner>/sano-sach-noi`.

## Cấu trúc

```
desktop/
├── main.go, app.go      # cửa sổ Wails + các hàm gọi từ giao diện (Version, CheckTTS, ChooseDocx, DescribeDocx)
├── maker.go             # Tạo sách: InspectDocx, Voices, PreviewClips, SpeakSample, StartRender/CancelRender (gọi internal/bookmaker)
├── books.go, media.go   # Thư viện ~/Sano/Sach + phát MP3/bìa qua /sano-media/
├── m4b.go               # Xuất M4B: ExportM4B/CancelM4B (sự kiện m4b:progress, m4b:finished), gọi internal/m4b
├── internal/library/    # đọc/ghi thư mục sách: mỗi cuốn một thư mục <slug>/ (MP3, .txt lời đọc, bìa, metadata.json, book-<slug>.zip)
├── version.go           # phiên bản: -ldflags "-X main.version=..." → file VERSION → "dev"
├── setup.go             # Cài/gỡ bộ đọc: SetupInfo, StartSetup/CancelSetup (sự kiện setup:progress), UninstallTTS
├── licenses/            # giấy phép bên thứ ba (VieNeu-TTS Apache-2.0, ffmpeg GPL...) — nhúng vào app
├── internal/setup/      # cài bộ đọc lần đầu: tải + kiểm SHA256, uv, Python, VieNeu, mô hình, ffmpeg, gỡ an toàn
├── internal/tts/        # tìm + kiểm tra bộ đọc: thư mục dữ liệu app, python, script nhúng, models.py check, ffmpeg
├── build/               # icon + Info.plist macOS (build/bin/ là file build ra, không commit)
└── frontend/            # Vue 3 + Tailwind + TypeScript (component ui/ kiểu shadcn-vue, design token ở src/style.css)
```

- `desktop/` là Go module riêng (`sano/desktop`) để thư viện Wails không lọt vào module gốc (CLI + thư viện tạo sách); vẫn import được `sano/internal/...` nhờ `replace sano => ../`.
- Giao diện: alias `@/` → `frontend/src`. Design token ở `frontend/src/style.css` + `frontend/tailwind.theme.ts`. Điều khoản và lời nhắc mẫu đọc thẳng từ `docs/` ở gốc repo (`?raw`) nên Vite cho phép đọc tới gốc repo.
- Wireframe đã duyệt nằm ở `frontend/src/wireframes/`, xem ở bản dev: `http://localhost:5390/?wireframe=desktop` (thêm `&screen=...&step=...`), `?wireframe=library`, `?wireframe=terms`. Không có trong bản build.
- Biến môi trường: `SANO_DATA_DIR` (thư mục dữ liệu app), `SANO_TTS_PYTHON` (python của venv VieNeu khác), `SANO_TTS_SCRIPTS` (thư mục chứa `models.py`, khi sửa script lúc dev), `SANO_FFMPEG`.
- Thử với thư viện tạm, không đụng sách thật: `SANO_HOME=/thư/mục/tạm make desktop-dev` (thay cho `~/Sano`, sách nằm ở `$SANO_HOME/Sach`).
- Thử Xuất M4B mà không qua hộp lưu file: `SANO_M4B_OUT=/thư/mục/tạm` (hoặc đường dẫn `.m4b`) — lưu thẳng vào đó, xong không tự mở thư mục.
