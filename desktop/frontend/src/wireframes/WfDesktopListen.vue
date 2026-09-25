<script setup lang="ts">
// D4 — Góp ý sau khi dùng (25/09): giọng đọc + nghe liền mạch. Bám khung D1 (1100×720).
// 1. Sách ghi tên giọng đọc: thẻ trong Thư viện + màn nghe (lấy từ manifest.voice_id của gói).
// 2. Chọn giọng: gom theo miền (Bắc 15 · Trung 2 · Nam 8) + lọc Nam/Nữ. Mặc định mở đúng
//    miền của giọng dùng lần trước; nhớ lựa chọn cho lần sau.
// 3. Thanh nghe nhỏ ở đáy cửa sổ: rời màn nghe (về Thư viện, Tạo sách, Cài đặt) vẫn nghe
//    tiếp. Bấm vào thanh → mở lại màn nghe. Bấm "Nghe mẫu" giọng thì tạm dừng sách.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển màn duyệt.
import { computed, ref } from 'vue'
import {
  Library, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Search, ChevronDown, ChevronLeft, Play, Pause,
  RotateCcw, RotateCw, SkipBack, SkipForward, X, Mic, Check, Sun, Moon, Gauge, Volume2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

type Screen = 'library' | 'voice' | 'player'
const q = new URLSearchParams(window.location.search)
const screen = ref<Screen>((q.get('mode') as Screen | null) || 'library')
const dark = ref(false)
const miniOn = ref(true) // đang có sách phát (thanh nghe nhỏ hiện khi rời màn nghe)
const playing = ref(true)

const books = [
  { title: 'Kỹ năng mềm cho người trẻ', author: 'Nguyễn Văn A', dur: '1 giờ 12 phút', voice: 'Hải Đăng', progress: 35 },
  { title: 'Lãnh đạo cho quản lý mới', author: 'Trần Thị B', dur: '58 phút', voice: 'Ngọc Huyền', progress: 100 },
  { title: 'Tài chính cá nhân cơ bản', author: 'Lê Văn C', dur: '1 giờ 05 phút', voice: 'Thiện Minh', progress: 0 },
  { title: 'Khởi nghiệp từ số 0', author: 'Phạm Thị D', dur: '47 phút', voice: 'Hải Đăng', progress: 12 },
  { title: 'Quản lý thời gian hiệu quả', author: 'Nguyễn Văn A', dur: '2 giờ 10 phút', voice: 'Thùy Dung', progress: 64 },
  { title: 'Đọc hiểu báo cáo tài chính', author: 'Hoàng Văn E', dur: '3 giờ 25 phút', voice: 'Trúc Ly', progress: 0 },
  { title: 'Bán hàng qua điện thoại', author: 'Trần Thị B', dur: '1 giờ 40 phút', voice: 'Hải Đăng', progress: 100 },
  { title: 'Ngủ ngon mỗi đêm', author: 'Đỗ Thị G', dur: '52 phút', voice: 'Mai Anh', progress: 0 },
]

// 25 giọng thật của VieNeu v3 Turbo (tên · giới tính · miền · phong cách)
const allVoices: [string, string, string, string][] = [
  ['Hải Đăng', 'Nam', 'Bắc', 'tự nhiên'], ['Trúc Ly', 'Nữ', 'Bắc', 'tự nhiên'], ['Thiện Minh', 'Nam', 'Bắc', 'kể chuyện'],
  ['Mai Anh', 'Nữ', 'Bắc', 'tin tức'], ['Adam bựa', 'Nam', 'Bắc', 'tự nhiên'], ['Thiền Tâm Đức', 'Nam', 'Bắc', 'kể chuyện'],
  ['Ngọc Huyền', 'Nữ', 'Bắc', 'tự nhiên'], ['Minh Đức', 'Nam', 'Bắc', 'tin tức'], ['Phạm Tuyên', 'Nam', 'Bắc', 'tự nhiên'],
  ['Xuân Vĩnh', 'Nam', 'Bắc', 'tự nhiên'], ['Thanh Bình', 'Nam', 'Bắc', 'kể chuyện'], ['Ngọc Linh', 'Nữ', 'Bắc', 'kể chuyện'],
  ['Đoan Trang', 'Nữ', 'Bắc', 'tự nhiên'], ['Quỳnh Anh', 'Nữ', 'Bắc', 'đọc truyện'], ['Quốc Tuấn', 'Nam', 'Bắc', 'tự nhiên'],
  ['Quang Sơn', 'Nam', 'Trung', 'tự nhiên'], ['Ngọc Trân', 'Nữ', 'Trung', 'tự nhiên'],
  ['Thùy Dung', 'Nữ', 'Nam', 'tin tức'], ['Thái Sơn', 'Nam', 'Nam', 'kể chuyện'], ['Thục Đoan', 'Nữ', 'Nam', 'kể chuyện'],
  ['Minh Triết', 'Nam', 'Nam', 'tin tức'], ['Mỹ Duyên', 'Nữ', 'Nam', 'đọc truyện'], ['Đức Trí', 'Nam', 'Nam', 'đọc truyện'],
  ['Kim Thanh', 'Nữ', 'Nam', 'đọc truyện'], ['Adam', 'Nam', 'Nam', 'tự nhiên'],
]
const regions = ['Bắc', 'Trung', 'Nam'] as const
const region = ref<string>('Bắc') // mặc định: miền của giọng dùng lần trước (Hải Đăng)
const gender = ref<'all' | 'Nam' | 'Nữ'>('all')
const voice = ref('Hải Đăng')
const sampling = ref<string | null>(null)
const lastUsed = 'Hải Đăng'
const count = (r: string) => allVoices.filter((v) => r === 'all' || v[2] === r).length
const shownVoices = computed(() =>
  allVoices.filter((v) => (region.value === 'all' || v[2] === region.value) && (gender.value === 'all' || v[1] === gender.value)),
)
function sample(name: string) {
  sampling.value = sampling.value === name ? null : name
  if (sampling.value) playing.value = false // nghe mẫu → tạm dừng sách đang phát
}

const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'voice', label: 'Tạo sách mới', icon: FilePlus2 },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
  { key: 'about', label: 'Giới thiệu', icon: Info },
]
const chapters = [
  { title: '1.1 Lắng nghe chủ động', dur: '8:12', done: true },
  { title: '1.2 Ba bước lắng nghe', dur: '3:26', current: true },
  { title: '1.3 Đặt câu hỏi mở', dur: '6:40' },
  { title: '2.1 Giao tiếp bằng văn bản', dur: '9:05' },
  { title: '2.2 Viết email ngắn gọn', dur: '7:18' },
  { title: '3.1 Quản lý cảm xúc', dur: '11:02' },
]
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <!-- Nút duyệt (ngoài khung) -->
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in ([
        ['library', '1 · Thư viện: tên giọng + thanh nghe nhỏ'], ['voice', '2 · Chọn giọng theo miền'], ['player', '3 · Màn nghe: tên giọng'],
      ] as [Screen, string][])" :key="m[0]" class="h-8 px-3 rounded-full border text-xs"
        :class="screen === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="screen = m[0]">{{ m[1] }}</button>
      <label class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground cursor-pointer">
        <input v-model="miniOn" type="checkbox" class="h-3.5 w-3.5" /> Đang có sách phát
      </label>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <div class="relative w-[1100px] h-[720px] rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground flex flex-col">
      <div class="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border bg-muted/50">
        <span class="h-3 w-3 rounded-full bg-rag-red"></span>
        <span class="h-3 w-3 rounded-full bg-rag-amber"></span>
        <span class="h-3 w-3 rounded-full bg-rag-green"></span>
        <span class="flex-1 text-center text-xs text-muted-foreground">Sano</span>
      </div>

      <div class="flex-1 flex min-h-0">
        <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
          <div class="px-4 py-4 flex items-center gap-2">
            <img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" />
            <span class="font-semibold tracking-tight">Sano</span>
          </div>
          <nav class="px-2 space-y-0.5">
            <button v-for="n in nav" :key="n.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
              :class="(screen === 'voice' ? n.key === 'voice' : n.key === 'library') ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
              @click="n.key === 'voice' ? (screen = 'voice') : (screen = 'library')">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </button>
          </nav>
          <div class="px-2 mt-2">
            <span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span>
          </div>
          <div class="flex-1"></div>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.7 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <div class="flex-1 min-w-0 flex flex-col">
          <!-- ─── 1 · Thư viện: thẻ sách có tên giọng ─── -->
          <main v-if="screen === 'library'" class="flex-1 min-h-0 overflow-auto p-6">
            <div class="flex items-center justify-between">
              <div>
                <h1 class="text-xl font-semibold tracking-tight">Thư viện</h1>
                <p class="text-sm text-muted-foreground">8 cuốn · 12 giờ 09 phút · lưu ở ~/Sano/Sach</p>
              </div>
              <Button @click="screen = 'voice'"><FilePlus2 class="w-4 h-4" /> Tạo sách mới</Button>
            </div>
            <div class="mt-4 flex items-center gap-2">
              <div class="flex-1 flex items-center h-9 rounded-md border border-input bg-background px-3 text-sm text-muted-foreground">
                <Search class="w-4 h-4 mr-2" /> Tìm theo tên sách, tác giả hoặc giọng đọc…
              </div>
              <button class="h-9 px-3 rounded-md border border-input bg-background text-sm flex items-center gap-1.5">
                <span class="text-muted-foreground">Sắp xếp:</span> Mới tạo nhất <ChevronDown class="w-4 h-4 text-muted-foreground" />
              </button>
            </div>
            <div class="mt-5 grid grid-cols-4 gap-5">
              <button v-for="b in books" :key="b.title" class="text-left group" @click="screen = 'player'">
                <div class="aspect-[3/4] rounded-lg overflow-hidden shadow-md group-hover:shadow-lg">
                  <WfBookCover :title="b.title" :author="b.author" />
                </div>
                <p class="mt-2 text-sm font-medium leading-snug line-clamp-2">{{ b.title }}</p>
                <p class="text-xs text-muted-foreground">{{ b.author }} · {{ b.dur }}</p>
                <!-- MỚI: tên giọng đọc -->
                <p class="mt-0.5 text-xs text-muted-foreground flex items-center gap-1"><Mic class="w-3 h-3" /> Giọng {{ b.voice }}</p>
                <div v-if="b.progress" class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: b.progress + '%' }"></div></div>
              </button>
            </div>
          </main>

          <!-- ─── 2 · Tạo sách · bước Giọng đọc: gom theo miền ─── -->
          <main v-else-if="screen === 'voice'" class="flex-1 min-h-0 overflow-auto p-6">
            <div class="max-w-2xl">
              <h1 class="text-xl font-semibold tracking-tight">Chọn giọng đọc</h1>
              <p class="text-sm text-muted-foreground">Bấm nghe để thử từng giọng với một câu trong sách của bạn.</p>

              <!-- MỚI: chọn miền (mặc định = miền giọng dùng lần trước) + lọc Nam/Nữ -->
              <div class="mt-5 flex items-center justify-between gap-3">
                <div class="inline-flex rounded-lg border border-border p-0.5 bg-muted/40">
                  <button v-for="r in [...regions, 'all']" :key="r" class="h-8 px-3.5 rounded-md text-sm whitespace-nowrap"
                    :class="region === r ? 'bg-background shadow-sm font-medium' : 'text-muted-foreground hover:text-foreground'"
                    @click="region = r">
                    {{ r === 'all' ? 'Tất cả' : `Miền ${r}` }} <span class="text-xs text-muted-foreground tabular-nums">{{ count(r) }}</span>
                  </button>
                </div>
                <div class="flex gap-1.5">
                  <span class="text-xs text-muted-foreground self-center mr-0.5">Giọng</span>
                  <button v-for="g in (['all', 'Nam', 'Nữ'] as const)" :key="g" class="h-8 px-3 rounded-full border text-xs whitespace-nowrap"
                    :class="gender === g ? 'border-primary bg-primary/10 text-primary font-medium' : 'border-border text-muted-foreground'"
                    @click="gender = g">{{ g === 'all' ? 'Tất cả' : g }}</button>
                </div>
              </div>

              <div class="mt-3 grid gap-2">
                <label v-for="v in shownVoices" :key="v[0]" class="flex items-center gap-3 rounded-lg border px-4 py-2.5 cursor-pointer"
                  :class="voice === v[0] ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/50'">
                  <input v-model="voice" type="radio" :value="v[0]" class="h-4 w-4 accent-[hsl(var(--primary))]" />
                  <span class="flex-1">
                    <span class="font-medium text-sm">{{ v[0] }}</span>
                    <span v-if="v[0] === lastUsed" class="ml-2 text-[11px] rounded-full bg-muted px-2 py-0.5 text-muted-foreground">Dùng lần trước</span>
                    <span class="block text-xs text-muted-foreground">{{ v[1] }} · {{ v[3] }}{{ region === 'all' ? ` · miền ${v[2]}` : '' }}</span>
                  </span>
                  <button class="h-8 px-3 rounded-full border border-border text-xs flex items-center gap-1.5 hover:bg-muted" @click.prevent="sample(v[0])">
                    <component :is="sampling === v[0] ? Pause : Play" class="w-3.5 h-3.5" /> Nghe mẫu
                  </button>
                </label>
              </div>
              <p v-if="miniOn && sampling" class="mt-2 text-xs text-muted-foreground">Đang nghe mẫu nên sách tạm dừng; bấm ▶ ở thanh dưới để nghe tiếp.</p>
              <div class="mt-5 text-sm">
                <p class="font-medium">Câu nghe mẫu</p>
                <input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" value="Bước thứ nhất, dừng việc đang làm và nhìn người nói." />
              </div>
              <div class="mt-6 flex justify-between"><Button variant="outline">Quay lại</Button><Button>Tiếp tục</Button></div>
            </div>
          </main>

          <!-- ─── 3 · Màn nghe: có tên giọng ─── -->
          <main v-else class="flex-1 min-h-0 flex">
            <div class="flex-1 flex flex-col p-6 min-w-0">
              <button class="text-sm text-muted-foreground flex items-center gap-1 hover:text-foreground w-fit" @click="screen = 'library'"><ChevronLeft class="w-4 h-4" /> Thư viện</button>
              <div class="flex-1 flex flex-col items-center justify-center">
                <div class="aspect-[3/4] w-44 rounded-xl shadow-2xl overflow-hidden">
                  <WfBookCover title="Kỹ năng mềm cho người trẻ" author="Nguyễn Văn A" size="lg" />
                </div>
                <h2 class="mt-5 text-lg font-semibold">Kỹ năng mềm cho người trẻ</h2>
                <p class="text-sm text-muted-foreground">Nguyễn Văn A · <span class="inline-flex items-center gap-1"><Mic class="w-3.5 h-3.5" /> Giọng Hải Đăng</span></p>
                <p class="text-sm text-muted-foreground">1.2 Ba bước lắng nghe</p>
                <div class="mt-5 w-full max-w-md">
                  <div class="h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full w-[30%] bg-primary rounded-full"></div></div>
                  <div class="flex justify-between text-[11px] text-muted-foreground mt-1 tabular-nums"><span>01:02</span><span>-02:24</span></div>
                </div>
                <div class="mt-4 flex items-center gap-4">
                  <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><SkipBack class="w-5 h-5" /></button>
                  <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><RotateCcw class="w-5 h-5" /></button>
                  <button class="h-14 w-14 grid place-items-center rounded-full bg-primary text-primary-foreground shadow" @click="playing = !playing">
                    <component :is="playing ? Pause : Play" class="w-6 h-6" />
                  </button>
                  <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><RotateCw class="w-5 h-5" /></button>
                  <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><SkipForward class="w-5 h-5" /></button>
                </div>
                <div class="mt-4 flex items-center gap-2 text-xs">
                  <button class="h-8 px-3 rounded-full border border-border flex items-center gap-1.5"><Gauge class="w-3.5 h-3.5" />1,25×</button>
                  <span class="text-muted-foreground">Rời màn này vẫn nghe tiếp ở thanh dưới</span>
                </div>
              </div>
            </div>
            <div class="w-64 shrink-0 border-l border-border overflow-auto">
              <div class="px-4 py-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Mục lục · 6 mục</div>
              <button v-for="c in chapters" :key="c.title" class="w-full flex items-center justify-between gap-2 px-4 py-2.5 text-sm text-left hover:bg-muted/60"
                :class="c.current ? 'bg-primary/10 text-primary font-medium' : c.done ? 'text-muted-foreground' : ''">
                <span class="truncate flex items-center gap-2">
                  <Volume2 v-if="c.current" class="w-3.5 h-3.5 shrink-0" /><Check v-else-if="c.done" class="w-3.5 h-3.5 shrink-0" /><span v-else class="w-3.5 shrink-0"></span>
                  {{ c.title }}
                </span>
                <span class="text-xs tabular-nums shrink-0">{{ c.dur }}</span>
              </button>
            </div>
          </main>

          <!-- ─── MỚI: thanh nghe nhỏ (mọi màn trừ màn nghe) ─── -->
          <div v-if="miniOn && screen !== 'player'" class="shrink-0 border-t border-border bg-background">
            <div class="h-0.5 bg-muted"><div class="h-full w-[30%] bg-primary"></div></div>
            <div class="h-16 flex items-center gap-3 px-4">
              <button class="flex items-center gap-3 min-w-0 flex-1 text-left" title="Mở màn nghe" @click="screen = 'player'">
                <div class="h-11 w-8 shrink-0 rounded overflow-hidden shadow"><WfBookCover title="Kỹ năng mềm cho người trẻ" size="sm" /></div>
                <span class="min-w-0">
                  <span class="block text-sm font-medium truncate">Kỹ năng mềm cho người trẻ</span>
                  <span class="block text-xs text-muted-foreground truncate">1.2 Ba bước lắng nghe · Giọng Hải Đăng</span>
                </span>
              </button>
              <span class="text-xs text-muted-foreground tabular-nums">01:02 / 03:26</span>
              <button class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" title="Lùi 15 giây"><RotateCcw class="w-4 h-4" /></button>
              <button class="h-10 w-10 grid place-items-center rounded-full bg-primary text-primary-foreground" @click="playing = !playing">
                <component :is="playing ? Pause : Play" class="w-5 h-5" />
              </button>
              <button class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" title="Tới 30 giây"><RotateCw class="w-4 h-4" /></button>
              <button class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted text-muted-foreground" title="Dừng nghe" @click="miniOn = false"><X class="w-4 h-4" /></button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
