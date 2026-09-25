<script setup lang="ts">
// Tạo sách mới: 6 bước, dữ liệu thật từ phần Go (bookmaker).
// Nghe thử không bắt buộc; chỉ cần xác nhận quyền dùng tài liệu là render được.
import { computed } from 'vue'
import { Check, ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { steps } from '../lib/mock'
import { canRender, rendering, selectedStems, startRender, state } from '../lib/store'
import StepFile from './create/StepFile.vue'
import StepToc from './create/StepToc.vue'
import StepVoice from './create/StepVoice.vue'
import StepIntro from './create/StepIntro.vue'
import StepPreview from './create/StepPreview.vue'
import StepRender from './create/StepRender.vue'

const locked = computed(() => rendering.value || !!state.render?.done)
const canNext = computed(() => {
  if (state.step === 1) return !!state.outline && !state.loading && state.title.trim() !== ''
  if (state.step === 2) return selectedStems.value.length > 0
  return true
})

function jump(n: number) {
  // Đang render/đã xong thì không nhảy ngược vào các bước chỉnh sửa
  if (locked.value) return
  if (n > 1 && !state.outline) return
  if (n === 6) {
    if (canRender.value) void startRender()
    return
  }
  state.step = n
}
</script>

<template>
  <section class="flex-1 flex flex-col min-h-0">
    <!-- Thanh bước -->
    <div class="shrink-0 border-b border-border px-6 h-14 flex items-center gap-1">
      <template v-for="(s, i) in steps" :key="s.n">
        <button class="flex items-center gap-2 px-2 h-8 rounded-md text-sm" :class="state.step === s.n ? 'text-foreground font-medium' : 'text-muted-foreground hover:text-foreground'" @click="jump(s.n)">
          <span class="h-6 w-6 rounded-full grid place-items-center text-xs"
            :class="state.step > s.n ? 'bg-primary/15 text-primary' : state.step === s.n ? 'bg-primary text-primary-foreground' : 'bg-muted'">
            <Check v-if="state.step > s.n" class="w-3.5 h-3.5" /><template v-else>{{ s.n }}</template>
          </span>
          {{ s.label }}
        </button>
        <ChevronRight v-if="i < steps.length - 1" class="w-4 h-4 text-muted-foreground/50" />
      </template>
    </div>

    <div class="flex-1 overflow-auto p-6">
      <StepFile v-if="state.step === 1" />
      <StepToc v-else-if="state.step === 2" />
      <StepVoice v-else-if="state.step === 3" />
      <StepIntro v-else-if="state.step === 4" />
      <StepPreview v-else-if="state.step === 5" />
      <StepRender v-else />
    </div>

    <!-- Chân: nút lùi / tiếp -->
    <div v-if="state.step < 6" class="shrink-0 border-t border-border px-6 h-16 flex items-center justify-between gap-4">
      <Button variant="ghost" :disabled="state.step === 1" @click="state.step--"><ChevronLeft class="w-4 h-4" /> Quay lại</Button>
      <p v-if="state.renderError" class="text-sm text-destructive truncate" :title="state.renderError">{{ state.renderError }}</p>
      <p v-else-if="state.step === 5 && !state.rightsConfirmedAt" class="text-xs text-muted-foreground">
        Tick xác nhận quyền dùng tài liệu để mở nút render
      </p>
      <Button v-if="state.step < 5" :disabled="!canNext" @click="state.step++">Tiếp tục <ChevronRight class="w-4 h-4" /></Button>
      <Button v-else :disabled="!canRender || state.previewing" @click="startRender">Nghe ổn, render cả cuốn <ChevronRight class="w-4 h-4" /></Button>
    </div>
  </section>
</template>
