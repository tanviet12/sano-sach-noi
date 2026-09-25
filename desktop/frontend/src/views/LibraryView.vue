<script setup lang="ts">
// Thư viện (wireframe D2 — thư viện dễ tìm): các cuốn đọc thật từ ~/Sano/Sach.
// Tìm theo tên/tác giả (không phân biệt dấu), sắp xếp (nhớ lựa chọn), lọc theo danh
// mục (chỉ hiện khi có từ 2 danh mục), hàng "Nghe tiếp", menu ⋯ trên bìa.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  Check, ChevronDown, FileArchive, FilePlus2, FolderOpen, Mic, MoreHorizontal, Pencil, Play, Search, Trash2, Upload, X,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import BookCover from '@/components/sano/BookCover.vue'
import { chooseBookZip, deleteBook, errText, onFileDrop, openBookFolder, openLibraryFolder, type LibraryBook } from '../lib/backend'
import {
  ago, categoryCounts, isListening, loadSort, matches, mergeCategory, saveSort, sortBooks, SORTS, type ShelfBook, type SortKey,
} from '../lib/find'
import { fmtLong, loadPosition } from '../lib/position'
import { go, openBook, refreshLibrary, state } from '../lib/store'
import { forgetBook } from '../lib/player'
import EditBookDialog from '../components/EditBookDialog.vue'
import ImportDialog from '../components/ImportDialog.vue'

onMounted(() => void refreshLibrary())

const categories = computed(() => categoryCounts(state.library?.books ?? []))
const books = computed<ShelfBook[]>(() =>
  (state.library?.books ?? []).map((b) => {
    const pos = loadPosition(b.slug)
    // cách viết danh mục khác hoa thường → hiện theo cách viết chung
    const category = mergeCategory(b.category ?? '', categories.value.map(([c]) => c))
    return { ...b, category, progress: Math.round(pos?.pct ?? 0), listenedAt: pos?.at ?? 0 }
  }),
)
const totalSec = computed(() => books.value.reduce((n, b) => n + b.durationSec, 0))
const dir = computed(() => (state.library?.dir ?? '~/Sano/Sach').replace(/^\/Users\/[^/]+|^\/home\/[^/]+/, '~'))

const uncategorized = computed(() => books.value.filter((b) => !b.category).length)
const listening = computed(() => books.value.filter(isListening))
const showChips = computed(() => categories.value.length >= 2)

// ── Tìm, lọc, sắp xếp ─────────────────────────────────────────────────────
const query = ref('')
const filter = ref<string>('all') // 'all' | 'listening' | '__none' | <danh mục>
const sort = ref<SortKey>(loadSort())
const sortOpen = ref(false)
const sortLabel = computed(() => SORTS.find((s) => s.key === sort.value)?.label ?? '')

function setSort(k: SortKey) {
  sort.value = k
  saveSort(k)
  sortOpen.value = false
}

// Danh mục đang lọc không còn (sửa/xoá sách) hoặc nút lọc bị ẩn → về Tất cả.
watch([showChips, categories], () => {
  if (!showChips.value) filter.value = 'all'
  else if (!['all', 'listening', '__none'].includes(filter.value) && !categories.value.some(([c]) => c === filter.value)) filter.value = 'all'
})

const chips = computed(() => [
  { key: 'all', label: `Tất cả · ${books.value.length}` },
  { key: 'listening', label: `Đang nghe · ${listening.value.length}` },
  ...categories.value.map(([c, n]) => ({ key: c, label: `${c} · ${n}` })),
  ...(uncategorized.value ? [{ key: '__none', label: `Chưa phân loại · ${uncategorized.value}` }] : []),
])

const shown = computed(() =>
  sortBooks(
    books.value.filter((b) => {
      if (!matches(b, query.value)) return false
      if (filter.value === 'listening') return isListening(b)
      if (filter.value === '__none') return !b.category
      if (filter.value !== 'all') return b.category === filter.value
      return true
    }),
    sort.value,
  ),
)
const continueList = computed(() => [...listening.value].sort((a, b) => b.listenedAt - a.listenedAt).slice(0, 3))
const showContinue = computed(() => filter.value === 'all' && !query.value.trim() && continueList.value.length > 0)

function clearSearch() {
  query.value = ''
  filter.value = 'all'
}

// ── Menu ⋯ trên bìa ───────────────────────────────────────────────────────
const menuFor = ref<string | null>(null)
const editing = ref<LibraryBook | null>(null)
const actionError = ref('')

function edit(b: LibraryBook) {
  menuFor.value = null
  editing.value = b
}

async function onSaved() {
  editing.value = null
  await refreshLibrary()
}

async function openFolder(b: LibraryBook) {
  menuFor.value = null
  actionError.value = ''
  try {
    await openBookFolder(b.slug)
  } catch (e) {
    actionError.value = errText(e)
  }
}

