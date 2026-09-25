<script setup lang="ts">
// Trang Tải về (wireframe WfDownload): 3 thẻ lớn Windows · macOS · Linux, thẻ đúng máy người xem
// có nhãn "Khuyên dùng cho máy bạn"; sau đó 3 bước sau khi tải, máy cần có, file khác.
// Link, phiên bản, dung lượng, ngày phát hành theo bản mới nhất trên GitHub (release.ts).
import { computed, ref } from 'vue'
import { withBase } from 'vitepress'
import {
  Apple, ArrowRight, BookOpen, Check, Copy, Cpu, Download, ExternalLink, FileAudio, FileCode2,
  HardDrive, Laptop, MemoryStick, Monitor, ShieldCheck, Smartphone, Wifi,
} from 'lucide-vue-next'
import { REPO, useRelease, type Asset } from './release'
import { useVisitorOS, type OS } from './visitorOS'

const { version, links, sizes, published } = useRelease()
const os = useVisitorOS()

const cards: { id: OS; name: string; icon: typeof Apple; req: string; asset: Asset; file: string; hint: string; alt?: { asset: Asset; label: string; hint: string } }[] = [
  {
    id: 'win', name: 'Windows', icon: Monitor, req: 'Windows 10 / 11 · 64-bit',
    asset: 'win', file: 'bộ cài .exe', hint: 'Cài vào thư mục của bạn, không cần quyền quản trị',
    alt: { asset: 'win-zip', label: 'Bản portable .zip', hint: 'giải nén là chạy, không cần cài' },
  },
  { id: 'mac', name: 'macOS', icon: Apple, req: 'macOS 10.13 trở lên · Apple Silicon và Intel', asset: 'mac', file: 'file .dmg', hint: 'Mở file, kéo Sano vào thư mục Applications' },
  { id: 'linux', name: 'Linux', icon: Laptop, req: 'x86_64 · Ubuntu 22.04+, Debian 12+, Fedora 36+', asset: 'linux', file: 'file .AppImage', hint: 'Cấp quyền chạy rồi mở, cần WebKitGTK 4.1' },
]

function mb(a: Asset) {
  const b = sizes.value[a]
  if (!b) return ''
  const v = b / 1e6
  return (v >= 10 ? Math.round(v).toString() : v.toFixed(1).replace('.', ',')) + ' MB'
}
const releasedOn = computed(() => {
  if (!published.value) return ''
  const d = new Date(published.value)
  return `${String(d.getDate()).padStart(2, '0')}/${String(d.getMonth() + 1).padStart(2, '0')}/${d.getFullYear()}`
})

const install = computed(() => {
  switch (os.value) {
    case 'mac': return 'Mở file .dmg vừa tải, kéo biểu tượng Sano vào thư mục Applications.'
    case 'linux': return 'Cấp quyền chạy cho file .AppImage (chmod +x) rồi mở. Thiếu WebKitGTK thì cài libwebkit2gtk-4.1-0.'
    case 'win': return 'Chạy file .exe vừa tải, bấm Tiếp tục đến hết. Sano có biểu tượng ở menu Start.'
    default: return 'Mở file cài trên máy tính: .exe (Windows), .dmg (macOS) hoặc .AppImage (Linux).'
  }
})

const copied = ref(false)
async function copyLink() {
  try {
    await navigator.clipboard.writeText(window.location.href)
    copied.value = true
  } catch {
    // trình duyệt chặn clipboard: người xem tự chép từ thanh địa chỉ
  }
}
</script>

