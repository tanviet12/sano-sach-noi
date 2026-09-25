<script setup lang="ts">
// Một đoạn, 12 giọng (trang /demo, wireframe WfDemo mục 3): cùng một đoạn văn, bấm từng giọng để so.
import { Pause, Play } from 'lucide-vue-next'
import { fmt } from './audioBus'
import { voiceAudio, voicePassage, voiceRegions } from './demoData'
import { useClip } from './useClip'

const { current, time, toggle } = useClip('demo-voices')
</script>

<template>
  <div>
    <blockquote class="passage">{{ voicePassage }}</blockquote>
    <div v-for="r in voiceRegions" :key="r.name" class="region">
      <p class="rn">{{ r.name }}</p>
      <div class="grid">
        <button
          v-for="v in r.voices"
          :key="v.file"
          type="button"
          class="voice sano-focus"
          :class="{ on: current === voiceAudio(v) }"
          :aria-pressed="current === voiceAudio(v)"
          :aria-label="`Nghe giọng ${v.name}, ${v.desc}`"
          @click="toggle(voiceAudio(v))"
        >
          <span class="circle"><Pause v-if="current === voiceAudio(v)" :size="16" /><Play v-else :size="16" class="nudge" /></span>
          <span class="txt"><span class="vn">{{ v.name }}</span><span class="vd sano-muted">{{ v.desc }}</span></span>
          <span class="dur sano-muted">{{ fmt(current === voiceAudio(v) ? time : v.sec) }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.passage {
  margin: 24px 0 0;
  max-width: 48rem;
  padding-left: 16px;
  border-left: 4px solid var(--vp-c-brand-1);
  font-style: italic;
  color: var(--vp-c-text-2);
  line-height: 1.65;
}
.region {
  margin-top: 28px;
}
.rn {
  margin: 0 0 8px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--vp-c-text-2);
}
.grid {
  display: grid;
  gap: 8px;
}
@media (min-width: 640px) {
  .grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (min-width: 1024px) {
  .grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
.voice {
  padding: 10px 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  text-align: left;
  border: 1px solid var(--vp-c-divider);
  border-radius: 10px;
  background: var(--vp-c-bg);
  transition: background-color 0.15s, border-color 0.15s;
}
.voice:hover {
  background: var(--vp-c-bg-soft);
}
.voice.on {
  border-color: var(--vp-c-brand-1);
  background: var(--vp-c-brand-soft);
}
.circle {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  border-radius: 9999px;
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand-1);
}
.voice.on .circle {
  background: var(--vp-button-brand-bg);
  color: #fff;
}
.nudge {
  margin-left: 2px;
}
.txt {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
}
.vn {
  font-weight: 500;
}
.vd {
  font-size: 12px;
}
.dur {
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}
</style>