async function trash(b: LibraryBook) {
  menuFor.value = null
  actionError.value = ''
  try {
    if (await deleteBook(b.slug)) {
      forgetBook(b.slug) // đang phát cuốn này → dừng
      await refreshLibrary()
    }
  } catch (e) {
    actionError.value = errText(e)
  }
}

// ── Nhập sách từ gói zip (wireframe D6) ─────────────────────────────────
const importPath = ref('')
const imported = ref<{ slug: string; title: string } | null>(null)
const dragging = ref(false)
let toastTimer = 0

async function pickZip() {
  actionError.value = ''
  try {
    const p = await chooseBookZip()
    if (p) importPath.value = p
  } catch (e) {
    actionError.value = errText(e)
  }
}
async function onImported(slug: string) {
  importPath.value = ''
  await refreshLibrary()
  query.value = ''
  filter.value = 'all'
  const b = books.value.find((x) => x.slug === slug)
  imported.value = { slug, title: b?.title ?? slug }
  window.clearTimeout(toastTimer)
  toastTimer = window.setTimeout(() => (imported.value = null), 10000)
}

// Kéo thả: chỉ nhận một file .zip; file khác báo lỗi nhẹ (Word thì vào Tạo sách mới).
let offDrop: (() => void) | null = null
let dragDepth = 0
const hasFiles = (e: DragEvent) => !!e.dataTransfer && [...e.dataTransfer.types].includes('Files')
function onDragEnter(e: DragEvent) {
  if (!hasFiles(e)) return
  dragDepth++
  dragging.value = true
}
function onDragLeave() {
  dragDepth = Math.max(0, dragDepth - 1)
  if (!dragDepth) dragging.value = false
}
function onDropHtml() {
  dragDepth = 0
  dragging.value = false
}
onMounted(() => {
  offDrop = onFileDrop((paths) => {
    dragging.value = false
    dragDepth = 0
    if (importPath.value || editing.value) return
    const zip = paths.find((p) => /\.zip$/i.test(p))
    if (zip) importPath.value = zip
    else if (paths.length) actionError.value = /\.docx$/i.test(paths[0]) ? 'File Word thì vào Tạo sách mới. Ở đây chỉ nhập gói sách .zip.' : 'Chỉ nhập được gói sách .zip.'
  })
  window.addEventListener('dragenter', onDragEnter)
  window.addEventListener('dragleave', onDragLeave)
  window.addEventListener('drop', onDropHtml)
})
onBeforeUnmount(() => {
  offDrop?.()
  window.removeEventListener('dragenter', onDragEnter)
  window.removeEventListener('dragleave', onDragLeave)
  window.removeEventListener('drop', onDropHtml)
  window.clearTimeout(toastTimer)
})

// Bấm ra ngoài hoặc Esc thì đóng menu ⋯ và menu sắp xếp.
function closeMenus(e: Event) {
  if (e instanceof KeyboardEvent) {
    if (e.key !== 'Escape' || editing.value || importPath.value) return
    menuFor.value = null
    sortOpen.value = false
    return
  }
  const t = e.target as HTMLElement
  if (!t.closest('[data-book-menu]')) menuFor.value = null
  if (!t.closest('[data-sort-menu]')) sortOpen.value = false
}
document.addEventListener('mousedown', closeMenus)
document.addEventListener('keydown', closeMenus)
onBeforeUnmount(() => {
  document.removeEventListener('mousedown', closeMenus)
  document.removeEventListener('keydown', closeMenus)
})

const progressText = (p: number) => (p >= 99 ? 'Đã nghe xong' : p === 0 ? 'Chưa nghe' : `Đã nghe ${p}%`)
</script>

