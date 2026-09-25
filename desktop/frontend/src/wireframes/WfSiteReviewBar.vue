<script setup lang="ts">
// Thanh duyệt NGOÀI khung trang (không thuộc trang thật): chuyển trang wireframe, sáng/tối,
// bật/tắt chú thích "phần nào VitePress có sẵn, phần nào tự làm", + nút riêng của từng trang (slot).
import { Eye, EyeOff, Moon, Sun } from 'lucide-vue-next'

defineProps<{ page: 'landing' | 'docs' | 'demo'; dark: boolean; notes: boolean }>()
defineEmits<{ 'toggle-dark': []; 'toggle-notes': [] }>()

const chip = 'h-7 px-2.5 rounded-full border text-xs flex items-center gap-1.5 transition-colors'
const on = 'border-primary bg-primary text-primary-foreground'
const off = 'border-border bg-background text-foreground hover:bg-muted'
</script>

<template>
  <div class="border-b border-dashed border-border bg-muted/70 text-foreground">
    <div class="mx-auto flex max-w-6xl flex-wrap items-center gap-2 px-4 py-2 sm:px-6">
      <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Wireframe · duyệt</span>
      <a href="?wireframe=landing" :class="[chip, page === 'landing' ? on : off]">Trang chủ</a>
      <a href="?wireframe=docs" :class="[chip, page === 'docs' ? on : off]">Trang hướng dẫn mẫu</a>
      <a href="?wireframe=demo" :class="[chip, page === 'demo' ? on : off]">Trang demo</a>
      <span class="h-4 w-px bg-border" />
      <slot />
      <button type="button" :class="[chip, off]" @click="$emit('toggle-dark')">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
      <button type="button" :class="[chip, notes ? on : off]" @click="$emit('toggle-notes')">
        <component :is="notes ? Eye : EyeOff" class="w-3.5 h-3.5" /> Chú thích VitePress
      </button>
    </div>
  </div>
</template>
