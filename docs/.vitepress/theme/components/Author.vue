<script setup lang="ts">
// Tác giả (wireframe WfLanding mục 8, phải). Ảnh đại diện: chữ viết tắt, chưa có ảnh thật.
import { ExternalLink } from 'lucide-vue-next'

const AUTHOR_FB = 'https://www.facebook.com/buitanviet'
// Dự án mã nguồn mở khác của tác giả. vbsec lên trước vì Sano được quét bằng vbsec.
const projects = [
  { name: 'vbsec', url: 'https://github.com/tanviet12/vbsec', desc: 'Quét bảo mật mã nguồn bằng AI. Sano được kiểm tra bằng vbsec.' },
  { name: 'Chat Quality Agent', url: 'https://tanviet12.github.io/chat-quality-agent/', desc: 'Dùng AI chấm chất lượng chăm sóc khách hàng qua Zalo OA, Facebook Messenger.' },
]
</script>

<template>
  <div>
    <h2 class="sano-h2">Tác giả</h2>
    <div class="sano-card box">
      <div class="head">
        <span class="avatar" aria-hidden="true">BV</span>
        <div>
          <a :href="AUTHOR_FB" target="_blank" rel="noopener" class="name">Bùi Tấn Việt <ExternalLink :size="14" aria-hidden="true" /></a>
          <p class="role sano-muted">CEO SePay và 123HOST</p>
        </div>
      </div>
      <p class="quote sano-muted">
        “Tôi học liên tục và muốn tranh thủ nghe lại tài liệu của mình lúc lái xe, lúc rảnh tay. Ngoài thị trường có nhiều app
        sách nói, nhưng không có cái nào đọc được tài liệu riêng của mình. Tôi tự làm một công cụ để dùng, thấy hữu ích nên mở mã
        nguồn cho ai cần.”
      </p>
      <p class="more-title">Dự án cùng tác giả</p>
      <ul class="more">
        <li v-for="p in projects" :key="p.name">
          <a :href="p.url" target="_blank" rel="noopener" class="more-name">{{ p.name }} <ExternalLink :size="12" aria-hidden="true" /></a>
          <span class="sano-muted"> — {{ p.desc }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.box {
  margin-top: 24px;
  padding: 20px;
}
.head {
  display: flex;
  align-items: center;
  gap: 12px;
}
.avatar {
  display: grid;
  place-items: center;
  width: 48px;
  height: 48px;
  flex-shrink: 0;
  border-radius: 9999px;
  background: var(--vp-button-brand-bg);
  color: #fff;
  font-weight: 600;
}
.name {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-weight: 600;
  color: var(--vp-c-text-1);
  text-decoration: none;
}
.name:hover {
  color: var(--vp-c-brand-1);
}
.role {
  margin: 0;
  font-size: 14px;
}
.quote {
  margin: 16px 0 0;
  font-size: 14px;
  line-height: 1.65;
}
.more-title {
  margin: 16px 0 0;
  padding-top: 16px;
  border-top: 1px solid var(--vp-c-divider);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--vp-c-text-2);
}
.more {
  margin: 8px 0 0;
  padding: 0;
  list-style: none;
  font-size: 14px;
  line-height: 1.6;
}
.more li + li {
  margin-top: 6px;
}
.more-name {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-weight: 600;
  color: var(--vp-c-text-1);
  text-decoration: none;
}
.more-name:hover {
  color: var(--vp-c-brand-1);
}
</style>
