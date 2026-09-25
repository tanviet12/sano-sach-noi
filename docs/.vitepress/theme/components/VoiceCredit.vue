<script setup lang="ts">
// Ghi công bộ đọc VieNeu-TTS (trang chủ mục Nghe thử + trang /demo): mọi giọng đọc trên trang
// do VieNeu-TTS tạo, chạy trên máy qua Sano.
import { ExternalLink } from 'lucide-vue-next'
</script>

<template>
  <div class="credit">
    <p class="t">Giọng đọc do <strong>VieNeu-TTS</strong> tạo</p>
    <p class="d sano-muted">
      Dự án mã nguồn mở chuyển văn bản thành giọng nói tiếng Việt của
      <a href="https://github.com/pnnbao97" target="_blank" rel="noopener">Pham Nguyen Ngoc Bao</a>, giấy phép Apache-2.0.
      Sano dùng mô hình VieNeu-TTS v3 Turbo chạy ngay trên máy, không chỉnh sửa mô hình. Cảm ơn tác giả đã chia sẻ.
    </p>
    <p class="links">
      <a href="https://github.com/pnnbao97/VieNeu-TTS" target="_blank" rel="noopener">GitHub VieNeu-TTS <ExternalLink :size="13" aria-hidden="true" /></a>
      <a href="https://huggingface.co/pnnbao-ump/VieNeu-TTS-v3-Turbo" target="_blank" rel="noopener">Mô hình trên Hugging Face <ExternalLink :size="13" aria-hidden="true" /></a>
      <a href="https://www.vieneu.io" target="_blank" rel="noopener">vieneu.io <ExternalLink :size="13" aria-hidden="true" /></a>
    </p>
  </div>
</template>

<style scoped>
.credit {
  margin-top: 24px;
  padding: 14px 16px;
  border: 1px dashed var(--vp-c-divider);
  border-radius: 10px;
  font-size: 14px;
}
.t {
  margin: 0;
  font-weight: 500;
}
.d {
  margin: 4px 0 0;
  line-height: 1.6;
}
.d a,
.links a {
  color: var(--vp-c-brand-1);
  text-decoration: none;
}
.d a:hover,
.links a:hover {
  text-decoration: underline;
}
.links {
  margin: 8px 0 0;
  display: flex;
  flex-wrap: wrap;
  gap: 6px 16px;
}
.links a {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
}
</style>
