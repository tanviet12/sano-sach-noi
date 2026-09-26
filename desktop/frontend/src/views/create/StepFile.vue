<script setup lang="ts">
// B1 Nạp file: hộp chọn file .docx (Wails) hoặc kéo thả → nạp thật: mục lục,
// số ký tự, cảnh báo lúc nạp (hình, bảng, tiêu đề gõ tay, viết tắt chưa có).
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { AlertTriangle, Check, CheckCircle2, Copy, Download, FileText, Loader2, ShieldCheck, Sparkles, Upload, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import BookCover from '@/components/sano/BookCover.vue'
import { chooseCover, chooseDocx, copyText, describeDocx, errText, onFileDrop, sampleDocx, saveSampleDocx } from '../../lib/backend'
import { DOCS } from '../../lib/mock'
import { SAMPLE_PROMPT } from '../../lib/prompt'
import { categoryCounts, seriesKey } from '../../lib/find'
import { clearFile, setFile, state } from '../../lib/store'
import CategoryPicker from '../../components/CategoryPicker.vue'

const copied = ref(false)
const picking = ref(false)
const coverError = ref('')

async function pick() {
  picking.value = true
  try {
    const f = await chooseDocx()
    if (f) await setFile(f)
  } catch (e) {
    state.fileError = errText(e)
  } finally {
    picking.value = false
  }
}

let offDrop = () => {}
onMounted(() => {
  offDrop = onFileDrop(async (paths) => {
    const p = paths[0]
    if (!p || state.loading) return
    try {
      await setFile(await describeDocx(p))
    } catch (e) {
      state.fileError = errText(e)
    }
  })
})
onBeforeUnmount(() => offDrop())

async function copyPrompt() {
  copied.value = await copyText(SAMPLE_PROMPT)
  setTimeout(() => (copied.value = false), 1500)
}

// Lưu file Word mẫu về máy để làm theo (có sẵn Heading 1/2, lời hướng dẫn).
const savedSample = ref('')
// "Mau-sach-noi-Sano.docx" + "Downloads" — không hiện cả đường dẫn dài
const savedName = computed(() => savedSample.value.split(/[\\/]/).pop() ?? '')
const savedDir = computed(() => savedSample.value.split(/[\\/]/).slice(-2, -1)[0] ?? '')
// Nạp đúng file vừa lưu để thử ngay (người dùng không phải đi tìm lại file).
async function useSaved() {
  picking.value = true
  try {
    await setFile(await describeDocx(savedSample.value))
  } catch (e) {
    state.fileError = errText(e)
  } finally {
    picking.value = false
  }
}
async function downloadSample() {
  state.fileError = ''
  try {
    savedSample.value = await saveSampleDocx()
  } catch (e) {
    state.fileError = errText(e)
  }
}

// Chưa có file Word: dùng tài liệu mẫu có sẵn để thử trọn luồng tạo sách.
async function useSample() {
  picking.value = true
  try {
    await setFile(await sampleDocx())
  } catch (e) {
    state.fileError = errText(e)
  } finally {
    picking.value = false
  }
}

async function pickCover() {
  coverError.value = ''
  try {
    const c = await chooseCover()
    if (c) {
      state.coverPath = c.path
      state.coverDataUrl = c.dataUrl
    }
  } catch (e) {
    coverError.value = errText(e)
  }
}

function fmtSize(bytes: number) {
  if (bytes >= 1024 * 1024) return (bytes / 1024 / 1024).toLocaleString('vi-VN', { maximumFractionDigits: 1 }) + ' MB'
  return Math.max(1, Math.round(bytes / 1024)).toLocaleString('vi-VN') + ' KB'
}

const categories = computed(() => categoryCounts(state.library?.books ?? []))
// Bộ sách đang có: [tên, số tập] + số tập đã dùng → gợi ý tập kế tiếp (wireframe D5).
const seriesVols = computed(() => {
  const m = new Map<string, [string, number[]]>()
  for (const b of state.library?.books ?? []) {
    if (!b.series) continue
    const g = m.get(seriesKey(b.series)) ?? [b.series, []]
    g[1].push(b.volume)
    m.set(seriesKey(b.series), g)
  }
  return [...m.values()]
})
const seriesGroups = computed<[string, number][]>(() => seriesVols.value.map(([n, v]) => [n, v.length]))
const taken = computed(() => seriesVols.value.find(([n]) => seriesKey(n) === seriesKey(state.series))?.[1] ?? [])
watch(() => state.series, (n, old) => {
  if (!n) state.volume = 0
  else if (seriesKey(n) !== seriesKey(old ?? '')) state.volume = Math.max(0, ...taken.value) + 1
})
const w = computed(() => state.outline?.warnings)
const chapterCount = computed(() => state.toc.filter((c) => c.kind === 'chapter').length)
const acronyms = computed(() => {
  const list = w.value?.unknownAcronyms ?? []
  const shown = list.slice(0, 8).map((a) => a.word).join(', ')
  return list.length > 8 ? `${shown}…` : shown
})
const hasWarnings = computed(() => !!w.value && (w.value.images + (w.value.skippedImages ?? 0) + w.value.tables + w.value.fakeHeadings.length + w.value.unknownAcronyms.length) > 0)
const fake = computed(() => (w.value?.fakeHeadings ?? []).slice(0, 2).map((s) => `«${s}»`).join(', '))
</script>

<template>
  <div class="max-w-2xl">
    <h1 class="text-xl font-semibold tracking-tight">Nạp file Word</h1>
    <p class="text-sm text-muted-foreground">
      Sano đọc mục lục từ kiểu Heading 1 / Heading 2 trong file.
      <a :href="DOCS + '/tao-sach-dau-tien#chuan-bi-file'" target="_blank" rel="noopener" class="text-primary hover:underline">Cách chuẩn bị file để đọc hay nhất</a>
    </p>

    <template v-if="!state.file">
      <button class="mt-5 w-full h-56 rounded-xl border-2 border-dashed border-border grid place-items-center hover:border-primary/50 hover:bg-primary/5" :disabled="picking" @click="pick">
        <span class="text-center">
          <Upload class="w-8 h-8 mx-auto text-muted-foreground" />
          <span class="block mt-3 font-medium">Kéo file .docx vào đây</span>
          <span class="block text-sm text-muted-foreground">hoặc bấm để chọn file</span>
        </span>
      </button>
      <p v-if="state.fileError" class="mt-3 text-sm text-destructive">{{ state.fileError }}</p>
      <div class="mt-3 rounded-lg border border-border px-4 py-3 text-sm">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <span class="text-muted-foreground flex-1 min-w-[16rem]">Chưa có file Word? Tải file mẫu có sẵn mục lục và hướng dẫn để làm theo, hoặc thử ngay với file mẫu.</span>
          <div class="flex gap-2">
            <Button variant="outline" size="sm" :disabled="picking" @click="downloadSample"><Download class="w-4 h-4" /> Tải file Word mẫu</Button>
            <Button variant="outline" size="sm" :disabled="picking" @click="useSample"><Sparkles class="w-4 h-4" /> Thử với tài liệu mẫu</Button>
          </div>
        </div>
        <div v-if="savedSample" class="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-border pt-3">
          <p class="flex items-start gap-1.5 text-xs text-muted-foreground flex-1 min-w-[16rem]">
            <CheckCircle2 class="w-3.5 h-3.5 mt-px text-rag-green shrink-0" />
            <span>Đã lưu <span class="text-foreground font-medium">{{ savedName }}</span> vào thư mục {{ savedDir }}. Mở bằng Word, thay nội dung của bạn rồi nạp vào đây, hoặc nạp luôn để thử.</span>
          </p>
          <Button size="sm" :disabled="picking" @click="useSaved"><Upload class="w-4 h-4" /> Nạp file này</Button>
        </div>
      </div>
      <p class="mt-3 flex items-start gap-1.5 text-xs text-muted-foreground">
        <ShieldCheck class="w-3.5 h-3.5 mt-px shrink-0" />
        Chỉ dùng tài liệu của bạn hoặc tài liệu bạn có quyền sử dụng. File có mật khẩu hoặc khoá bảo vệ sẽ bị từ chối.
      </p>
    </template>

    <template v-else>
      <div class="mt-5 rounded-lg border border-border p-4 flex items-center gap-3">
        <FileText class="w-8 h-8 text-primary shrink-0" />
        <div class="flex-1 min-w-0">
          <p class="font-medium truncate" :title="state.file.path">{{ state.file.name }}</p>
          <p v-if="state.loading" class="text-xs text-muted-foreground flex items-center gap-1.5"><Loader2 class="w-3 h-3 animate-spin" /> Đang đọc mục lục…</p>
          <p v-else-if="state.outline" class="text-xs text-muted-foreground">
            {{ fmtSize(state.file.size) }} · {{ chapterCount }} chương · {{ state.outline.sections }} tiểu mục · {{ state.outline.chars.toLocaleString('vi-VN') }} ký tự
          </p>
        </div>
        <Button variant="ghost" size="sm" @click="clearFile"><X class="w-4 h-4" /> Chọn file khác</Button>
      </div>

      <template v-if="state.outline">
        <div v-if="hasWarnings" class="mt-4 rounded-lg border border-rag-amber/40 bg-rag-amber/10 p-4 text-sm">
          <p class="font-medium flex items-center gap-2 text-rag-amber"><AlertTriangle class="w-4 h-4" /> Có phần sẽ không được đọc trọn vẹn</p>
          <ul class="mt-2 space-y-1 text-foreground/80 list-disc pl-5">
            <li v-if="w!.tables">{{ w!.tables }} bảng — nội dung bảng được đọc phẳng từng ô, mất hàng/cột</li>
            <li v-if="w!.images">{{ w!.images }} hình — không có lời tả, người nghe sẽ không biết nội dung hình</li>
            <li v-if="w!.skippedImages">{{ w!.skippedImages }} hình quá lớn hoặc vượt giới hạn số hình — đã bỏ qua, không trích ra</li>
            <li v-if="w!.fakeHeadings.length">{{ w!.fakeHeadings.length }} đoạn chữ to đậm có vẻ là tiêu đề nhưng không dùng kiểu Heading ({{ fake }})</li>
            <li v-if="w!.unknownAcronyms.length">{{ w!.unknownAcronyms.length }} từ viết tắt chưa có cách đọc: {{ acronyms }} — bộ đọc có thể đọc sai</li>
          </ul>
          <p class="mt-3 text-foreground/80">
            Muốn đọc đủ: dán file vào ChatGPT, Gemini hoặc Claude cùng lời nhắc mẫu để biến bảng, hình thành lời văn, rồi nạp lại.
          </p>
          <Button variant="outline" size="sm" class="mt-3" @click="copyPrompt">
            <component :is="copied ? Check : Copy" class="w-4 h-4" /> {{ copied ? 'Đã sao chép' : 'Sao chép lời nhắc mẫu' }}
          </Button>
        </div>
        <div v-else class="mt-4 rounded-lg border border-rag-green/40 bg-rag-green/10 p-4 text-sm flex items-center gap-2">
          <CheckCircle2 class="w-4 h-4 text-rag-green shrink-0" /> Không thấy bảng, hình hay tiêu đề gõ tay — file sẵn sàng để đọc.
        </div>

        <div class="mt-5 grid grid-cols-2 gap-4">
          <label class="text-sm">Tên sách
            <input v-model="state.title" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" />
          </label>
          <label class="text-sm">Tác giả
            <input v-model="state.author" placeholder="Không bắt buộc" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" />
          </label>
        </div>
        <div class="mt-4 text-sm">
          <span>Danh mục <span class="text-muted-foreground">· không bắt buộc</span></span>
          <CategoryPicker v-model="state.category" :categories="categories" class="mt-1 w-1/2 pr-2" />
          <p class="mt-1.5 text-xs text-muted-foreground">Dùng để lọc trong Thư viện.</p>
        </div>
        <div class="mt-4 text-sm flex gap-3 w-1/2 pr-2">
          <div class="flex-1 min-w-0">
            <span>Bộ sách <span class="text-muted-foreground">· không bắt buộc</span></span>
            <CategoryPicker v-model="state.series" kind="series" :categories="seriesGroups" class="mt-1" />
          </div>
          <label class="w-24 block shrink-0" :class="!state.series && 'opacity-40'">Tập số
            <input :value="state.volume || ''" @input="state.volume = Math.floor(Number(($event.target as HTMLInputElement).value)) || 0" type="number" min="1" max="999" :disabled="!state.series" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" />
          </label>
        </div>
        <p v-if="state.series && taken.length" class="mt-1.5 text-xs text-muted-foreground">Bộ "{{ state.series }}" đang có tập {{ [...taken].sort((a, b) => a - b).join(', ') }}.</p>
        <div class="mt-4 flex items-center gap-4">
          <div class="h-24 w-[72px] rounded-md shadow-sm shrink-0 overflow-hidden">
            <img v-if="state.coverDataUrl" :src="state.coverDataUrl" alt="Ảnh bìa" class="h-full w-full object-cover" />
            <BookCover v-else :title="state.title" :author="state.author" class="h-full w-full rounded-md shadow-none" />
          </div>
          <div class="text-sm">
            <p class="font-medium">Ảnh bìa</p>
            <p class="text-xs text-muted-foreground">{{ state.coverPath ? state.coverPath.split(/[\\/]/).pop() : 'Chưa có ảnh — Sano tự tạo bìa theo tên sách' }}</p>
            <div class="mt-2 flex gap-2">
              <Button variant="outline" size="sm" @click="pickCover">Chọn ảnh bìa</Button>
              <Button v-if="state.coverPath" variant="ghost" size="sm" @click="state.coverPath = ''; state.coverDataUrl = ''">Bỏ ảnh</Button>
            </div>
            <p v-if="coverError" class="mt-1 text-xs text-destructive">{{ coverError }}</p>
          </div>
        </div>
      </template>
      <p v-if="state.fileError" class="mt-3 text-sm text-destructive">{{ state.fileError }}</p>
    </template>
  </div>
</template>
