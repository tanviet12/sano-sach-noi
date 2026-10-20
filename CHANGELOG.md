# Changelog

## v0.1.25 (20/10/2026)

### Tính năng
- **Bản macOS có chữ ký số Apple:** Sano ký bằng chứng chỉ Developer ID và qua Apple kiểm duyệt (notarize). Tải về mở thẳng, không còn phải vào Quyền riêng tư & Bảo mật bấm "Vẫn mở". Lần đầu mở bản mới, macOS có thể hỏi lại quyền vào thư mục một lần
- **Nạp sách điện tử EPUB và file PDF có chữ** (từ 0.1.23): tên sách, chương, mục lấy từ mục lục hoặc bookmark có sẵn. PDF dễ đọc sai thì Sano báo ngay lúc nạp, kèm số trang

## v0.1.24 (06/10/2026)

### Tính năng
- **Nạp sách điện tử EPUB:** ngoài file Word, Sano nhận thẳng file `.epub`. Tên sách, chương, mục lấy từ mục lục có sẵn trong sách. Không đọc trang bìa, trang mục lục, chú thích cuối trang. Sách có DRM bị từ chối
- **Nạp file PDF có chữ:** chương, mục lấy từ bookmark (không có thì đoán theo cỡ chữ tiêu đề), tự bỏ đầu trang, chân trang, số trang. PDF dễ đọc sai thì Sano báo ngay lúc nạp, kèm số trang: chữ lỗi phông (TCVN3, VNI), trang chỉ có ảnh, dàn nhiều cột. Bản scan, PDF có mật khẩu hoặc cấm trích chữ bị từ chối

### Sửa lỗi
- Các màn tạo sách, Thư viện, Giới thiệu ghi đủ Word, EPUB, PDF thay vì chỉ "file Word"

## v0.1.23 (05/10/2026)

### Tính năng
- **Nạp sách điện tử EPUB:** ngoài file Word, Sano nhận thẳng file `.epub`. Tên sách, chương, mục lấy từ mục lục có sẵn trong sách. Không đọc trang bìa, trang mục lục, chú thích cuối trang. Sách có DRM bị từ chối
- **Nạp file PDF có chữ:** chương, mục lấy từ bookmark (không có thì đoán theo cỡ chữ tiêu đề), tự bỏ đầu trang, chân trang, số trang. PDF dễ đọc sai thì Sano báo ngay lúc nạp, kèm số trang: chữ lỗi phông (TCVN3, VNI), trang chỉ có ảnh, dàn nhiều cột. Bản scan, PDF có mật khẩu hoặc cấm trích chữ bị từ chối

## v0.1.21 (30/09/2026)

### Tính năng
- **Kết nối AI (MCP):** Claude Desktop, Claude Code, Codex, Cursor xem, tạo, sửa sách trong Sano bằng lời nói thường. Mục MCP mới ở thanh bên: thêm vào Claude Desktop một chạm, lệnh chép sẵn cho Claude Code, Codex, tắt được, có nhật ký. AI không có lệnh xoá sách
- **Sách AI tạo:** tạo trước, cam kết sau. Sách nằm ở khu chờ, tick cam kết một lần cho nhiều cuốn là vào Thư viện
- **AI sửa sách, hoàn tác được:** sửa lời, tìm và thay, đổi giọng, bìa, thông tin, cách đọc từ. Mỗi lần sửa Sano giữ bản trước, bấm Hoàn tác ở mục MCP
- **Tìm trong Mục lục** ở màn nghe: gõ tên chương, mục (có dấu hay không đều được) để nhảy tới chỗ đã nghe

## v0.1.20 (29/09/2026)

### Tính năng
- **Tìm và thay trong cả cuốn:** ở Sửa sách, bấm "Tìm và thay" (⌘F) để đổi một cụm chữ trong lời đọc mọi mục, tick thêm để đổi cả tiêu đề mục. Chỉ đọc lại các mục bị thay, không đọc lại cả cuốn

## v0.1.19 (29/09/2026)

