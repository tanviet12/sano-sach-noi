<script setup lang="ts">
// Bìa sách mặc định (khi sách chưa có ảnh bìa). Màu chọn cố định theo tên sách để
// cùng một cuốn luôn cùng màu. Họa tiết sóng âm lấy từ logo Sano.
// Palette dùng class Tailwind cố định (giống avatar palette) — không hex.
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{ title: string; author?: string; size?: 'sm' | 'md' | 'lg' }>(),
  { author: '', size: 'md' },
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
const p = computed(() => {
  // FNV-1a: phân bố màu đều hơn giữa các tên sách gần giống nhau
  let h = 2166136261
  for (const ch of props.title) h = Math.imul(h ^ (ch.codePointAt(0) ?? 0), 16777619) >>> 0
  return palettes[h % palettes.length]
})

const s = computed(
  () =>
    ({
      sm: { pad: 'p-2', label: 'text-[6px]', title: 'text-[10px]', author: 'text-[6px]', bars: 'h-3 gap-px', bar: 'w-[2px]' },
      md: { pad: 'p-4', label: 'text-[9px]', title: 'text-lg', author: 'text-[10px]', bars: 'h-8 gap-[3px]', bar: 'w-1' },
      lg: { pad: 'p-5', label: 'text-[10px]', title: 'text-2xl', author: 'text-xs', bars: 'h-10 gap-1', bar: 'w-1.5' },
    })[props.size],
)
const bars = [35, 70, 100, 60, 40]
</script>

<template>
  <div :class="[p.bg, s.pad]" class="relative h-full w-full overflow-hidden rounded-[inherit] text-white flex flex-col">
    <!-- Gáy sách: dải tối bên trái + vệt sáng mảnh -->
    <div class="absolute inset-y-0 left-0 w-[7%] bg-black/25"></div>
    <div class="absolute inset-y-0 left-[7%] w-px bg-white/15"></div>
    <!-- Vầng sáng góc trên -->
    <div class="absolute -top-1/4 -right-1/4 h-3/4 w-3/4 rounded-full bg-white/10 blur-2xl"></div>

    <div class="relative pl-[8%] flex-1 flex flex-col">
      <div class="flex items-start justify-between gap-2">
        <div>
          <p :class="[s.label, p.accent]" class="font-semibold uppercase tracking-[0.2em]">Sách nói</p>
          <div :class="p.accent" class="mt-1.5 h-px w-8 bg-current opacity-60"></div>
        </div>
        <!-- Họa tiết sóng âm (motif logo) -->
        <div :class="[s.bars, p.accent]" class="flex items-center shrink-0 opacity-70">
          <span v-for="(h, i) in bars" :key="i" :class="s.bar" class="rounded-full bg-current" :style="{ height: h + '%' }"></span>
        </div>
      </div>

      <h3 :class="s.title" class="mt-auto font-serif font-semibold leading-[1.15] text-balance line-clamp-4">{{ title }}</h3>

      <p v-if="author" :class="[s.author, p.sub]" class="mt-3 uppercase tracking-wider truncate">{{ author }}</p>
    </div>
  </div>
</template>
