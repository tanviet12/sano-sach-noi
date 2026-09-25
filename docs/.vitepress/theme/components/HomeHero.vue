<script setup lang="ts">
// Hero trang chủ (wireframe WfLanding mục 1). Nút tải đổi theo hệ điều hành người xem.
// Lúc build (SSR) chưa biết máy → hiện "Tải Sano"; vào trình duyệt mới nhận diện.
import { computed, onMounted, ref } from 'vue'
import { withBase } from 'vitepress'
import { Apple, BookOpen, Check, Download, ExternalLink, Laptop, Monitor } from 'lucide-vue-next'
import libraryShot from '../../../images/app/thu-vien.jpg'

const REPO = 'https://github.com/tanviet12/sano-sach-noi'

// Sano dùng để làm gì — cùng nội dung với trang Cài đặt, mục Giới thiệu trong app và README
const uses = [
  { t: 'Tạo sách nói', d: 'từ tài liệu Word của chính bạn' },
  { t: 'Nghe trên máy tính', d: 'ngay trong phần mềm, nhớ chỗ nghe dở' },
  { t: 'Nghe trên điện thoại', d: 'xuất một file M4B, nghe bằng app BookPlayer (miễn phí, có cho iPhone và Android)' },
  { t: 'Nghe khi lái xe ô tô', d: 'BookPlayer chạy trên CarPlay và Android Auto: chọn sách, chọn chương ngay trên màn hình xe' },
  { t: '25 giọng đọc AI tiếng Việt', d: 'nam, nữ, giọng Bắc, giọng Nam' },
]
const RELEASES = REPO + '/releases/latest'

type OS = 'mac' | 'win' | 'linux'
const builds: Record<OS, { label: string; file: string; note: string; icon: typeof Apple }> = {
  mac: { label: 'macOS', file: 'Sano-<phiên bản>-macos-universal.dmg', note: 'Apple Silicon và Intel', icon: Apple },
  win: { label: 'Windows', file: 'Sano-<phiên bản>-windows-amd64-setup.exe', note: 'Windows 10/11 · không cần quyền quản trị', icon: Monitor },
  linux: { label: 'Linux', file: 'Sano-<phiên bản>-linux-amd64.AppImage', note: 'x86_64 · cần WebKitGTK 4.1', icon: Laptop },
}

// null = chưa biết (SSR) · 'mobile' = điện thoại/máy tính bảng (Sano chỉ cài trên máy tính)
const os = ref<OS | 'mobile' | null>(null)
onMounted(() => {
  const nav = navigator as Navigator & { userAgentData?: { platform?: string; mobile?: boolean } }
  const platform = nav.userAgentData?.platform || nav.platform || ''
  const ua = nav.userAgent
  const touchMac = /Mac/.test(platform) && nav.maxTouchPoints > 1 // iPadOS báo là Mac
  if (nav.userAgentData?.mobile || /Android|iPhone|iPad|iPod/i.test(ua) || touchMac) os.value = 'mobile'
  else if (/Mac/i.test(platform) || /Macintosh/.test(ua)) os.value = 'mac'
  else if (/Win/i.test(platform) || /Windows/.test(ua)) os.value = 'win'
  else if (/Linux|X11|CrOS/i.test(platform + ua)) os.value = 'linux'
})

const build = computed(() => (os.value && os.value !== 'mobile' ? builds[os.value] : null))
const others = computed(() => (Object.keys(builds) as OS[]).filter((k) => k !== os.value))
</script>

