<script setup lang="ts">
// D5 — Chữ chạy theo lời đọc (cách A: ước lượng theo độ dài câu). Bám khung D1 (1100×720).
// - Màn nghe: dưới tên tiểu mục có ô "lời đọc" 2 dòng (câu đang đọc + câu kế). Bấm → xem lời.
// - Xem lời: phủ vùng nội dung (giữ thanh bên), chữ lớn, câu đang đọc đậm, câu đã đọc mờ,
//   tự cuộn giữ câu đang đọc ở khoảng 1/3 trên. Bấm một câu → nghe từ câu đó.
//   Người dùng tự cuộn → tạm ngừng tự cuộn, hiện nút "Về câu đang đọc".
// - Chữ lấy từ original_text của tiểu mục (đúng như trong file Word). Thời điểm từng câu
//   = thời lượng tiểu mục chia theo số ký tự → ghi rõ "vị trí chữ là ước lượng".
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Câu tự chạy ~2,5s/câu để xem hiệu ứng.
import { computed, onBeforeUnmount, ref, watch, nextTick } from 'vue'
import {
  Library, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, ChevronLeft, ChevronDown, Play, Pause, RotateCcw, RotateCw,
  SkipBack, SkipForward, Mic, Check, Volume2, Gauge, Sun, Moon, Minimize2, Crosshair, Type,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

type Mode = 'player' | 'lyrics' | 'lyrics-scrolled'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'player')
const dark = ref(false)
const playing = ref(true)
const big = ref(false) // cỡ chữ lớn hơn

const title = 'Kỹ năng mềm cho người trẻ'
const section = '1.2 Ba bước lắng nghe'
const paragraphs = [
  [
    'Lắng nghe không chỉ là im lặng khi người khác nói.',
    'Đó là một kỹ năng, và như mọi kỹ năng, nó cần được luyện tập mỗi ngày.',
  ],
  [
    'Bước thứ nhất, dừng việc đang làm và nhìn người nói.',
    'Đặt điện thoại xuống, quay người về phía họ, để họ biết bạn đang thật sự có mặt.',
    'Chỉ một hành động nhỏ, nhưng người đối diện sẽ cảm nhận được ngay.',
  ],
  [
    'Bước thứ hai, nghe để hiểu, không phải nghe để trả lời.',
    'Khi đang nghe mà đã nghĩ sẵn câu đáp, bạn sẽ bỏ lỡ điều quan trọng nhất họ muốn nói.',
  ],
  [
    'Bước thứ ba, nhắc lại bằng lời của mình.',
    'Hãy nói, nếu mình hiểu đúng thì ý bạn là như vậy, đúng không?',
    'Câu hỏi đơn giản này giúp hai bên chắc chắn đã hiểu nhau, trước khi đi tiếp.',
  ],
]
const sentences = paragraphs.flatMap((p, pi) => p.map((t) => ({ t, pi })))
const cur = ref(3)

let timer: number | undefined
watch(
  [playing, mode],
  () => {
    clearInterval(timer)
    if (playing.value) timer = window.setInterval(() => (cur.value = (cur.value + 1) % sentences.length), 2500)
  },
  { immediate: true },
)
onBeforeUnmount(() => clearInterval(timer))

const lyricsEl = ref<HTMLElement | null>(null)
watch([cur, mode], async () => {
  if (mode.value !== 'lyrics') return
  await nextTick()
  const el = lyricsEl.value?.querySelector<HTMLElement>(`[data-i="${cur.value}"]`)
  const box = lyricsEl.value
  if (el && box) box.scrollTo({ top: el.offsetTop - box.clientHeight / 3, behavior: 'smooth' })
})

