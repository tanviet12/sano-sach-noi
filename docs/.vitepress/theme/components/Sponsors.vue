<script setup lang="ts">
// Đơn vị tài trợ (wireframe WfLanding mục 8, trái). Logo dùng chung file với app.
import { ExternalLink } from 'lucide-vue-next'
import sepayLogo from '../../../../desktop/frontend/src/assets/sponsors/sepay.svg'
import hostLogo from '../../../../desktop/frontend/src/assets/sponsors/123host.svg'

const utm = '?utm_source=sano&utm_medium=docs&utm_campaign=tai-tro'
const sponsors = [
  {
    name: 'SePay',
    logo: sepayLogo,
    url: 'https://sepay.vn' + utm,
    site: 'sepay.vn',
    desc: 'Nền tảng Open Banking: tự động xác nhận thanh toán chuyển khoản, kết nối API với các ngân hàng Việt Nam cho website và phần mềm bán hàng.',
  },
  {
    name: '123HOST',
    logo: hostLogo,
    url: 'https://123host.vn' + utm,
    site: '123host.vn',
    desc: 'Hosting, VPS, máy chủ và tên miền cho doanh nghiệp, nhà phát triển Việt Nam.',
  },
]
</script>

<template>
  <div>
    <h2 class="sano-h2">Đơn vị tài trợ</h2>
    <p class="sano-lead">Sano miễn phí và mã nguồn mở nhờ sự tài trợ của</p>
    <div class="grid">
      <a v-for="sp in sponsors" :key="sp.name" :href="sp.url" target="_blank" rel="noopener" class="sano-card card sano-focus">
        <span class="logo"><img :src="sp.logo" :alt="sp.name" /></span>
        <span class="d sano-muted">{{ sp.desc }}</span>
        <span class="site">{{ sp.site }} <ExternalLink :size="14" aria-hidden="true" /></span>
      </a>
    </div>
  </div>
</template>

<style scoped>
.grid {
  margin-top: 24px;
  display: grid;
  gap: 16px;
}
@media (min-width: 640px) {
  .grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
.card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 20px;
  text-decoration: none;
  color: inherit;
  transition: border-color 0.15s;
}
.card:hover {
  border-color: hsl(var(--sano-primary) / 0.5);
}
.logo {
  display: grid;
  place-items: center;
  height: 64px;
  padding: 0 16px;
  border-radius: 6px;
  background: #fff;
}
.logo img {
  max-height: 40px;
  max-width: 170px;
  object-fit: contain;
}
.d {
  flex: 1;
  font-size: 14px;
  line-height: 1.6;
}
.site {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  font-weight: 500;
  color: var(--vp-c-brand-1);
}
</style>
