<script setup lang="ts">
// Thư viện (wireframe D2 — thư viện dễ tìm): các cuốn đọc thật từ ~/Sano/Sach.
// Tìm theo tên/tác giả (không phân biệt dấu), sắp xếp (nhớ lựa chọn), lọc theo danh
// mục (chỉ hiện khi có từ 2 danh mục), hàng "Nghe tiếp", menu ⋯ trên bìa.
// Wireframe D5: bộ sách gom các tập thành một thẻ (bấm mở trang bộ sách), "Tự sắp xếp"
// kéo thả thẻ để đổi chỗ, nút "Quản lý" đổi tên / xoá danh mục và bộ sách.
// Wireframe D6: xem dạng lưới hoặc danh sách (nút cạnh "Sắp xếp", nhớ lựa chọn).
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  Check, ChevronDown, ChevronLeft, FileArchive, FilePlus2, FolderOpen, GripVertical, Layers, Mic, MoreHorizontal, Pencil, Play, Search,
  Settings2, Trash2, Upload, X, ArrowUpDown, LayoutGrid, List, History,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import BookCover from '@/components/sano/BookCover.vue'
import {
  chooseBookZip, deleteBook, errText, libraryOrder, onFileDrop, openBookFolder, openLibraryFolder, setLibraryOrder, type LibraryBook,
} from '../lib/backend'
import {
  ago, categoryCounts, isListening, loadSort, loadView, saveView, matches, mergeCategory, moveItem, saveSort, seriesKey, shelfItems, sortShelf, SORTS,
  type ShelfBook, type ShelfItem, type ShelfView, type SortKey,
} from '../lib/find'
import { clearPositions, fmtLong, loadPosition } from '../lib/position'
import { go, openBook, refreshLibrary, state } from '../lib/store'
import { forgetBook } from '../lib/player'
import { openEdit } from '../lib/edit'
import ImportDialog from '../components/ImportDialog.vue'
import ManageShelfDialog from '../components/ManageShelfDialog.vue'

onMounted(() => void refreshLibrary())

