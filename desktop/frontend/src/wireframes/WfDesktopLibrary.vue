<script setup lang="ts">
// D2 — Thư viện dễ tìm (P3-15) cho phần mềm desktop. Bám khung D1 (1100×720). Đã duyệt 24/09.
// - Tìm theo tên sách / tác giả, sắp xếp, lọc theo danh mục.
// - Danh mục là ô tuỳ chọn khi tạo sách (ghi vào trường `category` của gói zip) —
//   KHÔNG có màn quản lý danh mục riêng. Nút lọc chỉ hiện khi có từ 2 danh mục.
// - Hàng "Nghe tiếp" gom các cuốn đang nghe dở.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Search, X, ChevronDown, Check, Play,
  MoreHorizontal, Pencil, FolderOpen, Trash2, Tag, Plus, Sun, Moon, ArrowUpCircle, FileText,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

type Mode = 'many' | 'few' | 'empty-search' | 'create' | 'edit'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'many')
const dark = ref(false)

interface Book { title: string; author: string; dur: string; chapters: number; progress: number; cat: string; listened: string }
const allBooks: Book[] = [
  { title: 'Kỹ năng mềm cho người trẻ', author: 'Nguyễn Văn A', dur: '1 giờ 12 phút', chapters: 4, progress: 35, cat: 'Kỹ năng', listened: 'hôm nay' },
  { title: 'Lãnh đạo cho quản lý mới', author: 'Trần Thị B', dur: '58 phút', chapters: 4, progress: 100, cat: 'Kinh doanh', listened: '3 ngày trước' },
  { title: 'Tài chính cá nhân cơ bản', author: 'Lê Văn C', dur: '1 giờ 05 phút', chapters: 4, progress: 0, cat: 'Tài chính', listened: '' },
  { title: 'Khởi nghiệp từ số 0', author: 'Phạm Thị D', dur: '47 phút', chapters: 4, progress: 12, cat: 'Kinh doanh', listened: 'hôm qua' },
  { title: 'Quản lý thời gian hiệu quả', author: 'Nguyễn Văn A', dur: '2 giờ 10 phút', chapters: 8, progress: 64, cat: 'Kỹ năng', listened: 'hôm qua' },
  { title: 'Đọc hiểu báo cáo tài chính', author: 'Hoàng Văn E', dur: '3 giờ 25 phút', chapters: 12, progress: 0, cat: 'Tài chính', listened: '' },
  { title: 'Bán hàng qua điện thoại', author: 'Trần Thị B', dur: '1 giờ 40 phút', chapters: 6, progress: 100, cat: 'Kinh doanh', listened: 'tuần trước' },
  { title: 'Ngủ ngon mỗi đêm', author: 'Đỗ Thị G', dur: '52 phút', chapters: 5, progress: 0, cat: 'Sức khoẻ', listened: '' },
  { title: 'Giao tiếp nơi công sở', author: 'Lê Văn C', dur: '1 giờ 18 phút', chapters: 5, progress: 0, cat: 'Kỹ năng', listened: '' },
  { title: 'Ghi chép cuộc họp', author: 'Phạm Thị D', dur: '34 phút', chapters: 3, progress: 0, cat: '', listened: '' },
  { title: 'Đầu tư cho người mới', author: 'Hoàng Văn E', dur: '2 giờ 02 phút', chapters: 7, progress: 20, cat: 'Tài chính', listened: '5 ngày trước' },
  { title: 'Xây dựng đội nhóm', author: 'Trần Thị B', dur: '1 giờ 33 phút', chapters: 6, progress: 0, cat: 'Kinh doanh', listened: '' },
]

const categories = computed(() => {
  const m = new Map<string, number>()
  for (const b of books.value) m.set(b.cat || '', (m.get(b.cat || '') ?? 0) + 1)
  return [...m.entries()].filter(([k]) => k).sort((a, b) => b[1] - a[1])
})
const uncategorized = computed(() => books.value.filter((b) => !b.cat).length)

