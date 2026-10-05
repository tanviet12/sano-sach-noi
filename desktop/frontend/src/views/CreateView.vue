<script setup lang="ts">
// Tạo sách nói: 7 bước, dữ liệu thật từ phần Go (bookmaker).
// Bước 1 Cách đọc (wireframe D8): cấp 1 sang thẳng Nạp file, cấp 2/3 qua màn nhờ AI.
// Nghe thử không bắt buộc; bấm render thì hiện popup cam kết (D19), tick đủ mới render.
import { computed, ref } from 'vue'
import { Check, ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { steps } from '../lib/mock'
import { canRender, rendering, selectedStems, startRender, state } from '../lib/store'
import StepLevel from './create/StepLevel.vue'
import StepFile from './create/StepFile.vue'
import StepToc from './create/StepToc.vue'
import StepVoice from './create/StepVoice.vue'
import StepIntro from './create/StepIntro.vue'
import StepPreview from './create/StepPreview.vue'
import StepRender from './create/StepRender.vue'
import RenderPledgeDialog from '../components/RenderPledgeDialog.vue'

// Cam kết trước khi render: đã cam kết cho file này (nạp file mới thì xoá) thì render luôn,
// chưa thì mở popup.
const pledgeOpen = ref(false)
const fileName = computed(() => state.file?.name ?? state.title)
function requestRender() {
  if (canRender.value) void startRender()
  else pledgeOpen.value = true
}
function pledged() {
  state.rightsConfirmedAt = new Date().toISOString()
  pledgeOpen.value = false
  void startRender()
}

const locked = computed(() => rendering.value || !!state.render?.done)
const canNext = computed(() => {
  if (state.step === 1) return state.level > 0 && (state.levelScreen === 'choose' || !!state.aiTool)
  if (state.step === 2) return !!state.outline && !state.loading && state.title.trim() !== ''
  if (state.step === 3) return selectedStems.value.length > 0
  return true
})
// Nút chính ghi rõ bước kế tiếp (wireframe D8).
const nextLabel = computed(() => {
  if (state.step === 1 && state.levelScreen === 'choose') {
    if (state.level === 1) return 'Tiếp: nạp file'
    if (state.level > 1) return state.level === 3 ? 'Tiếp: nhờ AI viết lại' : 'Tiếp: nhờ AI làm mượt'
    return 'Tiếp'
  }
  if (state.step === 1) return state.aiTool === 'gemini' ? 'Tiếp: dán văn bản AI trả về' : 'Tiếp: nạp file AI tạo'
  if (state.step === 2) return 'Tiếp: mục lục'
  return 'Tiếp tục'
})

function next() {
  if (state.step === 1 && state.levelScreen === 'choose' && state.level > 1) {
    state.levelScreen = 'ai'
    return
  }
  state.step++
}
function back() {
  if (state.step === 1) state.levelScreen = 'choose'
  else if (state.step === 2) {
    state.step = 1
    state.levelScreen = state.level > 1 ? 'ai' : 'choose'
  } else state.step--
}

function jump(n: number) {
  // Đang render/đã xong thì không nhảy ngược vào các bước chỉnh sửa
  if (locked.value) return
  if (n > 2 && !state.outline) return
  if (n === 2 && !state.level) return
  if (n === 7) {
    if (!state.previewing) requestRender()
    return
  }
  if (n === 1) state.levelScreen = 'choose' // bấm "Cách đọc" trên thanh bước: về màn chọn cấp
  state.step = n
}
</script>

<template>
  <section class="relative flex-1 flex flex-col min-h-0">
    <!-- Thanh bước -->
    <div class="shrink-0 border-b border-border px-6 h-14 flex items-center gap-0.5">
      <template v-for="(s, i) in steps" :key="s.n">
        <button class="flex items-center gap-1.5 px-1.5 h-8 rounded-md text-sm whitespace-nowrap" :class="state.step === s.n ? 'text-foreground font-medium' : 'text-muted-foreground hover:text-foreground'" @click="jump(s.n)">
          <span class="h-6 w-6 rounded-full grid place-items-center text-xs"
            :class="state.step > s.n ? 'bg-primary/15 text-primary' : state.step === s.n ? 'bg-primary text-primary-foreground' : 'bg-muted'">
            <Check v-if="state.step > s.n" class="w-3.5 h-3.5" /><template v-else>{{ s.n }}</template>
          </span>
          {{ s.label }}
        </button>
        <ChevronRight v-if="i < steps.length - 1" class="w-3.5 h-3.5 text-muted-foreground/50 shrink-0" />
      </template>
    </div>

    <div class="flex-1 overflow-auto p-6">
      <StepLevel v-if="state.step === 1" />
      <StepFile v-else-if="state.step === 2" />
      <StepToc v-else-if="state.step === 3" />
      <StepVoice v-else-if="state.step === 4" />
      <StepIntro v-else-if="state.step === 5" />
      <StepPreview v-else-if="state.step === 6" />
      <StepRender v-else />
    </div>

    <!-- Chân: nút lùi / tiếp -->
    <div v-if="state.step < 7" class="shrink-0 border-t border-border px-6 h-16 flex items-center justify-between gap-4">
      <Button variant="ghost" :disabled="state.step === 1 && state.levelScreen === 'choose'" @click="back"><ChevronLeft class="w-4 h-4" /> Quay lại</Button>
      <p v-if="state.renderError" class="text-sm text-destructive truncate" :title="state.renderError">{{ state.renderError }}</p>
      <p v-else-if="state.step === 1 && !state.level" class="text-xs text-muted-foreground">Chọn một cách để tiếp tục</p>
      <p v-else-if="state.step === 1 && state.levelScreen === 'ai' && !state.aiTool" class="text-xs text-muted-foreground">Chọn AI để tiếp tục</p>
      <Button v-if="state.step < 6" :disabled="!canNext" @click="next">{{ nextLabel }} <ChevronRight class="w-4 h-4" /></Button>
      <Button v-else :disabled="state.previewing" @click="requestRender">Nghe ổn, render cả cuốn <ChevronRight class="w-4 h-4" /></Button>
    </div>
    <RenderPledgeDialog v-if="pledgeOpen" :file-name="fileName" @close="pledgeOpen = false" @confirm="pledged" />
  </section>
</template>
