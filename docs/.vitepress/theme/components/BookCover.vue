<script setup lang="ts">
// Bìa mặc định giống phần mềm (desktop/frontend/src/components/sano/BookCover.vue), viết lại bằng CSS
// thường vì trang tài liệu không dùng Tailwind. Màu chọn cố định theo tên sách (FNV-1a).
import { computed } from 'vue'

const props = defineProps<{ title: string; author?: string }>()

// Bảng màu bìa (trang trí, cùng bộ với app): nền · chữ nhấn · chữ phụ
const palettes = [
  ['#1e1b4b', '#a5b4fc', 'rgb(199 210 254 / .8)'],
  ['#292524', '#fcd34d', 'rgb(231 229 228 / .8)'],
  ['#881337', '#fda4af', 'rgb(255 228 230 / .8)'],
  ['#92400e', '#fde68a', 'rgb(254 243 199 / .8)'],
  ['#1e293b', '#7dd3fc', 'rgb(226 232 240 / .8)'],
  ['#4c1d95', '#c4b5fd', 'rgb(237 233 254 / .8)'],
  ['#115e59', '#99f6e4', 'rgb(204 251 241 / .8)'],
  ['#991b1b', '#fecaca', 'rgb(254 226 226 / .8)'],
]
const style = computed(() => {
  let h = 2166136261
  for (const ch of props.title) h = Math.imul(h ^ (ch.codePointAt(0) ?? 0), 16777619) >>> 0
  const [bg, accent, sub] = palettes[h % palettes.length]
  return { '--c-bg': bg, '--c-accent': accent, '--c-sub': sub }
})
const bars = [35, 70, 100, 60, 40]
</script>

<template>
  <div class="cover" :style="style" role="img" :aria-label="title">
    <div class="spine" />
    <div class="glow" />
    <div class="inner">
      <div class="top">
        <div class="meta">
          <p class="label">Sách nói</p>
          <div class="rule" />
        </div>
        <div class="bars">
          <span v-for="(b, i) in bars" :key="i" class="bar" :style="{ height: b + '%' }" />
        </div>
      </div>
      <p class="title">{{ title }}</p>
      <p v-if="author" class="author">{{ author }}</p>
    </div>
  </div>
</template>

<style scoped>
.cover {
  position: relative;
  overflow: hidden;
  border-radius: 8px;
  aspect-ratio: 3 / 4;
  background: var(--c-bg);
  color: #fff;
  container-type: inline-size;
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.1);
}
.spine {
  position: absolute;
  inset: 0 auto 0 0;
  width: 7%;
  background: rgb(0 0 0 / 0.25);
  border-right: 1px solid rgb(255 255 255 / 0.15);
}
.glow {
  position: absolute;
  top: -25%;
  right: -25%;
  width: 75%;
  height: 75%;
  border-radius: 9999px;
  background: rgb(255 255 255 / 0.1);
  filter: blur(24px);
}
.inner {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 9cqw 9cqw 9cqw 16cqw;
}
.top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 4px;
}
.label {
  margin: 0;
  color: var(--c-accent);
  font-size: 5.5cqw;
  font-weight: 600;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  line-height: 1.3;
}
.rule {
  margin-top: 3cqw;
  width: 18cqw;
  height: 1px;
  background: var(--c-accent);
  opacity: 0.6;
}
.bars {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  height: 16cqw;
  gap: 2.2cqw;
  opacity: 0.7;
}
.bar {
  width: 2.6cqw;
  border-radius: 9999px;
  background: var(--c-accent);
}
.title {
  margin: auto 0 0;
  font-family: ui-serif, Georgia, 'Times New Roman', serif;
  font-weight: 600;
  font-size: 12.5cqw;
  line-height: 1.15;
  text-wrap: balance;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.author {
  margin: 5cqw 0 0;
  color: var(--c-sub);
  font-size: 5.5cqw;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
@container (max-width: 72px) {
  .meta,
  .author {
    display: none;
  }
  .inner {
    padding: 10cqw 8cqw 10cqw 14cqw;
  }
  .title {
    font-size: 15cqw;
    -webkit-line-clamp: 3;
  }
}
</style>