const books = computed(() => (mode.value === 'few' ? allBooks.slice(0, 4).map((b) => ({ ...b, cat: b.cat === 'Kinh doanh' ? 'Kinh doanh' : '' })) : allBooks))
const query = ref('')
const filter = ref<string>('all') // 'all' | 'listening' | '__none' | <danh mục>
const sorts = ['Mới tạo nhất', 'Nghe gần đây', 'Tên A–Z', 'Tác giả A–Z', 'Dài nhất']
const sort = ref(sorts[0])
const sortOpen = ref(false)
const menuFor = ref<string | null>(null)

const listening = computed(() => books.value.filter((b) => b.progress > 0 && b.progress < 100))
const showChips = computed(() => categories.value.length >= 2)
const shown = computed(() => {
  if (mode.value === 'empty-search') return []
  const q = query.value.trim().toLowerCase()
  return books.value.filter((b) => {
    if (q && !b.title.toLowerCase().includes(q) && !b.author.toLowerCase().includes(q)) return false
    if (filter.value === 'listening') return b.progress > 0 && b.progress < 100
    if (filter.value === '__none') return !b.cat
    if (filter.value !== 'all') return b.cat === filter.value
    return true
  })
})
const showContinue = computed(() => filter.value === 'all' && !query.value && mode.value === 'many' && listening.value.length > 0)

function setMode(m: Mode) {
  mode.value = m
  query.value = m === 'empty-search' ? 'tiếng nhật' : ''
  filter.value = 'all'
  menuFor.value = null
}

// Ô danh mục ở bước 1 (tạo sách) và hộp sửa thông tin: dạng ô chọn (bấm là mở danh sách),
// nút "Tạo danh mục mới" luôn nằm cuối danh sách — không cần gõ gì mới thấy.
const catValue = ref('')
const catOpen = ref(true)
const catCreating = ref(false)
const catNew = ref('')
const firstTime = ref(false) // chưa có cuốn nào có danh mục
const catList = computed(() => (firstTime.value ? [] : categories.value))
function pickCat(c: string) {
  catValue.value = c
  catOpen.value = false
  catCreating.value = false
}
function addCat() {
  if (catNew.value.trim()) pickCat(catNew.value.trim())
  catNew.value = ''
}

