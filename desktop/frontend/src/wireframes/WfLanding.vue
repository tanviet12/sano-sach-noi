<script setup lang="ts">
// P4 — Trang chủ trang tài liệu Sano (wireframe). Trang thật: VitePress trên GitHub Pages,
// `docs/index.md` layout home + component Vue tự làm trong `.vitepress/theme/components/`.
// Cùng bộ nhận diện với app (đỏ thương hiệu, logo, phông, bìa sách) nhưng KHÔNG làm lại giao diện app.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API; audio, ảnh app, video là chỗ trống ghi rõ "làm sau".
// Mở ở bản dev: ?wireframe=landing (thêm &os=mac|win|linux, &theme=dark để chụp nhanh).
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import {
  Apple, ArrowRight, AudioLines, Ban, BookOpen, Check, Download, ExternalLink, FileText, FileUp, Github, Headphones,
  Laptop, ListTree, Mic, Monitor, Pause, Play, Server, ShieldCheck, Smartphone, UserRoundX, Video, Wand2, WifiOff,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import BookCover from '@/components/sano/BookCover.vue'
import WfSiteNav from './WfSiteNav.vue'
import WfSiteReviewBar from './WfSiteReviewBar.vue'
import sepayLogo from '@/assets/sponsors/sepay.svg'
import hostLogo from '@/assets/sponsors/123host.svg'
import bpLibrary from '../../../../docs/images/nghe-tren-dien-thoai/2-thu-vien.jpg'
import bpPlayer from '../../../../docs/images/nghe-tren-dien-thoai/3-trinh-phat.jpg'
import bpChapters from '../../../../docs/images/nghe-tren-dien-thoai/4-muc-luc-chuong.jpg'

const REPO = 'https://github.com/tanviet12/sano-sach-noi'
const RELEASES = REPO + '/releases/latest'
const AUTHOR_FB = 'https://www.facebook.com/buitanviet'

// ── Trạng thái duyệt (ngoài khung) ──
const q = new URLSearchParams(window.location.search)
const dark = ref(q.get('theme') === 'dark')
watch(dark, (v) => document.documentElement.classList.toggle('dark', v), { immediate: true })
const notes = ref(q.get('notes') !== '0')

// ── Nút Tải về tự gợi ý theo hệ điều hành ──
type OS = 'mac' | 'win' | 'linux'
const ua = navigator.userAgent
const detected: OS = /Mac|iPhone|iPad/.test(ua) ? 'mac' : /Win/.test(ua) ? 'win' : 'linux'
const override = ref<OS | 'auto'>((q.get('os') as OS | null) ?? 'auto')
const os = computed<OS>(() => (override.value === 'auto' ? detected : override.value))
const builds: Record<OS, { label: string; file: string; note: string; icon: typeof Apple }> = {
  mac: { label: 'macOS', file: 'Sano-<phiên bản>-macos-universal.dmg', note: 'Apple Silicon và Intel', icon: Apple },
  win: { label: 'Windows', file: 'Sano-<phiên bản>-windows-amd64-setup.exe', note: 'Windows 10/11 · không cần quyền quản trị', icon: Monitor },
  linux: { label: 'Linux', file: 'Sano-<phiên bản>-linux-amd64.AppImage', note: 'x86_64 · cần WebKitGTK 4.1', icon: Laptop },
}
const build = computed(() => builds[os.value])
const others = computed(() => (Object.keys(builds) as OS[]).filter((k) => k !== os.value))

// ── Nghe thử: trình phát nhỏ giống trình phát trong app (audio giả, đếm giờ bằng setInterval) ──
const tracks = [
  { title: 'Lời mở đầu', sec: 38 },
  { title: '1.1 Lắng nghe chủ động', sec: 74 },
  { title: '1.2 Đặt câu hỏi mở', sec: 66 },
  { title: '2.1 Nói ngắn, ý rõ', sec: 58 },
]
const current = ref(1)
const time = ref(21)
const playing = ref(false)
let timer: number | undefined
function stop() {
  playing.value = false
  window.clearInterval(timer)
}
function toggle() {
  if (playing.value) return stop()
  playing.value = true
  timer = window.setInterval(() => {
    time.value++
    if (time.value >= tracks[current.value].sec) {
      if (current.value < tracks.length - 1) {
        current.value++
        time.value = 0
      } else stop()
    }
  }, 1000)
}
function pick(i: number) {
  current.value = i
  time.value = 0
  if (!playing.value) toggle()
}
onBeforeUnmount(stop)
const fmt = (s: number) => `${Math.floor(s / 60)}:${String(Math.floor(s % 60)).padStart(2, '0')}`
const pct = computed(() => (time.value / tracks[current.value].sec) * 100)

// Trước / sau khi chuẩn hoá văn nói — ví dụ lấy từ luật thật trong internal/bookmaker (normalize.go, spoken.go)
const normRows = [
  { before: 'Chương IV. Mô hình 4P', after: 'Chương bốn. Mô hình bốn Pê' },
  { before: 'PGS. TS. Lê Văn C, TP.HCM', after: 'Phó giáo sư Tiến sĩ Lê Văn C, Thành phố Hồ Chí Minh' },
  { before: 'Đọc sách → hiểu mình → đổi thói quen', after: 'Đọc sách dẫn tới hiểu mình dẫn tới đổi thói quen' },
  { before: 'Học 2-3 giờ/ngày, vd. sáng & tối', after: 'Học 2 đến 3 giờ một ngày, ví dụ sáng và tối' },
]
const normPick = ref<'before' | 'after'>('after')
const normPlaying = ref(false)

// ── Cách hoạt động ──
const steps = [
  { icon: FileUp, title: 'Nạp file Word', desc: 'Sano đọc mục lục theo Heading, cảnh báo bảng, hình, chữ viết tắt lạ.' },
  { icon: Headphones, title: 'Nghe thử & sửa lời đọc', desc: 'Nghe vài đoạn, sửa cách đọc từ riêng trước khi tạo cả cuốn.' },
  { icon: AudioLines, title: 'Tạo sách (chạy nền)', desc: 'Giọng đọc AI chạy trên máy bạn. Vẫn làm việc khác được trong lúc chờ.' },
  { icon: Smartphone, title: 'Nghe trên máy hoặc điện thoại', desc: 'Nghe ngay trong Sano, hoặc xuất một file M4B chép sang điện thoại.' },
]

// ── Điểm nổi bật ──
const highlights = [
  { icon: WifiOff, title: 'Chạy offline trên máy', desc: 'Tài liệu không gửi lên mạng. Cài bộ đọc xong là dùng được không cần internet.' },
  { icon: Mic, title: '25 giọng Việt', desc: 'Nam, nữ, giọng Bắc, giọng Nam. Nghe thử từng giọng trước khi chọn.' },
  { icon: Wand2, title: 'Tự chuẩn hoá văn nói', desc: 'Số, viết tắt, mũi tên, ký hiệu được đọc thành lời tự nhiên.' },
  { icon: ListTree, title: 'Mục lục chương', desc: 'Lấy từ Heading trong Word. Nhảy chương trong app và trong file M4B.' },
  { icon: UserRoundX, title: 'Không cần tài khoản', desc: 'Tải về, cài, dùng. Không đăng ký, không đăng nhập.' },
  { icon: Github, title: 'Mã nguồn mở MIT', desc: 'Miễn phí, xem và sửa được toàn bộ mã nguồn trên GitHub.' },
]

// ── Bản quyền: 3 ý như màn đồng ý điều khoản trong app ──
const rights = [
  { icon: FileText, title: 'Tài liệu bạn có quyền dùng', desc: 'Tài liệu của bạn, sách hết bảo hộ, hoặc được tác giả cho phép.' },
  { icon: Ban, title: 'Không dùng sách còn bản quyền', desc: 'Trừ khi được tác giả hoặc chủ sở hữu cho phép.' },
  { icon: ShieldCheck, title: 'Chạy trên máy bạn', desc: 'Tài liệu không gửi đi đâu. Bạn tự chịu trách nhiệm nội dung.' },
]

// ── Tài trợ (mô tả giống màn Giới thiệu trong app) ──
const utm = '?utm_source=sano&utm_medium=docs&utm_campaign=tai-tro'
const sponsors = [
  { name: 'SePay', logo: sepayLogo, url: 'https://sepay.vn' + utm, site: 'sepay.vn', desc: 'Nền tảng Open Banking: tự động xác nhận thanh toán chuyển khoản, kết nối API với các ngân hàng Việt Nam cho website và phần mềm bán hàng.' },
  { name: '123HOST', logo: hostLogo, url: 'https://123host.vn' + utm, site: '123host.vn', desc: 'Hosting, VPS, máy chủ và tên miền cho doanh nghiệp, nhà phát triển Việt Nam.' },
]

const noteCls = 'inline-flex items-center gap-1 rounded border border-dashed border-amber-500/70 bg-amber-50 px-1.5 py-0.5 text-[11px] font-medium text-amber-800 dark:bg-amber-950/40 dark:text-amber-300'
const phCls = 'grid place-items-center rounded-lg border-2 border-dashed border-border bg-muted/40 text-center text-xs text-muted-foreground'
</script>

<template>
  <div class="min-h-dvh bg-background text-foreground">
    <WfSiteReviewBar page="landing" :dark="dark" :notes="notes" @toggle-dark="dark = !dark" @toggle-notes="notes = !notes">
      <span class="text-xs text-muted-foreground">Gợi ý tải:</span>
      <button v-for="o in ([['auto', 'Tự nhận'], ['mac', 'macOS'], ['win', 'Windows'], ['linux', 'Linux']] as const)" :key="o[0]" type="button"
        class="h-7 px-2.5 rounded-full border text-xs"
        :class="override === o[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background hover:bg-muted'"
        @click="override = o[0]">{{ o[1] }}</button>
      <span class="h-4 w-px bg-border" />
    </WfSiteReviewBar>

    <WfSiteNav current="home" :dark="dark" @toggle-dark="dark = !dark" />

    <main>
      <!-- ═══ 1. Hero ═══ -->
      <section id="tai-ve" class="ambient-soft border-b border-border">
        <div class="mx-auto grid max-w-6xl items-center gap-10 px-4 py-12 sm:px-6 lg:grid-cols-[1fr_1.1fr] lg:py-20">
          <div>
            <p v-if="notes" :class="noteCls" class="mb-3">VitePress layout: home · hero tự làm (HomeHero.vue) vì nút tải đổi theo hệ điều hành</p>
            <Badge variant="outline" class="gap-1.5 bg-background/70"><span class="h-1.5 w-1.5 rounded-full bg-primary" /> Mã nguồn mở · Miễn phí</Badge>
            <p v-if="notes" :class="noteCls" class="mb-3">SEO — &lt;title&gt;: "Sano – Tạo sách nói bằng AI từ file Word" · meta description: "Sano giúp bạn tự làm sách nói bằng AI từ file Word của chính mình: giọng đọc tiếng Việt chạy trên máy, có mục lục chương, xuất M4B nghe trên điện thoại. Miễn phí, mã nguồn mở."</p>
            <h1 class="mt-4 text-4xl font-semibold leading-tight tracking-tight text-balance sm:text-5xl">
              Tạo <span class="text-primary">sách nói bằng AI</span> từ file Word
            </h1>
            <p class="mt-4 max-w-xl text-base leading-relaxed text-muted-foreground sm:text-lg">
              Biến tài liệu của chính bạn thành sách nói. Giọng đọc AI tiếng Việt chạy ngay trên máy, miễn phí, mã nguồn mở. Nghe trong phần mềm, hoặc chép một file sang điện thoại nghe khi lái xe, lúc rảnh tay.
            </p>

            <div class="mt-7 flex flex-col gap-3 sm:flex-row sm:items-center">
              <a :href="RELEASES" target="_blank" rel="noopener" class="focus-ring rounded-md">
                <Button size="lg" class="h-auto w-full px-6 py-3 sm:w-auto" tabindex="-1">
                  <Download class="w-5 h-5" />
                  <span class="flex flex-col items-start leading-tight">
                    <span class="text-base">Tải cho {{ build.label }}</span>
                    <span class="text-xs font-normal opacity-80">{{ build.note }}</span>
                  </span>
                </Button>
              </a>
              <a href="?wireframe=docs" class="focus-ring rounded-md">
                <Button variant="outline" size="lg" class="w-full sm:w-auto" tabindex="-1"><BookOpen class="w-4 h-4" /> Xem hướng dẫn</Button>
              </a>
            </div>
            <p class="mt-4 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted-foreground">
              <span>Bản khác:</span>
              <a v-for="k in others" :key="k" :href="RELEASES" target="_blank" rel="noopener" class="flex items-center gap-1 font-medium text-foreground hover:text-primary">
                <component :is="builds[k].icon" class="w-3.5 h-3.5" /> {{ builds[k].label }}
              </a>
              <span class="hidden sm:inline">·</span>
              <a :href="REPO + '/releases'" target="_blank" rel="noopener" class="hover:text-primary">Mọi phiên bản <ExternalLink class="inline w-3 h-3" /></a>
            </p>
            <p class="mt-2 text-xs text-muted-foreground">
              File <code class="rounded bg-muted px-1">{{ build.file }}</code> trên GitHub Releases. Bản cài chưa ký số —
              <a href="?wireframe=docs" class="underline hover:text-primary">cách mở lần đầu</a>.
            </p>
          </div>

          <!-- Ảnh chụp app: chỗ trống, chụp thật sau -->
          <div class="relative">
            <div class="overflow-hidden rounded-xl border border-border bg-background shadow-2xl">
              <div class="flex h-8 items-center gap-1.5 border-b border-border bg-muted/60 px-3">
                <span class="h-2.5 w-2.5 rounded-full bg-rag-red" /><span class="h-2.5 w-2.5 rounded-full bg-rag-amber" /><span class="h-2.5 w-2.5 rounded-full bg-rag-green" />
              </div>
              <div class="flex aspect-[16/10] opacity-60">
                <div class="w-1/5 space-y-2 border-r border-border bg-muted/40 p-3">
                  <div class="h-2.5 w-3/4 rounded bg-primary/40" /><div class="h-2.5 w-2/3 rounded bg-muted-foreground/20" /><div class="h-2.5 w-3/5 rounded bg-muted-foreground/20" /><div class="h-2.5 w-2/3 rounded bg-muted-foreground/20" />
                </div>
                <div class="grid flex-1 grid-cols-4 content-start gap-3 p-4">
                  <BookCover v-for="t in ['Kỹ năng mềm cho người trẻ', 'Ghi chép cuộc họp', 'Tài chính cá nhân cơ bản', 'Khởi nghiệp từ số 0']" :key="t" :title="t" class="aspect-[3/4]" />
                </div>
              </div>
            </div>
            <div class="absolute inset-0 grid place-items-center">
              <span class="rounded-md border-2 border-dashed border-primary/60 bg-background/90 px-3 py-1.5 text-xs font-medium text-primary shadow-sm">Ảnh chụp app — sẽ chụp thật</span>
            </div>
          </div>
        </div>
      </section>

      <!-- ═══ 2. Nghe thử ═══ -->
      <section id="nghe-thu" class="border-b border-border">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <p v-if="notes" :class="noteCls" class="mb-3">Component tự làm: SamplePlayer.vue + NormalizeCompare.vue (audio mp3 trong docs/public/)</p>
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Nghe thử</h2>
          <p class="mt-2 max-w-2xl text-muted-foreground">Một cuốn sách mẫu Sano tạo từ file Word, giọng Thiện Minh, chạy hoàn toàn trên máy tính.</p>

          <div class="mt-8 grid gap-6 lg:grid-cols-5">
            <!-- Trình phát nhỏ -->
            <div class="rounded-xl border border-border bg-card p-5 shadow-sm lg:col-span-3">
              <div class="flex gap-5">
                <BookCover title="Kỹ năng mềm cho người trẻ" author="Sách mẫu Sano" class="aspect-[3/4] w-24 shrink-0 shadow-lg sm:w-32" />
                <div class="min-w-0 flex-1">
                  <p class="font-semibold leading-snug">Kỹ năng mềm cho người trẻ</p>
                  <p class="text-sm text-muted-foreground">Giọng Thiện Minh · 4 đoạn mẫu</p>
                  <p class="mt-3 truncate text-sm text-primary">{{ tracks[current].title }}</p>
                  <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-muted" role="slider" aria-label="Vị trí nghe" :aria-valuenow="Math.round(pct)">
                    <div class="h-full rounded-full bg-primary transition-all duration-300" :style="{ width: pct + '%' }" />
                  </div>
                  <div class="mt-1 flex justify-between text-[11px] tabular-nums text-muted-foreground"><span>{{ fmt(time) }}</span><span>-{{ fmt(tracks[current].sec - time) }}</span></div>
                  <button type="button" :aria-label="playing ? 'Dừng' : 'Phát'" class="mt-2 grid h-12 w-12 place-items-center rounded-full bg-primary text-primary-foreground shadow focus-ring" @click="toggle">
                    <Pause v-if="playing" class="w-5 h-5" /><Play v-else class="ml-0.5 w-5 h-5" />
                  </button>
                </div>
              </div>
              <div class="mt-4 divide-y divide-border rounded-lg border border-border">
                <button v-for="(t, i) in tracks" :key="t.title" type="button"
                  class="flex w-full items-center justify-between gap-2 px-3 py-2.5 text-left text-sm hover:bg-muted/60 focus-ring"
                  :class="i === current ? 'bg-primary/10 font-medium text-primary' : ''" @click="pick(i)">
                  <span class="flex min-w-0 items-center gap-2">
                    <AudioLines v-if="i === current" class="w-3.5 h-3.5 shrink-0" /><span v-else class="w-3.5 shrink-0 text-center text-xs text-muted-foreground">{{ i + 1 }}</span>
                    <span class="truncate">{{ t.title }}</span>
                  </span>
                  <span class="shrink-0 text-xs tabular-nums">{{ fmt(t.sec) }}</span>
                </button>
              </div>
              <p class="mt-3 text-xs text-muted-foreground"><span :class="noteCls">Audio mẫu — sẽ render bằng giọng Thiện Minh</span></p>
            </div>

            <!-- Trước / Sau chuẩn hoá -->
            <div class="rounded-xl border border-border bg-card p-5 shadow-sm lg:col-span-2">
              <p class="font-semibold">Trước / sau khi chuẩn hoá văn nói</p>
              <p class="mt-1 text-sm text-muted-foreground">Văn viết để nhìn. Sano tự chuyển thành lời đọc tự nhiên, không cần AI trên mạng.</p>
              <div class="mt-4 inline-flex rounded-lg border border-border p-0.5 text-sm">
                <button v-for="o in ([['before', 'Chưa chuẩn hoá'], ['after', 'Đã chuẩn hoá']] as const)" :key="o[0]" type="button"
                  class="rounded-md px-3 py-1.5 focus-ring" :class="normPick === o[0] ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground'"
                  @click="normPick = o[0]">{{ o[1] }}</button>
              </div>
              <button type="button" class="mt-3 flex w-full items-center gap-3 rounded-lg border border-border px-3 py-2 text-left text-sm hover:bg-muted/60 focus-ring" @click="normPlaying = !normPlaying">
                <span class="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-primary text-primary-foreground"><Pause v-if="normPlaying" class="w-4 h-4" /><Play v-else class="ml-0.5 w-4 h-4" /></span>
                <span class="flex-1">Nghe bản {{ normPick === 'before' ? 'chưa chuẩn hoá' : 'đã chuẩn hoá' }}</span>
                <span class="text-xs tabular-nums text-muted-foreground">0:18</span>
              </button>
              <div class="mt-4 space-y-2.5">
                <div v-for="r in normRows" :key="r.before" class="rounded-lg bg-muted/50 px-3 py-2 text-sm">
                  <p :class="normPick === 'before' ? 'font-medium' : 'text-muted-foreground line-through decoration-muted-foreground/40'">{{ r.before }}</p>
                  <p class="mt-0.5 flex gap-1.5" :class="normPick === 'after' ? 'font-medium text-foreground' : 'text-muted-foreground'">
                    <ArrowRight class="mt-0.5 w-3.5 h-3.5 shrink-0 text-primary" />{{ r.after }}
                  </p>
                </div>
              </div>
              <p class="mt-3 text-xs"><span :class="noteCls">2 audio giả — sẽ render cùng đoạn bằng giọng Thiện Minh</span></p>
            </div>
          </div>
        </div>
      </section>

      <!-- ═══ 3. Cách hoạt động ═══ -->
      <section id="cach-hoat-dong" class="border-b border-border bg-muted/30">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <p v-if="notes" :class="noteCls" class="mb-3">Component tự làm: HowItWorks.vue (ảnh trong docs/public/screenshots/)</p>
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Cách hoạt động</h2>
          <p class="mt-2 max-w-2xl text-muted-foreground">Bốn bước, từ file Word tới sách nói có mục lục chương.</p>
          <ol class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <li v-for="(s, i) in steps" :key="s.title" class="flex flex-col rounded-xl border border-border bg-card p-4 shadow-sm">
              <div :class="phCls" class="aspect-[4/3] w-full"><span>Ảnh bước {{ i + 1 }}<br />chụp từ app</span></div>
              <div class="mt-4 flex items-center gap-2">
                <span class="grid h-7 w-7 shrink-0 place-items-center rounded-full bg-primary text-xs font-semibold text-primary-foreground">{{ i + 1 }}</span>
                <component :is="s.icon" class="w-4 h-4 text-primary" />
                <p class="font-medium">{{ s.title }}</p>
              </div>
              <p class="mt-2 text-sm leading-relaxed text-muted-foreground">{{ s.desc }}</p>
            </li>
          </ol>
          <div :class="phCls" class="mx-auto mt-8 aspect-video max-w-3xl bg-background">
            <div class="flex flex-col items-center gap-2">
              <span class="grid h-14 w-14 place-items-center rounded-full bg-primary/10 text-primary"><Video class="w-6 h-6" /></span>
              <span class="text-sm font-medium text-foreground">Video giới thiệu ~1 phút</span>
              <span>Sẽ quay sau · nhúng YouTube hoặc mp4 trong docs/public/</span>
            </div>
          </div>
        </div>
      </section>

      <!-- ═══ 4. Nghe trên điện thoại ═══ -->
      <section id="dien-thoai" class="border-b border-border">
        <div class="mx-auto grid max-w-6xl items-center gap-10 px-4 py-14 sm:px-6 lg:grid-cols-2">
          <div>
            <p v-if="notes" :class="noteCls" class="mb-3">Component tự làm: PhoneSection.vue · ảnh có sẵn docs/images/nghe-tren-dien-thoai/</p>
            <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Nghe trên điện thoại, trên xe</h2>
            <p class="mt-3 leading-relaxed text-muted-foreground">
              Sano xuất cả cuốn thành <strong class="text-foreground">một file M4B</strong> có mục lục chương, tên sách và bìa. Chép sang điện thoại là nghe được, không cần mạng. App nghe tự nhớ chỗ nghe dở, trên xe hiện tên chương.
            </p>
            <ul class="mt-5 space-y-2 text-sm">
              <li class="flex gap-2"><Check class="mt-0.5 w-4 h-4 shrink-0 text-primary" /><span><strong>iPhone:</strong> app BookPlayer (miễn phí, mã nguồn mở) hoặc app Sách của Apple</span></li>
              <li class="flex gap-2"><Check class="mt-0.5 w-4 h-4 shrink-0 text-primary" /><span><strong>Android:</strong> app nghe sách nói bất kỳ đọc được M4B, có Android Auto</span></li>
              <li class="flex gap-2"><Check class="mt-0.5 w-4 h-4 shrink-0 text-primary" /><span>Khoảng 29 MB cho mỗi giờ nghe</span></li>
            </ul>
            <div class="mt-6 flex flex-col gap-3 sm:flex-row">
              <a :href="RELEASES" target="_blank" rel="noopener" class="focus-ring rounded-md"><Button class="w-full sm:w-auto" tabindex="-1"><Download class="w-4 h-4" /> Tải M4B mẫu</Button></a>
              <a href="?wireframe=docs" class="focus-ring rounded-md"><Button variant="outline" class="w-full sm:w-auto" tabindex="-1">Hướng dẫn chi tiết <ArrowRight class="w-4 h-4" /></Button></a>
            </div>
            <p class="mt-2 text-xs"><span :class="noteCls">M4B mẫu — sẽ đính kèm vào Release khi phát hành</span></p>
          </div>
          <div class="grid grid-cols-3 gap-3 sm:gap-4">
            <figure v-for="(img, i) in [bpLibrary, bpPlayer, bpChapters]" :key="img" class="text-center">
              <div class="overflow-hidden rounded-[1.4rem] border-4 border-foreground/80 bg-black shadow-xl">
                <img :src="img" :alt="['Thư viện BookPlayer', 'Đang nghe, bìa và tên chương', 'Mục lục chương'][i]" class="w-full" />
              </div>
              <figcaption class="mt-2 text-xs text-muted-foreground">{{ ['Thư viện', 'Đang nghe', 'Mục lục chương'][i] }}</figcaption>
            </figure>
          </div>
        </div>
      </section>

      <!-- ═══ 5. Điểm nổi bật ═══ -->
      <section class="border-b border-border bg-muted/30">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <p v-if="notes" :class="noteCls" class="mb-3">VitePress có sẵn: frontmatter features (icon lucide dạng SVG)</p>
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Điểm nổi bật</h2>
          <div class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <div v-for="h in highlights" :key="h.title" class="rounded-xl border border-border bg-card p-5">
              <span class="grid h-10 w-10 place-items-center rounded-lg bg-primary/10 text-primary"><component :is="h.icon" class="w-5 h-5" /></span>
              <p class="mt-4 font-medium">{{ h.title }}</p>
              <p class="mt-1 text-sm leading-relaxed text-muted-foreground">{{ h.desc }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- ═══ 6. Bản quyền ═══ -->
      <section id="ban-quyen" class="border-b border-border">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <p v-if="notes" :class="noteCls" class="mb-3">Markdown + component nhỏ RightsSummary.vue (cùng nội dung màn đồng ý trong app)</p>
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Bản quyền</h2>
          <p class="mt-2 max-w-2xl text-muted-foreground">Sano là công cụ cho tài liệu của chính bạn. Trước khi dùng, phần mềm yêu cầu đồng ý điều khoản sử dụng.</p>
          <div class="mt-6 grid gap-3 md:grid-cols-3">
            <div v-for="r in rights" :key="r.title" class="flex gap-3 rounded-lg border border-border p-4">
              <span class="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-primary/10 text-primary"><component :is="r.icon" class="w-4 h-4" /></span>
              <div class="min-w-0">
                <p class="text-sm font-medium">{{ r.title }}</p>
                <p class="text-sm leading-relaxed text-muted-foreground">{{ r.desc }}</p>
              </div>
            </div>
          </div>
          <a href="?wireframe=docs" class="mt-4 inline-flex items-center gap-1 text-sm font-medium text-primary hover:underline">Đọc điều khoản sử dụng đầy đủ <ArrowRight class="w-4 h-4" /></a>
        </div>
      </section>

      <!-- ═══ 7. Kế hoạch tiếp theo ═══ -->
      <section class="border-b border-border">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Kế hoạch tiếp theo</h2>
          <div class="mt-6 flex flex-col gap-4 rounded-xl border border-dashed border-border p-5 sm:flex-row sm:items-start">
            <span class="grid h-10 w-10 shrink-0 place-items-center rounded-lg bg-muted text-muted-foreground"><Server class="w-5 h-5" /></span>
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <p class="font-medium">Máy chủ nghe sách riêng</p>
                <Badge variant="secondary">Giai đoạn sau</Badge>
              </div>
              <p class="mt-1 text-sm leading-relaxed text-muted-foreground">
                Dành cho gia đình và nhóm nhỏ: tự dựng trên máy chủ của bạn, nghe qua web và app điện thoại, nhớ vị trí nghe giữa các máy. Bản hiện tại là phần mềm máy tính.
              </p>
            </div>
          </div>
        </div>
      </section>

      <!-- ═══ 8. Tài trợ + Tác giả ═══ -->
      <section class="border-b border-border bg-muted/30">
        <div class="mx-auto grid max-w-6xl gap-10 px-4 py-14 sm:px-6 lg:grid-cols-[1.4fr_1fr]">
          <div>
            <p v-if="notes" :class="noteCls" class="mb-3">Component tự làm: Sponsors.vue (logo có sẵn trong app)</p>
            <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Đơn vị tài trợ</h2>
            <p class="mt-2 text-muted-foreground">Sano miễn phí và mã nguồn mở nhờ sự tài trợ của</p>
            <div class="mt-6 grid gap-4 sm:grid-cols-2">
              <a v-for="sp in sponsors" :key="sp.name" :href="sp.url" target="_blank" rel="noopener"
                class="group flex flex-col gap-3 rounded-xl border border-border bg-card p-5 transition-colors hover:border-primary/50 focus-ring">
                <div class="grid h-16 place-items-center rounded-md bg-white px-4"><img :src="sp.logo" :alt="sp.name" class="max-h-10 max-w-[170px] object-contain" /></div>
                <p class="flex-1 text-sm leading-relaxed text-muted-foreground">{{ sp.desc }}</p>
                <span class="flex items-center gap-1 text-sm font-medium text-primary">{{ sp.site }} <ExternalLink class="w-3.5 h-3.5" /></span>
              </a>
            </div>
          </div>
          <div>
            <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Tác giả</h2>
            <div class="mt-6 rounded-xl border border-border bg-card p-5">
              <div class="flex items-center gap-3">
                <span class="grid h-12 w-12 shrink-0 place-items-center rounded-full bg-primary text-base font-semibold text-primary-foreground">BV</span>
                <div>
                  <a :href="AUTHOR_FB" target="_blank" rel="noopener" class="font-semibold hover:text-primary">Bùi Tấn Việt <ExternalLink class="inline w-3.5 h-3.5" /></a>
                  <p class="text-sm text-muted-foreground">CEO SePay và 123HOST</p>
                </div>
              </div>
              <p class="mt-4 text-sm leading-relaxed text-muted-foreground">
                “Tôi học liên tục và muốn tranh thủ nghe lại tài liệu của mình lúc lái xe, lúc rảnh tay. Ngoài thị trường có nhiều app sách nói, nhưng không có cái nào đọc được tài liệu riêng của mình. Tôi tự làm một công cụ để dùng, thấy hữu ích nên mở mã nguồn cho ai cần.”
              </p>
              <p v-if="notes" :class="noteCls" class="mt-3">Chỗ ảnh đại diện: đang dùng chữ viết tắt, thay ảnh thật nếu anh muốn</p>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- ═══ 9. Chân trang (VitePress: themeConfig.footer) ═══ -->
    <footer class="border-t border-border">
      <div class="mx-auto max-w-6xl px-4 py-8 text-center text-sm text-muted-foreground sm:px-6">
        <p v-if="notes" :class="noteCls" class="mb-3">VitePress có sẵn: themeConfig.footer (message + copyright, cho phép HTML)</p>
        <p class="flex flex-wrap items-center justify-center gap-x-3 gap-y-1">
          <a :href="REPO" target="_blank" rel="noopener" class="flex items-center gap-1 hover:text-primary"><Github class="w-4 h-4" /> GitHub</a>
          <span>·</span>
          <span>Phát hành theo giấy phép <a :href="REPO + '/blob/main/LICENSE'" target="_blank" rel="noopener" class="underline hover:text-primary">MIT</a></span>
          <span>·</span>
          <span>Giọng đọc <a href="https://github.com/pnnbao97/VieNeu-TTS" target="_blank" rel="noopener" class="underline hover:text-primary">VieNeu-TTS</a> (Apache-2.0)</span>
        </p>
        <p class="mt-3 flex items-center justify-center">
          <a href="https://github.com/tanviet12/vbsec" target="_blank" rel="noopener" class="inline-flex items-center overflow-hidden rounded-md border border-border text-xs font-medium hover:border-rag-green/60">
            <span class="flex items-center gap-1.5 bg-muted px-2 py-1 text-foreground"><ShieldCheck class="w-3.5 h-3.5" /> vbsec</span>
            <span class="bg-rag-green/15 px-2 py-1 text-rag-green">đã quét bảo mật · đạt</span>
          </a>
        </p>
        <p v-if="notes" :class="noteCls" class="mt-2">Chỉ gắn khi lượt quét vbsec trên bản phát hành cho kết quả ĐẠT; ghi kèm ngày quét. Link tới github.com/tanviet12/vbsec</p>
        <p class="mt-2">Copyright © 2026 Bùi Tấn Việt</p>
      </div>
    </footer>
  </div>
</template>
