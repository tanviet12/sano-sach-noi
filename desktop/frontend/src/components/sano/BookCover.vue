<script setup lang="ts">
// Bìa sách: dùng ảnh `cover_image` nếu có; nếu không → bìa mặc định (wireframe WfBookCover):
// màu chọn cố định theo tên sách, gáy sách, tên chữ có chân, motif sóng âm của logo.
// Chữ co giãn theo bề rộng bìa (container query) nên dùng được từ mini player 36px tới bìa lớn.
// Palette dùng class Tailwind cố định (giống avatar palette) — không hex.
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<{
    title: string
    author?: string
    coverImage?: string | null
    /** Giữ để tương thích chỗ gọi cũ; màu giờ chọn theo tên sách. */
    seed?: number
    class?: string
  }>(),
  { author: '', seed: 0 },
)

const palettes = [
  { bg: 'bg-indigo-950', accent: 'text-indigo-300', sub: 'text-indigo-200/80' },
  { bg: 'bg-stone-800', accent: 'text-amber-300', sub: 'text-stone-200/80' },
  { bg: 'bg-rose-900', accent: 'text-rose-300', sub: 'text-rose-100/80' },
  { bg: 'bg-amber-800', accent: 'text-amber-200', sub: 'text-amber-100/80' },
  { bg: 'bg-slate-800', accent: 'text-sky-300', sub: 'text-slate-200/80' },
  { bg: 'bg-violet-900', accent: 'text-violet-300', sub: 'text-violet-100/80' },
  { bg: 'bg-teal-800', accent: 'text-teal-200', sub: 'text-teal-100/80' },
  { bg: 'bg-red-800', accent: 'text-red-200', sub: 'text-red-100/80' },
]
const palette = computed(() => {
  // FNV-1a: phân bố màu đều giữa các tên sách gần giống nhau
  let h = 2166136261
  for (const ch of props.title) h = Math.imul(h ^ (ch.codePointAt(0) ?? 0), 16777619) >>> 0
  return palettes[h % palettes.length]
})
const bars = [35, 70, 100, 60, 40]
</script>

<template>
  <div :class="cn('relative overflow-hidden rounded-lg shadow-sm bg-muted', $props.class)">
    <img
      v-if="coverImage"
      :src="coverImage"
      :alt="title"
      class="h-full w-full object-cover"
      loading="lazy"
    />
    <div v-else :class="palette.bg" class="cover relative h-full w-full text-white" role="img" :aria-label="title">
      <!-- Gáy sách + vầng sáng góc -->
      <div class="absolute inset-y-0 left-0 w-[7%] bg-black/25" />
      <div class="absolute inset-y-0 left-[7%] w-px bg-white/15" />
      <div class="absolute -top-1/4 -right-1/4 h-3/4 w-3/4 rounded-full bg-white/10 blur-2xl" />

      <div class="inner relative h-full flex flex-col">
        <div class="flex items-start justify-between gap-1">
          <div class="meta">
            <p :class="palette.accent" class="label font-semibold uppercase">Sách nói</p>
            <div :class="palette.accent" class="rule h-px bg-current opacity-60" />
          </div>
          <div :class="palette.accent" class="bars flex items-center shrink-0 opacity-70">
            <span v-for="(h, i) in bars" :key="i" class="bar rounded-full bg-current" :style="{ height: h + '%' }" />
          </div>
        </div>
        <p class="title mt-auto font-serif font-semibold text-balance">{{ title }}</p>
        <p v-if="author" :class="palette.sub" class="author uppercase truncate">{{ author }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Kích thước theo % bề rộng bìa (cqw) — Tailwind 3.4 chưa có container query sẵn. */
.cover { container-type: inline-size; }
.inner { padding: 9cqw 9cqw 9cqw 16cqw; }
.label { font-size: 5.5cqw; letter-spacing: 0.2em; }
.rule { margin-top: 3cqw; width: 18cqw; }
.bars { height: 16cqw; gap: 2.2cqw; }
.bar { width: 2.6cqw; }
.title {
  font-size: 12.5cqw;
  line-height: 1.15;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.author { margin-top: 5cqw; font-size: 5.5cqw; letter-spacing: 0.08em; }

/* Bìa nhỏ (danh sách): bỏ nhãn + tác giả, giữ tên sách */
@container (max-width: 72px) {
  .meta, .author { display: none; }
  .inner { padding: 10cqw 8cqw 10cqw 14cqw; }
  .title { font-size: 15cqw; -webkit-line-clamp: 3; }
}
/* Bìa rất nhỏ (mini player 36px): chỉ còn màu + sóng âm ở giữa */
@container (max-width: 44px) {
  .title { display: none; }
  .inner { justify-content: center; align-items: center; padding: 0 0 0 7cqw; }
  .inner > div:first-child { justify-content: center; }
  .bars { height: 40cqw; gap: 5cqw; }
  .bar { width: 6cqw; }
}
</style>
