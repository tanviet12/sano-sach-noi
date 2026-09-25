<script setup lang="ts">
// Màn nghe: mục lục tiểu mục, tua −15s/+30s, đổi tốc độ, nhớ vị trí nghe; xuất M4B
// (tiến độ ngay dưới hàng nút), xuất gói zip, xoá vào Thùng rác. Việc phát nằm ở
// lib/player.ts (dùng chung với thanh nghe nhỏ) nên rời màn này vẫn nghe tiếp.
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  Check, ChevronDown, ChevronLeft, Download, FolderOpen, Gauge, Loader2, Mic, Package, Pause, Play, RotateCcw, RotateCw,
  SkipBack, SkipForward, Trash2, Volume2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import BookCover from '@/components/sano/BookCover.vue'
import M4BProgress from '../components/M4BProgress.vue'
import LyricsPanel from '../components/LyricsPanel.vue'
import { deleteBook, errText, openBookFolder, revealBookZip } from '../lib/backend'
import { fmtClock, fmtLong } from '../lib/position'
import { go, state } from '../lib/store'
import { m4bBusy, startM4B, useM4B } from '../lib/m4b'
import {
  SPEEDS as speeds, bookHasLyrics, forgetBook, lyricIndex, lyrics, pause, pctTrack, pick, player, seek, seekFrac, setSpeed as applySpeed, skip, toggle, totalSec,
  track, tracks,
} from '../lib/player'

useM4B()

const actionError = ref('')
const lyricsOpen = ref(false)
watch(() => state.playerSlug, () => (lyricsOpen.value = false))
const tocEl = ref<HTMLElement | null>(null)

// Mục lục tự cuộn tới tiểu mục đang phát (mở màn nghe, chuyển tiểu mục, sang cuốn khác).
function scrollToCurrent(smooth: boolean) {
  void nextTick(() => {
    const el = tocEl.value?.querySelector<HTMLElement>('[data-current="true"]')
    el?.scrollIntoView({ block: 'center', behavior: smooth ? 'smooth' : 'auto' })
  })
}
onMounted(() => scrollToCurrent(false))
watch(() => player.current, () => scrollToCurrent(true))
watch(() => player.detail?.slug, () => scrollToCurrent(false))
watch(lyricsOpen, (open) => !open && scrollToCurrent(false))
const speedOpen = ref(false)

function seekTo(e: MouseEvent) {
  const r = (e.currentTarget as HTMLElement).getBoundingClientRect()
  seekFrac((e.clientX - r.left) / r.width)
}

function setSpeed(v: number) {
  applySpeed(v)
  speedOpen.value = false
}

// Đóng menu tốc độ khi bấm ra ngoài hoặc nhấn Esc.
function closeSpeed(e: Event) {
  if (e instanceof KeyboardEvent ? e.key === 'Escape' : !(e.target as HTMLElement).closest('[data-speed-menu]')) speedOpen.value = false
}
document.addEventListener('mousedown', closeSpeed)
document.addEventListener('keydown', closeSpeed)
onBeforeUnmount(() => {
  document.removeEventListener('mousedown', closeSpeed)
  document.removeEventListener('keydown', closeSpeed)
})

const fmtSpeed = (s: number) => s.toLocaleString('vi-VN') + '×'

async function removeBook() {
  actionError.value = ''
  try {
    pause()
    const slug = state.playerSlug
    const where = await deleteBook(slug)
    if (!where) return // người dùng huỷ
    forgetBook(slug)
    go('library')
  } catch (e) {
    actionError.value = errText(e)
  }
}

async function act(fn: (slug: string) => Promise<void>) {
  try {
    await fn(state.playerSlug)
  } catch (e) {
    actionError.value = errText(e)
  }
}
</script>

