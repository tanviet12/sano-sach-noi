<script setup lang="ts">
// Dải "Nghe giọng khác" ở trang chủ, dưới trình phát nghe thử (wireframe WfDemo, phần đổi trang chủ):
// 5 giọng khác Hải Đăng + link sang trang /demo.
import { withBase } from 'vitepress'
import { ArrowRight, Pause, Play } from 'lucide-vue-next'
import { allVoices, voiceAudio } from './demoData'
import { useClip } from './useClip'

const picks = ['Trúc Ly', 'Thái Sơn', 'Ngọc Trân', 'Thục Đoan', 'Ngọc Huyền']
const voices = picks.map((n) => allVoices.find((v) => v.name === n)!)
const { current, toggle } = useClip('home-voices')
</script>

<template>
  <div class="sano-card strip">
    <span class="label">Nghe giọng khác:</span>
    <button
      v-for="v in voices"
      :key="v.file"
      type="button"
      class="chip sano-focus"
      :class="{ on: current === voiceAudio(v) }"
      :aria-pressed="current === voiceAudio(v)"
      :aria-label="`Nghe giọng ${v.name}, ${v.desc}`"
      @click="toggle(voiceAudio(v))"
    >
      <Pause v-if="current === voiceAudio(v)" :size="14" /><Play v-else :size="14" /> {{ v.name }}
    </button>
    <a :href="withBase('/demo')" class="more">Nghe thêm 5 cuốn, 12 giọng <ArrowRight :size="16" aria-hidden="true" /></a>
  </div>
</template>

<style scoped>
.strip {
  margin-top: 24px;
  padding: 14px 16px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.label {
  margin-right: 4px;
  font-size: 14px;
  font-weight: 500;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 9999px;
  font-size: 14px;
  transition: background-color 0.15s, border-color 0.15s;
}
.chip:hover {
  background: var(--vp-c-bg-soft);
}
.chip.on {
  border-color: var(--vp-c-brand-1);
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand-1);
}
.more {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  font-weight: 500;
  color: var(--vp-c-brand-1);
  text-decoration: none;
}
.more:hover {
  text-decoration: underline;
}
@media (max-width: 639px) {
  .more {
    margin-left: 0;
    width: 100%;
  }
}
</style>
