<script setup lang="ts">
// Tủ sách nghe thử (trang /demo, wireframe WfDemo mục 2): 5 bìa một hàng (điện thoại vuốt ngang),
// trình phát cuốn đang chọn bên dưới, danh sách 3 chương; hết chương tự sang chương kế.
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { withBase } from 'vitepress'
import { AudioLines, Pause, Play } from 'lucide-vue-next'
import { announcePlay, fmt, onOtherPlay } from './audioBus'
import { bookAudio, bookCover, demoBooks } from './demoData'

const ID = 'demo-shelf'
const audio = ref<HTMLAudioElement>()
const bookIdx = ref(0)
const chIdx = ref(0)
const time = ref(0)
const playing = ref(false)
const book = computed(() => demoBooks[bookIdx.value])
const chapter = computed(() => book.value.chapters[chIdx.value])
const duration = ref(chapter.value.sec)
const src = computed(() => withBase(bookAudio(book.value, chIdx.value)))

function play() {
  const a = audio.value
  if (!a) return
  announcePlay(ID)
  void a.play().catch(() => (playing.value = false))
}
function toggle() {
  if (audio.value && !audio.value.paused) audio.value.pause()
  else play()
}
// Đổi chương/cuốn rồi phát: đợi Vue gắn src mới (không gọi load(), trình duyệt tự nạp lại)
function go(b: number, c: number) {
  if (b === bookIdx.value && c === chIdx.value) return toggle()
  audio.value?.pause()
  bookIdx.value = b
  chIdx.value = c
  time.value = 0
  duration.value = demoBooks[b].chapters[c].sec
  requestAnimationFrame(play)
}
function onEnded() {
  if (chIdx.value < book.value.chapters.length - 1) go(bookIdx.value, chIdx.value + 1)
  else {
    playing.value = false
    time.value = 0
  }
}
function seek(e: Event) {
  const v = Number((e.target as HTMLInputElement).value)
  if (audio.value) audio.value.currentTime = v
  time.value = v
}

let off: (() => void) | undefined
onMounted(() => (off = onOtherPlay(ID, () => audio.value?.pause())))
onBeforeUnmount(() => off?.())
</script>

<template>
  <div class="shelf">
    <audio
      ref="audio"
      :src="src"
      preload="none"
      @play="playing = true"
      @pause="playing = false"
      @ended="onEnded"
      @timeupdate="time = audio?.currentTime ?? 0"
      @loadedmetadata="duration = audio?.duration || chapter.sec"
    />
    <div class="books" role="list" aria-label="Chọn sách để nghe">
      <button
        v-for="(b, i) in demoBooks"
        :key="b.slug"
        type="button"
        role="listitem"
        class="book sano-focus"
        :class="{ on: i === bookIdx }"
        :aria-current="i === bookIdx ? 'true' : undefined"
        @click="go(i, 0)"
      >
        <img :src="withBase(bookCover(b))" :alt="`Bìa sách ${b.title}, do Sano tự vẽ`" loading="lazy" width="200" height="267" />
        <span class="bt">{{ b.title }}</span>
        <span class="bv sano-muted"><AudioLines v-if="i === bookIdx && playing" :size="12" class="ico" aria-hidden="true" />{{ b.voice }}</span>
      </button>
    </div>

    <div class="sano-card player">
      <img :src="withBase(bookCover(book))" alt="" class="cover" width="128" height="171" />
      <div class="info">
        <p class="title">{{ book.title }}</p>
        <p class="sano-muted meta">Giọng <strong>{{ book.voice }}</strong> · {{ book.voiceDesc }}</p>
        <p class="now">Chương {{ chIdx + 1 }}. {{ chapter.title }}</p>
        <input
          class="range"
          type="range"
          min="0"
          :max="duration || 1"
          step="0.1"
          :value="time"
          aria-label="Vị trí nghe"
          :aria-valuetext="`${fmt(time)} / ${fmt(duration)}`"
          :style="{ '--pct': (duration ? (time / duration) * 100 : 0) + '%' }"
          @input="seek"
        />
        <div class="times"><span>{{ fmt(time) }}</span><span>-{{ fmt(duration - time) }}</span></div>
        <button type="button" class="play sano-focus" :aria-label="playing ? 'Dừng' : 'Phát'" @click="toggle">
          <Pause v-if="playing" :size="20" /><Play v-else :size="20" class="nudge" />
        </button>
      </div>
      <ol class="list">
        <li v-for="(c, i) in book.chapters" :key="c.title">
          <button type="button" class="row sano-focus" :class="{ active: i === chIdx }" :aria-current="i === chIdx ? 'true' : undefined" @click="go(bookIdx, i)">
            <span class="left">
              <AudioLines v-if="i === chIdx" :size="14" class="ico" aria-hidden="true" />
              <span v-else class="num">{{ i + 1 }}</span>
              <span class="name">{{ c.title }}</span>
            </span>
            <span class="dur">{{ fmt(c.sec) }}</span>
          </button>
        </li>
      </ol>
    </div>
  </div>