<template>
  <LyricsPanel v-if="lyricsOpen && player.detail" @close="lyricsOpen = false" />
  <section v-else class="flex-1 flex min-h-0">
    <div class="flex-1 flex flex-col p-6 min-w-0">
      <button class="text-sm text-muted-foreground flex items-center gap-1 hover:text-foreground w-fit" @click="go('library')"><ChevronLeft class="w-4 h-4" /> Thư viện</button>
      <div v-if="!player.detail || player.slug !== state.playerSlug" class="flex-1 grid place-items-center text-sm text-muted-foreground">
        <span v-if="player.error" class="text-destructive">{{ player.error }}</span>
        <span v-else-if="!state.playerSlug">Chọn một cuốn trong thư viện để nghe.</span>
        <span v-else class="flex items-center gap-2"><Loader2 class="w-4 h-4 animate-spin" /> Đang mở sách…</span>
      </div>
      <div v-else class="flex-1 flex flex-col items-center justify-center">
        <div class="aspect-[3/4] rounded-xl shadow-2xl overflow-hidden" :class="bookHasLyrics ? 'w-40' : 'w-48'">
          <img v-if="player.detail.coverUrl" :src="player.detail.coverUrl" :alt="player.detail.title" class="h-full w-full object-cover" />
          <BookCover v-else :title="player.detail.title" :author="player.detail.author" class="h-full w-full rounded-xl shadow-none" />
        </div>
        <h2 class="mt-5 text-lg font-semibold">{{ player.detail.title }}</h2>
        <p v-if="player.detail.author || player.detail.voice" class="text-sm text-muted-foreground flex items-center gap-1">
          {{ player.detail.author }}<template v-if="player.detail.author && player.detail.voice"> ·</template>
          <span v-if="player.detail.voice" class="inline-flex items-center gap-1"><Mic class="w-3.5 h-3.5" /> Giọng {{ player.detail.voice }}</span>
        </p>
        <p class="text-sm text-muted-foreground max-w-md truncate" :title="track?.title">{{ track?.title }}</p>
        <!-- Chiều cao cố định (2 dòng câu đang đọc + 1 dòng câu kế) để không đẩy nút phía dưới. -->
        <button v-if="bookHasLyrics" class="group mt-4 w-full max-w-md rounded-lg border border-border bg-muted/30 hover:bg-muted/60 px-4 py-3 text-left disabled:cursor-default disabled:hover:bg-muted/30"
          :disabled="!lyrics" @click="lyricsOpen = true">
          <span class="flex items-center justify-between text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
            Lời đọc <span v-if="lyrics" class="normal-case font-normal tracking-normal text-primary opacity-0 group-hover:opacity-100">Xem cả lời →</span>
          </span>
          <template v-if="lyrics">
            <span class="mt-1 block h-[2.75em] text-sm font-medium leading-snug line-clamp-2">{{ lyrics.sentences[lyricIndex]?.text }}</span>
            <span class="mt-0.5 block h-[1.375em] text-sm text-muted-foreground leading-snug line-clamp-1">{{ lyrics.sentences[lyricIndex + 1]?.text }}</span>
          </template>
          <template v-else>
            <span class="mt-1 block h-[2.75em] text-sm text-muted-foreground leading-snug">Tiểu mục này không có chữ để hiện.</span>
            <span class="mt-0.5 block h-[1.375em]"></span>
          </template>
        </button>
        <div class="mt-5 w-full max-w-md">
          <div class="h-1.5 rounded-full bg-muted overflow-hidden cursor-pointer" role="slider" aria-label="Vị trí nghe" :aria-valuenow="Math.round(pctTrack)" @click="seekTo">
            <div class="h-full bg-primary rounded-full" :style="{ width: pctTrack + '%' }"></div>
          </div>
          <div class="flex justify-between text-[11px] text-muted-foreground mt-1 tabular-nums"><span>{{ fmtClock(player.time) }}</span><span>-{{ fmtClock(Math.max(0, player.duration - player.time)) }}</span></div>
        </div>
        <div class="mt-4 flex items-center gap-4">
          <button aria-label="Tiểu mục trước" class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted" @click="skip(-1)"><SkipBack class="w-5 h-5" /></button>
          <button aria-label="Lùi 15 giây" class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted" @click="seek(-15)"><RotateCcw class="w-5 h-5" /></button>
          <button :aria-label="player.playing ? 'Dừng' : 'Phát'" class="h-14 w-14 grid place-items-center rounded-full bg-primary text-primary-foreground shadow" @click="toggle">
            <Pause v-if="player.playing" class="w-6 h-6" /><Play v-else class="w-6 h-6 ml-0.5" />
          </button>
          <button aria-label="Tới 30 giây" class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted" @click="seek(30)"><RotateCw class="w-5 h-5" /></button>
          <button aria-label="Tiểu mục sau" class="h-10 w-10 grid place-items-center rounded-full hover:bg-muted" @click="skip(1)"><SkipForward class="w-5 h-5" /></button>
        </div>
        <div class="mt-4 flex items-center gap-2 text-xs">
          <div class="relative" data-speed-menu>
            <button class="h-8 px-3 rounded-full border border-border flex items-center gap-1.5 hover:bg-muted" aria-haspopup="listbox" :aria-expanded="speedOpen" @click="speedOpen = !speedOpen">
              <Gauge class="w-3.5 h-3.5" />{{ fmtSpeed(player.speed) }}<ChevronDown class="w-3 h-3 text-muted-foreground" />
            </button>
            <div v-if="speedOpen" role="listbox" aria-label="Tốc độ phát" class="absolute bottom-full left-0 mb-2 w-40 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-20">
              <div class="px-3 py-1.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Tốc độ phát</div>
              <button v-for="v in speeds" :key="v" role="option" :aria-selected="v === player.speed"
                class="w-full flex items-center justify-between px-3 py-1.5 text-sm text-left hover:bg-muted"
                :class="v === player.speed && 'text-primary font-medium'" @click="setSpeed(v)">
                {{ fmtSpeed(v) }}<Check v-if="v === player.speed" class="w-4 h-4" />
              </button>
            </div>
          </div>
          <span class="text-muted-foreground">Tua −15s / +30s · nhớ vị trí nghe · rời màn này vẫn nghe tiếp</span>
        </div>
        <p v-if="player.error || actionError" class="mt-3 text-sm text-destructive">{{ player.error || actionError }}</p>
      </div>
      <div class="flex flex-wrap items-center gap-2 border-t border-border pt-4">
        <Button variant="outline" size="sm" :disabled="!player.detail || m4bBusy()" title="Một file có mục lục chương và bìa — chép sang điện thoại nghe bằng app sách nói bất kỳ" @click="startM4B(state.playerSlug)"><Download class="w-4 h-4" /> Xuất M4B</Button>
        <Button variant="outline" size="sm" :disabled="!player.detail?.zip" title="Sao lưu hoặc chuyển sách sang máy khác" @click="act(revealBookZip)"><Package class="w-4 h-4" /> Xuất gói zip</Button>
        <Button variant="ghost" size="sm" :disabled="!player.detail" @click="act(openBookFolder)"><FolderOpen class="w-4 h-4" /> Mở thư mục</Button>
        <Button variant="ghost" size="sm" class="ml-auto text-destructive hover:text-destructive" :disabled="!player.detail" title="Chuyển cuốn sách vào Thùng rác (lấy lại được)" @click="removeBook"><Trash2 class="w-4 h-4" /> Xoá</Button>
        <M4BProgress v-if="state.playerSlug" :slug="state.playerSlug" />
      </div>
    </div>
    <div ref="tocEl" class="w-72 shrink-0 border-l border-border overflow-auto">
      <div class="px-4 py-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Mục lục · {{ tracks.length }} mục · {{ fmtLong(totalSec) }}</div>
      <button v-for="(c, i) in tracks" :key="c.file" :data-current="i === player.current"
        class="w-full flex items-center justify-between gap-2 px-4 py-2.5 text-sm text-left hover:bg-muted/60"
        :class="i === player.current ? 'bg-primary/10 text-primary font-medium' : i < player.current ? 'text-muted-foreground' : ''"
        :title="c.chapter !== c.title ? c.chapter : ''"
        @click="pick(i)">
        <span class="truncate flex items-center gap-2">
          <Volume2 v-if="i === player.current" class="w-3.5 h-3.5 shrink-0" />
          <Check v-else-if="i < player.current" class="w-3.5 h-3.5 shrink-0" />
          <span v-else class="w-3.5 shrink-0"></span>
          {{ c.title }}
        </span>
        <span class="text-xs tabular-nums shrink-0">{{ fmtClock(c.durationSec) }}</span>
      </button>
    </div>
  </section>
</template>
