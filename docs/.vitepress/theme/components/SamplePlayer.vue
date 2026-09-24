<script setup lang="ts">
// Trình phát nghe thử (wireframe WfLanding mục 2), giống trình phát trong app: bìa, tên đoạn,
// thanh vị trí kéo được, nút phát, danh sách đoạn; hết đoạn tự sang đoạn kế.
// Audio thật: sách mẫu docs/demo-books/ky-nang-mem-cho-nguoi-tre, giọng Thiện Minh, lời đọc ở audio/loi-doc.txt.
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { withBase } from 'vitepress'
import { AudioLines, Pause, Play } from 'lucide-vue-next'
import { announcePlay, fmt, onOtherPlay } from './audioBus'

const ID = 'sample'
const tracks = [
  { title: 'Chương 1. Lắng nghe chủ động', src: '/audio/ky-nang-mem-1.mp3', sec: 30.8 },
  { title: 'Chương 2. Giao tiếp hiệu quả', src: '/audio/ky-nang-mem-2.mp3', sec: 22.4 },
  { title: 'Chương 3. Làm việc nhóm', src: '/audio/ky-nang-mem-3.mp3', sec: 22.6 },
  { title: 'Chương 4. Quản lý thời gian', src: '/audio/ky-nang-mem-4.mp3', sec: 26.1 },
]

const audio = ref<HTMLAudioElement>()
const current = ref(0)
const time = ref(0)
const duration = ref(tracks[0].sec)
const playing = ref(false)

const src = computed(() => withBase(tracks[current.value].src))

function play() {
  const a = audio.value
  if (!a) return
  announcePlay(ID)
  void a.play().catch(() => (playing.value = false))
}
function toggle() {
  if (playing.value) audio.value?.pause()
  else play()
}
function pick(i: number) {
  if (i === current.value) return toggle()
  current.value = i
  time.value = 0
  duration.value = tracks[i].sec
  // đợi Vue gắn src mới rồi phát
  requestAnimationFrame(() => {
    audio.value?.load()
    play()
  })
}
function onEnded() {
  if (current.value < tracks.length - 1) pick(current.value + 1)
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
  <div class="sano-card player">
    <audio
      ref="audio"
      :src="src"
      preload="none"
      @play="playing = true"
      @pause="playing = false"
      @ended="onEnded"
      @timeupdate="time = audio?.currentTime ?? 0"
      @loadedmetadata="duration = audio?.duration || tracks[current].sec"
    />
    <div class="top">
      <img :src="withBase('/audio/ky-nang-mem-bia.jpg')" alt="Bìa sách mẫu Kỹ năng mềm cho người trẻ, do Sano tự vẽ" class="cover" width="128" height="171" />
      <div class="info">
        <p class="book">Kỹ năng mềm cho người trẻ</p>
        <p class="sano-muted meta">Giọng Thiện Minh · {{ tracks.length }} đoạn mẫu</p>
        <p class="now">{{ tracks[current].title }}</p>
        <input
          class="range"
          type="range"
          min="0"
          :max="duration"
          step="0.1"
          :value="time"
          aria-label="Vị trí nghe"
          :aria-valuetext="`${fmt(time)} / ${fmt(duration)}`"
          :style="{ '--pct': (time / duration) * 100 + '%' }"
          @input="seek"
        />
        <div class="times"><span>{{ fmt(time) }}</span><span>-{{ fmt(duration - time) }}</span></div>
        <button type="button" class="play sano-focus" :aria-label="playing ? 'Dừng' : 'Phát'" @click="toggle">
          <Pause v-if="playing" :size="20" /><Play v-else :size="20" class="nudge" />
        </button>
      </div>
    </div>
    <ol class="list">
      <li v-for="(t, i) in tracks" :key="t.src">
        <button type="button" class="row sano-focus" :class="{ active: i === current }" :aria-current="i === current ? 'true' : undefined" @click="pick(i)">
          <span class="left">
            <AudioLines v-if="i === current" :size="14" class="ico" aria-hidden="true" />
            <span v-else class="num">{{ i + 1 }}</span>
            <span class="name">{{ t.title }}</span>
          </span>
          <span class="dur">{{ fmt(t.sec) }}</span>
        </button>
      </li>
    </ol>
    <p class="note sano-muted">
      Tạo bằng Sano từ một file Word mẫu tự viết, chạy trên máy tính.
      <a :href="withBase('/audio/loi-doc.txt')" target="_blank" rel="noopener">Xem lời đọc</a>
    </p>
  </div>
</template>

<style scoped>
.player {
  padding: 20px;
}
.top {
  display: flex;
  gap: 20px;
}
.cover {
  align-self: flex-start;
  width: 96px;
  height: auto;
  aspect-ratio: 3 / 4;
  flex-shrink: 0;
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
  flex: 1;
}
.book {
  margin: 0;
  font-weight: 600;
  line-height: 1.35;
}
.meta {
  margin: 0;
  font-size: 14px;
}
.now {
  margin: 12px 0 0;
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
  display: grid;
  place-items: center;
  width: 48px;
  height: 48px;
  border-radius: 9999px;
  background: var(--vp-button-brand-bg);
  color: #fff;
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.2);
}
.play:hover {
  background: var(--vp-button-brand-hover-bg);
}
.nudge {
  margin-left: 2px;
}
.list {
  margin: 16px 0 0;
  padding: 0;
  list-style: none;
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  overflow: hidden;
}
.list li + li {
  border-top: 1px solid var(--vp-c-divider);
}
.row {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 12px;
  text-align: left;
  font-size: 14px;
}
.row:hover {
  background: hsl(var(--sano-muted) / 0.6);
}
.row.active {
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand-1);
  font-weight: 500;
}
.left {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
}
.ico {
  flex-shrink: 0;
}
.num {
  width: 14px;
  flex-shrink: 0;
  text-align: center;
  font-size: 12px;
  color: var(--vp-c-text-2);
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
}
.note {
  margin: 12px 0 0;
  font-size: 12px;
}
.note a {
  color: var(--vp-c-brand-1);
}
</style>
