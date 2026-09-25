<script setup lang="ts">
// Giới thiệu: Sano là gì, tài trợ, dự án cùng tác giả, điều khoản sử dụng, hướng dẫn, báo lỗi, mã nguồn, giấy phép.
import { computed, ref } from 'vue'
import { Bug, Check, ExternalLink, FileText, Github, LifeBuoy } from 'lucide-vue-next'
import logoUrl from '@/assets/logo.svg'
import ThirdPartyLicenses from '../components/ThirdPartyLicenses.vue'
import TermsDialog from '../components/TermsDialog.vue'
import { AUTHOR_FB, DOCS, REPO } from '../lib/mock'
import { state } from '../lib/store'
import { TERMS_VERSION } from '../lib/terms'
import sepayLogo from '../assets/sponsors/sepay.svg'
import hostLogo from '../assets/sponsors/123host.svg'

// Sano dùng để làm gì — cùng nội dung với trang chủ, trang Cài đặt và README
const uses = [
  { t: 'Tạo sách nói', d: 'từ tài liệu Word của chính bạn' },
  { t: 'Nghe trên máy tính', d: 'ngay trong phần mềm, nhớ chỗ nghe dở' },
  { t: 'Nghe trên điện thoại', d: 'xuất một file M4B, nghe bằng app BookPlayer (miễn phí, có cho iPhone và Android)' },
  { t: 'Nghe trên ô tô', d: 'BookPlayer chạy trên CarPlay và Android Auto: chọn sách, chọn chương ngay trên màn hình xe' },
  { t: '25 giọng đọc AI tiếng Việt', d: 'nam, nữ, giọng Bắc, giọng Nam' },
]

// Đơn vị tài trợ: Sano miễn phí nhờ hai đơn vị này. utm để biết lượt ghé đến từ app.
const utm = '?utm_source=sano&utm_medium=app&utm_campaign=tai-tro'
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

// Dự án mã nguồn mở khác của cùng tác giả. vbsec lên trước vì Sano được quét bằng vbsec.
const projects = [
  {
    name: 'vbsec',
    url: 'https://github.com/tanviet12/vbsec',
    desc: 'Quét bảo mật mã nguồn bằng AI, phát hiện hơn 20 loại lỗ hổng phổ biến. Sano được kiểm tra bằng vbsec.',
  },
  {
    name: 'Chat Quality Agent',
    url: 'https://tanviet12.github.io/chat-quality-agent/',
    desc: 'Dùng AI chấm chất lượng chăm sóc khách hàng qua Zalo OA, Facebook Messenger và gửi cảnh báo.',
  },
]

const showTerms = ref(false)
const termsLine = computed(() => {
  const t = state.terms
  if (!t?.acceptedAt || t.acceptedVersion > 1e9) return `Phiên bản ${TERMS_VERSION}`
  return `Đã đồng ý ngày ${new Date(t.acceptedAt).toLocaleDateString('vi-VN')} · phiên bản ${t.acceptedVersion}`
})
</script>

