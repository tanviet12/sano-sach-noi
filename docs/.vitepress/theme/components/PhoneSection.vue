<script setup lang="ts">
// Nghe trên điện thoại, trên xe (wireframe WfLanding mục 4). Ảnh chụp thật có sẵn ở docs/images/.
import { withBase } from 'vitepress'
import { ArrowRight, Check, Download } from 'lucide-vue-next'
import imgLibrary from '../../../images/nghe-tren-dien-thoai/2-thu-vien.jpg'
import imgPlayer from '../../../images/nghe-tren-dien-thoai/3-trinh-phat.jpg'
import imgChapters from '../../../images/nghe-tren-dien-thoai/4-muc-luc-chuong.jpg'

const RELEASES = 'https://github.com/tanviet12/sano-sach-noi/releases/latest'
const shots = [
  { src: imgLibrary, alt: 'Sách nói Sano trong thư viện BookPlayer', cap: 'Thư viện' },
  { src: imgPlayer, alt: 'Đang nghe, hiện bìa và tên chương', cap: 'Đang nghe' },
  { src: imgChapters, alt: 'Mục lục chương của file M4B', cap: 'Mục lục chương' },
]
</script>

<template>
  <section id="dien-thoai" class="sano-section">
    <div class="sano-container wrap">
      <div>
        <h2 class="sano-h2">Nghe trên điện thoại, trên xe</h2>
        <p class="text sano-muted">
          Sano xuất cả cuốn thành <strong>một file M4B</strong> có mục lục chương, tên sách và bìa. Chép sang điện thoại là nghe
          được, không cần mạng. App nghe tự nhớ chỗ nghe dở, trên xe hiện tên chương.
        </p>
        <ul class="list">
          <li><Check :size="16" class="ck" aria-hidden="true" /><span><strong>iPhone:</strong> app BookPlayer (miễn phí, mã nguồn mở) hoặc app Sách của Apple</span></li>
          <li><Check :size="16" class="ck" aria-hidden="true" /><span><strong>Trên ô tô có CarPlay:</strong> BookPlayer hiện thư viện sách, chọn sách, chọn chương ngay trên màn hình xe</span></li>
          <li><Check :size="16" class="ck" aria-hidden="true" /><span><strong>Android:</strong> Voice (mã nguồn mở) hoặc app nghe sách nói bất kỳ đọc được M4B, có Android Auto</span></li>
          <li><Check :size="16" class="ck" aria-hidden="true" /><span>Khoảng 29 MB cho mỗi giờ nghe</span></li>
        </ul>
        <div class="actions">
          <a :href="RELEASES" target="_blank" rel="noopener" class="sano-btn brand block-sm"><Download :size="16" aria-hidden="true" /> Tải M4B mẫu</a>
          <a :href="withBase('/nghe-tren-dien-thoai')" class="sano-btn outline block-sm">Hướng dẫn chi tiết <ArrowRight :size="16" aria-hidden="true" /></a>
        </div>
        <p class="note sano-muted">File M4B mẫu (sách “Kỹ năng mềm cho người trẻ”, khoảng 4 phút 30 giây) đính kèm trong bản phát hành.</p>
      </div>
      <div class="shots">
        <figure v-for="s in shots" :key="s.cap">
          <div class="phone"><img :src="s.src" :alt="s.alt" /></div>
          <figcaption class="sano-muted">{{ s.cap }}</figcaption>
        </figure>
      </div>
    </div>
  </section>
</template>

<style scoped>
.wrap {
  display: grid;
  gap: 40px;
  align-items: center;
}
@media (min-width: 1024px) {
  .wrap {
    grid-template-columns: 1fr 1fr;
  }
}
.text {
  margin: 12px 0 0;
  line-height: 1.65;
}
.text strong {
  color: var(--vp-c-text-1);
}
.list {
  margin: 20px 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 14px;
  line-height: 1.5;
}
.list li {
  display: flex;
  gap: 8px;
}
.ck {
  margin-top: 2px;
  flex-shrink: 0;
  color: var(--vp-c-brand-1);
}
.actions {
  margin-top: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
@media (min-width: 640px) {
  .actions {
    flex-direction: row;
  }
}
.note {
  margin: 8px 0 0;
  font-size: 12px;
}
.shots {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}
@media (min-width: 640px) {
  .shots {
    gap: 16px;
  }
}
figure {
  margin: 0;
  text-align: center;
}
.phone {
  overflow: hidden;
  border: 4px solid hsl(var(--sano-foreground) / 0.8);
  border-radius: 1.4rem;
  background: #000;
  box-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.2);
}
.phone img {
  display: block;
  width: 100%;
}
figcaption {
  margin-top: 8px;
  font-size: 12px;
}
</style>