const categories = computed(() => categoryCounts(state.library?.books ?? []))
// Tăng sau khi xoá lịch sử nghe: vị trí nghe đọc từ bộ nhớ trình duyệt, không tự báo đổi.
const historyTick = ref(0)
const books = computed<ShelfBook[]>(() =>
  (void historyTick.value, state.library?.books ?? []).map((b) => {
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

// Bộ sách: [tên, số tập].
const seriesGroups = computed<[string, number][]>(() => {
  const m = new Map<string, [string, number]>()
  for (const b of books.value) {
    if (!b.series) continue
    const k = seriesKey(b.series)
    const g = m.get(k) ?? [b.series, 0]
    g[1]++
    m.set(k, g)
  }
  return [...m.values()].sort((a, b) => a[0].localeCompare(b[0], 'vi'))
})
const showManage = computed(() => categories.value.length > 0 || seriesGroups.value.length > 0)
const manageTab = ref<'cat' | 'series' | null>(null)
async function onShelfChanged() {
  await refreshLibrary()
  order.value = await libraryOrder().catch(() => order.value)
}

// ── Tìm, lọc, sắp xếp ─────────────────────────────────────────────────────
const query = ref('')
const filter = ref<string>('all') // 'all' | 'listening' | '__none' | <danh mục>
const sort = ref<SortKey>(loadSort())
const sortOpen = ref(false)
const sortLabel = computed(() => SORTS.find((s) => s.key === sort.value)?.label ?? '')

const view = ref<ShelfView>(loadView())
function setView(v: ShelfView) {
  view.value = v
  saveView(v)
  menuFor.value = null
}
// Cuốn / bộ đang mở ở trình phát: hàng trong danh sách luôn hiện nút nghe đỏ.
const isPlaying = (i: ShelfItem) => !!state.playerSlug && (i.kind === 'book' ? i.book.slug === state.playerSlug : i.vols.some((v) => v.slug === state.playerSlug))
function playItem(i: ShelfItem) {
  if (Date.now() - dragEndedAt < 400) return
  openBook(i.kind === 'book' ? i.book.slug : i.current.slug, true)
}

function setSort(k: SortKey) {
  if (k === 'manual') return startArrange()
  sort.value = k
  saveSort(k)
  sortOpen.value = false
  doneArrange()
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

const bookMatch = (b: ShelfBook) => {
  if (!matches(b, query.value) && !(b.series && matches({ title: b.series, author: '' }, query.value))) return false
  if (filter.value === 'listening') return isListening(b)
  if (filter.value === '__none') return !b.category
  if (filter.value !== 'all') return b.category === filter.value
  return true
}
const allItems = computed(() => shelfItems(books.value))
// Thẻ bộ sách hiện khi có ít nhất một tập khớp tìm kiếm / bộ lọc.
const shown = computed(() =>
  sortShelf(allItems.value.filter((i) => (i.kind === 'book' ? bookMatch(i.book) : i.vols.some(bookMatch))), sort.value, order.value),
)

// ── Tự sắp xếp (kéo thả) ──────────────────────────────────────────────────
const order = ref<string[]>([])
const orderBefore = ref<string[] | null>(null) // thứ tự lúc vào chế độ tự sắp xếp → "Về thứ tự cũ"
const manual = computed(() => sort.value === 'manual')
// Đang kéo thả: chỉ bật khi bấm nút Sắp xếp (hoặc chọn "Tự sắp xếp"), bấm Xong là tắt;
// thứ tự vẫn giữ kiểu "Tự sắp xếp".
const arranging = ref(false)
const dragKey = ref<string | null>(null)
const overKey = ref<string | null>(null)
let dragEndedAt = 0
onMounted(async () => {
  order.value = await libraryOrder().catch(() => [])
})
function startArrange() {
  if (!manual.value) {
    // chưa từng tự sắp xếp → lấy thứ tự đang thấy làm gốc
    if (!order.value.length) void saveOrder(sortShelf(allItems.value, sort.value, order.value).map((i) => i.key))
    sort.value = 'manual'
    saveSort('manual')
  }
  sortOpen.value = false
  menuFor.value = null
  orderBefore.value = [...order.value]
  arranging.value = true
}
function doneArrange() {
  arranging.value = false
  orderBefore.value = null
}
function endDrag() {
  dragKey.value = overKey.value = null
  dragEndedAt = Date.now()
}

// Kéo bằng chuột (pointer), KHÔNG dùng kéo thả HTML5: trong app, Wails chặn sự kiện thả
// của WebView để nhận file (DisableWebViewDrop) nên kéo thả HTML5 không chạy.
const dragDelta = ref({ x: 0, y: 0 })
let press: { key: string; x: number; y: number; id: number } | null = null
function onPointerDown(e: PointerEvent, key: string) {
  if (!arranging.value || e.button !== 0) return
  press = { key, x: e.clientX, y: e.clientY, id: e.pointerId }
  window.addEventListener('pointermove', onPointerMove)
  window.addEventListener('pointerup', onPointerUp)
  window.addEventListener('pointercancel', onPointerCancel)
}
function onPointerMove(e: PointerEvent) {
  if (!press || e.pointerId !== press.id) return
  const dx = e.clientX - press.x
  const dy = e.clientY - press.y
  if (!dragKey.value) {
    if (Math.hypot(dx, dy) < 6) return // chưa đủ xa: vẫn là bấm
    dragKey.value = press.key
  }
  e.preventDefault()
  dragDelta.value = { x: dx, y: dy }
  // thẻ đang kéo có pointer-events: none → phần tử dưới con trỏ là thẻ đích
  const el = document.elementFromPoint(e.clientX, e.clientY)?.closest<HTMLElement>('[data-shelf-key]')
  const k = el?.dataset.shelfKey ?? null
  overKey.value = k && k !== dragKey.value ? k : null
}
function stopPointer() {
  press = null
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('pointerup', onPointerUp)
  window.removeEventListener('pointercancel', onPointerCancel)
}
function onPointerUp() {
  const target = overKey.value
  const was = !!dragKey.value
  stopPointer()
  dragDelta.value = { x: 0, y: 0 }
  if (was && target) onDrop(target)
  else if (was) endDrag()
}
function onPointerCancel() {
  stopPointer()
  dragDelta.value = { x: 0, y: 0 }
  endDrag()
}
onBeforeUnmount(stopPointer)
async function saveOrder(keys: string[]) {
  order.value = keys
  try {
    await setLibraryOrder(keys)
  } catch (e) {
    actionError.value = errText(e)
  }
}
function onDrop(target: string) {
  const from = dragKey.value
  endDrag()
  if (!from || from === target) return
  // đổi chỗ trên thứ tự của CẢ kệ (kể cả thẻ đang bị lọc ẩn)
  void saveOrder(moveItem(sortShelf(allItems.value, 'manual', order.value).map((i) => i.key), from, target))
}
const canRestore = computed(() => !!orderBefore.value && orderBefore.value.join('|') !== order.value.join('|'))
function restoreOrder() {
  if (orderBefore.value) void saveOrder([...orderBefore.value])
}

// ── Trang một bộ sách ─────────────────────────────────────────────────────
const openSeries = ref<string | null>(null) // khoá "s:..."
const seriesItem = computed(() => allItems.value.find((i): i is Extract<ShelfItem, { kind: 'series' }> => i.kind === 'series' && i.key === openSeries.value) ?? null)
watch(seriesItem, (it) => {
  if (openSeries.value && !it) openSeries.value = null // bộ vừa bị xoá / đổi tên
})
function openItem(i: ShelfItem) {
  // Vừa kéo xong thì không tính là bấm mở; không kéo thì vẫn bấm mở được như thường.
  if (Date.now() - dragEndedAt < 400) return
  if (i.kind === 'book') openBook(i.book.slug)
  else openSeries.value = i.key
}
const continueList = computed(() => [...listening.value].sort((a, b) => b.listenedAt - a.listenedAt).slice(0, 3))
const showContinue = computed(() => filter.value === 'all' && !query.value.trim() && continueList.value.length > 0)

// Xoá lịch sử nghe: mọi cuốn về "Chưa nghe", hàng "Nghe tiếp" trống. Sách giữ nguyên,
// vị trí cũ cất sang lịch sử (clearPositions) cho thống kê sau này.
const confirmClear = ref(false)
const listenedCount = computed(() => books.value.filter((b) => b.progress > 0 || b.listenedAt > 0).length)
function clearHistory() {
  clearPositions()
  historyTick.value++
  confirmClear.value = false
  if (filter.value === 'listening') filter.value = 'all'
}

function clearSearch() {
  query.value = ''
  filter.value = 'all'
}

// ── Menu ⋯ trên bìa ───────────────────────────────────────────────────────
const menuFor = ref<string | null>(null)
const actionError = ref('')

// "Sửa sách" (wireframe D11): mở màn Sửa sách ở tab Nội dung.
function edit(b: LibraryBook) {
  menuFor.value = null
  openEdit(b.slug)
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

// Kéo thả: chỉ nhận một file .zip; file khác báo lỗi nhẹ (Word thì vào Tạo sách nói).
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
    // kéo bìa để sắp xếp cũng bắn sự kiện thả của cửa sổ → bỏ qua
    if (importPath.value || arranging.value || dragKey.value || Date.now() - dragEndedAt < 1500) return
    const zip = paths.find((p) => /\.zip$/i.test(p))
    if (zip) importPath.value = zip
    else if (paths.length) actionError.value = /\.(docx|epub)$/i.test(paths[0]) ? 'File Word, EPUB thì vào Tạo sách nói. Ở đây chỉ nhập gói sách .zip.' : 'Chỉ nhập được gói sách .zip.'
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
    if (e.key !== 'Escape' || importPath.value || manageTab.value) return
    if (arranging.value) doneArrange()
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
    <!-- ─── Trang một bộ sách (wireframe D5, trạng thái 2) ─── -->
    <div v-if="seriesItem" class="flex-1 min-h-0 overflow-auto p-6">
      <button class="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground" @click="openSeries = null"><ChevronLeft class="w-4 h-4" /> Thư viện</button>
      <div class="mt-4 flex gap-6">
        <div class="relative w-36 shrink-0">
          <div class="absolute inset-0 translate-x-3 -translate-y-2 rounded-lg bg-muted-foreground/25"></div>
          <div class="absolute inset-0 translate-x-1.5 -translate-y-1 rounded-lg bg-muted-foreground/40"></div>
          <div class="relative aspect-[3/4] rounded-lg shadow-lg overflow-hidden">
            <BookCover :cover-image="seriesItem.thumbUrl || seriesItem.coverUrl" :title="seriesItem.name" :author="seriesItem.author" class="h-full w-full shadow-none" />
          </div>
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-xs font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5"><Layers class="w-3.5 h-3.5" /> Bộ sách · {{ seriesItem.vols.length }} tập</p>
          <h1 class="mt-1 text-2xl font-semibold tracking-tight">{{ seriesItem.name }}</h1>
          <p class="text-sm text-muted-foreground">{{ [seriesItem.author, fmtLong(seriesItem.durationSec), seriesItem.vols[0].category].filter(Boolean).join(' · ') }}</p>
          <div class="mt-4 flex items-center gap-2">
            <Button @click="openBook(seriesItem.current.slug, true)"><Play class="w-4 h-4" /> {{ seriesItem.progress === 0 ? 'Nghe' : 'Nghe tiếp' }} tập {{ seriesItem.current.volume }}</Button>
            <Button variant="outline" size="icon" aria-label="Đổi tên hoặc xoá bộ sách" title="Đổi tên hoặc xoá bộ sách" @click="manageTab = 'series'"><Pencil class="w-4 h-4" /></Button>
          </div>
        </div>
      </div>
      <div class="mt-6 rounded-lg border border-border divide-y divide-border">
        <div v-for="v in seriesItem.vols" :key="v.slug" class="relative flex items-center gap-4 px-4 py-3 hover:bg-muted/40">
          <span class="w-14 text-sm font-semibold text-muted-foreground shrink-0">Tập {{ v.volume }}</span>
          <div class="h-12 w-9 shrink-0 rounded shadow-sm overflow-hidden">
            <BookCover :cover-image="v.thumbUrl || v.coverUrl" :title="v.title" :author="v.author" class="h-full w-full rounded shadow-none" />
          </div>
          <button class="flex-1 min-w-0 text-left" @click="openBook(v.slug)">
            <p class="text-sm font-medium truncate">{{ v.title }}</p>
            <p class="text-xs text-muted-foreground">{{ fmtLong(v.durationSec) }} · {{ progressText(v.progress) }}</p>
            <div class="mt-1.5 h-1 w-48 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: v.progress + '%' }"></div></div>
          </button>
          <button class="h-8 w-8 grid place-items-center rounded-full shrink-0" :class="v.slug === seriesItem.current.slug ? 'bg-primary text-primary-foreground' : 'border border-border text-foreground hover:bg-muted'"
            :aria-label="`Nghe tập ${v.volume}`" @click="openBook(v.slug, true)"><Play class="w-4 h-4 ml-0.5" /></button>
          <div data-book-menu class="relative">
            <button class="h-8 w-8 grid place-items-center rounded-md text-muted-foreground hover:bg-muted" aria-label="Thao tác với tập này" aria-haspopup="menu" :aria-expanded="menuFor === v.slug"
              @click="menuFor = menuFor === v.slug ? null : v.slug"><MoreHorizontal class="w-4 h-4" /></button>
            <div v-if="menuFor === v.slug" role="menu" class="absolute top-9 right-0 w-44 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-20 text-sm">
              <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted" @click="edit(v)"><Pencil class="w-4 h-4" /> Sửa sách</button>
              <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted" @click="openFolder(v)"><FolderOpen class="w-4 h-4" /> Mở thư mục</button>
              <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted text-destructive" @click="trash(v)"><Trash2 class="w-4 h-4" /> Chuyển vào Thùng rác</button>
            </div>
          </div>
        </div>
      </div>
      <p class="mt-3 text-xs text-muted-foreground">Các tập luôn xếp theo số tập. Thêm tập mới: khi tạo sách, chọn bộ này ở ô "Bộ sách". Nghe hết một tập thì tự chuyển sang tập sau.</p>
    </div>

    <template v-else>
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
          <Button @click="go('create')"><FilePlus2 class="w-4 h-4" /> Tạo sách nói</Button>
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
            <div v-if="sortOpen" role="listbox" aria-label="Sắp xếp" class="absolute right-0 top-full mt-1 w-56 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-20">
              <button v-for="s in SORTS" :key="s.key" role="option" :aria-selected="s.key === sort"
                class="w-full flex items-center justify-between px-3 py-1.5 text-sm text-left hover:bg-muted" :class="[s.key === sort && 'text-primary font-medium', s.key === 'manual' && 'border-b border-border mb-1 pb-2']" @click="setSort(s.key)">
                <span>{{ s.label }}<span v-if="s.key === 'manual'" class="block text-[11px] font-normal text-muted-foreground">Kéo bìa sách để đổi chỗ</span></span>
                <Check v-if="s.key === sort" class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <!-- Nút lọc: chỉ hiện khi có từ 2 danh mục -->
        <div v-if="showChips || showManage" class="mt-3 flex flex-wrap items-center gap-1.5">
          <template v-if="showChips">
            <button v-for="c in chips" :key="c.key" class="h-7 px-3 rounded-full text-xs border"
              :class="filter === c.key ? 'border-primary bg-primary/10 text-primary font-medium' : 'border-border text-muted-foreground hover:text-foreground'"
              :aria-pressed="filter === c.key" @click="filter = c.key">{{ c.label }}</button>
          </template>
          <button v-if="showManage" class="h-7 px-2.5 rounded-full text-xs text-muted-foreground hover:text-foreground flex items-center gap-1" @click="manageTab = 'cat'">
            <Settings2 class="w-3.5 h-3.5" /> Quản lý danh mục, bộ sách
          </button>
        </div>

        <!-- Dải nhắc khi đang tự sắp xếp -->
        <div v-if="arranging" class="mt-3 flex items-center gap-3 rounded-lg bg-primary/5 border border-primary/20 pl-3 pr-2 py-1.5 text-sm">
          <GripVertical class="w-4 h-4 text-primary shrink-0" />
          <span class="flex-1">{{ view === 'list' ? 'Kéo hàng lên xuống để đổi chỗ.' : 'Kéo bìa sách để đổi chỗ.' }} Thứ tự được lưu lại, lần sau mở vẫn giữ nguyên.</span>
          <button v-if="canRestore" class="text-xs text-muted-foreground hover:text-foreground" @click="restoreOrder">Về thứ tự cũ</button>
          <Button size="sm" @click="doneArrange"><Check class="w-4 h-4" /> Xong</Button>
        </div>
      </template>
    </div>

    <div class="flex-1 overflow-auto px-6 pb-6">
      <!-- Thư viện trống -->
      <div v-if="state.library && !books.length" class="mt-10 rounded-xl border-2 border-dashed border-border p-10 text-center">
        <p class="font-medium">Chưa có cuốn nào</p>
        <p class="mt-1 text-sm text-muted-foreground">Tạo sách nói từ file Word, hoặc nhập gói sách (.zip) người khác gửi cho bạn.</p>
        <div class="mt-4 flex justify-center gap-2">
          <Button @click="go('create')"><FilePlus2 class="w-4 h-4" /> Tạo sách nói</Button>
          <Button variant="outline" @click="pickZip"><Upload class="w-4 h-4" /> Nhập sách</Button>
          <Button variant="ghost" @click="openLibraryFolder().catch(() => {})"><FolderOpen class="w-4 h-4" /> Mở thư mục</Button>
        </div>
      </div>

      <template v-else-if="books.length">
        <!-- Nghe tiếp -->
        <div v-if="showContinue" class="mt-5">
          <div class="flex items-center justify-between">
            <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Nghe tiếp</h2>
            <button v-if="!confirmClear" class="h-7 px-2.5 rounded-md text-xs flex items-center gap-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
              title="Xoá vị trí nghe của mọi cuốn" @click="confirmClear = true"><History class="w-3.5 h-3.5" /> Xoá lịch sử nghe</button>
          </div>
          <div v-if="confirmClear" role="alertdialog" aria-label="Xoá lịch sử nghe" class="mt-2 flex items-center gap-3 rounded-lg border border-border bg-muted/40 pl-3 pr-2 py-2 text-sm">
            <History class="w-4 h-4 text-muted-foreground shrink-0" />
            <span class="flex-1">Xoá lịch sử nghe của {{ listenedCount }} cuốn? Tiến độ về 0, lần sau nghe lại từ đầu. Sách không bị xoá.</span>
            <Button size="sm" variant="outline" @click="confirmClear = false">Huỷ</Button>
            <Button size="sm" variant="destructive" @click="clearHistory">Xoá lịch sử</Button>
          </div>
          <div class="mt-2 grid grid-cols-3 gap-3">
            <button v-for="b in continueList" :key="b.slug" class="flex items-center gap-3 rounded-lg border border-border p-2.5 text-left hover:bg-muted/50"
              :aria-label="`Nghe tiếp ${b.title}`" @click="openBook(b.slug, true)">
              <div class="h-14 w-[42px] shrink-0 rounded shadow-sm overflow-hidden">
                <BookCover :cover-image="b.thumbUrl || b.coverUrl" :title="b.title" :author="b.author" class="h-full w-full rounded shadow-none" />
              </div>
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium truncate">{{ b.title }}</p>
                <p class="text-xs text-muted-foreground truncate">Đã nghe {{ b.progress }}%<template v-if="b.listenedAt"> · {{ ago(b.listenedAt) }}</template></p>
                <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: b.progress + '%' }"></div></div>
              </div>
              <span class="h-8 w-8 grid place-items-center rounded-full bg-primary text-primary-foreground shrink-0"><Play class="w-4 h-4 ml-0.5" /></span>
            </button>
          </div>
        </div>
        <div class="flex items-center justify-between" :class="showContinue ? 'mt-6' : 'mt-4'">
          <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Tất cả sách</h2>
          <div class="flex items-center gap-2">
          <button v-if="shown.length > 1 && !arranging" class="h-7 px-2.5 rounded-md text-xs flex items-center gap-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
            title="Kéo sách để sắp xếp theo ý bạn" @click="startArrange"><ArrowUpDown class="w-3.5 h-3.5" /> Sắp xếp</button>
            <div class="flex rounded-md border border-border p-0.5" role="radiogroup" aria-label="Kiểu xem">
              <button role="radio" :aria-checked="view === 'grid'" title="Xem dạng lưới" aria-label="Xem dạng lưới" class="h-6 w-7 grid place-items-center rounded"
                :class="view === 'grid' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'" @click="setView('grid')"><LayoutGrid class="w-3.5 h-3.5" /></button>
              <button role="radio" :aria-checked="view === 'list'" title="Xem dạng danh sách" aria-label="Xem dạng danh sách" class="h-6 w-7 grid place-items-center rounded"
                :class="view === 'list' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'" @click="setView('list')"><List class="w-3.5 h-3.5" /></button>
            </div>
          </div>
        </div>

        <!-- Lưới sách: sách lẻ + thẻ bộ sách -->
        <div v-if="shown.length && view === 'grid'" class="mt-6 grid grid-cols-5 gap-x-5 gap-y-7 pr-3">
          <div v-for="it in shown" :key="it.key" class="text-left group relative rounded-lg"
            :class="[
              it.kind === 'book' && imported?.slug === it.book.slug && 'ring-2 ring-primary ring-offset-4 ring-offset-background',
              arranging && 'cursor-grab select-none touch-none', dragKey === it.key && 'opacity-80 shadow-2xl cursor-grabbing',
              overKey === it.key && dragKey !== it.key && 'ring-2 ring-primary ring-offset-4 ring-offset-background',
            ]"
            :data-shelf-key="it.key" :draggable="false"
            :style="dragKey === it.key ? { transform: `translate(${dragDelta.x}px, ${dragDelta.y}px) scale(1.04)`, zIndex: 40, pointerEvents: 'none' } : undefined"
            @pointerdown="onPointerDown($event, it.key)" @dragstart.prevent>
            <div class="relative">
              <!-- bộ sách: hai tập phía sau nhô lên rõ, như chồng sách -->
              <template v-if="it.kind === 'series'">
                <div class="absolute inset-0 translate-x-[14px] -translate-y-[14px] rounded-lg bg-slate-400 dark:bg-slate-600 border-2 border-background shadow"></div>
                <div class="absolute inset-0 translate-x-[7px] -translate-y-[7px] rounded-lg bg-slate-600 dark:bg-slate-400 border-2 border-background shadow"></div>
              </template>
              <button class="relative block w-full aspect-[3/4] rounded-lg shadow-md group-hover:shadow-xl transition overflow-hidden"
                :class="!arranging && 'group-hover:-translate-y-0.5'" :aria-label="it.kind === 'book' ? `Nghe ${it.book.title}` : `Mở bộ sách ${it.name}`" @click="openItem(it)">
                <BookCover :cover-image="it.kind === 'book' ? it.book.thumbUrl || it.book.coverUrl : it.thumbUrl || it.coverUrl" :title="it.kind === 'book' ? it.book.title : it.name" :author="it.kind === 'book' ? it.book.author : it.author" class="h-full w-full shadow-none" />
                <span v-if="it.kind === 'series'" class="absolute top-2 right-2 flex items-center gap-1 rounded-full bg-primary text-primary-foreground text-[11px] font-semibold px-2 py-0.5 shadow-md"><Layers class="w-3 h-3" /> Bộ {{ it.vols.length }} tập</span>
                <span v-if="arranging" class="absolute top-1.5 left-1.5 h-7 w-7 grid place-items-center rounded-md bg-black/50 text-white"><GripVertical class="w-4 h-4" /></span>
              </button>
            </div>
            <template v-if="it.kind === 'book'">
              <!-- Menu ⋯ trên bìa -->
              <div v-if="!arranging" data-book-menu>
                <button class="absolute top-1.5 right-1.5 h-7 w-7 grid place-items-center rounded-full bg-black/40 text-white opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
                  :class="menuFor === it.book.slug && 'opacity-100'" aria-label="Thao tác với sách" aria-haspopup="menu" :aria-expanded="menuFor === it.book.slug"
                  @click="menuFor = menuFor === it.book.slug ? null : it.book.slug"><MoreHorizontal class="w-4 h-4" /></button>
                <div v-if="menuFor === it.book.slug" role="menu" class="absolute top-10 right-1.5 w-44 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-20 text-sm">
                  <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted" @click="edit(it.book)"><Pencil class="w-4 h-4" /> Sửa sách</button>
                  <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted" @click="openFolder(it.book)"><FolderOpen class="w-4 h-4" /> Mở thư mục</button>
                  <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted text-destructive" @click="trash(it.book)"><Trash2 class="w-4 h-4" /> Chuyển vào Thùng rác</button>
                </div>
              </div>
              <p class="mt-2 text-sm font-medium truncate" :title="it.book.title">{{ it.book.title }}</p>
              <p class="text-xs text-muted-foreground truncate">{{ it.book.author ? `${it.book.author} · ` : '' }}{{ fmtLong(it.book.durationSec) }}</p>
              <p v-if="it.book.voice" class="text-xs text-muted-foreground truncate flex items-center gap-1"><Mic class="w-3 h-3 shrink-0" /> Giọng {{ it.book.voice }}</p>
              <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: it.book.progress + '%' }"></div></div>
              <p class="mt-1 text-[11px] text-muted-foreground flex items-center gap-1.5">
                <span class="shrink-0">{{ progressText(it.book.progress) }}</span>
                <template v-if="it.book.category && showChips && filter === 'all'"><span>·</span><span class="truncate">{{ it.book.category }}</span></template>
              </p>
            </template>
            <template v-else>
              <p class="mt-2 text-sm font-medium truncate flex items-center gap-1.5" :title="it.name"><Layers class="w-3.5 h-3.5 text-primary shrink-0" /><span class="truncate">{{ it.name }}</span></p>
              <p class="text-xs text-muted-foreground truncate">Bộ sách · {{ it.vols.length }} tập · {{ fmtLong(it.durationSec) }}</p>
              <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: it.progress + '%' }"></div></div>
              <p class="mt-1 text-[11px] text-muted-foreground truncate">{{ it.progress >= 99 ? 'Đã nghe hết bộ' : isListening(it.current) ? `Đang nghe tập ${it.current.volume}` : it.progress === 0 ? 'Chưa nghe' : `Nghe tiếp tập ${it.current.volume}` }}</p>
            </template>
          </div>
        </div>

        <!-- Danh sách: mỗi sách / bộ sách một hàng -->
        <div v-else-if="shown.length" class="mt-3 mr-3 rounded-lg border border-border">
          <div class="flex items-center gap-4 px-3 h-8 text-[11px] font-medium uppercase tracking-wider text-muted-foreground border-b border-border bg-muted/30 rounded-t-lg">
            <span v-if="arranging" class="w-4"></span>
            <span class="w-9"></span>
            <span class="flex-1">Tên sách</span>
            <span class="w-28">Danh mục</span>
            <span class="w-28 text-right">Thời lượng</span>
            <span class="w-36">Tiến độ</span>
            <span class="w-[68px]"></span>
          </div>
          <div v-for="(it, i) in shown" :key="it.key" class="group relative flex items-center gap-4 px-3 py-2 bg-background hover:bg-muted/40"
            :class="[
              i < shown.length - 1 && 'border-b border-border', i === shown.length - 1 && 'rounded-b-lg',
              !arranging && 'cursor-pointer', arranging && 'cursor-grab select-none touch-none', dragKey === it.key && 'opacity-90 shadow-2xl cursor-grabbing',
              overKey === it.key && dragKey !== it.key && 'bg-primary/5 shadow-[inset_0_2px_0_hsl(var(--primary))]',
              it.kind === 'book' && imported?.slug === it.book.slug && 'bg-primary/5',
            ]"
            :data-shelf-key="it.key" :draggable="false"
            :style="dragKey === it.key ? { transform: `translateY(${dragDelta.y}px)`, zIndex: 40, pointerEvents: 'none' } : undefined"
            @pointerdown="onPointerDown($event, it.key)" @dragstart.prevent @click="!arranging && openItem(it)">
            <GripVertical v-if="arranging" class="w-4 h-4 text-muted-foreground shrink-0" />
            <div class="relative h-12 w-9 shrink-0">
              <template v-if="it.kind === 'series'">
                <div class="absolute inset-0 translate-x-[4px] -translate-y-[4px] rounded bg-slate-400 dark:bg-slate-600 border border-background"></div>
                <div class="absolute inset-0 translate-x-[2px] -translate-y-[2px] rounded bg-slate-600 dark:bg-slate-400 border border-background"></div>
              </template>
              <div class="relative h-full w-full rounded shadow-sm overflow-hidden">
                <BookCover :cover-image="it.kind === 'book' ? it.book.thumbUrl || it.book.coverUrl : it.thumbUrl || it.coverUrl" :title="it.kind === 'book' ? it.book.title : it.name" :author="it.kind === 'book' ? it.book.author : it.author" class="h-full w-full rounded shadow-none" />
              </div>
            </div>
            <template v-if="it.kind === 'book'">
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium truncate" :title="it.book.title">{{ it.book.title }}</p>
                <p class="text-xs text-muted-foreground truncate flex items-center gap-1">
                  <template v-if="it.book.author">{{ it.book.author }}<template v-if="it.book.voice"> ·</template></template>
                  <template v-if="it.book.voice"><Mic class="w-3 h-3 shrink-0" /> {{ it.book.voice }}</template>
                </p>
              </div>
              <span class="w-28 text-xs text-muted-foreground truncate" :title="it.book.category">{{ it.book.category || '—' }}</span>
              <span class="w-28 text-right text-xs text-muted-foreground tabular-nums">{{ fmtLong(it.book.durationSec) }}</span>
              <div class="w-36">
                <div class="h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: it.book.progress + '%' }"></div></div>
                <p class="mt-1 text-[11px] text-muted-foreground truncate">{{ progressText(it.book.progress) }}</p>
              </div>
            </template>
            <template v-else>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium truncate flex items-center gap-1.5">
                  <span class="truncate" :title="it.name">{{ it.name }}</span>
                  <span class="shrink-0 flex items-center gap-1 rounded-full bg-primary/10 text-primary text-[11px] font-semibold px-1.5 py-px"><Layers class="w-3 h-3" /> Bộ {{ it.vols.length }} tập</span>
                </p>
                <p class="text-xs text-muted-foreground truncate flex items-center gap-1">
                  <template v-if="it.author">{{ it.author }}<template v-if="it.current.voice"> ·</template></template>
                  <template v-if="it.current.voice"><Mic class="w-3 h-3 shrink-0" /> {{ it.current.voice }}</template>
                </p>
              </div>
              <span class="w-28 text-xs text-muted-foreground truncate" :title="it.current.category">{{ it.current.category || '—' }}</span>
              <span class="w-28 text-right text-xs text-muted-foreground tabular-nums">{{ fmtLong(it.durationSec) }}</span>
              <div class="w-36">
                <div class="h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: it.progress + '%' }"></div></div>
                <p class="mt-1 text-[11px] text-muted-foreground truncate">{{ it.progress >= 99 ? 'Đã nghe hết bộ' : isListening(it.current) ? `Đang nghe tập ${it.current.volume}` : it.progress === 0 ? 'Chưa nghe' : `Nghe tiếp tập ${it.current.volume}` }}</p>
              </div>
            </template>
            <div class="w-[68px] flex items-center justify-end gap-1">
              <template v-if="!arranging">
                <button class="h-8 w-8 grid place-items-center rounded-full focus-visible:opacity-100"
                  :class="isPlaying(it) ? 'bg-primary text-primary-foreground' : 'border border-border bg-background text-foreground hover:bg-muted opacity-0 group-hover:opacity-100'"
                  :aria-label="it.kind === 'book' ? `Nghe ${it.book.title}` : `Nghe tiếp bộ ${it.name}`" @click.stop="playItem(it)"><Play class="w-4 h-4 ml-0.5" /></button>
                <div v-if="it.kind === 'book'" data-book-menu @click.stop>
                  <button class="h-7 w-7 grid place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:opacity-100"
                    :class="menuFor === it.book.slug ? 'bg-muted text-foreground' : 'opacity-0 group-hover:opacity-100'"
                    aria-label="Thao tác với sách" aria-haspopup="menu" :aria-expanded="menuFor === it.book.slug"
                    @click="menuFor = menuFor === it.book.slug ? null : it.book.slug"><MoreHorizontal class="w-4 h-4" /></button>
                  <div v-if="menuFor === it.book.slug" role="menu" class="absolute top-11 right-3 w-48 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-20 text-sm">
                    <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted" @click="edit(it.book)"><Pencil class="w-4 h-4" /> Sửa sách</button>
                    <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted" @click="openFolder(it.book)"><FolderOpen class="w-4 h-4" /> Mở thư mục</button>
                    <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted text-destructive" @click="trash(it.book)"><Trash2 class="w-4 h-4" /> Chuyển vào Thùng rác</button>
                  </div>
                </div>
                <span v-else class="w-7"></span>
              </template>
            </div>
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

    </template>

    <ManageShelfDialog v-if="manageTab" :categories="categories" :series="seriesGroups" :tab="manageTab" @close="manageTab = null" @changed="onShelfChanged" />
    <ImportDialog v-if="importPath" :path="importPath" @close="importPath = ''" @imported="onImported" />

    <div v-if="dragging && !importPath" class="pointer-events-none absolute inset-3 rounded-xl border-2 border-dashed border-primary bg-primary/5 grid place-items-center z-30">
      <div class="text-center">
        <FileArchive class="w-10 h-10 mx-auto text-primary" />
        <p class="mt-3 font-medium">Thả gói sách (.zip) để nhập vào thư viện</p>
        <p class="mt-1 text-sm text-muted-foreground">File .docx, .epub thì vào Tạo sách nói</p>
      </div>
    </div>

    <div v-if="imported" role="status" class="absolute bottom-5 left-1/2 -translate-x-1/2 flex items-center gap-3 rounded-lg border border-border bg-background shadow-lg pl-4 pr-2 py-2 text-sm z-20">
      <Check class="w-4 h-4 text-rag-green" /> <span class="max-w-72 truncate">Đã nhập “{{ imported.title }}”</span>
      <Button size="sm" @click="openBook(imported.slug, true); imported = null"><Play class="w-4 h-4" /> Nghe ngay</Button>
      <button aria-label="Đóng thông báo" class="h-8 w-8 grid place-items-center rounded-md text-muted-foreground hover:bg-muted" @click="imported = null"><X class="w-4 h-4" /></button>
    </div>
  </section>
</template>