const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'create', label: 'Tạo sách mới', icon: FilePlus2 },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
  { key: 'about', label: 'Giới thiệu', icon: Info },
]
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <!-- Nút duyệt trạng thái (ngoài khung, chỉ để xem wireframe) -->
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in ([
        ['many', 'Nhiều sách (12)'], ['few', 'Ít sách (4) · 1 danh mục'], ['empty-search', 'Tìm không thấy'],
        ['create', 'Tạo sách · ô Danh mục'], ['edit', 'Sửa thông tin sách'],
      ] as [Mode, string][])" :key="m[0]"
        class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="setMode(m[0])">{{ m[1] }}</button>
      <label v-if="mode === 'create'" class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground cursor-pointer">
        <input v-model="firstTime" type="checkbox" class="h-3.5 w-3.5" /> Lần đầu (chưa có danh mục)
      </label>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <!-- Khung cửa sổ ứng dụng -->
    <div class="relative w-[1100px] h-[720px] rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground flex flex-col">
      <div class="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border bg-muted/50">
        <span class="h-3 w-3 rounded-full bg-rag-red"></span>
        <span class="h-3 w-3 rounded-full bg-rag-amber"></span>
        <span class="h-3 w-3 rounded-full bg-rag-green"></span>
        <span class="flex-1 text-center text-xs text-muted-foreground">Sano</span>
      </div>

      <div class="flex-1 flex min-h-0">
        <!-- Thanh bên (như app hiện tại) -->
        <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
          <div class="px-4 py-4 flex items-center gap-2">
            <img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" />
            <span class="font-semibold tracking-tight">Sano</span>
          </div>
          <nav class="px-2 space-y-0.5">
            <button v-for="n in nav" :key="n.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
              :class="(mode === 'create' ? n.key === 'create' : n.key === 'library') ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
              @click="n.key === 'create' ? setMode('create') : setMode('many')">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </button>
          </nav>
          <div class="px-2 mt-2">
            <span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span>
          </div>
          <div class="flex-1"></div>
          <span class="mx-3 mb-2 flex items-center gap-2 rounded-md px-2 h-8 text-xs text-primary"><ArrowUpCircle class="w-4 h-4" /> Có bản mới 0.2.0</span>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.0 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <!-- ─── Tạo sách · bước 1 với ô Danh mục ─── -->
        <main v-if="mode === 'create'" class="flex-1 min-w-0 overflow-auto p-6">
          <div class="max-w-2xl">
            <div class="flex items-center gap-3 rounded-lg border border-border px-4 py-3">
              <FileText class="w-5 h-5 text-primary" />
              <div class="flex-1 text-sm"><p class="font-medium">quan-ly-thoi-gian.docx</p><p class="text-xs text-muted-foreground">8 chương · 42.300 ký tự · đọc khoảng 2 giờ</p></div>
              <Button variant="outline" size="sm">Chọn file khác</Button>
            </div>
            <div class="mt-5 grid grid-cols-2 gap-4">
              <label class="text-sm">Tên sách
                <input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" value="Quản lý thời gian hiệu quả" />
              </label>
              <label class="text-sm">Tác giả
                <input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" value="Nguyễn Văn A" />
              </label>
            </div>

            <!-- Ô danh mục: dạng ô chọn, bấm là thấy danh sách + nút tạo mới -->
            <div class="mt-4 text-sm">
              <span>Danh mục <span class="text-muted-foreground">· không bắt buộc</span></span>
              <div class="relative mt-1 w-1/2 pr-2">
                <button class="w-full flex items-center h-9 rounded-md border bg-background px-3 text-left"
                  :class="catOpen ? 'border-ring ring-2 ring-ring/20' : 'border-input'" @click="catOpen = !catOpen">
                  <Tag class="w-4 h-4 text-muted-foreground mr-2" />
                  <span class="flex-1" :class="!catValue && 'text-muted-foreground'">{{ catValue || 'Chọn danh mục' }}</span>
                  <ChevronDown class="w-4 h-4 text-muted-foreground" />
                </button>
                <div v-if="catOpen" class="absolute left-0 right-2 top-full mt-1 rounded-lg border border-border bg-popover shadow-lg py-1 z-10">
                  <button class="w-full flex items-center justify-between px-3 py-1.5 text-left hover:bg-muted" @click="pickCat('')">
                    <span class="text-muted-foreground">Không phân loại</span><Check v-if="!catValue" class="w-4 h-4 text-primary" />
                  </button>
                  <template v-if="catList.length">
                    <button v-for="[c, n] in catList" :key="c" class="w-full flex items-center justify-between px-3 py-1.5 text-left hover:bg-muted" @click="pickCat(c)">
                      <span :class="catValue === c && 'text-primary font-medium'">{{ c }}</span>
                      <span class="flex items-center gap-2 text-xs text-muted-foreground">{{ n }} cuốn <Check v-if="catValue === c" class="w-4 h-4 text-primary" /></span>
                    </button>
                  </template>
                  <p v-else class="px-3 py-1.5 text-xs text-muted-foreground">Chưa có danh mục nào. Tạo danh mục đầu tiên để dễ lọc sách sau này.</p>
                  <div class="border-t border-border mt-1 pt-1">
                    <div v-if="catCreating" class="flex items-center gap-2 px-2 py-1">
                      <input v-model="catNew" class="flex-1 h-8 rounded-md border border-input bg-background px-2" placeholder="Tên danh mục, vd: Kỹ năng" @keydown.enter="addCat" />
                      <Button size="sm" @click="addCat">Thêm</Button>
                    </div>
                    <button v-else class="w-full flex items-center gap-2 px-3 py-1.5 text-left text-primary hover:bg-muted" @click="catCreating = true">
                      <Plus class="w-4 h-4" /> Tạo danh mục mới
                    </button>
                  </div>
                </div>
              </div>
              <p class="mt-1.5 text-xs text-muted-foreground">Dùng để lọc trong Thư viện.</p>
            </div>

            <div class="mt-24 flex justify-end"><Button>Tiếp tục</Button></div>
          </div>
        </main>

        <!-- ─── Thư viện ─── -->
        <main v-else class="flex-1 min-w-0 flex flex-col">
          <div class="px-6 pt-6">
            <div class="flex items-center justify-between">
              <div>
                <h1 class="text-xl font-semibold tracking-tight">Thư viện</h1>
                <p class="text-sm text-muted-foreground">{{ books.length }} cuốn · {{ mode === 'few' ? '4 giờ 02 phút' : '19 giờ 16 phút' }} · lưu ở ~/Sano/Sach</p>
              </div>
              <Button @click="setMode('create')"><FilePlus2 class="w-4 h-4" /> Tạo sách mới</Button>
            </div>

            <!-- Thanh công cụ: tìm + sắp xếp -->
            <div class="mt-4 flex items-center gap-2">
              <div class="flex-1 flex items-center h-9 rounded-md border border-input bg-background px-3">
                <Search class="w-4 h-4 text-muted-foreground mr-2" />
                <input v-model="query" class="flex-1 bg-transparent outline-none text-sm" placeholder="Tìm theo tên sách hoặc tác giả…" />
                <button v-if="query" class="text-muted-foreground" @click="query = ''"><X class="w-4 h-4" /></button>
              </div>
              <div class="relative">
                <button class="h-9 px-3 rounded-md border border-input bg-background text-sm flex items-center gap-1.5" @click="sortOpen = !sortOpen">
                  <span class="text-muted-foreground">Sắp xếp:</span> {{ sort }} <ChevronDown class="w-4 h-4 text-muted-foreground" />
                </button>
                <div v-if="sortOpen" class="absolute right-0 top-full mt-1 w-44 rounded-lg border border-border bg-popover shadow-lg py-1 z-20">
                  <button v-for="s in sorts" :key="s" class="w-full flex items-center justify-between px-3 py-1.5 text-sm text-left hover:bg-muted" :class="s === sort && 'text-primary font-medium'" @click="sort = s; sortOpen = false">
                    {{ s }} <Check v-if="s === sort" class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Nút lọc: chỉ hiện khi có ≥2 danh mục -->
            <div v-if="showChips" class="mt-3 flex flex-wrap gap-1.5">
              <button v-for="c in ([
                ['all', `Tất cả · ${books.length}`],
                ['listening', `Đang nghe · ${listening.length}`],
                ...categories.map(([k, n]) => [k, `${k} · ${n}`]),
                ...(uncategorized ? [['__none', `Chưa phân loại · ${uncategorized}`]] : []),
              ] as string[][])" :key="c[0]"
                class="h-7 px-3 rounded-full text-xs border"
                :class="filter === c[0] ? 'border-primary bg-primary/10 text-primary font-medium' : 'border-border text-muted-foreground hover:text-foreground'"
                @click="filter = c[0]">{{ c[1] }}</button>
            </div>
          </div>

          <div class="flex-1 overflow-auto px-6 pb-6">
            <!-- Nghe tiếp -->
            <div v-if="showContinue" class="mt-5">
              <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Nghe tiếp</h2>
              <div class="mt-2 grid grid-cols-3 gap-3">
                <button v-for="b in listening.slice(0, 3)" :key="b.title" class="flex items-center gap-3 rounded-lg border border-border p-2.5 text-left hover:bg-muted/50">
                  <div class="h-14 w-[42px] shrink-0 rounded shadow-sm"><WfBookCover :title="b.title" :author="b.author" size="sm" /></div>
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-medium truncate">{{ b.title }}</p>
                    <p class="text-xs text-muted-foreground">Đã nghe {{ b.progress }}% · {{ b.listened }}</p>
                    <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: b.progress + '%' }"></div></div>
                  </div>
                  <span class="h-8 w-8 grid place-items-center rounded-full bg-primary text-primary-foreground shrink-0"><Play class="w-4 h-4 ml-0.5" /></span>
                </button>
              </div>
              <h2 class="mt-6 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Tất cả sách</h2>
            </div>

            <!-- Lưới sách -->
            <div v-if="shown.length" class="mt-3 grid grid-cols-5 gap-4">
              <div v-for="b in shown" :key="b.title" class="text-left group relative">
                <div class="aspect-[3/4] rounded-lg shadow-md group-hover:shadow-xl group-hover:-translate-y-0.5 transition cursor-pointer">
                  <WfBookCover :title="b.title" :author="b.author" size="sm" />
                </div>
                <!-- Menu ⋯ trên bìa -->
                <button class="absolute top-1.5 right-1.5 h-7 w-7 grid place-items-center rounded-full bg-black/40 text-white opacity-0 group-hover:opacity-100"
                  :class="menuFor === b.title && 'opacity-100'" @click="menuFor = menuFor === b.title ? null : b.title"><MoreHorizontal class="w-4 h-4" /></button>
                <div v-if="menuFor === b.title" class="absolute top-10 right-1.5 w-44 rounded-lg border border-border bg-popover shadow-lg py-1 z-20 text-sm">
                  <button class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted" @click="setMode('edit')"><Pencil class="w-4 h-4" /> Sửa thông tin</button>
                  <button class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted"><FolderOpen class="w-4 h-4" /> Mở thư mục</button>
                  <button class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted text-destructive"><Trash2 class="w-4 h-4" /> Chuyển vào Thùng rác</button>
                </div>
                <p class="mt-2 text-sm font-medium truncate">{{ b.title }}</p>
                <p class="text-xs text-muted-foreground truncate">{{ b.author }} · {{ b.dur }}</p>
                <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: b.progress + '%' }"></div></div>
                <p class="mt-1 text-[11px] text-muted-foreground flex items-center gap-1.5">
                  <span>{{ b.progress === 100 ? 'Đã nghe xong' : b.progress === 0 ? 'Chưa nghe' : `Đã nghe ${b.progress}%` }}</span>
                  <template v-if="b.cat && showChips && filter === 'all'"><span>·</span><span class="truncate">{{ b.cat }}</span></template>
                </p>
              </div>
            </div>

            <!-- Tìm không thấy -->
            <div v-else class="mt-16 flex flex-col items-center text-center">
              <Search class="w-8 h-8 text-muted-foreground" />
              <p class="mt-3 font-medium">Không có cuốn nào khớp "{{ query }}"</p>
              <p class="text-sm text-muted-foreground">Thử tên khác, tên tác giả, hoặc bỏ bộ lọc.</p>
              <Button variant="outline" size="sm" class="mt-4" @click="query = ''; filter = 'all'; mode = 'many'">Xoá tìm kiếm</Button>
            </div>
          </div>
        </main>
      </div>

      <!-- ═══ Hộp sửa thông tin sách ═══ -->
      <div v-if="mode === 'edit'" class="absolute inset-0 bg-black/40 grid place-items-center z-30">
        <div class="w-[440px] rounded-xl border border-border bg-background shadow-2xl p-5">
          <div class="flex items-start justify-between">
            <h2 class="font-semibold">Sửa thông tin sách</h2>
            <button class="text-muted-foreground" @click="setMode('many')"><X class="w-4 h-4" /></button>
          </div>
          <div class="mt-4 space-y-3 text-sm">
            <label class="block">Tên sách<input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" value="Ghi chép cuộc họp" /></label>
            <label class="block">Tác giả<input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" value="Phạm Thị D" /></label>
            <label class="block">Danh mục <span class="text-muted-foreground">· không bắt buộc</span>
              <div class="mt-1 flex items-center h-9 rounded-md border border-input bg-background px-3 cursor-pointer"><Tag class="w-4 h-4 text-muted-foreground mr-2" /><span class="flex-1">Kỹ năng</span><ChevronDown class="w-4 h-4 text-muted-foreground" /></div>
              <span class="mt-1 block text-xs text-muted-foreground">Bấm để chọn danh mục khác hoặc tạo danh mục mới (giống bước tạo sách)</span>
            </label>
            <p class="text-xs text-muted-foreground">Chỉ đổi thông tin hiển thị và gói zip. Lời giới thiệu đầu sách đã đọc giữ nguyên, muốn đọc lại tên mới thì tạo lại sách.</p>
          </div>
          <div class="mt-5 flex justify-end gap-2">
            <Button variant="outline" @click="setMode('many')">Huỷ</Button>
            <Button @click="setMode('many')">Lưu</Button>
          </div>
        </div>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D2 · Thư viện dễ tìm — tìm kiếm + sắp xếp luôn có; nút lọc danh mục chỉ hiện khi có từ 2 danh mục; danh mục đặt ở bước 1 khi tạo sách hoặc qua "Sửa thông tin" (menu ⋯ trên bìa). Không có màn quản lý danh mục riêng.
    </p>
  </div>
</template>
