<script setup lang="ts">
// D1 — Phần mềm tạo sách (desktop, Wails). Khung cửa sổ ~1100×720, thanh bên + luồng 6 bước.
// Một cuốn render một lúc: đang render thì bấm "Tạo sách mới" sẽ mở màn tiến độ của cuốn đó.
// Sáng/tối theo hệ điều hành (nút đổi ở góc để duyệt cả hai).
// Wireframe tĩnh: dữ liệu giả, bấm để chuyển màn, KHÔNG gọi API.
import { computed, ref } from 'vue'
import {
  Library, FilePlus2, Settings, Upload, FileText, AlertTriangle, Copy, Check, ChevronRight, ChevronLeft,
  Play, Pause, Loader2, Download, FolderOpen, Trash2, Sun, Moon, HardDrive, RefreshCw, SkipBack,
  SkipForward, RotateCcw, RotateCw, Gauge, BookOpen, Package, X, Info, Cpu, Clock, Volume2, Pencil,
  ShieldCheck, ArrowUpCircle, LifeBuoy, ExternalLink, Github, Bug, Scale,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import WfBookCover from './WfBookCover.vue'

type View = 'setup' | 'library' | 'create' | 'player' | 'settings'
const q = new URLSearchParams(window.location.search)
const view = ref<View>((q.get('screen') as View | null) || 'library')
const step = ref(Number(q.get('step')) || 1)
const dark = ref(false)
const rendering = ref(true) // có 1 cuốn đang render → hiện thẻ tiến độ ở thanh bên
const fileLoaded = ref(true)
const copied = ref(false)
const selectedVoice = ref('Ly')
const playingSample = ref<string | null>(null)
const renderDone = ref(false)
// Link ngoài (app thật mở bằng trình duyệt mặc định)
const DOCS = 'https://tanviet12.github.io/sano-sach-noi'
const REPO = 'https://github.com/tanviet12/sano-sach-noi'
const AUTHOR_FB = 'https://www.facebook.com/buitanviet'
// Cập nhật app: 'closed' | 'info' | 'downloading' | 'ready'
type UpdateState = 'closed' | 'info' | 'downloading' | 'ready'
const update = ref<UpdateState>(q.get('update') ? 'info' : 'closed')
const updateAfterRender = ref(false)
const changelog = [
  'Nghe thử được bất kỳ đoạn nào, không giới hạn 3 đoạn',
  'Xuất M4B nhanh hơn khoảng 2 lần',
  'Sửa lỗi đọc sai số La Mã trong tiêu đề chương',
]

const steps = [
  { n: 1, label: 'Nạp file' },
  { n: 2, label: 'Mục lục' },
  { n: 3, label: 'Giọng đọc' },
  { n: 4, label: 'Lời mở đầu' },
  { n: 5, label: 'Nghe thử' },
  { n: 6, label: 'Render' },
]

const nav = [
  { key: 'library' as View, label: 'Thư viện', icon: Library },
  { key: 'create' as View, label: 'Tạo sách mới', icon: FilePlus2 },
  { key: 'settings' as View, label: 'Cài đặt', icon: Settings },
]

// Sách đã tạo (thư viện)
const books = [
  { title: 'Kỹ năng mềm cho người trẻ', author: 'Nguyễn Văn A', dur: '1 giờ 12 phút', chapters: 4, progress: 35, date: '22/09/2026' },
  { title: 'Lãnh đạo cho quản lý mới', author: 'Trần Thị B', dur: '58 phút', chapters: 4, progress: 100, date: '20/09/2026' },
  { title: 'Tài chính cá nhân cơ bản', author: 'Lê Văn C', dur: '1 giờ 05 phút', chapters: 4, progress: 0, date: '18/09/2026' },
  { title: 'Khởi nghiệp từ số 0', author: 'Phạm Thị D', dur: '47 phút', chapters: 4, progress: 12, date: '15/09/2026' },
]

// Mục lục đọc được từ Heading 1/2 của file Word
const toc = ref([
  { title: 'Trang bìa', kind: 'skip', chars: 120, sections: [] as { title: string; chars: number; on: boolean }[], on: false, note: 'Gợi ý bỏ: trang bìa' },
  { title: 'Mục lục', kind: 'skip', chars: 640, sections: [], on: false, note: 'Gợi ý bỏ: mục lục gốc' },
  { title: 'Chương 1. Lắng nghe chủ động', kind: 'chapter', chars: 8200, on: true, note: '', sections: [
    { title: '1.1 Vì sao lắng nghe khó', chars: 2900, on: true },
    { title: '1.2 Ba bước lắng nghe', chars: 3100, on: true },
    { title: '1.3 Bài tập mỗi ngày', chars: 2200, on: true },
  ] },
  { title: 'Chương 2. Giao tiếp rõ ràng', kind: 'chapter', chars: 9100, on: true, note: '', sections: [
    { title: '2.1 Nói ngắn, ý rõ', chars: 4200, on: true },
    { title: '2.2 Hỏi để hiểu', chars: 4900, on: true },
  ] },
  { title: 'Chương 3. Quản lý thời gian', kind: 'chapter', chars: 7600, on: true, note: '', sections: [
    { title: '3.1 Việc quan trọng trước', chars: 3800, on: true },
    { title: '3.2 Nói không đúng lúc', chars: 3800, on: true },
  ] },
  { title: 'Phụ lục: bảng tự đánh giá', kind: 'chapter', chars: 1500, on: false, note: 'Chủ yếu là bảng — nội dung bảng không được đọc', sections: [] },
])
const totalChars = computed(() =>
  toc.value.filter((c) => c.on).reduce((s, c) => s + c.chars, 0),
)
// ~70 ký tự/giây render trên máy tham khảo; ~15 ký tự/giây nghe
const estListen = computed(() => Math.round(totalChars.value / 15 / 60))
const estRender = computed(() => Math.round(totalChars.value / 70 / 60))

const voices = [
  { id: 'Ly', name: 'Trúc Ly', desc: 'Nữ · miền Bắc', rec: true },
  { id: 'Ngoc', name: 'Bích Ngọc', desc: 'Nữ · miền Bắc', rec: false },
  { id: 'Tuyen', name: 'Phạm Tuyên', desc: 'Nam · miền Bắc', rec: false },
  { id: 'Doan', name: 'Thục Đoan', desc: 'Nữ · miền Nam', rec: false },
  { id: 'Vinh', name: 'Xuân Vĩnh', desc: 'Nam · miền Nam', rec: false },
]

const samples = [
  { id: 's1', title: 'Lời mở đầu', text: 'Bạn đang nghe sách nói. Cuốn sách: Kỹ năng giao tiếp. Tác giả: Nguyễn Văn A.' },
  { id: 's2', title: '1.2 Ba bước lắng nghe', text: 'Bước thứ nhất, dừng việc đang làm và nhìn người nói. Bước thứ hai, nhắc lại ý chính bằng lời của bạn. Bước thứ ba, hỏi thêm một câu để hiểu rõ hơn.' },
  { id: 's3', title: '2.1 Nói ngắn, ý rõ', text: 'Một ý, một câu. Khi cần nói nhiều ý, hãy báo trước có mấy ý, ví dụ: tôi có ba điều muốn chia sẻ.' },
]

const renderChapters = [
  { title: 'Lời mở đầu', pct: 100 },
  { title: 'Chương 1. Lắng nghe chủ động', pct: 100 },
  { title: 'Chương 2. Giao tiếp rõ ràng', pct: 38 },
  { title: 'Chương 3. Quản lý thời gian', pct: 0 },
]

const playerChapters = [
  { title: 'Lời mở đầu', dur: '0:18', done: true },
  { title: '1.1 Vì sao lắng nghe khó', dur: '3:12', done: true },
  { title: '1.2 Ba bước lắng nghe', dur: '3:26', current: true },
  { title: '1.3 Bài tập mỗi ngày', dur: '2:27' },
  { title: '2.1 Nói ngắn, ý rõ', dur: '4:40' },
  { title: '2.2 Hỏi để hiểu', dur: '5:26' },
  { title: '3.1 Việc quan trọng trước', dur: '4:13' },
  { title: '3.2 Nói không đúng lúc', dur: '4:13' },
]

function go(v: View) {
  view.value = v
  if (v === 'create') step.value = rendering.value && !renderDone.value ? 6 : 1
}
function copyPrompt() {
  copied.value = true
  setTimeout(() => (copied.value = false), 1500)
}
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex items-center justify-center p-6">
    <!-- Nút duyệt sáng/tối (ngoài khung, chỉ để xem wireframe) -->
    <div class="fixed top-3 right-3 flex gap-2 z-10">
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs text-foreground" @click="view = 'setup'">Màn cài lần đầu</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs text-foreground" @click="update = 'info'">Hộp cập nhật</button>
    </div>

    <!-- Khung cửa sổ ứng dụng -->
    <div class="relative w-[1100px] h-[720px] rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground flex flex-col">
      <!-- Thanh tiêu đề hệ điều hành (giả) -->
      <div class="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border bg-muted/50">
        <span class="h-3 w-3 rounded-full bg-rag-red"></span>
        <span class="h-3 w-3 rounded-full bg-rag-amber"></span>
        <span class="h-3 w-3 rounded-full bg-rag-green"></span>
        <span class="flex-1 text-center text-xs text-muted-foreground">Sano</span>
      </div>

      <!-- ═══ Màn cài bộ đọc lần đầu ═══ -->
      <div v-if="view === 'setup'" class="flex-1 grid place-items-center p-10">
        <div class="w-full max-w-lg">
          <img src="@/assets/logo.svg" alt="" class="h-12 w-12 rounded-xl" />
          <h1 class="mt-5 text-2xl font-semibold tracking-tight">Chào mừng đến với Sano</h1>
          <p class="mt-2 text-sm text-muted-foreground">
            Để đọc sách thành giọng nói, Sano cần tải bộ đọc tiếng Việt về máy. Việc này chỉ làm một lần, sau đó dùng không cần mạng.
          </p>

          <div class="mt-6 rounded-lg border border-border divide-y divide-border text-sm">
            <div class="flex items-center justify-between px-4 py-3"><span class="flex items-center gap-2"><HardDrive class="w-4 h-4 text-muted-foreground" />Dung lượng cần</span><span class="font-medium tabular-nums">~3,2 GB</span></div>
            <div class="flex items-center justify-between px-4 py-3"><span class="flex items-center gap-2"><Clock class="w-4 h-4 text-muted-foreground" />Thời gian tải (mạng 50 Mbps)</span><span class="font-medium tabular-nums">~10 phút</span></div>
            <div class="flex items-center justify-between px-4 py-3"><span class="flex items-center gap-2"><Cpu class="w-4 h-4 text-muted-foreground" />Máy của bạn</span><span class="font-medium">Apple M2 · 16 GB RAM · đủ</span></div>
          </div>

          <div class="mt-6 space-y-3">
            <div v-for="(it, i) in [
              { label: 'Python', pct: 100 },
              { label: 'Bộ đọc VieNeu-TTS', pct: 100 },
              { label: 'Mô hình giọng đọc (2,6 GB)', pct: 46 },
              { label: 'Kiểm tra đọc thử', pct: 0 },
            ]" :key="i">
              <div class="flex justify-between text-xs mb-1">
                <span class="flex items-center gap-1.5">
                  <Check v-if="it.pct === 100" class="w-3.5 h-3.5 text-rag-green" />
                  <Loader2 v-else-if="it.pct > 0" class="w-3.5 h-3.5 animate-spin text-primary" />
                  <span v-else class="w-3.5 h-3.5 rounded-full border border-border inline-block"></span>
                  {{ it.label }}
                </span>
                <span class="tabular-nums text-muted-foreground">{{ it.pct }}%</span>
              </div>
              <div class="h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary rounded-full" :style="{ width: it.pct + '%' }"></div></div>
            </div>
          </div>

          <div class="mt-6 flex items-center justify-between">
            <span class="text-xs text-muted-foreground">Còn khoảng 5 phút · có thể đóng cửa sổ, lần sau mở sẽ tải tiếp<br /><a :href="DOCS + '/cai-dat'" target="_blank" rel="noopener" class="text-primary hover:underline">Cài không được? Xem hướng dẫn</a></span>
            <Button @click="view = 'library'">Bỏ qua (xem wireframe)</Button>
          </div>
        </div>
      </div>

      <!-- ═══ Khung chính: thanh bên + nội dung ═══ -->
      <div v-else class="flex-1 flex min-h-0">
        <!-- Thanh bên -->
        <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
          <div class="px-4 py-4 flex items-center gap-2">
            <img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" />
            <span class="font-semibold tracking-tight">Sano</span>
          </div>
          <nav class="px-2 space-y-0.5">
            <button
              v-for="n in nav" :key="n.key"
              class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
              :class="view === n.key || (n.key === 'library' && view === 'player') ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
              @click="go(n.key)"
            >
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </button>
          </nav>

          <div class="px-2 mt-2">
            <a :href="DOCS" target="_blank" rel="noopener" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground hover:bg-muted hover:text-foreground">
              <LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" />
            </a>
          </div>

          <div class="flex-1"></div>

          <!-- Thẻ tiến độ render (hiện ở mọi màn khi đang render) -->
          <button v-if="rendering && !renderDone && view !== 'create'" class="m-3 rounded-lg border border-border bg-background p-3 text-left hover:border-primary/50" @click="view = 'create'; step = 6">
            <div class="flex items-center gap-1.5 text-xs font-medium"><Loader2 class="w-3.5 h-3.5 animate-spin text-primary" /> Đang render</div>
            <p class="mt-1 text-xs text-muted-foreground truncate">Kỹ năng giao tiếp</p>
            <div class="mt-2 h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full w-[42%] bg-primary rounded-full"></div></div>
            <div class="mt-1 flex justify-between text-[11px] text-muted-foreground tabular-nums"><span>42%</span><span>còn ~4 phút</span></div>
          </button>
          <button class="mx-3 mb-2 flex items-center gap-2 rounded-md px-2 h-8 text-xs text-primary hover:bg-primary/10" @click="update = 'info'">
            <ArrowUpCircle class="w-4 h-4" /> Có bản mới 0.2.0
          </button>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">
            Phiên bản 0.1.0 · mã nguồn mở<br />
            Tác giả <a :href="AUTHOR_FB" target="_blank" rel="noopener" class="hover:text-foreground underline-offset-2 hover:underline">Bùi Tấn Việt</a>
          </div>
        </aside>

        <!-- Nội dung -->
        <main class="flex-1 min-w-0 flex flex-col">
          <!-- ─── Thư viện ─── -->
          <section v-if="view === 'library'" class="flex-1 overflow-auto p-6">
            <div class="flex items-center justify-between">
              <div>
                <h1 class="text-xl font-semibold tracking-tight">Thư viện</h1>
                <p class="text-sm text-muted-foreground">4 cuốn · 4 giờ 02 phút · lưu ở ~/Sano/Sach</p>
              </div>
              <Button @click="go('create')"><FilePlus2 class="w-4 h-4" /> Tạo sách mới</Button>
            </div>

            <div class="mt-6 grid grid-cols-4 gap-5">
              <button v-for="b in books" :key="b.title" class="text-left group" @click="view = 'player'">
                <div class="aspect-[3/4] rounded-lg shadow-md group-hover:shadow-xl group-hover:-translate-y-0.5 transition">
                  <WfBookCover :title="b.title" :author="b.author" />
                </div>
                <p class="mt-2 text-sm font-medium truncate">{{ b.title }}</p>
                <p class="text-xs text-muted-foreground">{{ b.dur }} · {{ b.chapters }} chương</p>
                <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: b.progress + '%' }"></div></div>
                <p class="mt-1 text-[11px] text-muted-foreground">{{ b.progress === 100 ? 'Đã nghe xong' : b.progress === 0 ? 'Chưa nghe' : `Đã nghe ${b.progress}%` }}</p>
              </button>
            </div>
          </section>

          <!-- ─── Trình phát ─── -->
          <section v-else-if="view === 'player'" class="flex-1 flex min-h-0">
            <div class="flex-1 flex flex-col p-6 min-w-0">
              <button class="text-sm text-muted-foreground flex items-center gap-1 hover:text-foreground w-fit" @click="view = 'library'"><ChevronLeft class="w-4 h-4" /> Thư viện</button>
              <div class="flex-1 flex flex-col items-center justify-center">
                <div class="aspect-[3/4] w-48 rounded-xl shadow-2xl">
                  <WfBookCover title="Kỹ năng mềm cho người trẻ" author="Nguyễn Văn A" size="lg" />
                </div>
                <h2 class="mt-5 text-lg font-semibold">Kỹ năng mềm cho người trẻ</h2>
                <p class="text-sm text-muted-foreground">1.2 Ba bước lắng nghe</p>
                <div class="mt-5 w-full max-w-md">
                  <div class="h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full w-[30%] bg-primary rounded-full"></div></div>
                  <div class="flex justify-between text-[11px] text-muted-foreground mt-1 tabular-nums"><span>01:02</span><span>-02:24</span></div>
                </div>
                <div class="mt-4 flex items-center gap-4">
                  <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><SkipBack class="w-5 h-5" /></button>
                  <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><RotateCcw class="w-5 h-5" /></button>
                  <button class="h-14 w-14 grid place-items-center rounded-full bg-primary text-primary-foreground shadow"><Play class="w-6 h-6 ml-0.5" /></button>
                  <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><RotateCw class="w-5 h-5" /></button>
                  <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><SkipForward class="w-5 h-5" /></button>
                </div>
                <div class="mt-4 flex items-center gap-2 text-xs">
                  <button class="h-8 px-3 rounded-full border border-border flex items-center gap-1.5"><Gauge class="w-3.5 h-3.5" />1,25×</button>
                  <span class="text-muted-foreground">Tua −15s / +30s · nhớ vị trí nghe</span>
                </div>
              </div>
              <div class="flex items-center gap-2 border-t border-border pt-4">
                <Button variant="outline" size="sm"><Download class="w-4 h-4" /> Xuất M4B (nghe trên điện thoại)</Button>
                <Button variant="outline" size="sm" title="Sao lưu hoặc chuyển sách sang máy khác"><Package class="w-4 h-4" /> Xuất gói zip</Button>
                <Button variant="ghost" size="sm"><FolderOpen class="w-4 h-4" /> Mở thư mục</Button>
                <div class="flex-1"></div>
                <Button variant="ghost" size="sm" class="text-destructive"><Trash2 class="w-4 h-4" /> Xoá</Button>
              </div>
            </div>
            <div class="w-72 shrink-0 border-l border-border overflow-auto">
              <div class="px-4 py-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Mục lục · 8 mục · 1 giờ 12 phút</div>
              <button v-for="c in playerChapters" :key="c.title"
                class="w-full flex items-center justify-between gap-2 px-4 py-2.5 text-sm text-left hover:bg-muted/60"
                :class="c.current ? 'bg-primary/10 text-primary font-medium' : c.done ? 'text-muted-foreground' : ''">
                <span class="truncate flex items-center gap-2">
                  <Volume2 v-if="c.current" class="w-3.5 h-3.5 shrink-0" />
                  <Check v-else-if="c.done" class="w-3.5 h-3.5 shrink-0" />
                  <span v-else class="w-3.5 shrink-0"></span>
                  {{ c.title }}
                </span>
                <span class="text-xs tabular-nums shrink-0">{{ c.dur }}</span>
              </button>
            </div>
          </section>

          <!-- ─── Cài đặt ─── -->
          <section v-else-if="view === 'settings'" class="flex-1 overflow-auto p-6 max-w-2xl">
            <h1 class="text-xl font-semibold tracking-tight">Cài đặt</h1>
            <div class="mt-6 space-y-6">
              <div>
                <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Lưu trữ</h2>
                <div class="rounded-lg border border-border divide-y divide-border text-sm">
                  <div class="flex items-center justify-between px-4 py-3"><span>Thư mục lưu sách</span><span class="flex items-center gap-2 text-muted-foreground">~/Sano/Sach <Button variant="outline" size="sm">Đổi</Button></span></div>
                  <div class="flex items-center justify-between px-4 py-3"><span>Dung lượng sách đã tạo</span><span class="text-muted-foreground tabular-nums">412 MB</span></div>
                </div>
              </div>
              <div>
                <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Bộ đọc</h2>
                <div class="rounded-lg border border-border divide-y divide-border text-sm">
                  <div class="flex items-center justify-between px-4 py-3"><span>VieNeu-TTS v2 · mô hình đã tải</span><Badge variant="secondary">3,2 GB</Badge></div>
                  <div class="flex items-center justify-between px-4 py-3"><span>Giọng mặc định</span><span class="text-muted-foreground">Trúc Ly · Nữ · miền Bắc</span></div>
                  <div class="flex items-center justify-between px-4 py-3"><span>Kiểm tra bộ đọc</span><Button variant="outline" size="sm"><RefreshCw class="w-4 h-4" /> Chạy kiểm tra</Button></div>
                  <div class="flex items-center justify-between px-4 py-3"><span class="text-muted-foreground">Gỡ bộ đọc và mô hình (giải phóng 3,2 GB)</span><Button variant="ghost" size="sm" class="text-destructive"><Trash2 class="w-4 h-4" /> Gỡ</Button></div>
                </div>
              </div>
              <div>
                <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Ứng dụng</h2>
                <div class="rounded-lg border border-border divide-y divide-border text-sm">
                  <div class="flex items-center justify-between px-4 py-3"><span>Giao diện</span><span class="text-muted-foreground">Theo hệ điều hành</span></div>
                  <label class="flex items-center justify-between px-4 py-3 cursor-pointer"><span>Tự kiểm tra bản mới mỗi ngày</span><input type="checkbox" checked class="h-4 w-4 accent-[hsl(var(--primary))]" /></label>
                  <div class="flex items-center justify-between px-4 py-3">
                    <span>Phiên bản 0.1.0</span>
                    <span class="flex items-center gap-2"><Badge class="bg-rag-amber/15 text-rag-amber border-0">Có bản 0.2.0</Badge><Button size="sm" @click="update = 'info'">Cập nhật</Button></span>
                  </div>
                </div>
              </div>
              <div>
                <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Giới thiệu</h2>
                <div class="rounded-lg border border-border p-4 flex gap-4">
                  <img src="@/assets/logo.svg" alt="" class="h-12 w-12 rounded-xl shrink-0" />
                  <div class="text-sm min-w-0">
                    <p><span class="font-medium">Sano</span> — tự làm sách nói từ file Word, chạy trên máy bạn. Mã nguồn mở, miễn phí.</p>
                    <p class="mt-2 text-muted-foreground">
                      Do <a :href="AUTHOR_FB" target="_blank" rel="noopener" class="text-foreground font-medium hover:underline">Bùi Tấn Việt</a> làm — CEO
                      <a href="https://sepay.vn" target="_blank" rel="noopener" class="text-foreground hover:underline">SePay</a> và
                      <a href="https://123host.vn" target="_blank" rel="noopener" class="text-foreground hover:underline">123HOST</a>.
                    </p>
                  </div>
                </div>
                <div class="mt-2 rounded-lg border border-border divide-y divide-border text-sm">
                  <a :href="DOCS" target="_blank" rel="noopener" class="flex items-center gap-2.5 px-4 py-3 hover:bg-muted/50"><LifeBuoy class="w-4 h-4 text-muted-foreground" /><span class="flex-1">Hướng dẫn sử dụng</span><ExternalLink class="w-3.5 h-3.5 text-muted-foreground" /></a>
                  <a :href="REPO + '/issues'" target="_blank" rel="noopener" class="flex items-center gap-2.5 px-4 py-3 hover:bg-muted/50"><Bug class="w-4 h-4 text-muted-foreground" /><span class="flex-1">Báo lỗi / góp ý</span><ExternalLink class="w-3.5 h-3.5 text-muted-foreground" /></a>
                  <a :href="REPO" target="_blank" rel="noopener" class="flex items-center gap-2.5 px-4 py-3 hover:bg-muted/50"><Github class="w-4 h-4 text-muted-foreground" /><span class="flex-1">Mã nguồn trên GitHub · giấy phép MIT</span><ExternalLink class="w-3.5 h-3.5 text-muted-foreground" /></a>
                  <div class="flex items-center gap-2.5 px-4 py-3"><Scale class="w-4 h-4 text-muted-foreground" /><span class="flex-1">Giấy phép bên thứ ba</span><span class="text-xs text-muted-foreground">VieNeu-TTS (Apache-2.0) · ffmpeg · …</span></div>
                </div>
              </div>
            </div>
          </section>

          <!-- ─── Tạo sách mới: 6 bước ─── -->
          <section v-else class="flex-1 flex flex-col min-h-0">
            <!-- Thanh bước -->
            <div class="shrink-0 border-b border-border px-6 h-14 flex items-center gap-1">
              <template v-for="(s, i) in steps" :key="s.n">
                <button class="flex items-center gap-2 px-2 h-8 rounded-md text-sm" :class="step === s.n ? 'text-foreground font-medium' : 'text-muted-foreground hover:text-foreground'" @click="step = s.n">
                  <span class="h-6 w-6 rounded-full grid place-items-center text-xs"
                    :class="step > s.n ? 'bg-primary/15 text-primary' : step === s.n ? 'bg-primary text-primary-foreground' : 'bg-muted'">
                    <Check v-if="step > s.n" class="w-3.5 h-3.5" /><template v-else>{{ s.n }}</template>
                  </span>
                  {{ s.label }}
                </button>
                <ChevronRight v-if="i < steps.length - 1" class="w-4 h-4 text-muted-foreground/50" />
              </template>
            </div>

            <div class="flex-1 overflow-auto p-6">
              <!-- B1 Nạp file -->
              <div v-if="step === 1" class="max-w-2xl">
                <h1 class="text-xl font-semibold tracking-tight">Nạp file Word</h1>
                <p class="text-sm text-muted-foreground">
                  Sano đọc mục lục từ kiểu Heading 1 / Heading 2 trong file.
                  <a :href="DOCS + '/chuan-bi-file'" target="_blank" rel="noopener" class="text-primary hover:underline">Cách chuẩn bị file để đọc hay nhất</a>
                </p>

                <button v-if="!fileLoaded" class="mt-5 w-full h-56 rounded-xl border-2 border-dashed border-border grid place-items-center hover:border-primary/50 hover:bg-primary/5" @click="fileLoaded = true">
                  <span class="text-center">
                    <Upload class="w-8 h-8 mx-auto text-muted-foreground" />
                    <span class="block mt-3 font-medium">Kéo file .docx vào đây</span>
                    <span class="block text-sm text-muted-foreground">hoặc bấm để chọn file</span>
                  </span>
                </button>

                <template v-else>
                  <div class="mt-5 rounded-lg border border-border p-4 flex items-center gap-3">
                    <FileText class="w-8 h-8 text-primary shrink-0" />
                    <div class="flex-1 min-w-0">
                      <p class="font-medium truncate">ky-nang-giao-tiep.docx</p>
                      <p class="text-xs text-muted-foreground">1,4 MB · 3 chương · 7 tiểu mục · 26.500 ký tự</p>
                    </div>
                    <Button variant="ghost" size="sm" @click="fileLoaded = false"><X class="w-4 h-4" /> Chọn file khác</Button>
                  </div>

                  <div class="mt-4 rounded-lg border border-rag-amber/40 bg-rag-amber/10 p-4 text-sm">
                    <p class="font-medium flex items-center gap-2 text-rag-amber"><AlertTriangle class="w-4 h-4" /> Có phần sẽ không được đọc</p>
                    <ul class="mt-2 space-y-1 text-foreground/80 list-disc pl-5">
                      <li>2 bảng (Phụ lục) — nội dung trong bảng bị bỏ qua</li>
                      <li>5 hình — chỉ đọc chú thích, không mô tả hình</li>
                      <li>1 đoạn chữ to đậm có vẻ là tiêu đề nhưng không dùng kiểu Heading</li>
                    </ul>
                    <p class="mt-3 text-foreground/80">
                      Muốn đọc đủ: dán file vào ChatGPT, Gemini hoặc Claude cùng lời nhắc mẫu để biến bảng, hình thành lời văn, rồi nạp lại.
                    </p>
                    <Button variant="outline" size="sm" class="mt-3" @click="copyPrompt">
                      <component :is="copied ? Check : Copy" class="w-4 h-4" /> {{ copied ? 'Đã sao chép' : 'Sao chép lời nhắc mẫu' }}
                    </Button>
                  </div>

                  <div class="mt-5 grid grid-cols-2 gap-4">
                    <label class="text-sm">Tên sách
                      <input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" value="Kỹ năng giao tiếp" />
                    </label>
                    <label class="text-sm">Tác giả
                      <input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" value="Nguyễn Văn A" />
                    </label>
                  </div>
                  <div class="mt-4 flex items-center gap-4">
                    <div class="h-24 w-[72px] rounded-md shadow-sm shrink-0"><WfBookCover title="Kỹ năng giao tiếp" author="Nguyễn Văn A" size="sm" /></div>
                    <div class="text-sm">
                      <p class="font-medium">Ảnh bìa</p>
                      <p class="text-xs text-muted-foreground">Chưa có ảnh — Sano tự tạo bìa theo tên sách</p>
                      <Button variant="outline" size="sm" class="mt-2">Chọn ảnh bìa</Button>
                    </div>
                  </div>
                </template>
              </div>

              <!-- B2 Mục lục -->
              <div v-else-if="step === 2" class="max-w-3xl">
                <div class="flex items-end justify-between">
                  <div>
                    <h1 class="text-xl font-semibold tracking-tight">Chọn phần sẽ đọc</h1>
                    <p class="text-sm text-muted-foreground">Bỏ tick trang bìa, mục lục gốc hay phần không cần nghe.</p>
                  </div>
                  <div class="text-right text-sm">
                    <p class="tabular-nums"><span class="font-medium">{{ totalChars.toLocaleString('vi-VN') }}</span> ký tự</p>
                    <p class="text-xs text-muted-foreground">Nghe ~{{ estListen }} phút · render ~{{ estRender }} phút trên máy này</p>
                  </div>
                </div>
                <div class="mt-5 rounded-lg border border-border divide-y divide-border">
                  <div v-for="c in toc" :key="c.title">
                    <label class="flex items-center gap-3 px-4 py-3 cursor-pointer" :class="!c.on && 'text-muted-foreground'">
                      <input v-model="c.on" type="checkbox" class="h-4 w-4 accent-[hsl(var(--primary))]" />
                      <span class="flex-1 font-medium text-sm">{{ c.title }}</span>
                      <span v-if="c.note" class="text-xs text-rag-amber">{{ c.note }}</span>
                      <span class="text-xs tabular-nums text-muted-foreground w-24 text-right">{{ c.chars.toLocaleString('vi-VN') }} ký tự</span>
                    </label>
                    <label v-for="s in c.sections" :key="s.title" class="flex items-center gap-3 pl-11 pr-4 py-2 cursor-pointer text-sm" :class="(!c.on || !s.on) && 'text-muted-foreground'">
                      <input v-model="s.on" type="checkbox" :disabled="!c.on" class="h-4 w-4 accent-[hsl(var(--primary))]" />
                      <span class="flex-1">{{ s.title }}</span>
                      <span class="text-xs tabular-nums text-muted-foreground w-24 text-right">{{ s.chars.toLocaleString('vi-VN') }} ký tự</span>
                    </label>
                  </div>
                </div>
              </div>

              <!-- B3 Giọng đọc -->
              <div v-else-if="step === 3" class="max-w-2xl">
                <h1 class="text-xl font-semibold tracking-tight">Chọn giọng đọc</h1>
                <p class="text-sm text-muted-foreground">Bấm nghe để thử từng giọng với một câu trong sách của bạn.</p>
                <div class="mt-5 grid gap-2">
                  <label v-for="v in voices" :key="v.id" class="flex items-center gap-3 rounded-lg border px-4 py-3 cursor-pointer"
                    :class="selectedVoice === v.id ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/50'">
                    <input v-model="selectedVoice" type="radio" :value="v.id" class="h-4 w-4 accent-[hsl(var(--primary))]" />
                    <span class="flex-1">
                      <span class="font-medium text-sm">{{ v.name }}</span>
                      <Badge v-if="v.rec" variant="secondary" class="ml-2">Hay dùng cho sách</Badge>
                      <span class="block text-xs text-muted-foreground">{{ v.desc }}</span>
                    </span>
                    <button class="h-8 px-3 rounded-full border border-border text-xs flex items-center gap-1.5 hover:bg-muted" @click.prevent="playingSample = playingSample === v.id ? null : v.id">
                      <component :is="playingSample === v.id ? Pause : Play" class="w-3.5 h-3.5" /> Nghe mẫu
                    </button>
                  </label>
                </div>
                <div class="mt-5 text-sm">
                  <p class="font-medium">Câu nghe mẫu</p>
                  <input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" value="Bước thứ nhất, dừng việc đang làm và nhìn người nói." />
                </div>
              </div>

              <!-- B4 Lời mở đầu -->
              <div v-else-if="step === 4" class="max-w-2xl">
                <h1 class="text-xl font-semibold tracking-tight">Lời mở đầu</h1>
                <p class="text-sm text-muted-foreground">Đoạn đọc đầu tiên, trước chương 1. Mỗi dòng đọc cách nhau một nhịp nghỉ.</p>
                <label class="mt-5 flex items-center gap-2 text-sm"><input type="checkbox" checked class="h-4 w-4 accent-[hsl(var(--primary))]" /> Có lời mở đầu</label>
                <textarea class="mt-3 w-full h-36 rounded-md border border-input bg-background p-3 text-sm leading-relaxed">Bạn đang nghe sách nói.

