<script setup lang="ts">
// B1 Nạp file: hộp chọn file .docx / .epub (Wails) hoặc kéo thả → nạp thật: mục lục,
// số ký tự, cảnh báo lúc nạp (hình, bảng, tiêu đề gõ tay, viết tắt chưa có).
// Cấp 2/3 (bước Cách đọc): nạp file AI tạo, hoặc dán văn bản AI trả về (Gemini).
// Cấp 1 không nhắc tới AI.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { AlertTriangle, Check, CheckCircle2, ChevronLeft, ClipboardPaste, Copy, Download, FileText, Loader2, ShieldCheck, Sparkles, Upload, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import BookCover from '@/components/sano/BookCover.vue'
import { chooseCover, chooseDocx, copyText, describeDocx, errText, onFileDrop, pastedText, sampleDocx, saveSampleDocx } from '../../lib/backend'
import { DOCS } from '../../lib/mock'
import { promptFor } from '../../lib/prompt'
import { categoryCounts, seriesKey } from '../../lib/find'
import { addBookWord, clearFile, setFile, state } from '../../lib/store'
import { globalReading, loadGlobalDict } from '../../lib/dict'
import DictWordPopover from '../../components/DictWordPopover.vue'
import CategoryPicker from '../../components/CategoryPicker.vue'