const pct = computed(() => Math.round(((cur.value + 0.5) / sentences.length) * 100))
const clock = computed(() => {
  const s = Math.round((206 * pct.value) / 100)
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`
})

const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'create', label: 'Tạo sách mới', icon: FilePlus2 },
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
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in ([
        ['player', '1 · Màn nghe: ô lời đọc'], ['lyrics', '2 · Xem lời (chữ chạy theo)'], ['lyrics-scrolled', '3 · Xem lời: đã tự cuộn đi chỗ khác'],
      ] as [Mode, string][])" :key="m[0]" class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="mode = m[0]">{{ m[1] }}</button>
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
            <span v-for="n in nav" :key="n.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
              :class="n.key === 'library' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </span>
          </nav>
          <div class="px-2 mt-2">
            <span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span>
          </div>
          <div class="flex-1"></div>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.7 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <!-- ─── 1 · Màn nghe (như hiện tại) + ô lời đọc ─── -->
        <main v-if="mode === 'player'" class="flex-1 min-w-0 flex">
          <div class="flex-1 flex flex-col p-6 min-w-0">
            <span class="text-sm text-muted-foreground flex items-center gap-1"><ChevronLeft class="w-4 h-4" /> Thư viện</span>
            <div class="flex-1 flex flex-col items-center justify-center">
              <div class="aspect-[3/4] w-40 rounded-xl shadow-2xl overflow-hidden"><WfBookCover :title="title" author="Nguyễn Văn A" size="lg" /></div>
              <h2 class="mt-4 text-lg font-semibold">{{ title }}</h2>
              <p class="text-sm text-muted-foreground flex items-center gap-1">Nguyễn Văn A · <Mic class="w-3.5 h-3.5" /> Giọng Hải Đăng</p>
              <p class="text-sm text-muted-foreground">{{ section }}</p>

              <!-- MỚI: ô lời đọc — bấm mở Xem lời -->
              <button class="group mt-4 w-full max-w-md rounded-lg border border-border bg-muted/30 hover:bg-muted/60 px-4 py-3 text-left" @click="mode = 'lyrics'">
                <span class="flex items-center justify-between text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                  Lời đọc <span class="normal-case font-normal tracking-normal text-primary opacity-0 group-hover:opacity-100">Xem cả lời →</span>
                </span>
                <span class="mt-1 block text-sm font-medium leading-snug line-clamp-2">{{ sentences[cur].t }}</span>
                <span class="mt-0.5 block text-sm text-muted-foreground leading-snug line-clamp-1">{{ sentences[cur + 1]?.t }}</span>
              </button>

              <div class="mt-4 w-full max-w-md">
                <div class="h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary rounded-full" :style="{ width: pct + '%' }"></div></div>
                <div class="flex justify-between text-[11px] text-muted-foreground mt-1 tabular-nums"><span>{{ clock }}</span><span>3:26</span></div>
              </div>
              <div class="mt-3 flex items-center gap-4">
                <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><SkipBack class="w-5 h-5" /></button>
                <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><RotateCcw class="w-5 h-5" /></button>
                <button class="h-14 w-14 grid place-items-center rounded-full bg-primary text-primary-foreground shadow" @click="playing = !playing">
                  <component :is="playing ? Pause : Play" class="w-6 h-6" />
                </button>
                <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><RotateCw class="w-5 h-5" /></button>
                <button class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted"><SkipForward class="w-5 h-5" /></button>
              </div>
              <div class="mt-3 flex items-center gap-2 text-xs">
                <span class="h-8 px-3 rounded-full border border-border flex items-center gap-1.5"><Gauge class="w-3.5 h-3.5" />1×<ChevronDown class="w-3 h-3 text-muted-foreground" /></span>
                <span class="text-muted-foreground">Tua −15s / +30s · nhớ vị trí nghe</span>
              </div>
            </div>
          </div>
          <div class="w-64 shrink-0 border-l border-border overflow-auto">
            <div class="px-4 py-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Mục lục · 6 mục</div>
            <span v-for="c in chapters" :key="c.title" class="w-full flex items-center justify-between gap-2 px-4 py-2.5 text-sm"
              :class="c.current ? 'bg-primary/10 text-primary font-medium' : c.done ? 'text-muted-foreground' : ''">
              <span class="truncate flex items-center gap-2">
                <Volume2 v-if="c.current" class="w-3.5 h-3.5 shrink-0" /><Check v-else-if="c.done" class="w-3.5 h-3.5 shrink-0" /><span v-else class="w-3.5 shrink-0"></span>{{ c.title }}
              </span>
              <span class="text-xs tabular-nums shrink-0">{{ c.dur }}</span>
            </span>
          </div>
        </main>

        <!-- ─── 2/3 · Xem lời: phủ vùng nội dung ─── -->
        <main v-else class="flex-1 min-w-0 flex flex-col bg-gradient-to-b from-primary/5 to-background">
          <!-- đầu: sách + tiểu mục + thu nhỏ -->
          <div class="shrink-0 h-16 flex items-center gap-3 px-6 border-b border-border/60">
            <div class="h-11 w-8 shrink-0 rounded overflow-hidden shadow"><WfBookCover :title="title" size="sm" /></div>
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium truncate">{{ section }}</p>
              <p class="text-xs text-muted-foreground truncate">{{ title }} · Giọng Hải Đăng</p>
            </div>
            <button class="h-8 px-2.5 rounded-md text-xs text-muted-foreground hover:bg-muted flex items-center gap-1.5" title="Cỡ chữ" @click="big = !big"><Type class="w-4 h-4" /> {{ big ? 'Chữ vừa' : 'Chữ lớn' }}</button>
            <button class="h-8 px-2.5 rounded-md text-xs text-muted-foreground hover:bg-muted flex items-center gap-1.5" @click="mode = 'player'"><Minimize2 class="w-4 h-4" /> Thu nhỏ</button>
          </div>

          <!-- lời: câu đang đọc đậm, đã đọc mờ -->
          <div ref="lyricsEl" class="relative flex-1 overflow-auto px-14 py-10">
            <div class="max-w-2xl mx-auto space-y-6" :class="big ? 'text-[26px] leading-[1.5]' : 'text-[21px] leading-[1.55]'">
              <p v-for="(p, pi) in paragraphs" :key="pi">
                <template v-for="(s, si) in sentences" :key="si">
                  <span v-if="s.pi === pi" :data-i="si" class="cursor-pointer rounded transition-colors duration-300 hover:bg-muted/70"
                    :class="si === cur ? 'font-semibold text-foreground' : si < cur ? 'text-muted-foreground/70' : 'text-muted-foreground'"
                    title="Nghe từ câu này" @click="cur = si">{{ s.t + ' ' }}</span>
                </template>
              </p>
              <p class="pt-6 text-sm text-muted-foreground">Hết tiểu mục · tiếp theo: 1.3 Đặt câu hỏi mở</p>
            </div>
            <!-- 3 · người dùng tự cuộn → ngừng tự cuộn, hiện nút quay về -->
            <button v-if="mode === 'lyrics-scrolled'" class="sticky bottom-4 left-full mr-2 h-9 px-3.5 rounded-full bg-primary text-primary-foreground text-sm shadow-lg flex items-center gap-1.5 ml-auto" @click="mode = 'lyrics'">
              <Crosshair class="w-4 h-4" /> Về câu đang đọc
            </button>
          </div>

          <!-- điều khiển gọn ở đáy -->
          <div class="shrink-0 border-t border-border bg-background/80">
            <div class="h-0.5 bg-muted"><div class="h-full bg-primary" :style="{ width: pct + '%' }"></div></div>
            <div class="h-16 flex items-center gap-3 px-6">
              <button class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted"><SkipBack class="w-4 h-4" /></button>
              <button class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted"><RotateCcw class="w-4 h-4" /></button>
              <button class="h-11 w-11 grid place-items-center rounded-full bg-primary text-primary-foreground" @click="playing = !playing">
                <component :is="playing ? Pause : Play" class="w-5 h-5" />
              </button>
              <button class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted"><RotateCw class="w-4 h-4" /></button>
              <button class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted"><SkipForward class="w-4 h-4" /></button>
              <span class="text-xs text-muted-foreground tabular-nums">{{ clock }} / 3:26</span>
              <span class="flex-1"></span>
              <span class="text-[11px] text-muted-foreground">Vị trí chữ là ước lượng · bấm một câu để nghe từ câu đó</span>
              <Button variant="outline" size="sm" class="h-8"><Gauge class="w-3.5 h-3.5" /> 1×</Button>
            </div>
          </div>
        </main>
      </div>
    </div>
  </div>
</template>