### Tính năng
- **Cam kết trước khi tạo:** bấm render hoặc tạo video, Sano hiện bảng cam kết (quyền dùng tài liệu, không vi phạm pháp luật, không mạo danh, không phát tán tác phẩm của người khác, tự chịu trách nhiệm). Tick đủ mới tạo được
- **Điều khoản sử dụng phiên bản 2:** rõ hơn về nội dung vi phạm pháp luật, video và ảnh chia sẻ, thêm mục Bồi hoàn. Mở app sẽ được hỏi đồng ý lại một lần

### Sửa lỗi
- Video và ảnh khung dọc 9:16 không còn bị Reels, TikTok, Story che logo, chữ và tên sách

## v0.1.18 (28/09/2026)

### Tính năng
- **Chia sẻ câu hay:** ảnh có lời từ đoạn đang nghe, đăng Facebook, Zalo, Story. Có Sao chép ảnh, AirDrop
- **Tạo video:** video ngắn có tiếng đọc cho Reels, TikTok; hoặc video cả cuốn có chữ chạy cho YouTube, kèm thumbnail, phụ đề, mô tả có mốc chương
- **Lời đọc phóng to** ngay trong màn nghe, là chế độ mặc định. Rê chuột vào lời đọc để chia sẻ đúng câu đó
- **Sửa sách:** thêm Dịch giả, Nhà xuất bản

### Sửa lỗi
- Chữ chạy không còn lệch một câu ở tiểu mục có dấu ":"
- Giọng Thiền Tâm Đức đọc đúng "chánh" (chánh niệm, chánh kiến)
- Mục lục màn nghe hiện tên chương

## v0.1.17 (27/09/2026)

### Tính năng
- **Sửa sách:** sửa chữ từng mục rồi Lưu & đọc lại, không phải xoá tạo lại. Đổi tên thì bìa và lời giới thiệu tự làm lại. Đổi giọng cả cuốn chạy nền
- **Từ điển cách đọc:** dạy Sano đọc đúng tên riêng, viết tắt (vd Nielsen → Niu-sen), cho một cuốn hoặc mọi sách
- **Phím tắt khi nghe:** Space dừng / nghe tiếp, ← → sang tiểu mục trước / sau

### Sửa lỗi
- Dấu ":" nghỉ rõ, không đọc liền; đổi giọng ở Nghe thử không mất chỗ đã sửa
- Bìa trong Thư viện không còn lúc hiện lúc trắng

## v0.1.16 (27/09/2026)