const copied = ref(false)
const levelTitles = ['', 'Đọc nguyên văn', 'Làm mượt', 'Viết lại thành văn sách nói']
const viaAI = computed(() => state.level > 1)
// Gemini không tạo được file Word → mở sẵn ô dán văn bản.
if (!state.file && viaAI.value && state.aiTool === 'gemini') state.pasteMode = true
const pasted = ref('')
// Đếm nhanh để người dùng thấy Sano nhận ra bao nhiêu chương, mục trước khi nạp.
const pasteCount = computed(() => {
  const lines = pasted.value.split('\n').map((l) => l.trim())
  return { chapters: lines.filter((l) => /^#(?!#)/.test(l)).length, sections: lines.filter((l) => l.startsWith('##')).length }
})
async function usePasted() {
  picking.value = true
  state.fileError = ''
  try {
    await setFile(await pastedText(pasted.value))
    // AI hay chép tên sách IN HOA theo bản gốc; bộ đọc bỏ qua từ điển với dòng in hoa.
    if (state.title && state.title === state.title.toUpperCase() && state.title !== state.title.toLowerCase()) {
      const t = state.title.toLowerCase()
      state.title = t.charAt(0).toUpperCase() + t.slice(1)
    }
  } catch (e) {
    state.fileError = errText(e)
  } finally {
    picking.value = false
  }
}
function changeLevel() {
  state.step = 1
  state.levelScreen = 'choose'
}
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
  copied.value = await copyText(promptFor(2, state.aiTool || 'claude'))
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
// Từ viết tắt chưa có cách đọc (D12): bấm từng từ để dạy Sano đọc. Từ đã thêm (của cuốn
// hoặc từ điển chung) hiện ✓ kèm cách đọc.
onMounted(() => void loadGlobalDict())
const acronymList = computed(() => (w.value?.unknownAcronyms ?? []).slice(0, 20).map((a) => ({
  ...a, reading: state.bookDict[a.word] || globalReading(a.word),
})))
const acronymLeft = computed(() => acronymList.value.filter((a) => !a.reading).length)
const teaching = ref('')
const teachBusy = ref(false)
const teachError = ref('')
async function teach(reading: string, scope: 'book' | 'global') {
  teachBusy.value = true
  teachError.value = ''
  try {
    await addBookWord(teaching.value, reading, scope, false)
    teaching.value = ''
  } catch (e) {
    teachError.value = errText(e)
  } finally {
    teachBusy.value = false
  }
}
function closeTeach(e: MouseEvent) {
  if (teaching.value && !(e.target as HTMLElement).closest('[data-teach]')) teaching.value = ''
}
onMounted(() => document.addEventListener('mousedown', closeTeach))
onBeforeUnmount(() => document.removeEventListener('mousedown', closeTeach))
const hasWarnings = computed(() => !!w.value && (w.value.images + (w.value.skippedImages ?? 0) + w.value.tables + w.value.fakeHeadings.length + w.value.unknownAcronyms.length) > 0)
// Bảng, hình, tiêu đề gõ tay còn sót (khác chữ viết tắt: AI không cần làm lại).
const layoutLeft = computed(() => !!w.value && (w.value.images + (w.value.skippedImages ?? 0) + w.value.tables + w.value.fakeHeadings.length) > 0)
const fake = computed(() => (w.value?.fakeHeadings ?? []).slice(0, 2).map((s) => `«${s}»`).join(', '))
</script>

<template>
  <div class="max-w-2xl">
    <h1 class="text-xl font-semibold tracking-tight">{{ viaAI ? 'Nạp file AI tạo' : 'Nạp file Word' }}</h1>
    <p class="text-sm text-muted-foreground">
      Cấp {{ state.level }} · {{ levelTitles[state.level] }} <button class="text-primary hover:underline ml-1" @click="changeLevel">Đổi</button>
      <template v-if="!viaAI"> · Sano đọc mục lục từ kiểu Heading 1 / Heading 2 trong file Word, hoặc mục lục có sẵn của sách EPUB.
        <a :href="DOCS + '/tao-sach-dau-tien#chuan-bi-file'" target="_blank" rel="noopener" class="text-primary hover:underline">Cách chuẩn bị file để đọc hay nhất</a></template>
    </p>

    <template v-if="!state.file && state.pasteMode">
      <button class="mt-4 text-sm text-muted-foreground hover:text-foreground flex items-center gap-1" @click="state.pasteMode = false"><ChevronLeft class="w-4 h-4" /> Nạp file Word thay vì dán</button>
      <textarea v-model="pasted" aria-label="Văn bản AI trả về" class="mt-3 w-full h-56 rounded-lg border border-input bg-background p-3 text-sm font-mono leading-relaxed" spellcheck="false"
        placeholder="% Tên sách&#10;# Chương 1. Tên chương&#10;## Tên mục&#10;Nội dung mục…"></textarea>
      <p class="mt-2 text-xs text-muted-foreground">Dán nguyên kết quả AI trả về. Dòng <code class="font-mono">#</code> là chương, <code class="font-mono">##</code> là mục, <code class="font-mono">%</code> là tên sách (nếu có).</p>
      <div class="mt-3 flex items-center justify-between gap-3">
        <span class="text-sm text-muted-foreground">{{ pasted.trim() ? (pasteCount.chapters ? `Nhận ra ${pasteCount.chapters} chương · ${pasteCount.sections} mục` : 'Chưa thấy dòng chương nào (dòng bắt đầu bằng #)') : '' }}</span>
        <Button :disabled="picking || !pasteCount.chapters" @click="usePasted"><ClipboardPaste class="w-4 h-4" /> Nạp văn bản này</Button>
      </div>
      <p v-if="state.fileError" class="mt-3 text-sm text-destructive">{{ state.fileError }}</p>
    </template>

    <template v-else-if="!state.file">
      <button class="mt-5 w-full h-56 rounded-xl border-2 border-dashed border-border grid place-items-center hover:border-primary/50 hover:bg-primary/5" :disabled="picking" @click="pick">
        <span class="text-center">
          <Upload class="w-8 h-8 mx-auto text-muted-foreground" />
          <span class="block mt-3 font-medium">{{ viaAI ? 'Kéo file Word AI tạo vào đây' : 'Kéo file .docx hoặc .epub vào đây' }}</span>
          <span class="block text-sm text-muted-foreground">hoặc bấm để chọn file</span>
        </span>
      </button>
      <p v-if="state.fileError" class="mt-3 text-sm text-destructive">{{ state.fileError }}</p>
      <button v-if="viaAI" class="mt-3 text-sm text-muted-foreground hover:text-foreground flex items-center gap-1.5" @click="state.pasteMode = true">
        <ClipboardPaste class="w-4 h-4" /> AI không tạo được file Word? <span class="text-primary">Dán văn bản AI trả về</span>
      </button>
      <div v-else class="mt-3 rounded-lg border border-border px-4 py-3 text-sm">
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
        <Button variant="ghost" size="sm" @click="clearFile"><X class="w-4 h-4" /> {{ state.pasteMode ? 'Dán lại' : 'Chọn file khác' }}</Button>
      </div>

      <template v-if="state.outline">
        <div v-if="hasWarnings" class="mt-4 rounded-lg border border-rag-amber/40 bg-rag-amber/10 p-4 text-sm">
          <p class="font-medium flex items-center gap-2 text-rag-amber"><AlertTriangle class="w-4 h-4" /> Có phần sẽ không được đọc trọn vẹn</p>
          <ul class="mt-2 space-y-1 text-foreground/80 list-disc pl-5">
            <li v-if="w!.tables">{{ w!.tables }} bảng — nội dung bảng được đọc phẳng từng ô, mất hàng/cột</li>
            <li v-if="w!.images">{{ w!.images }} hình — không có lời tả, người nghe sẽ không biết nội dung hình</li>
            <li v-if="w!.skippedImages">{{ w!.skippedImages }} hình quá lớn hoặc vượt giới hạn số hình — đã bỏ qua, không trích ra</li>
            <li v-if="w!.fakeHeadings.length">{{ w!.fakeHeadings.length }} đoạn chữ to đậm có vẻ là tiêu đề nhưng không dùng kiểu Heading ({{ fake }})</li>
          </ul>
          <!-- Từ chưa có cách đọc: từng từ là một nút, bấm để dạy Sano đọc (D12) -->
          <div v-if="acronymList.length" class="mt-3 border-t border-rag-amber/30 pt-3" data-teach>
            <p class="text-foreground/80">
              <template v-if="acronymLeft"><b>{{ acronymLeft }} từ viết tắt chưa có cách đọc</b>, bộ đọc có thể đọc sai. Bấm từng từ để dạy Sano cách đọc:</template>
              <template v-else><b>Đã có cách đọc cho các từ viết tắt.</b> Bấm một từ để sửa lại.</template>
            </p>
            <div class="mt-2 flex flex-wrap gap-2 relative">
              <button v-for="a in acronymList" :key="a.word" type="button" class="h-8 px-3 rounded-full border text-sm flex items-center gap-1.5 bg-background"
                :class="teaching === a.word ? 'border-primary ring-2 ring-primary/30' : a.reading ? 'border-rag-green/50' : 'border-border hover:border-primary/50'"
                @click="teaching = teaching === a.word ? '' : a.word; teachError = ''">
                <Check v-if="a.reading" class="w-3.5 h-3.5 text-rag-green" /><b>{{ a.word }}</b>
                <span class="text-xs text-muted-foreground">{{ a.reading ? '→ ' + a.reading : a.count + ' lần' }}</span>
              </button>
              <span v-if="acronymLeft" class="h-8 px-1 text-sm text-muted-foreground flex items-center">Bỏ qua cũng được, sửa sau ở bước Nghe thử</span>
              <DictWordPopover v-if="teaching" :key="teaching" class="absolute top-10 left-0" :word="teaching" :voice="state.voice"
                :reading="state.bookDict[teaching] || globalReading(teaching)" :count="acronymList.find((a) => a.word === teaching)?.count ?? null"
                :busy="teachBusy" :error="teachError" @add="teach" @cancel="teaching = ''" />
            </div>
          </div>
          <template v-if="viaAI && layoutLeft">
            <p class="mt-3 text-foreground/80">AI còn để sót bảng, hình hoặc tiêu đề gõ tay. Gửi lại file cho AI cùng prompt làm mượt rồi nạp lại, hoặc cứ tiếp tục nếu chấp nhận được.</p>
            <Button variant="outline" size="sm" class="mt-3" @click="copyPrompt">
              <component :is="copied ? Check : Copy" class="w-4 h-4" /> {{ copied ? 'Đã sao chép' : 'Sao chép prompt làm mượt' }}
            </Button>
          </template>
          <template v-else-if="!viaAI && layoutLeft">
            <p class="mt-3 text-foreground/80">Muốn đọc đủ: đổi sang cấp 2 Làm mượt ở bước Cách đọc để biến bảng, hình thành lời văn, rồi nạp lại.</p>
            <Button variant="outline" size="sm" class="mt-3" @click="changeLevel">Đổi cách làm</Button>
          </template>
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
        <div class="mt-4 grid grid-cols-2 gap-4 text-sm">
          <div>
            <span>Danh mục <span class="text-muted-foreground">· không bắt buộc</span></span>
            <CategoryPicker v-model="state.category" :categories="categories" class="mt-1" />
            <p class="mt-1.5 text-xs text-muted-foreground">Dùng để lọc trong Thư viện.</p>
          </div>
          <div>
            <div class="flex gap-3">
              <div class="flex-1 min-w-0">
                <span>Bộ sách <span class="text-muted-foreground">· không bắt buộc</span></span>
                <CategoryPicker v-model="state.series" kind="series" :categories="seriesGroups" class="mt-1" />
              </div>
              <label class="w-20 block shrink-0" :class="!state.series && 'opacity-40'">Tập số
                <input :value="state.volume || ''" @input="state.volume = Math.floor(Number(($event.target as HTMLInputElement).value)) || 0" type="number" min="1" max="999" :disabled="!state.series" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" />
              </label>
            </div>
            <p class="mt-1.5 text-xs text-muted-foreground">{{ state.series && taken.length ? `Bộ "${state.series}" đang có tập ${[...taken].sort((a, b) => a - b).join(', ')}.` : 'Sách nhiều tập thì gom thành một bộ.' }}</p>
          </div>
        </div>
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
