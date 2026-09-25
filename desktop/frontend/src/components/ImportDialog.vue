<script setup lang="ts">
// Hộp Nhập sách (wireframe D6): xem trước → (trùng: giữ cả hai / thay thế) → đang
// nhập (tiến độ, huỷ) → xong (báo về Thư viện) hoặc lỗi (nói rõ lý do, chưa thêm gì).
// Mọi kiểm tra an toàn gói zip làm ở Go (library/importzip.go).
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { AlertTriangle, Copy, Download, FileArchive, Loader2, Mic, Replace, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import BookCover from '@/components/sano/BookCover.vue'
import { cancelImport, errText, importBookZip, onEvent, previewBookZip, type ImportPreview, type ImportProgress } from '../lib/backend'
import { fmtLong } from '../lib/position'
import { forgetBook } from '../lib/player'

const props = defineProps<{ path: string }>()
const emit = defineEmits<{ close: []; imported: [slug: string] }>()

type Stage = 'loading' | 'preview' | 'importing' | 'error'
const stage = ref<Stage>('loading')
const pv = ref<ImportPreview | null>(null)
const error = ref('')
const choice = ref<'both' | 'replace'>('both')
const progress = ref<ImportProgress>({ done: 0, total: 0 })
const pct = computed(() => (progress.value.total ? Math.round((progress.value.done / progress.value.total) * 100) : 0))
const fileName = computed(() => pv.value?.fileName || props.path.split(/[\\/]/).pop() || '')

let off: (() => void) | null = null
onMounted(async () => {
  off = onEvent<ImportProgress>('import:progress', (p) => (progress.value = p))
  try {
    pv.value = await previewBookZip(props.path)
    stage.value = 'preview'
  } catch (e) {
    error.value = errText(e)
    stage.value = 'error'
  }
})
onBeforeUnmount(() => off?.())

const mb = (n: number) => (n >= 1 << 20 ? `${Math.round(n / (1 << 20)).toLocaleString('vi-VN')} MB` : `${Math.max(1, Math.round(n / 1024))} KB`)
const createdOn = computed(() => {
  const d = pv.value?.existingCreatedAt ? new Date(pv.value.existingCreatedAt) : null
  return d && !isNaN(+d) ? `${d.getDate()}/${d.getMonth() + 1}` : ''
})

async function start() {
  if (!pv.value) return
  const replaceSlug = pv.value.existingSlug && choice.value === 'replace' ? pv.value.existingSlug : ''
  if (replaceSlug) forgetBook(replaceSlug) // đang phát cuốn cũ → dừng trước khi thay
  stage.value = 'importing'
  progress.value = { done: 0, total: pv.value.sections }
  try {
    emit('imported', await importBookZip(props.path, replaceSlug))
  } catch (e) {
    error.value = errText(e)
    stage.value = 'error'
  }
}

function close() {
  if (stage.value === 'importing') void cancelImport()
  else emit('close')
}
function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && stage.value !== 'importing') emit('close')
}
document.addEventListener('keydown', onKey)
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="fixed inset-0 bg-black/40 grid place-items-center z-40" @mousedown.self="stage !== 'importing' && emit('close')">
    <div role="dialog" aria-modal="true" :aria-label="stage === 'error' ? 'Không nhập được gói này' : 'Nhập sách'" class="w-[480px] max-w-[calc(100vw-2rem)] rounded-xl border border-border bg-background shadow-2xl">
      <div class="flex items-center justify-between px-5 pt-4">
        <h2 class="font-semibold">{{ stage === 'error' ? 'Không nhập được gói này' : pv?.existingSlug && stage === 'preview' ? 'Thư viện đã có cuốn này' : 'Nhập sách' }}</h2>
        <button v-if="stage !== 'importing'" aria-label="Đóng" class="text-muted-foreground hover:text-foreground" @click="emit('close')"><X class="w-4 h-4" /></button>
      </div>

      <div v-if="stage === 'loading'" class="px-5 py-8 text-sm text-muted-foreground flex items-center gap-2"><Loader2 class="w-4 h-4 animate-spin" /> Đang kiểm tra gói sách…</div>

      <div v-else-if="stage === 'error'" class="px-5 py-4 text-sm">
        <div class="flex gap-3 rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-3">
          <AlertTriangle class="w-5 h-5 text-destructive shrink-0" />
          <div class="min-w-0">
            <p class="font-medium truncate">{{ fileName }}</p>
            <p class="mt-1 text-muted-foreground break-words">{{ error }}</p>
          </div>
        </div>
        <p class="mt-3 text-xs text-muted-foreground">Chưa có gì được thêm vào thư viện.</p>
      </div>

      <template v-else-if="pv">
        <div class="px-5 pt-4 flex gap-4">
          <div class="w-20 shrink-0 aspect-[3/4] rounded-md overflow-hidden shadow">
            <img v-if="pv.coverDataUrl.startsWith('data:image/')" :src="pv.coverDataUrl" alt="" class="h-full w-full object-cover" />
            <BookCover v-else :title="pv.title" :author="pv.author" class="h-full w-full rounded-md shadow-none" />
          </div>
          <div class="min-w-0 text-sm">
            <p class="font-medium text-base leading-snug break-words">{{ pv.title }}</p>
            <p v-if="pv.author" class="text-muted-foreground truncate">{{ pv.author }}</p>
            <p v-if="pv.voice" class="mt-2 text-muted-foreground flex items-center gap-1"><Mic class="w-3.5 h-3.5" /> Giọng {{ pv.voice }}</p>
            <p class="text-muted-foreground">{{ pv.chapters }} chương · {{ pv.sections }} mục<template v-if="pv.durationSec"> · {{ fmtLong(pv.durationSec) }}</template> · {{ mb(pv.sizeBytes) }}</p>
            <p class="mt-1 text-xs text-muted-foreground truncate flex items-center gap-1"><FileArchive class="w-3.5 h-3.5 shrink-0" /> {{ fileName }}</p>
          </div>
        </div>

        <div v-if="pv.existingSlug && stage === 'preview'" class="px-5 pt-4 space-y-2 text-sm" role="radiogroup" aria-label="Xử lý cuốn trùng">
          <p class="text-muted-foreground">“{{ pv.existingTitle || pv.title }}” đã có trong thư viện<template v-if="createdOn"> (tạo ngày {{ createdOn }})</template>. Bạn muốn làm gì?</p>
          <label class="flex gap-3 rounded-lg border px-3 py-2.5 cursor-pointer" :class="choice === 'both' ? 'border-primary bg-primary/5' : 'border-border'">
            <input v-model="choice" type="radio" value="both" class="mt-0.5 accent-[hsl(var(--primary))]" />
            <span><span class="font-medium flex items-center gap-1.5"><Copy class="w-3.5 h-3.5" /> Giữ cả hai</span><span class="block text-xs text-muted-foreground">Cuốn mới thêm vào với tên “{{ pv.title }} (2)”.</span></span>
          </label>
          <label class="flex gap-3 rounded-lg border px-3 py-2.5 cursor-pointer" :class="choice === 'replace' ? 'border-primary bg-primary/5' : 'border-border'">
            <input v-model="choice" type="radio" value="replace" class="mt-0.5 accent-[hsl(var(--primary))]" />
            <span><span class="font-medium flex items-center gap-1.5"><Replace class="w-3.5 h-3.5" /> Thay thế cuốn đang có</span><span class="block text-xs text-muted-foreground">Cuốn cũ chuyển vào Thùng rác (lấy lại được). Vị trí đang nghe giữ nguyên.</span></span>
          </label>
        </div>

        <div v-if="stage === 'importing'" class="px-5 pt-5 text-sm">
          <div class="flex items-center justify-between text-muted-foreground">
            <span class="flex items-center gap-2"><Loader2 class="w-4 h-4 animate-spin" /> Đang giải nén âm thanh… {{ progress.done }}/{{ progress.total }} mục</span>
            <span class="tabular-nums">{{ pct }}%</span>
          </div>
          <div class="mt-2 h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary rounded-full transition-[width]" :style="{ width: pct + '%' }"></div></div>
        </div>

        <p v-if="stage === 'preview' && !pv.existingSlug" class="mx-5 mt-4 rounded-md bg-muted/60 px-3 py-2 text-xs text-muted-foreground">
          Chỉ nhập sách bạn có quyền nghe, ví dụ sách tự làm, sách được tác giả cho phép chia sẻ, hoặc tác phẩm đã hết bản quyền.
        </p>
      </template>

      <div class="flex justify-end gap-2 px-5 py-4">
        <Button v-if="stage === 'error'" @click="emit('close')">Đóng</Button>
        <Button v-else-if="stage === 'importing'" variant="outline" @click="close">Huỷ</Button>
        <template v-else-if="stage === 'preview'">
          <Button variant="outline" @click="emit('close')">Huỷ</Button>
          <Button @click="start"><Download class="w-4 h-4" /> {{ pv?.existingSlug ? 'Tiếp tục' : 'Nhập vào thư viện' }}</Button>
        </template>
      </div>
    </div>
  </div>
</template>
