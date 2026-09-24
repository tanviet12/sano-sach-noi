# Nghỉ giữa đoạn khi đọc giọng — không còn cần bản vá

**Trạng thái:** bản vá nghỉ đoạn **đã bỏ** khi chuyển sang VieNeu-TTS v3 Turbo. File này giữ lại để
giải thích vì sao, và để link cũ không gãy.

## Trước đây (VieNeu-TTS v2)

Pipeline `cmd/sano-docx2tts` xuất `reading_script` với **dòng trắng `\n\n`** làm ranh giới đoạn (sau
tiêu đề + giữa các đoạn). Script v2 xoá sạch dòng trắng khi chia chunk và nối mọi chunk với khoảng
lặng đều 80ms → audio đọc liền tù tì. Repo từng có bản vá trong `scripts/tts/` (`split_paragraphs` +
`concat_with_para_gaps`, nghỉ 550ms giữa đoạn) cùng vòng retry chunk lỗi.

## Bây giờ (VieNeu-TTS v3 Turbo)

`tts.infer(text, voice=...)` của v3 tự làm hết:

| Việc | v3 Turbo tự làm |
|---|---|
| Chia chunk | theo câu, tối đa 256 ký tự/chunk |
| Nghỉ theo ranh giới | ngắt đoạn **0,70s** > hết câu **0,50s** > ngắt trong câu **0,30s** (`V3_GAP_SILENCE` trong `vieneu_utils/core_utils.py` của upstream) |
| Chống lỗi "nói thêm" | babble guard: chunk ngắn sinh thừa thì sinh lại (`babble_retries`) |

Vì vậy `scripts/tts/audio_gen.py` chỉ đưa **nguyên văn bản** vào `tts.infer()`. Không tự chia chunk
nữa — chia nhỏ trước sẽ làm mất khoảng nghỉ theo ranh giới của v3. Dòng trắng / xuống dòng trong
`reading_script` vẫn là ranh giới đoạn như cũ, phía Go không phải đổi.

## Tinh chỉnh

Muốn đổi độ dài nghỉ thì phải sửa ở upstream (bảng `V3_GAP_SILENCE`) — hiện Sano dùng mặc định.

## Liên quan
- Cài đặt, ghim phiên bản, danh sách giọng: [`docs/tts-build-guide.md`](tts-build-guide.md).
