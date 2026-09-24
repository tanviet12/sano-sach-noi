<script setup lang="ts">
// Dòng trạng thái xuất M4B của một cuốn, hiện ngay dưới nút Xuất M4B: tiến độ +
// Huỷ khi đang xuất, nơi lưu + Mở thư mục khi xong, lỗi rõ khi hỏng.
import { computed } from 'vue'
import { Check, FolderOpen, Loader2, X } from 'lucide-vue-next'
import { cancelM4B, fmtSize, showM4B, useM4B } from '../lib/m4b'
import { fmtLong } from '../lib/position'

const props = defineProps<{ slug: string }>()
const m4b = useM4B()

const st = computed(() => (m4b.status?.slug === props.slug ? m4b.status : null))
const otherRunning = computed(() => m4b.status?.running && m4b.status.slug !== props.slug ? m4b.status : null)
const clickError = computed(() => (m4b.error && m4b.errorSlug === props.slug ? m4b.error : ''))
const pct = computed(() => st.value?.progress?.percent ?? 0)
const fileName = computed(() => st.value?.path.split(/[\\/]/).pop() ?? '')
</script>

<template>
  <div v-if="st?.running" class="w-full flex items-center gap-2 text-sm" role="status" aria-live="polite">
    <Loader2 class="w-4 h-4 animate-spin text-primary shrink-0" />
    <span class="shrink-0 tabular-nums">Đang xuất M4B {{ pct }}%</span>
    <span class="shrink-0 text-xs text-muted-foreground">
      <template v-if="st.progress.phase === 'mux'">ghép mục lục, bìa…</template>
      <template v-else>tiểu mục {{ st.progress.track }}/{{ st.progress.tracks }}</template>
    </span>
    <div class="flex-1 min-w-16 h-1.5 rounded-full bg-muted overflow-hidden">
      <div class="h-full bg-primary rounded-full transition-[width]" :style="{ width: pct + '%' }"></div>
    </div>
    <button class="shrink-0 text-xs text-destructive hover:underline flex items-center gap-1" @click="cancelM4B"><X class="w-3.5 h-3.5" /> Huỷ</button>
  </div>
  <div v-else-if="st?.done" class="w-full flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
    <Check class="w-4 h-4 text-rag-green shrink-0" />
    <span>Đã lưu <span class="font-medium" :title="st.path">{{ fileName }}</span></span>
    <span class="text-xs text-muted-foreground">{{ st.chapters }} mốc chương · {{ fmtLong(st.durationSec) }} · {{ fmtSize(st.size) }}</span>
    <button class="text-xs text-primary hover:underline flex items-center gap-1" @click="showM4B"><FolderOpen class="w-3.5 h-3.5" /> Mở thư mục</button>
  </div>
  <p v-else-if="st?.error" class="w-full text-sm text-destructive whitespace-pre-wrap break-words max-h-32 overflow-auto">Xuất M4B không thành công: {{ st.error }}</p>
  <p v-else-if="st?.cancelled" class="w-full text-xs text-muted-foreground">Đã huỷ xuất M4B.</p>
  <p v-if="clickError" class="w-full text-sm text-destructive">{{ clickError }}</p>
  <p v-if="otherRunning" class="w-full text-xs text-muted-foreground">Đang xuất M4B "{{ otherRunning.title }}" ({{ otherRunning.progress.percent }}%) — xong mới xuất được cuốn khác.</p>
</template>
