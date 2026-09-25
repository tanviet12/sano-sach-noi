<script setup lang="ts">
// Xem lời (wireframe D5): chữ lớn, câu đang đọc đậm, câu đã đọc mờ, tự cuộn giữ câu
// đang đọc ở ~1/3 trên. Người dùng tự cuộn → ngừng tự cuộn, hiện "Về câu đang đọc".
// Bấm một câu → nghe từ câu đó. Vị trí chữ là ước lượng (lib/lyrics.ts).
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { Crosshair, Gauge, Minimize2, Pause, Play, RotateCcw, RotateCw, SkipBack, SkipForward, Type } from 'lucide-vue-next'
import BookCover from '@/components/sano/BookCover.vue'
import { fmtClock } from '../lib/position'
import {
  SPEEDS, lyricIndex, lyrics, pctTrack, player, seek, seekTime, setSpeed, skip, toggle, track, tracks,
} from '../lib/player'

defineEmits<{ close: [] }>()

const BIG_KEY = 'sano.lyricsBig'
function loadBig() {
  try {
    return localStorage.getItem(BIG_KEY) === '1'
  } catch {
    return false
  }
}
const big = ref(loadBig())
function toggleBig() {
  big.value = !big.value
  try {
    localStorage.setItem(BIG_KEY, big.value ? '1' : '0')
  } catch {
    // không lưu được thì thôi
  }
}

const box = ref<HTMLElement | null>(null)
const follow = ref(true)
let auto = false // đang cuộn do app (không tính là người dùng tự cuộn)
let autoTimer = 0

function scrollToCurrent(smooth = true) {
  void nextTick(() => {
    const el = box.value?.querySelector<HTMLElement>(`[data-i="${lyricIndex.value}"]`)
    if (!el || !box.value) return
    auto = true
    box.value.scrollTo({ top: el.offsetTop - box.value.clientHeight / 3, behavior: smooth ? 'smooth' : 'auto' })
    window.clearTimeout(autoTimer)
    autoTimer = window.setTimeout(() => (auto = false), smooth ? 900 : 100)
  })
}
function userScrolled() {
  if (!auto) follow.value = false
}
function backToCurrent() {
  follow.value = true
  scrollToCurrent()
}

onMounted(() => scrollToCurrent(false))
watch(lyricIndex, () => follow.value && scrollToCurrent())
watch(
  () => player.current,
  () => {
    follow.value = true // sang tiểu mục mới: bám theo lại từ đầu
    scrollToCurrent(false)
  },
)

function pickSentence(start: number) {
  follow.value = true
  seekTime(start + 0.01)
}

const next = computed(() => tracks.value[player.current + 1])
const fmtSpeed = (s: number) => s.toLocaleString('vi-VN') + '×'
function cycleSpeed() {
  const i = SPEEDS.indexOf(player.speed)
  setSpeed(SPEEDS[(i + 1) % SPEEDS.length])
}
</script>

<template>
  <section class="flex-1 min-w-0 min-h-0 flex flex-col bg-gradient-to-b from-primary/5 to-background">
    <div class="shrink-0 h-16 flex items-center gap-3 px-6 border-b border-border/60">
      <span class="h-11 w-8 shrink-0 rounded overflow-hidden shadow">
        <img v-if="player.detail?.coverUrl" :src="player.detail.coverUrl" alt="" class="h-full w-full object-cover" />
        <BookCover v-else :title="player.detail?.title ?? ''" class="h-full w-full rounded shadow-none" />
      </span>
      <div class="min-w-0 flex-1">
        <p class="text-sm font-medium truncate">{{ track?.title }}</p>
        <p class="text-xs text-muted-foreground truncate">
          {{ player.detail?.title }}<template v-if="player.detail?.voice"> · Giọng {{ player.detail.voice }}</template>
        </p>
      </div>
      <button class="h-8 px-2.5 rounded-md text-xs text-muted-foreground hover:bg-muted flex items-center gap-1.5" @click="toggleBig">
        <Type class="w-4 h-4" /> {{ big ? 'Chữ vừa' : 'Chữ lớn' }}
      </button>
      <button class="h-8 px-2.5 rounded-md text-xs text-muted-foreground hover:bg-muted flex items-center gap-1.5" @click="$emit('close')">
        <Minimize2 class="w-4 h-4" /> Thu nhỏ
      </button>
    </div>

    <div ref="box" class="relative flex-1 overflow-auto px-14 py-10" @scroll.passive="userScrolled">
      <div v-if="lyrics" class="max-w-2xl mx-auto space-y-6" :class="big ? 'text-[26px] leading-[1.5]' : 'text-[21px] leading-[1.55]'">
        <p v-for="(p, pi) in lyrics.paragraphs" :key="pi">
          <span v-for="s in p" :key="s.index" :data-i="s.index" role="button" tabindex="0" title="Nghe từ câu này"
            class="cursor-pointer rounded transition-colors duration-300 hover:bg-muted/70"
            :class="s.index === lyricIndex ? 'font-semibold text-foreground' : s.index < lyricIndex ? 'text-muted-foreground/70' : 'text-muted-foreground'"
            @click="pickSentence(s.start)" @keydown.enter="pickSentence(s.start)">{{ s.text + ' ' }}</span>
        </p>
        <p class="pt-6 text-sm text-muted-foreground">Hết tiểu mục<template v-if="next"> · tiếp theo: {{ next.title }}</template></p>
      </div>
      <p v-else class="text-sm text-muted-foreground text-center mt-20">Tiểu mục này không có chữ để hiện.</p>
      <button v-if="!follow && lyrics" class="sticky bottom-4 ml-auto h-9 px-3.5 rounded-full bg-primary text-primary-foreground text-sm shadow-lg flex items-center gap-1.5" @click="backToCurrent">
        <Crosshair class="w-4 h-4" /> Về câu đang đọc
      </button>
    </div>

    <div class="shrink-0 border-t border-border bg-background/80">
      <div class="h-0.5 bg-muted"><div class="h-full bg-primary" :style="{ width: pctTrack + '%' }"></div></div>
      <div class="h-16 flex items-center gap-3 px-6">
        <button aria-label="Tiểu mục trước" class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" @click="skip(-1)"><SkipBack class="w-4 h-4" /></button>
        <button aria-label="Lùi 15 giây" class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" @click="seek(-15)"><RotateCcw class="w-4 h-4" /></button>
        <button :aria-label="player.playing ? 'Dừng' : 'Phát'" class="h-11 w-11 grid place-items-center rounded-full bg-primary text-primary-foreground" @click="toggle">
          <Pause v-if="player.playing" class="w-5 h-5" /><Play v-else class="w-5 h-5 ml-0.5" />
        </button>
        <button aria-label="Tới 30 giây" class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" @click="seek(30)"><RotateCw class="w-4 h-4" /></button>
        <button aria-label="Tiểu mục sau" class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" @click="skip(1)"><SkipForward class="w-4 h-4" /></button>
        <span class="text-xs text-muted-foreground tabular-nums">{{ fmtClock(player.time) }} / {{ fmtClock(player.duration) }}</span>
        <span class="flex-1"></span>
        <span class="text-[11px] text-muted-foreground">Vị trí chữ là ước lượng · bấm một câu để nghe từ câu đó</span>
        <button class="h-8 px-3 rounded-md border border-border text-xs flex items-center gap-1.5 hover:bg-muted" title="Đổi tốc độ phát" @click="cycleSpeed">
          <Gauge class="w-3.5 h-3.5" /> {{ fmtSpeed(player.speed) }}
        </button>
      </div>
    </div>
  </section>
</template>