<template>
  <section class="dl-hero sano-ambient">
    <div class="sano-container">
      <p class="eyebrow">Miễn phí · Mã nguồn mở</p>
      <h1 class="h1">Tải Sano</h1>
      <p class="lead">Một phần mềm, chạy trên cả <strong>Windows, macOS và Linux</strong>. Chọn đúng máy của bạn, bấm là tải.</p>
      <p class="meta">
        <span class="pill">Bản mới nhất {{ version }}</span>
        <span v-if="releasedOn">phát hành {{ releasedOn }}</span>
        <span v-if="releasedOn" aria-hidden="true">·</span>
        <a :href="withBase('/nhat-ky-thay-doi')" class="link">Có gì mới <ArrowRight :size="14" aria-hidden="true" /></a>
      </p>

      <div v-if="os === 'mobile'" class="sano-card mobile">
        <Smartphone :size="24" class="mobile-icon" aria-hidden="true" />
        <p>Bạn đang xem trên điện thoại. Sano cài trên máy tính; nghe trên điện thoại bằng file M4B xuất từ Sano. Gửi link trang này sang máy tính để tải.</p>
        <button type="button" class="sano-btn outline" @click="copyLink">
          <component :is="copied ? Check : Copy" :size="16" aria-hidden="true" /> {{ copied ? 'Đã sao chép' : 'Sao chép link' }}
        </button>
      </div>

      <div class="cards">
        <div v-for="c in cards" :key="c.id" class="os-card" :class="{ on: os === c.id }">
          <span v-if="os === c.id" class="badge"><Check :size="14" aria-hidden="true" /> Khuyên dùng cho máy bạn</span>
          <div class="os-icon"><component :is="c.icon" :size="32" aria-hidden="true" /></div>
          <h2 class="os-name">{{ c.name }}</h2>
          <p class="req">{{ c.req }}</p>
          <a :href="links[c.asset].url" class="sano-btn dl-btn" :class="os === c.id ? 'brand' : 'outline'">
            <Download :size="20" aria-hidden="true" /> Tải {{ c.file }}<span v-if="mb(c.asset)" class="size">· {{ mb(c.asset) }}</span>
          </a>
          <p class="hint">{{ c.hint }}</p>
          <p v-if="c.alt" class="alt">
            <a :href="links[c.alt.asset].url" class="link">{{ c.alt.label }}</a>
            <span class="sano-muted"><template v-if="mb(c.alt.asset)"> · {{ mb(c.alt.asset) }}</template> · {{ c.alt.hint }}</span>
          </p>
        </div>
      </div>

      <p class="trust">
        <ShieldCheck :size="16" class="ok" aria-hidden="true" />
        <span>
          File cài build tự động trên GitHub từ mã nguồn công khai, kèm SHA256SUMS. Chưa ký số nên lần đầu mở máy sẽ cảnh báo —
          <a :href="withBase('/mo-app-lan-dau')" class="link">cách mở lần đầu</a>.
        </span>
      </p>
    </div>
  </section>

  <section class="sano-section">
    <div class="sano-container">
      <h2 class="sano-h2">Sau khi tải</h2>
      <ol class="grid3 steps">
        <li class="sano-card box">
          <span class="num">1</span>
          <h3>Cài Sano</h3>
          <p class="sano-muted">{{ install }}</p>
        </li>
        <li class="sano-card box">
          <span class="num">2</span>
          <h3>Mở lần đầu</h3>
          <p class="sano-muted">Windows hoặc macOS có thể hỏi lại vì bản cài chưa ký số. <a :href="withBase('/mo-app-lan-dau')" class="link">Xem cách mở</a>.</p>
        </li>
        <li class="sano-card box">
          <span class="num">3</span>
          <h3>Sano tự cài bộ đọc</h3>
          <p class="sano-muted">Tải khoảng 1 GB một lần (mô hình giọng đọc AI). Xong là dùng được không cần mạng. Thư viện có sẵn một cuốn mẫu để nghe thử.</p>
        </li>
      </ol>
      <a :href="withBase('/cai-dat')" class="link more"><BookOpen :size="16" aria-hidden="true" /> Hướng dẫn cài đặt đầy đủ</a>
    </div>
  </section>

  <section class="sano-section">
    <div class="sano-container">
      <h2 class="sano-h2">Máy cần có</h2>
      <div class="grid4">
        <div class="sano-card box"><MemoryStick :size="20" class="ic" aria-hidden="true" /><p class="t">RAM 4 GB trở lên</p><p class="sano-muted">Ít hơn vẫn chạy, đọc chậm hơn</p></div>
        <div class="sano-card box"><HardDrive :size="20" class="ic" aria-hidden="true" /><p class="t">Ổ trống 2,5 GB</p><p class="sano-muted">Cài xong bộ đọc chiếm khoảng 1,5 GB</p></div>
        <div class="sano-card box"><Cpu :size="20" class="ic" aria-hidden="true" /><p class="t">Không cần card đồ hoạ</p><p class="sano-muted">Giọng đọc AI chạy bằng CPU</p></div>
        <div class="sano-card box"><Wifi :size="20" class="ic" aria-hidden="true" /><p class="t">Mạng cho lần đầu</p><p class="sano-muted">Sau đó dùng không cần mạng</p></div>
      </div>
    </div>
  </section>

  <section class="sano-section last">
    <div class="sano-container">
      <h2 class="sano-h2">File khác</h2>
      <div class="grid3">
        <a :href="links.m4b.url" class="sano-card box tile">
          <FileAudio :size="20" class="ic" aria-hidden="true" />
          <p class="t">Sách nói mẫu M4B<template v-if="mb('m4b')"> · {{ mb('m4b') }}</template></p>
          <p class="sano-muted">"Kỹ năng mềm cho người trẻ", giọng Hải Đăng. Chép sang điện thoại nghe thử trước khi cài.</p>
        </a>
        <a :href="REPO + '/releases'" target="_blank" rel="noopener" class="sano-card box tile">
          <ExternalLink :size="20" class="ic" aria-hidden="true" />
          <p class="t">Các bản trước + SHA256SUMS</p>
          <p class="sano-muted">Mọi phiên bản trên GitHub Releases, kèm mã kiểm tra file.</p>
        </a>
        <a :href="REPO" target="_blank" rel="noopener" class="sano-card box tile">
          <FileCode2 :size="20" class="ic" aria-hidden="true" />
          <p class="t">Mã nguồn + dòng lệnh</p>
          <p class="sano-muted">Tự build từ mã nguồn, hoặc dùng công cụ dòng lệnh sano-docx2tts.</p>
        </a>
      </div>
    </div>
  </section>