Cuốn sách: Kỹ năng giao tiếp.

Tác giả: Nguyễn Văn A.</textarea>
                <p class="mt-2 text-xs text-muted-foreground flex items-center gap-1.5"><Info class="w-3.5 h-3.5" /> Đặt chữ "Cuốn sách:" trước tên sách để bộ đọc không nuốt mất tên ở đầu câu.</p>
                <Button variant="outline" size="sm" class="mt-4"><Play class="w-4 h-4" /> Nghe lời mở đầu</Button>
              </div>

              <!-- B5 Nghe thử -->
              <div v-else-if="step === 5" class="max-w-3xl">
                <h1 class="text-xl font-semibold tracking-tight">Nghe thử trước khi render</h1>
                <p class="text-sm text-muted-foreground">
                  Render cả cuốn mất khoảng {{ estRender }} phút và không dừng giữa chừng được. Hãy nghe vài đoạn để chắc giọng và cách đọc đã ổn.
                </p>
                <div class="mt-5 space-y-3">
                  <div v-for="s in samples" :key="s.id" class="rounded-lg border border-border p-4">
                    <div class="flex items-center gap-3">
                      <button class="h-9 w-9 grid place-items-center rounded-full bg-primary text-primary-foreground shrink-0" @click="playingSample = playingSample === s.id ? null : s.id">
                        <component :is="playingSample === s.id ? Pause : Play" class="w-4 h-4" />
                      </button>
                      <span class="flex-1 font-medium text-sm">{{ s.title }}</span>
                      <Badge variant="secondary">Đã render · 0:{{ 10 + s.text.length % 40 }}</Badge>
                    </div>
                    <div class="mt-3 text-sm">
                      <p class="text-xs font-medium text-muted-foreground flex items-center gap-1.5"><Pencil class="w-3 h-3" /> Lời đọc (sửa được — sửa xong bấm render lại đoạn này)</p>
                      <textarea class="mt-1 w-full h-16 rounded-md border border-input bg-background p-2 text-sm" :value="s.text"></textarea>
                    </div>
                  </div>
                </div>
                <Button variant="outline" size="sm" class="mt-3">+ Chọn thêm đoạn khác để nghe thử</Button>
              </div>

              <!-- B6 Render -->
              <div v-else class="max-w-2xl">
                <template v-if="!renderDone">
                  <h1 class="text-xl font-semibold tracking-tight">Đang render cả cuốn</h1>
                  <p class="text-sm text-muted-foreground">Có thể dùng máy bình thường hoặc thu nhỏ cửa sổ. Nếu đóng Sano, bạn sẽ được hỏi có dừng render không.</p>
                  <div class="mt-6">
                    <div class="flex justify-between text-sm"><span class="font-medium">42%</span><span class="text-muted-foreground tabular-nums">còn ~4 phút · 10.500 / 24.900 ký tự</span></div>
                    <div class="mt-2 h-2.5 rounded-full bg-muted overflow-hidden"><div class="h-full w-[42%] bg-primary rounded-full"></div></div>
                  </div>
                  <div class="mt-6 rounded-lg border border-border divide-y divide-border">
                    <div v-for="c in renderChapters" :key="c.title" class="flex items-center gap-3 px-4 py-3 text-sm">
                      <Check v-if="c.pct === 100" class="w-4 h-4 text-rag-green" />
                      <Loader2 v-else-if="c.pct > 0" class="w-4 h-4 animate-spin text-primary" />
                      <span v-else class="w-4 h-4 rounded-full border border-border"></span>
                      <span class="flex-1" :class="c.pct === 0 && 'text-muted-foreground'">{{ c.title }}</span>
                      <span class="text-xs tabular-nums text-muted-foreground">{{ c.pct }}%</span>
                    </div>
                  </div>
                  <div class="mt-5 flex gap-2">
                    <Button variant="outline" class="text-destructive">Huỷ render</Button>
                    <Button variant="ghost" @click="renderDone = true">Xem màn đã xong (wireframe)</Button>
                  </div>
                </template>

                <template v-else>
                  <div class="h-12 w-12 rounded-full bg-rag-green/15 grid place-items-center"><Check class="w-6 h-6 text-rag-green" /></div>
                  <h1 class="mt-4 text-xl font-semibold tracking-tight">Đã tạo xong "Kỹ năng giao tiếp"</h1>
                  <p class="text-sm text-muted-foreground">28 phút nghe · 3 chương · đã thêm vào thư viện</p>
                  <div class="mt-6 grid gap-3">
                    <button class="flex items-center gap-3 rounded-lg border border-border p-4 text-left hover:bg-muted/50" @click="view = 'player'">
                      <BookOpen class="w-5 h-5 text-primary" /><span class="flex-1"><span class="block font-medium text-sm">Nghe ngay</span><span class="block text-xs text-muted-foreground">Mở trong thư viện</span></span><ChevronRight class="w-4 h-4 text-muted-foreground" />
                    </button>
                    <button class="flex items-center gap-3 rounded-lg border border-border p-4 text-left hover:bg-muted/50">
                      <Download class="w-5 h-5 text-primary" /><span class="flex-1"><span class="block font-medium text-sm">Xuất file M4B</span><span class="block text-xs text-muted-foreground">Một file có mục lục chương, chép sang điện thoại nghe bằng app sách nói bất kỳ</span></span><ChevronRight class="w-4 h-4 text-muted-foreground" />
                    </button>
                    <button class="flex items-center gap-3 rounded-lg border border-border p-4 text-left hover:bg-muted/50">
                      <Package class="w-5 h-5 text-primary" /><span class="flex-1"><span class="block font-medium text-sm">Xuất gói zip</span><span class="block text-xs text-muted-foreground">Sao lưu hoặc chuyển sách sang máy khác</span></span><ChevronRight class="w-4 h-4 text-muted-foreground" />
                    </button>
                  </div>
                </template>
              </div>
            </div>

            <!-- Chân: nút lùi / tiếp -->
            <div v-if="step < 6" class="shrink-0 border-t border-border px-6 h-16 flex items-center justify-between">
              <Button variant="ghost" :disabled="step === 1" @click="step--"><ChevronLeft class="w-4 h-4" /> Quay lại</Button>
              <Button v-if="step < 5" @click="step++">Tiếp tục <ChevronRight class="w-4 h-4" /></Button>
              <Button v-else @click="step = 6">Nghe ổn, render cả cuốn <ChevronRight class="w-4 h-4" /></Button>
            </div>
          </section>
        </main>
      </div>
      <!-- ═══ Hộp cập nhật ═══ -->
      <div v-if="update !== 'closed'" class="absolute inset-0 bg-background/70 backdrop-blur-sm grid place-items-center z-20">
        <div role="dialog" aria-modal="true" aria-labelledby="upd-title" class="w-[460px] rounded-xl border border-border bg-card text-card-foreground shadow-2xl p-6">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
              <div class="h-10 w-10 rounded-full bg-primary/10 grid place-items-center"><ArrowUpCircle class="w-5 h-5 text-primary" /></div>
              <div>
                <h2 id="upd-title" class="font-semibold">{{ update === 'ready' ? 'Sẵn sàng cập nhật' : 'Có bản mới 0.2.0' }}</h2>
                <p class="text-xs text-muted-foreground">Đang dùng 0.1.0 · phát hành 20/10/2026 · 18 MB</p>
              </div>
            </div>
            <button v-if="update !== 'downloading'" aria-label="Đóng" class="h-8 w-8 grid place-items-center rounded-md hover:bg-muted" @click="update = 'closed'"><X class="w-4 h-4" /></button>
          </div>

          <template v-if="update === 'info'">
            <p class="mt-5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Có gì mới</p>
            <ul class="mt-2 space-y-1.5 text-sm list-disc pl-5">
              <li v-for="c in changelog" :key="c">{{ c }}</li>
            </ul>
            <a class="mt-2 inline-block text-xs text-primary hover:underline" href="#">Xem đầy đủ thay đổi</a>

            <div v-if="rendering && !renderDone" class="mt-4 rounded-lg border border-rag-amber/40 bg-rag-amber/10 p-3 text-sm">
              <p class="font-medium text-rag-amber flex items-center gap-1.5"><AlertTriangle class="w-4 h-4" /> Đang render "Kỹ năng giao tiếp"</p>
              <p class="mt-1 text-foreground/80">Cập nhật cần khởi động lại Sano. Chọn cập nhật khi render xong để không mất tiến độ.</p>
              <label class="mt-2 flex items-center gap-2"><input v-model="updateAfterRender" type="checkbox" class="h-4 w-4 accent-[hsl(var(--primary))]" /> Tự cập nhật khi render xong</label>
            </div>

            <div class="mt-6 flex justify-end gap-2">
              <Button variant="outline" @click="update = 'closed'">Để sau</Button>
              <Button v-if="rendering && !renderDone" :disabled="!updateAfterRender" @click="update = 'closed'">Hẹn cập nhật</Button>
              <Button v-else @click="update = 'downloading'">Cập nhật ngay</Button>
            </div>
            <button v-if="rendering && !renderDone" class="mt-3 w-full text-center text-[11px] text-muted-foreground hover:text-foreground" @click="update = 'downloading'">(wireframe) Xem bước tải khi không render</button>
          </template>

          <template v-else-if="update === 'downloading'">
            <div class="mt-6 space-y-3 text-sm">
              <div>
                <div class="flex justify-between text-xs mb-1"><span class="flex items-center gap-1.5"><Loader2 class="w-3.5 h-3.5 animate-spin text-primary" />Đang tải bản 0.2.0</span><span class="tabular-nums text-muted-foreground">11 / 18 MB</span></div>
                <div class="h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full w-[61%] bg-primary rounded-full"></div></div>
              </div>
              <p class="flex items-center gap-1.5 text-xs text-muted-foreground"><span class="w-3.5 h-3.5 rounded-full border border-border inline-block"></span>Kiểm tra chữ ký bản phát hành</p>
            </div>
            <div class="mt-6 flex justify-between">
              <Button variant="ghost" @click="update = 'info'">Huỷ</Button>
              <Button variant="ghost" @click="update = 'ready'">(wireframe) Tải xong</Button>
            </div>
          </template>

          <template v-else>
            <div class="mt-5 rounded-lg border border-border p-3 text-sm space-y-2">
              <p class="flex items-center gap-2"><Check class="w-4 h-4 text-rag-green" /> Đã tải bản 0.2.0</p>
              <p class="flex items-center gap-2"><ShieldCheck class="w-4 h-4 text-rag-green" /> Chữ ký hợp lệ — bản phát hành chính thức của Sano</p>
            </div>
            <p class="mt-3 text-xs text-muted-foreground">Sách, tiến độ nghe và bộ đọc giữ nguyên. Sano sẽ tự mở lại sau vài giây.</p>
            <div class="mt-6 flex justify-end gap-2">
              <Button variant="outline" @click="update = 'closed'">Khởi động lại sau</Button>
              <Button @click="update = 'closed'"><RefreshCw class="w-4 h-4" /> Khởi động lại</Button>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>
