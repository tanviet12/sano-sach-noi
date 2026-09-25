<script setup lang="ts">
// Điểm nổi bật (wireframe WfLanding mục 5). Tự làm thay cho frontmatter `features` của VitePress
// để giữ đúng thứ tự mục như wireframe (features của VitePress luôn nằm ngay dưới hero).
import { Github, ListTree, Mic, UserRoundX, Wand2, WifiOff } from 'lucide-vue-next'

const items = [
  { icon: WifiOff, title: 'Chạy offline trên máy', desc: 'Tài liệu không gửi lên mạng. Cài bộ đọc xong là dùng được không cần internet.' },
  { icon: Mic, title: '25 giọng Việt', desc: 'Nam, nữ, giọng Bắc, Trung, Nam của bộ đọc mã nguồn mở VieNeu-TTS. Nghe thử từng giọng trước khi chọn.' },
  { icon: Wand2, title: 'Tự chuẩn hoá văn nói', desc: 'Số, viết tắt, mũi tên, ký hiệu được đọc thành lời tự nhiên.' },
  { icon: ListTree, title: 'Mục lục chương', desc: 'Lấy từ Heading trong Word. Nhảy chương trong app và trong file M4B.' },
  { icon: UserRoundX, title: 'Không cần API key, không tốn token', desc: 'Không điền key ChatGPT hay dịch vụ AI nào, không trả tiền theo lượt dùng. Mô hình AI chạy trên máy bạn, không đăng ký, không đăng nhập.' },
  { icon: Github, title: 'Mã nguồn mở MIT', desc: 'Miễn phí, xem và sửa được toàn bộ mã nguồn trên GitHub.' },
]
</script>

<template>
  <section class="sano-section alt">
    <div class="sano-container">
      <h2 class="sano-h2">Điểm nổi bật</h2>
      <div class="grid">
        <div v-for="h in items" :key="h.title" class="sano-card item">
          <span class="sano-icon-tile"><component :is="h.icon" :size="20" aria-hidden="true" /></span>
          <p class="t">{{ h.title }}</p>
          <p class="d sano-muted">{{ h.desc }}</p>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.grid {
  margin-top: 32px;
  display: grid;
  gap: 16px;
}
@media (min-width: 640px) {
  .grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (min-width: 1024px) {
  .grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
.item {
  padding: 20px;
  box-shadow: none;
}
.t {
  margin: 16px 0 0;
  font-weight: 500;
}
.d {
  margin: 4px 0 0;
  font-size: 14px;
  line-height: 1.6;
}
</style>
