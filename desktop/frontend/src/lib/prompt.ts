// Lời nhắc mẫu (P2-6, P3-6): dán cùng file Word vào ChatGPT, Gemini hoặc Claude bản
// web để biến bảng, hình thành lời văn trước khi nạp vào Sano. Một nguồn duy nhất:
// docs/prompts/loi-nhac-mau.txt (hướng dẫn dùng ở docs/lam-muot-tai-lieu.md).
import raw from '../../../../docs/prompts/loi-nhac-mau.txt?raw'

export const SAMPLE_PROMPT = raw.trim()