</template>

<style scoped>
.books {
  margin: 32px -16px 0;
  padding: 0 16px 8px;
  display: flex;
  gap: 12px;
  overflow-x: auto;
  scroll-snap-type: x mandatory;
}
@media (min-width: 640px) {
  .books {
    margin: 32px 0 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    overflow: visible;
  }
}
.book {
  flex: 0 0 128px;
  scroll-snap-align: start;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  text-align: left;
  border: 1px solid var(--vp-c-divider);
  border-radius: 10px;
  background: var(--vp-c-bg);
  transition: background-color 0.15s, border-color 0.15s;
}
.book:hover {
  background: var(--vp-c-bg-soft);
}
.book.on {
  border-color: var(--vp-c-brand-1);
  background: var(--vp-c-brand-soft);
}
.book img {
  width: 100%;
  height: auto;
  aspect-ratio: 3 / 4;
  object-fit: cover;
  border-radius: 6px;
  margin-bottom: 6px;
}
.bt {
  font-size: 13px;
  font-weight: 600;
  line-height: 1.35;
}
.bv {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
}
.player {
  margin-top: 24px;
  padding: 20px;
  display: grid;
  gap: 20px;
  grid-template-columns: auto 1fr;
}
@media (min-width: 768px) {
  .player {
    grid-template-columns: auto 1fr 1fr;
    align-items: start;
  }
}
.cover {
  width: 96px;
  height: auto;
  aspect-ratio: 3 / 4;
  border-radius: 8px;
  box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.2);
  object-fit: cover;
}
@media (min-width: 640px) {
  .cover {
    width: 128px;
  }
}
.info {
  min-width: 0;
}
.title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  line-height: 1.35;
}
.meta {
  margin: 4px 0 0;
  font-size: 14px;
}
.meta strong {
  color: var(--vp-c-text-1);
  font-weight: 600;
}
.now {
  margin: 14px 0 0;
  font-size: 14px;
  color: var(--vp-c-brand-1);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.range {
  --pct: 0%;
  margin-top: 8px;
  width: 100%;
  height: 6px;
  appearance: none;
  border-radius: 9999px;
  background: linear-gradient(to right, var(--vp-c-brand-1) var(--pct), var(--vp-c-bg-soft) var(--pct));
  cursor: pointer;
}
.range::-webkit-slider-thumb {
  appearance: none;
  width: 14px;
  height: 14px;
  border-radius: 9999px;
  background: var(--vp-c-brand-1);
  border: 2px solid var(--vp-c-bg);
}
.range::-moz-range-thumb {
  width: 12px;
  height: 12px;
  border-radius: 9999px;
  background: var(--vp-c-brand-1);
  border: 2px solid var(--vp-c-bg);
}
.range:focus-visible {
  outline: 2px solid var(--vp-c-brand-1);
  outline-offset: 3px;
}
.times {
  margin-top: 4px;
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  font-variant-numeric: tabular-nums;
  color: var(--vp-c-text-2);
}
.play {
  margin-top: 8px;
  width: 48px;
  height: 48px;
  display: grid;
  place-items: center;
  border-radius: 9999px;
  background: var(--vp-button-brand-bg);
  color: #fff;
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.15);
}
.nudge {
  margin-left: 2px;
}
.list {
  grid-column: 1 / -1;
  margin: 0;
  padding: 0;
  list-style: none;
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  overflow: hidden;
  min-width: 0;
}
@media (min-width: 768px) {
  .list {
    grid-column: auto;
  }
}
.list li + li {
  border-top: 1px solid var(--vp-c-divider);
}
.row {
  width: 100%;
  padding: 10px 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  text-align: left;
  font-size: 14px;
}
.row:hover {
  background: var(--vp-c-bg-soft);
}
.row.active {
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand-1);
  font-weight: 500;
}
.left {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}
.num {
  width: 14px;
  text-align: center;
  font-size: 12px;
  color: var(--vp-c-text-2);
}
.ico {
  flex-shrink: 0;
  color: var(--vp-c-brand-1);
}
.name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.dur {
  flex-shrink: 0;
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  color: var(--vp-c-text-2);
}
</style>