</template>

<style scoped>
.dl-hero {
  border-bottom: 1px solid var(--vp-c-divider);
}
.eyebrow {
  margin: 0;
  font-size: 14px;
  font-weight: 500;
  color: var(--vp-c-brand-1);
}
.h1 {
  margin: 8px 0 0;
  font-size: 36px;
  line-height: 1.15;
  font-weight: 600;
  letter-spacing: -0.025em;
}
.lead {
  margin: 12px 0 0;
  max-width: 42rem;
  font-size: 18px;
  line-height: 1.6;
  color: var(--vp-c-text-2);
}
.lead strong {
  color: var(--vp-c-text-1);
}
.meta {
  margin: 12px 0 0;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 8px;
  font-size: 14px;
  color: var(--vp-c-text-2);
}
.pill {
  border-radius: 999px;
  padding: 2px 10px;
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand-1);
  font-weight: 500;
}
.link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
  color: var(--vp-c-brand-1);
  text-decoration: none;
}
.link:hover {
  text-decoration: underline;
}
.mobile {
  margin-top: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  font-size: 14px;
}
.mobile p {
  margin: 0;
  flex: 1;
}
.mobile-icon {
  color: var(--vp-c-brand-1);
  flex-shrink: 0;
}
.cards {
  margin-top: 32px;
  display: grid;
  gap: 20px;
}
.os-card {
  position: relative;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--vp-c-divider);
  border-radius: 16px;
  background: hsl(var(--sano-card));
  padding: 24px;
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.04);
}
.os-card.on {
  border-color: var(--vp-c-brand-1);
  box-shadow: 0 0 0 3px hsl(var(--sano-primary) / 0.2), 0 4px 12px rgb(0 0 0 / 0.06);
}
.badge {
  position: absolute;
  top: -12px;
  left: 24px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border-radius: 999px;
  padding: 4px 12px;
  background: var(--vp-button-brand-bg);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.15);
}
.os-icon {
  display: grid;
  place-items: center;
  width: 64px;
  height: 64px;
  border-radius: 16px;
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand-1);
}
.os-card.on .os-icon {
  background: var(--vp-button-brand-bg);
  color: #fff;
}
.os-name {
  margin: 16px 0 0;
  padding: 0;
  border: 0;
  font-size: 24px;
  font-weight: 600;
  letter-spacing: -0.02em;
}
.req {
  margin: 4px 0 0;
  min-height: 40px;
  font-size: 14px;
  line-height: 1.4;
  color: var(--vp-c-text-2);
}
.dl-btn {
  margin-top: 20px;
  min-height: 48px;
  font-size: 16px;
}
.size {
  font-weight: 400;
  opacity: 0.8;
}
.hint {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--vp-c-text-2);
}
.alt {
  margin: 16px 0 0;
  padding-top: 12px;
  border-top: 1px solid var(--vp-c-divider);
  font-size: 14px;
}
.trust {
  margin: 24px 0 0;
  display: flex;
  gap: 8px;
  align-items: flex-start;
  font-size: 14px;
  color: var(--vp-c-text-2);
}
.trust .ok {
  margin-top: 3px;
  flex-shrink: 0;
  color: hsl(var(--sano-green));
}
.grid3,
.grid4 {
  margin: 24px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 16px;
}
.box {
  padding: 20px;
  margin: 0;
}
.box h3 {
  margin: 12px 0 0;
  font-size: 16px;
  font-weight: 600;
}
.box p {
  margin: 4px 0 0;
  font-size: 14px;
  line-height: 1.5;
}
.num {
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  border-radius: 999px;
  background: var(--vp-button-brand-bg);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
}
.ic {
  color: var(--vp-c-brand-1);
}
.box .t {
  margin-top: 8px;
  font-weight: 500;
  color: var(--vp-c-text-1);
}
.tile {
  display: block;
  color: inherit;
  text-decoration: none;
  transition: border-color 0.15s;
}
.tile:hover {
  border-color: hsl(var(--sano-primary) / 0.5);
}
.tile:hover .t {
  color: var(--vp-c-brand-1);
}
.more {
  margin-top: 20px;
  font-size: 14px;
}
.last {
  border-bottom: 0;
}
@media (min-width: 640px) {
  .h1 {
    font-size: 44px;
  }
  .mobile {
    flex-direction: row;
    align-items: center;
  }
  .grid4 {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (min-width: 768px) {
  .cards,
  .grid3 {
    grid-template-columns: repeat(3, 1fr);
  }
}
@media (min-width: 1024px) {
  .grid4 {
    grid-template-columns: repeat(4, 1fr);
  }
}
</style>