### Tính năng
- **Nghe trên điện thoại:** nút mới dẫn từng bước, quét mã QR cài BookPlayer, có AirDrop trên Mac
- **Chỉnh quãng nghỉ** giữa các phần: chung ở Cài đặt hoặc riêng từng cuốn (nút Nghỉ)
- **Màn nghe vừa mọi cỡ cửa sổ**; Xuất zip, Xoá gom vào bánh răng góc trên
- **Báo bản mới rõ hơn**, tự hiện một lần mỗi phiên bản
- **Trang [Skill AI](https://sanobook.com/skill-ai):** tải skill làm sách nói cho Claude, ChatGPT

## v0.1.15 (27/09/2026)

### Sửa lỗi
- **M4B có quãng nghỉ giữa các phần:** trước đây nghe trên điện thoại đọc liền sang chương sau. Nay nghỉ 2 giây khi sang chương, 1,5 giây giữa các tiểu mục
- **Nghe trong Sano:** nghỉ khoảng 1,5 giây giữa các tiểu mục, bằng với file M4B

### Lưu ý khi nâng cấp
- Từ **0.1.2 – 0.1.14**: bấm **Cập nhật ngay**. Từ 0.1.0 / 0.1.1: tải bản 0.1.15 và cài đè một lần
- File M4B đã xuất trước đây: xuất lại để có quãng nghỉ

## v0.1.14 (27/09/2026)

### Tính năng
- **Cách đọc 3 cấp:** đọc nguyên văn, hoặc nhờ AI làm mượt, viết lại thành văn sách nói
- **Nhờ AI:** chọn Claude, ChatGPT hoặc Gemini, Sano đưa đúng prompt; tặng kèm skill làm sách nói
- **Hành trình nghe:** thời gian nghe, chuỗi ngày, sách nghe xong, mục tiêu mỗi ngày
- **Thư viện:** xem dạng danh sách, xoá lịch sử nghe

### Lưu ý khi nâng cấp
- Từ **0.1.2 – 0.1.12**: bấm **Cập nhật ngay**. Không có bản 0.1.13, các thay đổi ra chung trong 0.1.14

Chi tiết: [Ba cách đọc và skill](https://sanobook.com/lam-muot-tai-lieu) · [Hành trình nghe](https://sanobook.com/hanh-trinh-nghe). Nhỏ hơn: menu đổi thành Thư viện → Tạo sách nói → Hành trình nghe → Cài đặt → Giới thiệu; thanh bên có dòng nhà tài trợ; dấu ✓ trong mục lục chỉ hiện khi nghe thật từ 85% trở lên. Từ 0.1.0 / 0.1.1: tải bản 0.1.14 và cài đè một lần.

## v0.1.12 (27/09/2026)

### Sửa lỗi
- **Mac Intel cài được bộ đọc:** trước đây bước cài báo lỗi `onnxruntime ... doesn't have a source distribution or wheel for the current platform` vì onnxruntime 1.24 bỏ bản cho Mac Intel. Mac Intel giờ dùng onnxruntime 1.23.2. Apple Silicon, Windows, Linux không đổi
- **Báo rõ khi macOS quá cũ:** bộ đọc cần macOS 14 trở lên trên Apple Silicon, macOS 13 trở lên trên Mac Intel. Máy cũ hơn thấy thông báo cần cập nhật macOS ngay ở màn cài, thay cho lỗi khó hiểu

### Lưu ý khi nâng cấp
- Từ **0.1.2 – 0.1.11**: bấm **Cập nhật ngay** trong app. Từ **0.1.0 / 0.1.1**: tải bản 0.1.12 và cài đè một lần
- Mac Intel từng cài lỗi: cập nhật xong bấm **Thử lại** ở màn cài bộ đọc

## v0.1.11 (26/09/2026)

### Tính năng
- **Bộ sách nhiều tập:** ô **Bộ sách** + **Tập số** ở bước Nạp file và trong Sửa thông tin (tự điền tập kế tiếp, báo khi trùng số tập). Các tập gom thành một thẻ trên kệ, bấm vào mở trang bộ sách xếp theo số tập, có **Nghe tiếp** đúng tập đang dở. Nghe hết một tập thì tự chuyển sang tập sau. Gói zip mang theo tên bộ và số tập, nhập sang máy khác vẫn giữ
- **Tự sắp xếp:** nút **Sắp xếp** cạnh "Tất cả sách" (hoặc chọn Tự sắp xếp trong menu), nhấn giữ bìa kéo đổi chỗ, bấm **Xong**. Có **Về thứ tự cũ**. Thứ tự lưu ở `~/Sano/.thu-tu.json`
- **Quản lý danh mục và bộ sách:** đổi tên áp cho mọi cuốn (trùng danh mục có sẵn thì gộp), xoá thì sách giữ nguyên

### Lưu ý khi nâng cấp
- Từ **0.1.2 – 0.1.10**: bấm **Cập nhật ngay** trong app. Từ **0.1.0 / 0.1.1**: tải bản 0.1.11 và cài đè một lần

## v0.1.10 (26/09/2026)

### Tính năng
- **Giọng khuyên dùng:** bước chọn giọng mở sẵn tab **Khuyên dùng** với 3 giọng trong, không rè: Hải Đăng, Thiện Minh (nam, miền Bắc), Mỹ Duyên (nữ, miền Nam). Ở các tab miền, 3 giọng này nằm trên cùng, có nhãn Khuyên dùng

### Lưu ý khi nâng cấp
- Từ **0.1.2 – 0.1.9**: bấm **Cập nhật ngay** trong app. Từ **0.1.0 / 0.1.1**: tải bản 0.1.10 và cài đè một lần

## v0.1.9 (26/09/2026)

### Tính năng
- **3 sách mẫu mới có sẵn trong thư viện:** "Giới thiệu Sano" (3 phút, giọng Thiện Minh), "Nghe Để Nhớ" (12 phút, giọng Hải Đăng), "Tiệm Cà Phê Thứ Hai" (18 phút, giọng Mỹ Duyên), thay cho cuốn "Kỹ năng mềm cho người trẻ". Máy đã cài Sano cũng nhận 3 cuốn mới khi cập nhật; cuốn nào đã xoá thì không tự thêm lại. Cuốn mẫu cũ còn trong thư viện thì vẫn giữ

### Lưu ý khi nâng cấp
- Từ **0.1.2 – 0.1.8**: bấm **Cập nhật ngay** trong app. Từ **0.1.0 / 0.1.1**: tải bản 0.1.9 và cài đè một lần

## v0.1.8 (26/09/2026)

### Tính năng
- **Nhập sách từ gói zip:** nút **Nhập sách** ở Thư viện (hoặc kéo thả file .zip vào cửa sổ). Xem trước bìa, tên, giọng, số chương, thời lượng; trùng sách thì chọn Giữ cả hai / Thay thế. Sách nhập giữ đủ mục lục, bìa, tên giọng, chữ chạy theo
  - Gói zip là file người khác gửi nên được kiểm chặt: chỉ đọc đúng file của gói sách, chặn đường dẫn ra ngoài thư mục, tên trùng, file mã hoá, file quá lớn hoặc nén bất thường; mp3 và ảnh bìa phải đúng định dạng (không nhận SVG); giải nén vào thư mục tạm rồi đóng gói lại sạch

### Sửa lỗi
- Sách tên trùng tên thiết bị của Windows ("Con", "Nul"…) không tạo được thư mục; nay thêm hậu tố

### Lưu ý khi nâng cấp
- Từ **0.1.2 – 0.1.7**: bấm **Cập nhật ngay** trong app. Từ **0.1.0 / 0.1.1**: tải bản 0.1.8 và cài đè một lần

## v0.1.7 (25/09/2026)

### Tính năng
- **Chữ chạy theo lời đọc:** ô "Lời đọc" ở màn nghe hiện câu đang đọc; bấm vào mở màn Xem lời chữ lớn, câu đang đọc in đậm, tự cuộn theo, bấm một câu để nghe từ câu đó. Dùng được cho cả sách đã tạo: thời điểm từng câu ước lượng theo độ dài câu rồi khớp vào khoảng lặng thật trong âm thanh
- **Rời màn nghe vẫn nghe tiếp:** thanh nghe nhỏ ở đáy cửa sổ (phát/dừng, tua −15s/+30s, dừng hẳn); bấm vào tên sách để mở lại màn nghe. Nghe mẫu giọng khi đang tạo sách thì sách tự tạm dừng
- **Sách ghi tên giọng đọc** trên thẻ trong Thư viện và màn nghe (cả sách đã tạo trước đây); tìm sách được theo tên giọng
- **Chọn giọng theo miền:** Miền Bắc / Trung / Nam / Tất cả kèm lọc giọng Nam, Nữ; mở đúng miền đã chọn lần trước, giọng đã dùng có nhãn "Dùng lần trước"

### Sửa lỗi
- Mục lục ở màn nghe tự cuộn tới tiểu mục đang phát, không phải kéo tìm
- Chữ **P** đứng riêng ("chữ P đầu tiên", "P thứ hai là giá") và "4Ps" đọc là "pê" thay vì "phê"

### Lưu ý khi nâng cấp
- Từ **0.1.2 – 0.1.6**: bấm **Cập nhật ngay** trong app. Từ **0.1.0 / 0.1.1**: tải bản 0.1.7 và cài đè một lần. Tên giọng, chữ chạy theo có ngay cho sách đã tạo; cách đọc chữ P mới chỉ áp dụng cho sách tạo sau khi nâng cấp

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
- Trang [Chính sách ký số](https://sanobook.com/chinh-sach-ky-so): file nào được ký, ký ở đâu, ai duyệt; đang xin ký số Windows miễn phí qua SignPath Foundation

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