<template>
  <section id="tai-ve" class="hero sano-ambient">
    <div class="wrap">
      <div>
        <span class="badge"><span class="dot" /> Mã nguồn mở · Miễn phí</span>
        <h1 class="h1">Tạo <span class="accent">sách nói bằng AI</span> từ file Word</h1>
        <p class="lead">
          Biến tài liệu của chính bạn thành sách nói. Giọng đọc AI tiếng Việt chạy ngay trên máy, miễn phí, mã nguồn mở.
          Nghe trong phần mềm, hoặc chép một file sang điện thoại nghe khi lái xe, lúc rảnh tay.
        </p>
        <ul class="uses">
          <li v-for="u in uses" :key="u.t">
            <Check :size="16" class="ck" aria-hidden="true" /><span><strong>{{ u.t }}</strong> {{ u.d }}</span>
          </li>
        </ul>

        <div class="actions">
          <a :href="RELEASES" target="_blank" rel="noopener" class="sano-btn brand block-sm dl">
            <Download :size="20" aria-hidden="true" />
            <span class="dl-text">
              <span class="dl-main">{{ build ? `Tải cho ${build.label}` : os === 'mobile' ? 'Tải cho máy tính' : 'Tải Sano' }}</span>
              <span class="dl-note">{{ build ? build.note : 'Windows · macOS · Linux' }}</span>
            </span>
          </a>
          <a :href="withBase('/cai-dat')" class="sano-btn outline block-sm big">
            <BookOpen :size="16" aria-hidden="true" /> Xem hướng dẫn
          </a>
        </div>

        <p class="others">
          <template v-if="build">
            <span>Bản khác:</span>
            <a v-for="k in others" :key="k" :href="RELEASES" target="_blank" rel="noopener" class="os-link">
              <component :is="builds[k].icon" :size="14" aria-hidden="true" /> {{ builds[k].label }}
            </a>
            <span class="sep">·</span>
          </template>
          <a :href="REPO + '/releases'" target="_blank" rel="noopener" class="all">Mọi phiên bản <ExternalLink :size="12" aria-hidden="true" /></a>
        </p>
        <p class="file">
          <template v-if="os === 'mobile'">Sano cài trên máy tính Windows, macOS, Linux; nghe trên điện thoại bằng file M4B.</template>
          <template v-else-if="build">File <code>{{ build.file }}</code> trên GitHub Releases.</template>
          <template v-else>Bản cài trên GitHub Releases.</template>
          Bản cài chưa ký số —
          <a :href="withBase('/mo-app-lan-dau')">cách mở lần đầu</a>.
        </p>
      </div>

      <div class="shot">
        <div class="window">
          <div class="bar"><span class="r" /><span class="a" /><span class="g" /></div>
          <img :src="libraryShot" alt="Thư viện sách nói trong phần mềm Sano" class="body" width="1600" height="955" />
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero {
  border-bottom: 1px solid var(--vp-c-divider);
}
.wrap {
  margin: 0 auto;
  max-width: 1152px;
  padding: 48px 16px;
  display: grid;
  gap: 40px;
  align-items: center;
}
@media (min-width: 640px) {
  .wrap {
    padding-left: 24px;
    padding-right: 24px;
  }
}
@media (min-width: 1024px) {
  .wrap {
    grid-template-columns: 1fr 1.1fr;
    padding-top: 80px;
    padding-bottom: 80px;
  }
}
.badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 9999px;
  padding: 2px 10px;
  font-size: 12px;
  font-weight: 600;
  background: color-mix(in srgb, var(--vp-c-bg) 70%, transparent);
}
.dot {
  width: 6px;
  height: 6px;
  border-radius: 9999px;
  background: var(--vp-c-brand-1);
}
.h1 {
  margin: 16px 0 0;
  font-size: 36px;
  line-height: 1.15;
  font-weight: 600;
  letter-spacing: -0.025em;
  text-wrap: balance;
}
@media (min-width: 640px) {
  .h1 {
    font-size: 48px;
  }
}
.accent {
  color: var(--vp-c-brand-1);
}
.lead {
  margin: 16px 0 0;
  max-width: 36rem;
  font-size: 16px;
  line-height: 1.65;
  color: var(--vp-c-text-2);
}
@media (min-width: 640px) {
  .lead {
    font-size: 18px;
  }
}
.uses {
  margin: 20px 0 0;
  padding: 0;
  max-width: 36rem;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 15px;
  line-height: 1.55;
  color: var(--vp-c-text-2);
}
.uses li {
  display: flex;
  gap: 8px;
}
.uses strong {
  color: var(--vp-c-text-1);
  font-weight: 600;
}
.ck {
  margin-top: 3px;
  flex-shrink: 0;
  color: var(--vp-c-brand-1);
}
.actions {
  margin-top: 28px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
@media (min-width: 640px) {
  .actions {
    flex-direction: row;
    align-items: center;
  }
}
.dl {
  padding: 12px 24px;
}
.dl-text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  line-height: 1.25;
}
.dl-main {
  font-size: 16px;
}
.dl-note {
  font-size: 12px;
  font-weight: 400;
  opacity: 0.85;
}
.big {
  min-height: 44px;
  padding: 0 32px;
}
.others {
  margin: 16px 0 0;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 12px;
  font-size: 14px;
  color: var(--vp-c-text-2);
}
.os-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
  color: var(--vp-c-text-1);
  text-decoration: none;
}
.os-link:hover,
.all:hover {
  color: var(--vp-c-brand-1);
}
.all {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--vp-c-text-2);
  text-decoration: none;
}
@media (max-width: 639px) {
  .sep {
    display: none;
  }
}
.file {
  margin: 8px 0 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--vp-c-text-2);
}
.file code {
  border-radius: 4px;
  background: var(--vp-c-bg-soft);
  padding: 1px 4px;
  font-size: 11px;
}
.file a {
  color: inherit;
  text-decoration: underline;
}
.file a:hover {
  color: var(--vp-c-brand-1);
}

.shot {
  position: relative;
}
.window {
  overflow: hidden;
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  background: var(--vp-c-bg);
  box-shadow: 0 25px 50px -12px rgb(0 0 0 / 0.25);
}
.bar {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 12px;
  border-bottom: 1px solid var(--vp-c-divider);
  background: hsl(var(--sano-muted) / 0.6);
}
.bar span {
  width: 10px;
  height: 10px;
  border-radius: 9999px;
}
.bar .r {
  background: hsl(var(--sano-red));
}
.bar .a {
  background: hsl(var(--sano-amber));
}
.bar .g {
  background: hsl(var(--sano-green));
}
.body {
  display: block;
  width: 100%;
  height: auto;
}
</style>
