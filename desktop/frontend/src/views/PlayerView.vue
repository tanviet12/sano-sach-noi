<script setup lang="ts">
// Trình phát: phát MP3 thật của cuốn trong thư viện, mục lục tiểu mục, tua
// −15s/+30s, đổi tốc độ, nhớ vị trí nghe; xuất M4B (tiến độ ngay dưới hàng nút),
// xuất gói zip, xoá vào Thùng rác.
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import {
  Check, ChevronDown, ChevronLeft, Download, FolderOpen, Gauge, Loader2, Package, Pause, Play, RotateCcw, RotateCw,
  SkipBack, SkipForward, Trash2, Volume2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import BookCover from '@/components/sano/BookCover.vue'
import M4BProgress from '../components/M4BProgress.vue'
import { book, deleteBook, errText, openBookFolder, revealBookZip, type BookDetail } from '../lib/backend'
import { clearAudioSource, setAudioSource } from '../lib/audio'
import { fmtClock, fmtLong, loadPosition, savePosition } from '../lib/position'
import { go, state } from '../lib/store'
import { m4bBusy, startM4B, useM4B } from '../lib/m4b'

useM4B()

const detail = ref<BookDetail | null>(null)
const error = ref('')
const current = ref(0)
const playing = ref(false)
const time = ref(0)
const duration = ref(0)
const speeds = [0.75, 1, 1.25, 1.5, 1.75, 2]
const SPEED_KEY = 'sano.speed'
// Nhớ tốc độ đã chọn giữa các lần nghe (tiện ích riêng của máy, mất cũng không sao).
function loadSpeed() {
  try {
    const v = Number(localStorage.getItem(SPEED_KEY))
    return speeds.includes(v) ? v : 1
  } catch {
    return 1
  }
}
const speed = ref(loadSpeed())
const speedOpen = ref(false)

const audio = new Audio()
audio.preload = 'auto'
let pendingSeek = 0

const tracks = computed(() => detail.value?.tracks ?? [])
const track = computed(() => tracks.value[current.value])
const totalSec = computed(() => tracks.value.reduce((n, t) => n + t.durationSec, 0))

watch(
  () => state.playerSlug,
  async (slug) => {
    audio.pause()
    playing.value = false
    detail.value = null
    error.value = ''
    if (!slug) return
    try {
      detail.value = await book(slug)
      const pos = loadPosition(slug)
      load(Math.min(pos?.track ?? 0, Math.max(0, tracks.value.length - 1)), pos?.time ?? 0)
      if (state.playerAutoplay) void play() // bấm phát ở hàng Nghe tiếp
      state.playerAutoplay = false
    } catch (e) {
      error.value = errText(e)
    }
  },
  { immediate: true },
)

let loadSeq = 0
let loading: Promise<void> = Promise.resolve()

function load(i: number, at = 0) {
  const t = tracks.value[i]
  if (!t) return
  current.value = i
  time.value = at
  duration.value = t.durationSec
  pendingSeek = at
  const seq = ++loadSeq
  // Linux: nạp qua blob (bất đồng bộ, xem lib/audio.ts); play() đợi xong mới phát.
  loading = setAudioSource(audio, t.url)
    .then(() => {
      if (seq !== loadSeq) return // đã chọn tiểu mục khác
      audio.defaultPlaybackRate = speed.value // nạp file mới trình duyệt đặt lại tốc độ theo giá trị này
      audio.playbackRate = speed.value
    })
    .catch((e) => {
      if (seq === loadSeq) error.value = errText(e)
    })
}

async function play() {
  error.value = ''
  try {
    const seq = loadSeq
    await loading
    if (seq !== loadSeq || error.value) return
    await audio.play()
  } catch (e) {
    error.value = 'Không phát được: ' + errText(e)
  }
}

function toggle() {
  if (!track.value) return
  if (playing.value) audio.pause()
  else void play()
}

function pick(i: number) {
  load(i)
  void play()
}

function skip(delta: number) {
  const i = current.value + delta
  if (i < 0 || i >= tracks.value.length) return
  load(i)
  void play()
}

function seek(delta: number) {
  audio.currentTime = Math.min(Math.max(0, audio.currentTime + delta), audio.duration || duration.value)
}

function seekTo(e: MouseEvent) {
  const el = e.currentTarget as HTMLElement
  const r = el.getBoundingClientRect()
  const d = audio.duration || duration.value
  if (d) audio.currentTime = ((e.clientX - r.left) / r.width) * d
}

function setSpeed(v: number) {
  speed.value = v
  audio.defaultPlaybackRate = v
  audio.playbackRate = v
  speedOpen.value = false
  try {
    localStorage.setItem(SPEED_KEY, String(v))
  } catch {
    // không lưu được thì thôi
  }
}

// Đóng menu tốc độ khi bấm ra ngoài hoặc nhấn Esc.
function closeSpeed(e: Event) {
  if (e instanceof KeyboardEvent ? e.key === 'Escape' : !(e.target as HTMLElement).closest('[data-speed-menu]')) speedOpen.value = false
}
document.addEventListener('mousedown', closeSpeed)
document.addEventListener('keydown', closeSpeed)

let played = false // chỉ mở rồi thoát, chưa phát → không tính là "nghe gần đây"

function remember() {
  const slug = state.playerSlug
  if (!slug || !tracks.value.length || !played) return
  const before = tracks.value.slice(0, current.value).reduce((n, t) => n + t.durationSec, 0)
  const pct = totalSec.value ? Math.min(100, ((before + time.value) / totalSec.value) * 100) : 0
  savePosition(slug, { track: current.value, time: time.value, pct })
}

audio.addEventListener('loadedmetadata', () => {
  if (audio.duration && isFinite(audio.duration)) duration.value = audio.duration
  if (pendingSeek) audio.currentTime = pendingSeek
  pendingSeek = 0
})
audio.addEventListener('play', () => {
  playing.value = true
  played = true
})
audio.addEventListener('pause', () => {
  playing.value = false
  remember()
})
let lastSave = 0
audio.addEventListener('timeupdate', () => {
  time.value = audio.currentTime
  if (Date.now() - lastSave > 5000) {
    lastSave = Date.now()
    remember()
  }
})
audio.addEventListener('ended', () => {
  if (current.value < tracks.value.length - 1) {
    load(current.value + 1)
    void play()
  } else {
    time.value = duration.value
    remember()
  }
})
audio.addEventListener('error', () => {
  if (audio.src) error.value = 'Không đọc được file âm thanh của tiểu mục này'
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', closeSpeed)
  document.removeEventListener('keydown', closeSpeed)
  remember()
  audio.pause()
  clearAudioSource(audio)
})

const pctTrack = computed(() => (duration.value ? Math.min(100, (time.value / duration.value) * 100) : 0))
const fmtSpeed = (s: number) => s.toLocaleString('vi-VN') + '×'

async function removeBook() {
  error.value = ''
  try {
    audio.pause()
    const where = await deleteBook(state.playerSlug)
    if (!where) return // người dùng huỷ
    state.playerSlug = ''
    go('library')
  } catch (e) {
    error.value = errText(e)
  }
}

async function act(fn: (slug: string) => Promise<void>) {
  try {
    await fn(state.playerSlug)
  } catch (e) {
    error.value = errText(e)
  }
}
</script>

<template>
  <section class="flex-1 flex min-h-0">
    <div class="flex-1 flex flex-col p-6 min-w-0">
      <button class="text-sm text-muted-foreground flex items-center gap-1 hover:text-foreground w-fit" @click="go('library')"><ChevronLeft class="w-4 h-4" /> Thư viện</button>
      <div v-if="!detail" class="flex-1 grid place-items-center text-sm text-muted-foreground">
        <span v-if="error" class="text-destructive">{{ error }}</span>
        <span v-else-if="!state.playerSlug">Chọn một cuốn trong thư viện để nghe.</span>
        <span v-else class="flex items-center gap-2"><Loader2 class="w-4 h-4 animate-spin" /> Đang mở sách…</span>
      </div>
      <div v-else class="flex-1 flex flex-col items-center justify-center">
        <div class="aspect-[3/4] w-48 rounded-xl shadow-2xl overflow-hidden">
          <img v-if="detail.coverUrl" :src="detail.coverUrl" :alt="detail.title" class="h-full w-full object-cover" />
          <BookCover v-else :title="detail.title" :author="detail.author" class="h-full w-full rounded-xl shadow-none" />
        </div>
        <h2 class="mt-5 text-lg font-semibold">{{ detail.title }}</h2>
        <p class="text-sm text-muted-foreground">{{ track?.title }}</p>
        <div class="mt-5 w-full max-w-md">
          <div class="h-1.5 rounded-full bg-muted overflow-hidden cursor-pointer" role="slider" aria-label="Vị trí nghe" :aria-valuenow="Math.round(pctTrack)" @click="seekTo">
            <div class="h-full bg-primary rounded-full" :style="{ width: pctTrack + '%' }"></div>
          </div>
          <div class="flex justify-between text-[11px] text-muted-foreground mt-1 tabular-nums"><span>{{ fmtClock(time) }}</span><span>-{{ fmtClock(Math.max(0, duration - time)) }}</span></div>
        </div>
        <div class="mt-4 flex items-center gap-4">
          <button aria-label="Tiểu mục trước" class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted" @click="skip(-1)"><SkipBack class="w-5 h-5" /></button>
          <button aria-label="Lùi 15 giây" class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted" @click="seek(-15)"><RotateCcw class="w-5 h-5" /></button>
          <button :aria-label="playing ? 'Dừng' : 'Phát'" class="h-14 w-14 grid place-items-center rounded-full bg-primary text-primary-foreground shadow" @click="toggle">
            <Pause v-if="playing" class="w-6 h-6" /><Play v-else class="w-6 h-6 ml-0.5" />
          </button>
          <button aria-label="Tới 30 giây" class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted" @click="seek(30)"><RotateCw class="w-5 h-5" /></button>
          <button aria-label="Tiểu mục sau" class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted" @click="skip(1)"><SkipForward class="w-5 h-5" /></button>
        </div>
        <div class="mt-4 flex items-center gap-2 text-xs">
          <div class="relative" data-speed-menu>
            <button class="h-8 px-3 rounded-full border border-border flex items-center gap-1.5 hover:bg-muted" aria-haspopup="listbox" :aria-expanded="speedOpen" @click="speedOpen = !speedOpen">
              <Gauge class="w-3.5 h-3.5" />{{ fmtSpeed(speed) }}<ChevronDown class="w-3 h-3 text-muted-foreground" />
            </button>
            <div v-if="speedOpen" role="listbox" aria-label="Tốc độ phát" class="absolute bottom-full left-0 mb-2 w-40 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-20">
              <div class="px-3 py-1.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Tốc độ phát</div>
              <button v-for="v in speeds" :key="v" role="option" :aria-selected="v === speed"
                class="w-full flex items-center justify-between px-3 py-1.5 text-sm text-left hover:bg-muted"
                :class="v === speed && 'text-primary font-medium'" @click="setSpeed(v)">
                {{ fmtSpeed(v) }}<Check v-if="v === speed" class="w-4 h-4" />
              </button>
            </div>
          </div>
          <span class="text-muted-foreground">Tua −15s / +30s · nhớ vị trí nghe</span>
        </div>
        <p v-if="error" class="mt-3 text-sm text-destructive">{{ error }}</p>
      </div>
      <div class="flex flex-wrap items-center gap-2 border-t border-border pt-4">
        <Button variant="outline" size="sm" :disabled="!detail || m4bBusy()" title="Một file có mục lục chương và bìa — chép sang điện thoại nghe bằng app sách nói bất kỳ" @click="startM4B(state.playerSlug)"><Download class="w-4 h-4" /> Xuất M4B</Button>
        <Button variant="outline" size="sm" :disabled="!detail?.zip" title="Sao lưu hoặc chuyển sách sang máy khác" @click="act(revealBookZip)"><Package class="w-4 h-4" /> Xuất gói zip</Button>
        <Button variant="ghost" size="sm" :disabled="!detail" @click="act(openBookFolder)"><FolderOpen class="w-4 h-4" /> Mở thư mục</Button>
        <Button variant="ghost" size="sm" class="ml-auto text-destructive hover:text-destructive" :disabled="!detail" title="Chuyển cuốn sách vào Thùng rác (lấy lại được)" @click="removeBook"><Trash2 class="w-4 h-4" /> Xoá</Button>
        <M4BProgress v-if="state.playerSlug" :slug="state.playerSlug" />
      </div>
    </div>
    <div class="w-72 shrink-0 border-l border-border overflow-auto">
      <div class="px-4 py-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Mục lục · {{ tracks.length }} mục · {{ fmtLong(totalSec) }}</div>
      <button v-for="(c, i) in tracks" :key="c.file"
        class="w-full flex items-center justify-between gap-2 px-4 py-2.5 text-sm text-left hover:bg-muted/60"
        :class="i === current ? 'bg-primary/10 text-primary font-medium' : i < current ? 'text-muted-foreground' : ''"
        :title="c.chapter !== c.title ? c.chapter : ''"
        @click="pick(i)">
        <span class="truncate flex items-center gap-2">
          <Volume2 v-if="i === current" class="w-3.5 h-3.5 shrink-0" />
          <Check v-else-if="i < current" class="w-3.5 h-3.5 shrink-0" />
          <span v-else class="w-3.5 shrink-0"></span>
          {{ c.title }}
        </span>
        <span class="text-xs tabular-nums shrink-0">{{ fmtClock(c.durationSec) }}</span>
      </button>
    </div>
  </section>
</template>
