<script setup lang="ts">
// Trước / sau chuẩn hoá văn nói (wireframe WfLanding mục 2). Hai audio cùng một đoạn, giọng Thiện Minh:
// bản trước = văn bản gốc đưa thẳng vào bộ đọc; bản sau = cùng đoạn qua chuẩn hoá của Sano (internal/bookmaker).
// Cột "sau" là đúng chữ Sano đã tạo ra (--keep-txt), không viết tay.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { withBase } from 'vitepress'
import { ArrowRight, Pause, Play } from 'lucide-vue-next'
import { announcePlay, fmt, onOtherPlay } from './audioBus'

const ID = 'normalize'
const rows = [
  { before: 'Chương IV. Mô hình 4P', after: 'Chương bốn. Mô hình bốn Pê' },
  { before: 'PGS. TS. Lê Văn C, TP.HCM', after: 'Phó giáo sư Tiến sĩ Lê Văn C, Thành phố Hồ Chí Minh' },
  { before: 'đọc sách → hiểu mình → đổi thói quen', after: 'đọc sách, dẫn tới hiểu mình, dẫn tới đổi thói quen' },
  { before: 'học 2-3 giờ mỗi ngày, vd. sáng & tối', after: 'học 2 đến 3 giờ mỗi ngày, ví dụ sáng và tối' },
]
const versions = {
  before: { label: 'Chưa chuẩn hoá', src: '/audio/chuan-hoa-truoc.mp3', sec: 12.6 },
  after: { label: 'Đã chuẩn hoá', src: '/audio/chuan-hoa-sau.mp3', sec: 11.6 },
} as const
type Pick = keyof typeof versions

const pick = ref<Pick>('after')
const playing = ref(false)
const audio = ref<HTMLAudioElement>()
const v = computed(() => versions[pick.value])

function toggle() {
  const a = audio.value
  if (!a) return
  if (playing.value) return a.pause()
  announcePlay(ID)
  void a.play().catch(() => (playing.value = false))
}
// đổi bản khi đang phát → phát tiếp bản mới từ đầu
watch(pick, () => {
  const was = playing.value
  audio.value?.pause()
  requestAnimationFrame(() => {
    audio.value?.load()
    if (was) toggle()
  })
})

let off: (() => void) | undefined
onMounted(() => (off = onOtherPlay(ID, () => audio.value?.pause())))
onBeforeUnmount(() => off?.())
</script>

<template>
  <div class="sano-card box">
    <audio ref="audio" :src="withBase(v.src)" preload="none" @play="playing = true" @pause="playing = false" @ended="playing = false" />
    <p class="title">Trước / sau khi chuẩn hoá văn nói</p>
    <p class="sano-muted desc">Văn viết để nhìn. Sano tự chuyển thành lời đọc tự nhiên, không cần AI trên mạng.</p>

    <div class="seg" role="group" aria-label="Chọn bản nghe">
      <button
        v-for="(o, k) in versions"
        :key="k"
        type="button"
        class="sano-focus"
        :class="{ on: pick === k }"
        :aria-pressed="pick === k"
        @click="pick = k"
      >
        {{ o.label }}
      </button>
    </div>

    <button type="button" class="listen sano-focus" @click="toggle">
      <span class="circle"><Pause v-if="playing" :size="16" /><Play v-else :size="16" class="nudge" /></span>
      <span class="grow">Nghe bản {{ pick === 'before' ? 'chưa chuẩn hoá' : 'đã chuẩn hoá' }}</span>
      <span class="dur">{{ fmt(v.sec) }}</span>
    </button>

    <div class="rows">
      <div v-for="r in rows" :key="r.before" class="row">
        <p :class="pick === 'before' ? 'strong' : 'faded'">{{ r.before }}</p>
        <p class="after" :class="pick === 'after' ? 'strong' : 'dim'">
          <ArrowRight :size="14" class="arrow" aria-hidden="true" />{{ r.after }}
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.box {
  padding: 20px;
}
.title {
  margin: 0;
  font-weight: 600;
}
.desc {
  margin: 4px 0 0;
  font-size: 14px;
  line-height: 1.55;
}
.seg {
  margin-top: 16px;
  display: inline-flex;
  padding: 2px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  font-size: 14px;
}
.seg button {
  padding: 6px 12px;
  border-radius: 6px;
  color: var(--vp-c-text-2);
}
.seg button:hover {
  color: var(--vp-c-text-1);
}
.seg button.on {
  background: var(--vp-button-brand-bg);
  color: #fff;
}
.listen {
  margin-top: 12px;
  display: flex;
  width: 100%;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  text-align: left;
  font-size: 14px;
}
.listen:hover {
  background: hsl(var(--sano-muted) / 0.6);
}
.circle {
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border-radius: 9999px;
  background: var(--vp-button-brand-bg);
  color: #fff;
}
.nudge {
  margin-left: 2px;
}
.grow {
  flex: 1;
}
.dur {
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  color: var(--vp-c-text-2);
}
.rows {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.row {
  border-radius: 8px;
  background: hsl(var(--sano-muted) / 0.5);
  padding: 8px 12px;
  font-size: 14px;
  line-height: 1.5;
}
.row p {
  margin: 0;
}
.strong {
  font-weight: 500;
  color: var(--vp-c-text-1);
}
.faded {
  color: var(--vp-c-text-2);
  text-decoration: line-through;
  text-decoration-color: hsl(var(--sano-muted-foreground) / 0.4);
}
.dim {
  color: var(--vp-c-text-2);
}
.after {
  margin-top: 2px !important;
  display: flex;
  gap: 6px;
}
.arrow {
  margin-top: 3px;
  flex-shrink: 0;
  color: var(--vp-c-brand-1);
}
</style>