<template>
  <section class="flex-1 overflow-auto p-6 max-w-2xl">
    <h1 class="text-xl font-semibold tracking-tight">Giới thiệu</h1>
    <p class="text-sm text-muted-foreground">Phiên bản {{ state.version || '…' }}</p>
    <div class="mt-6">
      <div class="rounded-lg border border-border p-4 flex gap-4">
        <img :src="logoUrl" alt="" class="h-12 w-12 rounded-xl shrink-0" />
        <div class="text-sm min-w-0">
          <p><span class="font-medium">Sano</span> — biến tài liệu của chính bạn thành sách nói, chạy trên máy bạn. Mã nguồn mở, miễn phí.</p>
          <p class="mt-2 text-muted-foreground">
            Do <a :href="AUTHOR_FB" target="_blank" rel="noopener" class="text-foreground font-medium hover:underline">Bùi Tấn Việt</a> làm — CEO
            <a href="https://sepay.vn" target="_blank" rel="noopener" class="text-foreground hover:underline">SePay</a> và
            <a href="https://123host.vn" target="_blank" rel="noopener" class="text-foreground hover:underline">123HOST</a>.
          </p>
        </div>
      </div>

      <h2 class="mt-6 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Sano làm được gì</h2>
      <ul class="mt-3 rounded-lg border border-border divide-y divide-border text-sm">
        <li v-for="u in uses" :key="u.t" class="flex items-start gap-2.5 px-4 py-3">
          <Check class="w-4 h-4 mt-0.5 shrink-0 text-primary" />
          <span><span class="font-medium">{{ u.t }}</span> <span class="text-muted-foreground">{{ u.d }}</span></span>
        </li>
      </ul>

      <h2 class="mt-6 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Đơn vị tài trợ</h2>
      <p class="mt-1 text-sm text-muted-foreground">Sano miễn phí và mã nguồn mở nhờ sự tài trợ của</p>
      <div class="mt-3 grid grid-cols-1 sm:grid-cols-2 gap-3">
        <a v-for="sp in sponsors" :key="sp.name" :href="sp.url" target="_blank" rel="noopener"
          class="group rounded-lg border border-border p-4 flex flex-col gap-3 hover:border-primary/50 hover:bg-muted/30 transition-colors">
          <div class="h-14 rounded-md bg-white grid place-items-center px-4">
            <img :src="sp.logo" :alt="sp.name" class="max-h-10 max-w-[170px] object-contain" />
          </div>
          <p class="text-sm text-muted-foreground leading-relaxed flex-1">{{ sp.desc }}</p>
          <span class="text-sm font-medium text-primary flex items-center gap-1">{{ sp.site }} <ExternalLink class="w-3.5 h-3.5" /></span>
        </a>
      </div>

      <h2 class="mt-8 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Dự án cùng tác giả</h2>
      <div class="mt-3 rounded-lg border border-border divide-y divide-border text-sm">
        <a v-for="p in projects" :key="p.name" :href="p.url" target="_blank" rel="noopener" class="flex items-start gap-3 px-4 py-3 hover:bg-muted/50">
          <div class="flex-1 min-w-0">
            <p class="font-medium">{{ p.name }}</p>
            <p class="mt-0.5 text-muted-foreground leading-relaxed">{{ p.desc }}</p>
          </div>
          <ExternalLink class="w-3.5 h-3.5 mt-1 shrink-0 text-muted-foreground" />
        </a>
      </div>

      <h2 class="mt-8 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Giọng đọc</h2>
      <a href="https://github.com/pnnbao97/VieNeu-TTS" target="_blank" rel="noopener" class="mt-3 flex items-start gap-3 rounded-lg border border-border px-4 py-3 text-sm hover:bg-muted/50">
        <div class="flex-1 min-w-0">
          <p class="font-medium">VieNeu-TTS</p>
          <p class="mt-0.5 text-muted-foreground leading-relaxed">
            Mọi giọng đọc trong Sano do VieNeu-TTS tạo — dự án mã nguồn mở chuyển văn bản thành giọng nói tiếng Việt của
            Pham Nguyen Ngoc Bao, giấy phép Apache-2.0. Cảm ơn tác giả đã chia sẻ.
          </p>
        </div>
        <ExternalLink class="w-3.5 h-3.5 mt-1 shrink-0 text-muted-foreground" />
      </a>

      <h2 class="mt-8 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Thông tin</h2>
      <div class="mt-3 rounded-lg border border-border divide-y divide-border text-sm">
        <a :href="DOCS" target="_blank" rel="noopener" class="flex items-center gap-2.5 px-4 py-3 hover:bg-muted/50"><LifeBuoy class="w-4 h-4 text-muted-foreground" /><span class="flex-1">Hướng dẫn sử dụng</span><ExternalLink class="w-3.5 h-3.5 text-muted-foreground" /></a>
        <a :href="REPO + '/issues'" target="_blank" rel="noopener" class="flex items-center gap-2.5 px-4 py-3 hover:bg-muted/50"><Bug class="w-4 h-4 text-muted-foreground" /><span class="flex-1">Báo lỗi / góp ý</span><ExternalLink class="w-3.5 h-3.5 text-muted-foreground" /></a>
        <a :href="REPO" target="_blank" rel="noopener" class="flex items-center gap-2.5 px-4 py-3 hover:bg-muted/50"><Github class="w-4 h-4 text-muted-foreground" /><span class="flex-1">Mã nguồn trên GitHub · giấy phép MIT</span><ExternalLink class="w-3.5 h-3.5 text-muted-foreground" /></a>
        <button type="button" class="w-full flex items-center gap-2.5 px-4 py-3 hover:bg-muted/50 text-left" @click="showTerms = true">
          <FileText class="w-4 h-4 text-muted-foreground" /><span class="flex-1">Điều khoản sử dụng</span>
          <span class="text-xs text-muted-foreground flex items-center gap-1"><Check v-if="state.terms?.acceptedAt" class="w-3.5 h-3.5 text-rag-green" />{{ termsLine }}</span>
        </button>
        <ThirdPartyLicenses />
      </div>
    </div>
    <TermsDialog v-if="showTerms" @close="showTerms = false" />
  </section>
</template>
