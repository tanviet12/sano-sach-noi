<script setup lang="ts">
// Trang tải riêng /tai-ve (wireframe): 3 thẻ lớn Windows · macOS · Linux cho thấy ngay Sano chạy
// trên cả 3; thẻ đúng máy người xem có nhãn "Khuyên dùng cho máy bạn". Sau đó: 3 bước sau khi
// tải, máy cần có, file khác. Wireframe tĩnh, KHÔNG gọi API. Mở ở bản dev: ?wireframe=download
import { computed, ref, watch } from 'vue'
import { Apple, ArrowRight, BookOpen, Check, Copy, Cpu, Download, ExternalLink, FileAudio, FileCode2, HardDrive, Laptop, MemoryStick, Monitor, ShieldCheck, Smartphone, Wifi } from 'lucide-vue-next'
import WfSiteNav from './WfSiteNav.vue'
import WfSiteReviewBar from './WfSiteReviewBar.vue'

const q = new URLSearchParams(window.location.search)
const dark = ref(q.get('theme') === 'dark')
watch(dark, (v) => document.documentElement.classList.toggle('dark', v), { immediate: true })
const notes = ref(q.get('notes') !== '0')
const noteCls = 'inline-flex items-center gap-1 rounded border border-dashed border-amber-500/70 bg-amber-50 px-1.5 py-0.5 text-[11px] font-medium text-amber-800 dark:bg-amber-950/40 dark:text-amber-300'

// Giả lập máy người xem (trang thật tự nhận diện như nút tải ở trang chủ)
type OS = 'win' | 'mac' | 'linux'
const seen = ref<OS | 'mobile'>('win')
const seenOpts = [['win', 'Windows'], ['mac', 'macOS'], ['linux', 'Linux'], ['mobile', 'Điện thoại']] as const

const version = '0.1.1'
const released = '25/09/2026'
const cards: { id: OS; name: string; icon: typeof Apple; req: string; file: string; size: string; hint: string; alt?: { label: string; size: string; hint: string } }[] = [
  {
    id: 'win', name: 'Windows', icon: Monitor, req: 'Windows 10 / 11 · 64-bit',
    file: 'bộ cài .exe', size: '9,8 MB', hint: 'Cài vào thư mục của bạn, không cần quyền quản trị',
    alt: { label: 'Bản portable .zip', size: '7,8 MB', hint: 'giải nén là chạy, không cần cài' },
  },
  { id: 'mac', name: 'macOS', icon: Apple, req: 'macOS 10.13 trở lên · Apple Silicon và Intel', file: 'file .dmg', size: '17 MB', hint: 'Mở file, kéo Sano vào thư mục Applications' },
  { id: 'linux', name: 'Linux', icon: Laptop, req: 'x86_64 · Ubuntu 22.04+, Debian 12+, Fedora 36+', file: 'file .AppImage', size: '8,2 MB', hint: 'Cấp quyền chạy rồi mở, cần WebKitGTK 4.1' },
]
const install = computed(() => ({
  win: 'Chạy file .exe vừa tải, bấm Tiếp tục đến hết. Sano có biểu tượng ở menu Start.',
  mac: 'Mở file .dmg, kéo biểu tượng Sano vào thư mục Applications.',
  linux: 'chmod +x Sano-*.AppImage rồi mở file. Thiếu WebKitGTK: sudo apt install libwebkit2gtk-4.1-0',
  mobile: 'Mở file cài trên máy tính: .exe (Windows), .dmg (macOS) hoặc .AppImage (Linux).',
})[seen.value])
const copied = ref(false)
</script>