<template>
  <section class="relative flex-1 min-h-0 flex flex-col">
    <div class="px-6 pt-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-xl font-semibold tracking-tight">Thư viện</h1>
          <p class="text-sm text-muted-foreground">
            {{ books.length }} cuốn<template v-if="books.length"> · {{ fmtLong(totalSec) }}</template> · lưu ở
            <button class="hover:text-foreground hover:underline" :title="state.library?.dir" @click="openLibraryFolder().catch(() => {})">{{ dir }}</button>
          </p>
        </div>
        <div class="flex gap-2">
          <Button variant="outline" title="Nhập gói sách (.zip) người khác gửi hoặc bản sao lưu" @click="pickZip"><Upload class="w-4 h-4" /> Nhập sách</Button>
          <Button @click="go('create')"><FilePlus2 class="w-4 h-4" /> Tạo sách mới</Button>
        </div>
      </div>
      <p v-if="state.libraryError" class="mt-4 text-sm text-destructive">{{ state.libraryError }}</p>
      <p v-if="actionError" class="mt-4 text-sm text-destructive">{{ actionError }}</p>

      <template v-if="books.length">
        <!-- Thanh công cụ: tìm + sắp xếp -->
        <div class="mt-4 flex items-center gap-2">
          <div class="flex-1 flex items-center h-9 rounded-md border border-input bg-background px-3 focus-within:border-ring">
            <Search class="w-4 h-4 text-muted-foreground mr-2 shrink-0" />
            <input v-model="query" class="flex-1 min-w-0 bg-transparent outline-none text-sm" placeholder="Tìm theo tên sách, tác giả hoặc giọng đọc…" aria-label="Tìm sách" />
            <button v-if="query" class="text-muted-foreground hover:text-foreground" aria-label="Xoá chữ đang tìm" @click="query = ''"><X class="w-4 h-4" /></button>
          </div>
          <div class="relative" data-sort-menu>
            <button class="h-9 px-3 rounded-md border border-input bg-background text-sm flex items-center gap-1.5 hover:bg-muted/50" aria-haspopup="listbox" :aria-expanded="sortOpen" @click="sortOpen = !sortOpen">
              <span class="text-muted-foreground">Sắp xếp:</span> {{ sortLabel }} <ChevronDown class="w-4 h-4 text-muted-foreground" />
            </button>
            <div v-if="sortOpen" role="listbox" aria-label="Sắp xếp" class="absolute right-0 top-full mt-1 w-44 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-20">
              <button v-for="s in SORTS" :key="s.key" role="option" :aria-selected="s.key === sort"
                class="w-full flex items-center justify-between px-3 py-1.5 text-sm text-left hover:bg-muted" :class="s.key === sort && 'text-primary font-medium'" @click="setSort(s.key)">
                {{ s.label }} <Check v-if="s.key === sort" class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <!-- Nút lọc: chỉ hiện khi có từ 2 danh mục -->
        <div v-if="showChips" class="mt-3 flex flex-wrap gap-1.5">
          <button v-for="c in chips" :key="c.key" class="h-7 px-3 rounded-full text-xs border"
            :class="filter === c.key ? 'border-primary bg-primary/10 text-primary font-medium' : 'border-border text-muted-foreground hover:text-foreground'"
            :aria-pressed="filter === c.key" @click="filter = c.key">{{ c.label }}</button>
        </div>
      </template>
    </div>

    <div class="flex-1 overflow-auto px-6 pb-6">
      <!-- Thư viện trống -->
      <div v-if="state.library && !books.length" class="mt-10 rounded-xl border-2 border-dashed border-border p-10 text-center">
        <p class="font-medium">Chưa có cuốn nào</p>
        <p class="mt-1 text-sm text-muted-foreground">Tạo sách nói từ file Word, hoặc nhập gói sách (.zip) người khác gửi cho bạn.</p>
        <div class="mt-4 flex justify-center gap-2">
          <Button @click="go('create')"><FilePlus2 class="w-4 h-4" /> Tạo sách mới</Button>
          <Button variant="outline" @click="pickZip"><Upload class="w-4 h-4" /> Nhập sách</Button>
          <Button variant="ghost" @click="openLibraryFolder().catch(() => {})"><FolderOpen class="w-4 h-4" /> Mở thư mục</Button>
        </div>
      </div>

      <template v-else-if="books.length">
        <!-- Nghe tiếp -->
        <div v-if="showContinue" class="mt-5">
          <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Nghe tiếp</h2>
          <div class="mt-2 grid grid-cols-3 gap-3">
            <button v-for="b in continueList" :key="b.slug" class="flex items-center gap-3 rounded-lg border border-border p-2.5 text-left hover:bg-muted/50"
              :aria-label="`Nghe tiếp ${b.title}`" @click="openBook(b.slug, true)">
              <div class="h-14 w-[42px] shrink-0 rounded shadow-sm overflow-hidden">
                <img v-if="b.coverUrl" :src="b.coverUrl" alt="" class="h-full w-full object-cover" />
                <BookCover v-else :title="b.title" :author="b.author" class="h-full w-full rounded shadow-none" />
              </div>
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium truncate">{{ b.title }}</p>
                <p class="text-xs text-muted-foreground truncate">Đã nghe {{ b.progress }}%<template v-if="b.listenedAt"> · {{ ago(b.listenedAt) }}</template></p>
                <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: b.progress + '%' }"></div></div>
              </div>
              <span class="h-8 w-8 grid place-items-center rounded-full bg-primary text-primary-foreground shrink-0"><Play class="w-4 h-4 ml-0.5" /></span>
            </button>
          </div>
          <h2 class="mt-6 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Tất cả sách</h2>
        </div>

        <!-- Lưới sách -->
        <div v-if="shown.length" class="mt-3 grid grid-cols-5 gap-4">
          <div v-for="b in shown" :key="b.slug" class="text-left group relative rounded-lg" :class="imported?.slug === b.slug && 'ring-2 ring-primary ring-offset-4 ring-offset-background'">
            <button class="block w-full aspect-[3/4] rounded-lg shadow-md group-hover:shadow-xl group-hover:-translate-y-0.5 transition overflow-hidden" :aria-label="`Nghe ${b.title}`" @click="openBook(b.slug)">
              <img v-if="b.coverUrl" :src="b.coverUrl" :alt="b.title" class="h-full w-full object-cover" />
              <BookCover v-else :title="b.title" :author="b.author" class="h-full w-full shadow-none" />
            </button>
            <!-- Menu ⋯ trên bìa -->
            <div data-book-menu>
              <button class="absolute top-1.5 right-1.5 h-7 w-7 grid place-items-center rounded-full bg-black/40 text-white opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
                :class="menuFor === b.slug && 'opacity-100'" aria-label="Thao tác với sách" aria-haspopup="menu" :aria-expanded="menuFor === b.slug"
                @click="menuFor = menuFor === b.slug ? null : b.slug"><MoreHorizontal class="w-4 h-4" /></button>
              <div v-if="menuFor === b.slug" role="menu" class="absolute top-10 right-1.5 w-44 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-20 text-sm">
                <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted" @click="edit(b)"><Pencil class="w-4 h-4" /> Sửa thông tin</button>
                <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted" @click="openFolder(b)"><FolderOpen class="w-4 h-4" /> Mở thư mục</button>
                <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted text-destructive" @click="trash(b)"><Trash2 class="w-4 h-4" /> Chuyển vào Thùng rác</button>
              </div>
            </div>
            <p class="mt-2 text-sm font-medium truncate" :title="b.title">{{ b.title }}</p>
            <p class="text-xs text-muted-foreground truncate">{{ b.author ? `${b.author} · ` : '' }}{{ fmtLong(b.durationSec) }}</p>
            <p v-if="b.voice" class="text-xs text-muted-foreground truncate flex items-center gap-1"><Mic class="w-3 h-3 shrink-0" /> Giọng {{ b.voice }}</p>
            <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: b.progress + '%' }"></div></div>
            <p class="mt-1 text-[11px] text-muted-foreground flex items-center gap-1.5">
              <span class="shrink-0">{{ progressText(b.progress) }}</span>
              <template v-if="b.category && showChips && filter === 'all'"><span>·</span><span class="truncate">{{ b.category }}</span></template>
            </p>
          </div>
        </div>

        <!-- Tìm không thấy -->
        <div v-else class="mt-16 flex flex-col items-center text-center">
          <Search class="w-8 h-8 text-muted-foreground" />
          <p class="mt-3 font-medium">{{ query.trim() ? `Không có cuốn nào khớp "${query.trim()}"` : 'Không có cuốn nào trong bộ lọc này' }}</p>
          <p class="text-sm text-muted-foreground">Thử tên khác, tên tác giả, hoặc bỏ bộ lọc.</p>
          <Button variant="outline" size="sm" class="mt-4" @click="clearSearch">Xoá tìm kiếm</Button>
        </div>
      </template>
    </div>

    <EditBookDialog v-if="editing" :book="editing" :categories="categories" @close="editing = null" @saved="onSaved" />
    <ImportDialog v-if="importPath" :path="importPath" @close="importPath = ''" @imported="onImported" />

    <div v-if="dragging && !importPath" class="pointer-events-none absolute inset-3 rounded-xl border-2 border-dashed border-primary bg-primary/5 grid place-items-center z-30">
      <div class="text-center">
        <FileArchive class="w-10 h-10 mx-auto text-primary" />
        <p class="mt-3 font-medium">Thả gói sách (.zip) để nhập vào thư viện</p>
        <p class="mt-1 text-sm text-muted-foreground">File .docx thì vào Tạo sách mới</p>
      </div>
    </div>

    <div v-if="imported" role="status" class="absolute bottom-5 left-1/2 -translate-x-1/2 flex items-center gap-3 rounded-lg border border-border bg-background shadow-lg pl-4 pr-2 py-2 text-sm z-20">
      <Check class="w-4 h-4 text-rag-green" /> <span class="max-w-72 truncate">Đã nhập “{{ imported.title }}”</span>
      <Button size="sm" @click="openBook(imported.slug, true); imported = null"><Play class="w-4 h-4" /> Nghe ngay</Button>
      <button aria-label="Đóng thông báo" class="h-8 w-8 grid place-items-center rounded-md text-muted-foreground hover:bg-muted" @click="imported = null"><X class="w-4 h-4" /></button>
    </div>
  </section>
</template>
