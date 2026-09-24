<script setup lang="ts">
// Cách hoạt động (wireframe WfLanding mục 3): 4 bước + video giới thiệu (quay app thật, lồng tiếng Thiện Minh).
import { withBase } from 'vitepress'
import { AudioLines, FileUp, Headphones, Smartphone } from 'lucide-vue-next'
import shot1 from '../../../images/app/b1-nap-file.jpg'
import shot2 from '../../../images/app/b5-nghe-thu.jpg'
import shot3 from '../../../images/app/b6-render.jpg'
import shot4 from '../../../images/app/nghe.jpg'

const steps = [
  { icon: FileUp, shot: shot1, title: 'Nạp file Word', desc: 'Sano đọc mục lục theo Heading, cảnh báo bảng, hình, chữ viết tắt lạ.' },
  { icon: Headphones, shot: shot2, title: 'Nghe thử & sửa lời đọc', desc: 'Nghe vài đoạn, sửa cách đọc từ riêng trước khi tạo cả cuốn.' },
  { icon: AudioLines, shot: shot3, title: 'Tạo sách (chạy nền)', desc: 'Giọng đọc AI chạy trên máy bạn. Vẫn làm việc khác được trong lúc chờ.' },
  { icon: Smartphone, shot: shot4, title: 'Nghe trên máy hoặc điện thoại', desc: 'Nghe ngay trong Sano, hoặc xuất một file M4B chép sang điện thoại.' },
]
</script>

<template>
  <section id="cach-hoat-dong" class="sano-section alt">
    <div class="sano-container">
      <h2 class="sano-h2">Cách hoạt động</h2>
      <p class="sano-lead">Bốn bước, từ file Word tới sách nói có mục lục chương.</p>
      <ol class="steps">
        <li v-for="(s, i) in steps" :key="s.title" class="sano-card step">
          <div class="ph"><img :src="s.shot" :alt="`Bước ${i + 1}: ${s.title}`" width="1600" height="955" loading="lazy" /></div>
          <div class="head">
            <span class="num">{{ i + 1 }}</span>
            <component :is="s.icon" :size="16" class="ico" aria-hidden="true" />
            <p class="t">{{ s.title }}</p>
          </div>
          <p class="d sano-muted">{{ s.desc }}</p>
        </li>
      </ol>
      <video
        class="video"
        controls
        preload="none"
        playsinline
        :poster="withBase('/video/sano-gioi-thieu-poster.jpg')"
        aria-label="Video giới thiệu Sano: nạp file Word, nghe thử, tạo sách, nghe trên máy hoặc điện thoại"
      >
        <source :src="withBase('/video/sano-gioi-thieu.mp4')" type="video/mp4" />
      </video>
    </div>
  </section>
</template>

<style scoped>
.steps {
  margin: 32px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 16px;
}
@media (min-width: 640px) {
  .steps {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (min-width: 1024px) {
  .steps {
    grid-template-columns: repeat(4, 1fr);
  }
}
.step {
  display: flex;
  flex-direction: column;
  padding: 16px;
}
/* Ảnh chụp cả cửa sổ: phóng vào vùng nội dung chính (bỏ thanh bên) cho đọc được chữ */
.ph {
  aspect-ratio: 4 / 3;
  overflow: hidden;
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  background: var(--vp-c-bg);
}
.ph img {
  display: block;
  width: 165%;
  max-width: none;
  height: auto;
  margin-left: -29%;
}
.head {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.num {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  border-radius: 9999px;
  background: var(--vp-button-brand-bg);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
}
.ico {
  color: var(--vp-c-brand-1);
  flex-shrink: 0;
}
.t {
  margin: 0;
  font-weight: 500;
}
.d {
  margin: 8px 0 0;
  font-size: 14px;
  line-height: 1.6;
}
.video {
  display: block;
  width: 100%;
  max-width: 48rem;
  margin: 32px auto 0;
  aspect-ratio: 16 / 9;
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  background: var(--vp-c-bg);
}
</style>
