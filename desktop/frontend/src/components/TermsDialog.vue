<script setup lang="ts">
// Hộp xem lại điều khoản sử dụng (mở từ trang Giới thiệu). Esc hoặc bấm nền để đóng.
import { onBeforeUnmount, onMounted } from 'vue'
import { X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import TermsText from './TermsText.vue'
import { TERMS_VERSION } from '../lib/terms'

const emit = defineEmits<{ close: [] }>()
const onKey = (e: KeyboardEvent) => e.key === 'Escape' && emit('close')
onMounted(() => document.addEventListener('keydown', onKey))
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="absolute inset-0 z-20 grid place-items-center bg-background/70 backdrop-blur-sm" @click.self="emit('close')">
    <div role="dialog" aria-modal="true" aria-labelledby="terms-title" class="flex max-h-[85%] w-[640px] flex-col rounded-xl border border-border bg-card text-card-foreground shadow-2xl">
      <div class="flex items-center justify-between border-b border-border px-5 py-4">
        <h2 id="terms-title" class="font-semibold">Điều khoản sử dụng · phiên bản {{ TERMS_VERSION }}</h2>
        <button class="text-muted-foreground hover:text-foreground" aria-label="Đóng" @click="emit('close')"><X class="h-4 w-4" /></button>
      </div>
      <div class="flex-1 overflow-auto px-5 py-4"><TermsText /></div>
      <div class="flex justify-end border-t border-border px-5 py-3"><Button variant="outline" @click="emit('close')">Đóng</Button></div>
    </div>
  </div>
</template>