<template>
  <div class="min-h-dvh bg-background text-foreground">
    <WfSiteReviewBar page="download" :dark="dark" :notes="notes" @toggle-dark="dark = !dark" @toggle-notes="notes = !notes">
      <span class="text-xs text-muted-foreground">Máy người xem:</span>
      <button v-for="o in seenOpts" :key="o[0]" type="button" class="h-7 px-2.5 rounded-full border text-xs"
        :class="seen === o[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background hover:bg-muted'" @click="seen = o[0]">{{ o[1] }}</button>
      <span class="h-4 w-px bg-border" />
    </WfSiteReviewBar>
    <WfSiteNav current="download" :dark="dark" @toggle-dark="dark = !dark" />

    <main>
      <!-- 1. Mở đầu + 3 thẻ tải -->
      <section class="border-b border-border bg-gradient-to-b from-primary/5 to-transparent">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <p v-if="notes" :class="noteCls" class="mb-3">Trang mới docs/tai-ve.md (layout: page, tự làm) · menu "Tải về" trỏ về đây thay cho /#tai-ve</p>
          <p class="text-sm font-medium text-primary">Miễn phí · Mã nguồn mở</p>
          <h1 class="mt-2 text-3xl font-semibold tracking-tight sm:text-4xl">Tải Sano</h1>
          <p class="mt-3 max-w-2xl text-lg text-muted-foreground">Một phần mềm, chạy trên cả <strong class="text-foreground">Windows, macOS và Linux</strong>. Chọn đúng máy của bạn, bấm là tải.</p>
          <p class="mt-3 flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-muted-foreground">
            <span class="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2.5 py-0.5 font-medium text-primary">Bản mới nhất {{ version }}</span>
            <span>phát hành {{ released }}</span><span>·</span>
            <a href="#" class="inline-flex items-center gap-1 font-medium text-primary hover:underline">Có gì mới <ArrowRight class="w-3.5 h-3.5" /></a>
          </p>
          <p v-if="notes" :class="noteCls" class="mt-2">Số phiên bản, ngày, dung lượng từng file lấy từ GitHub Releases như nút tải trang chủ (release.ts)</p>

          <!-- Điện thoại: Sano chỉ cài trên máy tính -->
          <div v-if="seen === 'mobile'" class="mt-6 flex flex-col gap-3 rounded-xl border border-border bg-card p-4 sm:flex-row sm:items-center">
            <Smartphone class="w-6 h-6 shrink-0 text-primary" />
            <p class="flex-1 text-sm">Bạn đang xem trên điện thoại. Sano cài trên máy tính; nghe trên điện thoại bằng file M4B xuất từ Sano. Gửi link trang này sang máy tính để tải.</p>
            <button type="button" class="inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-muted" @click="copied = true">
              <component :is="copied ? Check : Copy" class="w-4 h-4" /> {{ copied ? 'Đã sao chép' : 'Sao chép link' }}
            </button>
          </div>

          <div class="mt-8 grid gap-5 md:grid-cols-3">
            <div v-for="c in cards" :key="c.id" class="relative flex flex-col rounded-2xl border bg-card p-6 shadow-sm transition-shadow"
              :class="seen === c.id ? 'border-primary ring-2 ring-primary/30 shadow-md' : 'border-border'">
              <span v-if="seen === c.id" class="absolute -top-3 left-6 inline-flex items-center gap-1 rounded-full bg-primary px-3 py-1 text-xs font-semibold text-primary-foreground shadow">
                <Check class="w-3.5 h-3.5" /> Khuyên dùng cho máy bạn
              </span>
              <div class="grid h-16 w-16 place-items-center rounded-2xl" :class="seen === c.id ? 'bg-primary text-primary-foreground' : 'bg-primary/10 text-primary'">
                <component :is="c.icon" class="w-8 h-8" />
              </div>
              <h2 class="mt-4 text-2xl font-semibold tracking-tight">{{ c.name }}</h2>
              <p class="mt-1 min-h-10 text-sm text-muted-foreground">{{ c.req }}</p>

              <a href="#" class="mt-6 inline-flex items-center justify-center gap-2 rounded-lg px-5 py-3 font-medium focus-ring"
                :class="seen === c.id ? 'bg-primary text-primary-foreground shadow hover:bg-primary/90' : 'border border-border bg-background hover:bg-muted'">
                <Download class="w-5 h-5" /> Tải {{ c.file }} <span class="font-normal opacity-80">· {{ c.size }}</span>
              </a>
              <p class="mt-2 text-xs text-muted-foreground">{{ c.hint }}</p>

              <div v-if="c.alt" class="mt-4 border-t border-border pt-3 text-sm">
                <a href="#" class="font-medium text-primary hover:underline">{{ c.alt.label }}</a>
                <span class="text-muted-foreground"> · {{ c.alt.size }} · {{ c.alt.hint }}</span>
              </div>
            </div>
          </div>
          <p v-if="notes" :class="noteCls" class="mt-3">Lúc build (SSR) chưa biết máy → không thẻ nào có nhãn, cả 3 nút cùng kiểu viền; vào trình duyệt mới gắn nhãn</p>

          <p class="mt-6 flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-muted-foreground">
            <ShieldCheck class="w-4 h-4 text-rag-green" />
            File cài build tự động trên GitHub từ mã nguồn công khai, kèm SHA256SUMS.
            Chưa ký số nên lần đầu mở máy sẽ cảnh báo —
            <a href="#" class="font-medium text-primary hover:underline">cách mở lần đầu</a>.
          </p>
        </div>
      </section>

      <!-- 2. Sau khi tải -->
      <section class="border-b border-border">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Sau khi tải</h2>
          <ol class="mt-6 grid gap-4 md:grid-cols-3">
            <li class="rounded-xl border border-border bg-card p-5">
              <span class="grid h-8 w-8 place-items-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">1</span>
              <h3 class="mt-3 font-semibold">Cài Sano</h3>
              <p class="mt-1 text-sm text-muted-foreground">{{ install }}</p>
            </li>
            <li class="rounded-xl border border-border bg-card p-5">
              <span class="grid h-8 w-8 place-items-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">2</span>
              <h3 class="mt-3 font-semibold">Mở lần đầu</h3>
              <p class="mt-1 text-sm text-muted-foreground">Windows hoặc macOS có thể hỏi lại vì bản cài chưa ký số. <a href="#" class="font-medium text-primary hover:underline">Xem cách mở</a>.</p>
            </li>
            <li class="rounded-xl border border-border bg-card p-5">
              <span class="grid h-8 w-8 place-items-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">3</span>
              <h3 class="mt-3 font-semibold">Sano tự cài bộ đọc</h3>
              <p class="mt-1 text-sm text-muted-foreground">Tải khoảng 1 GB một lần (mô hình giọng đọc AI). Xong là dùng được không cần mạng. Thư viện có sẵn một cuốn mẫu để nghe thử.</p>
            </li>
          </ol>
          <a href="#" class="mt-5 inline-flex items-center gap-1 text-sm font-medium text-primary hover:underline"><BookOpen class="w-4 h-4" /> Hướng dẫn cài đặt đầy đủ</a>
        </div>
      </section>

      <!-- 3. Máy cần có -->
      <section class="border-b border-border">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Máy cần có</h2>
          <div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div class="rounded-xl border border-border bg-card p-5"><MemoryStick class="w-5 h-5 text-primary" /><p class="mt-2 font-medium">RAM 4 GB trở lên</p><p class="text-sm text-muted-foreground">Ít hơn vẫn chạy, đọc chậm hơn</p></div>
            <div class="rounded-xl border border-border bg-card p-5"><HardDrive class="w-5 h-5 text-primary" /><p class="mt-2 font-medium">Ổ trống 2,5 GB</p><p class="text-sm text-muted-foreground">Cài xong bộ đọc chiếm khoảng 1,5 GB</p></div>
            <div class="rounded-xl border border-border bg-card p-5"><Cpu class="w-5 h-5 text-primary" /><p class="mt-2 font-medium">Không cần card đồ hoạ</p><p class="text-sm text-muted-foreground">Giọng đọc AI chạy bằng CPU</p></div>
            <div class="rounded-xl border border-border bg-card p-5"><Wifi class="w-5 h-5 text-primary" /><p class="mt-2 font-medium">Mạng cho lần đầu</p><p class="text-sm text-muted-foreground">Sau đó dùng không cần mạng</p></div>
          </div>
        </div>
      </section>

      <!-- 4. File khác -->
      <section>
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">File khác</h2>
          <div class="mt-6 grid gap-4 md:grid-cols-3">
            <a href="#" class="group rounded-xl border border-border bg-card p-5 hover:border-primary/50">
              <FileAudio class="w-5 h-5 text-primary" />
              <p class="mt-2 font-medium group-hover:text-primary">Sách nói mẫu M4B · 2,1 MB</p>
              <p class="text-sm text-muted-foreground">"Kỹ năng mềm cho người trẻ", giọng Hải Đăng. Chép sang điện thoại nghe thử trước khi cài.</p>
            </a>
            <a href="#" class="group rounded-xl border border-border bg-card p-5 hover:border-primary/50">
              <ExternalLink class="w-5 h-5 text-primary" />
              <p class="mt-2 font-medium group-hover:text-primary">Các bản trước + SHA256SUMS</p>
              <p class="text-sm text-muted-foreground">Mọi phiên bản trên GitHub Releases, kèm mã kiểm tra file.</p>
            </a>
            <a href="#" class="group rounded-xl border border-border bg-card p-5 hover:border-primary/50">
              <FileCode2 class="w-5 h-5 text-primary" />
              <p class="mt-2 font-medium group-hover:text-primary">Mã nguồn + dòng lệnh</p>
              <p class="text-sm text-muted-foreground">Tự build từ mã nguồn, hoặc dùng công cụ dòng lệnh sano-docx2tts.</p>
            </a>
          </div>
          <p v-if="notes" :class="noteCls" class="mt-4">Trang chủ giữ nguyên nút tải theo máy; link "Mọi phiên bản" dưới nút đổi sang /tai-ve. Trang Cài đặt giữ bảng tải hiện có.</p>
        </div>
      </section>
    </main>
  </div>
</template>
