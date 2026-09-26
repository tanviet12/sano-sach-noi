<script setup lang="ts">
// D5 — Bộ sách, tự sắp xếp, quản lý danh mục (bổ sung cho D2 Thư viện). Bám khung D1 (1100×720).
// - Bộ sách: các tập cùng bộ gom thành MỘT thẻ trên kệ, bấm vào thấy các tập theo số tập.
//   Đặt bộ + số tập ở hộp "Sửa thông tin" (và bước 1 khi tạo sách, cùng hai ô này).
// - Tự sắp xếp: thêm một kiểu trong menu Sắp xếp, kéo bìa để đổi chỗ. Một bộ sách di chuyển
//   như một thẻ. Thứ tự lưu trong ~/Sano (file nhỏ), không cần cơ sở dữ liệu.
// - Quản lý danh mục và bộ sách: đổi tên (áp cho mọi cuốn), xoá (sách giữ nguyên).
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Search, ChevronDown, Check, Play,
  MoreHorizontal, Pencil, Trash2, Tag, Plus, Sun, Moon, X, ChevronLeft, Layers, GripVertical,
  Settings2, ChevronRight,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

type Mode = 'shelf' | 'series' | 'edit' | 'manual' | 'manage'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'shelf')
const dark = ref(false)

interface Vol { n: number; title: string; dur: string; progress: number }
const series = {
  name: 'Kinh doanh cho người mới',
  author: 'Nguyễn Văn A',
  cat: 'Kinh doanh',
  vols: [
    { n: 1, title: 'Tìm đúng khách hàng', dur: '1 giờ 20 phút', progress: 100 },
    { n: 2, title: 'Bán hàng không ngại', dur: '1 giờ 35 phút', progress: 40 },
    { n: 3, title: 'Giữ chân khách cũ', dur: '1 giờ 15 phút', progress: 0 },
  ] as Vol[],
}
interface Item { id: string; kind: 'book' | 'series'; title: string; author: string; meta: string; progress: number; note: string; cat: string }
const initial: Item[] = [
  { id: 's1', kind: 'series', title: series.name, author: series.author, meta: '3 tập · 4 giờ 10 phút', progress: 47, note: 'Đang nghe tập 2', cat: 'Kinh doanh' },
  { id: 'b1', kind: 'book', title: 'Kỹ năng mềm cho người trẻ', author: 'Nguyễn Văn A', meta: '1 giờ 12 phút', progress: 35, note: 'Đã nghe 35%', cat: 'Kỹ năng' },
  { id: 'b2', kind: 'book', title: 'Tài chính cá nhân cơ bản', author: 'Lê Văn C', meta: '1 giờ 05 phút', progress: 0, note: 'Chưa nghe', cat: 'Tài chính' },
  { id: 'b3', kind: 'book', title: 'Ghi chép cuộc họp', author: 'Phạm Thị D', meta: '34 phút', progress: 0, note: 'Chưa nghe', cat: '' },
  { id: 'b4', kind: 'book', title: 'Quản lý thời gian hiệu quả', author: 'Nguyễn Văn A', meta: '2 giờ 10 phút', progress: 64, note: 'Đã nghe 64%', cat: 'Kỹ năng' },
  { id: 'b5', kind: 'book', title: 'Ngủ ngon mỗi đêm', author: 'Đỗ Thị G', meta: '52 phút', progress: 0, note: 'Chưa nghe', cat: 'Sức khoẻ' },
  { id: 'b6', kind: 'book', title: 'Đầu tư cho người mới', author: 'Hoàng Văn E', meta: '2 giờ 02 phút', progress: 20, note: 'Đã nghe 20%', cat: 'Tài chính' },
]
const items = ref<Item[]>([...initial])

const sorts = ['Tự sắp xếp', 'Mới tạo nhất', 'Nghe gần đây', 'Tên A–Z', 'Tác giả A–Z', 'Dài nhất']
const sort = ref(mode.value === 'manual' ? 'Tự sắp xếp' : 'Mới tạo nhất')
const sortOpen = ref(false)
const filter = ref('all')

// Kéo thả (chạy thật trong wireframe để cảm được thao tác)
const dragId = ref<string | null>(null)
const overId = ref<string | null>(null)
function onDrop(target: string) {
  const from = items.value.findIndex((i) => i.id === dragId.value)
  const to = items.value.findIndex((i) => i.id === target)
  if (from < 0 || to < 0 || from === to) return
  const list = [...items.value]
  const [m] = list.splice(from, 1)
  list.splice(to, 0, m)
  items.value = list
  dragId.value = overId.value = null
}

// Hộp quản lý
const tab = ref<'cat' | 'series'>('cat')
const cats = ref([
  { name: 'Kinh doanh', n: 5 },
  { name: 'Kỹ năng', n: 2 },
  { name: 'Tài chính', n: 2 },
  { name: 'Sức khoẻ', n: 1 },
])
const seriesList = ref([
  { name: 'Kinh doanh cho người mới', n: 3 },
  { name: 'Tiếng Anh giao tiếp', n: 2 },
])
const editing = ref<string | null>('Sức khoẻ')
const editValue = ref('Sức khoẻ và thể thao')
const confirmDel = ref<string | null>(null)

// Hộp sửa thông tin: ô bộ sách
const seriesOpen = ref(false)
const seriesValue = ref(series.name)
const volNo = ref(3)

function setMode(m: Mode) {
  mode.value = m
  sortOpen.value = false
  seriesOpen.value = false
  confirmDel.value = null
  if (m === 'manual') sort.value = 'Tự sắp xếp'
  else if (m === 'shelf') sort.value = 'Mới tạo nhất'
}
const manual = computed(() => sort.value === 'Tự sắp xếp')
const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'create', label: 'Tạo sách mới', icon: FilePlus2 },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
  { key: 'about', label: 'Giới thiệu', icon: Info },
]
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <!-- Nút duyệt trạng thái (ngoài khung) -->
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in ([
        ['shelf', '1. Kệ sách có bộ sách'], ['series', '2. Mở một bộ sách'], ['edit', '3. Sửa thông tin · ô Bộ sách'],
        ['manual', '4. Tự sắp xếp (kéo thả)'], ['manage', '5. Quản lý danh mục, bộ sách'],
      ] as [Mode, string][])" :key="m[0]"
        class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="setMode(m[0])">{{ m[1] }}</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <!-- Khung cửa sổ -->
    <div class="relative w-[1100px] h-[720px] rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground flex flex-col">
      <div class="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border bg-muted/50">
        <span class="h-3 w-3 rounded-full bg-rag-red"></span>
        <span class="h-3 w-3 rounded-full bg-rag-amber"></span>
        <span class="h-3 w-3 rounded-full bg-rag-green"></span>
        <span class="flex-1 text-center text-xs text-muted-foreground">Sano</span>
      </div>

      <div class="flex-1 flex min-h-0">
        <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
          <div class="px-4 py-4 flex items-center gap-2">
            <img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" />
            <span class="font-semibold tracking-tight">Sano</span>
          </div>
          <nav class="px-2 space-y-0.5">
            <button v-for="n in nav" :key="n.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
              :class="n.key === 'library' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
              @click="setMode('shelf')">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </button>
          </nav>
          <div class="px-2 mt-2">
            <span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span>
          </div>
          <div class="flex-1"></div>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.10 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <!-- ─── 2. Trong một bộ sách ─── -->
        <main v-if="mode === 'series'" class="flex-1 min-w-0 overflow-auto p-6">
          <button class="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground" @click="setMode('shelf')"><ChevronLeft class="w-4 h-4" /> Thư viện</button>
          <div class="mt-4 flex gap-6">
            <div class="relative w-36 shrink-0">
              <div class="absolute inset-0 translate-x-3 -translate-y-2 rounded-lg bg-muted-foreground/25"></div>
              <div class="absolute inset-0 translate-x-1.5 -translate-y-1 rounded-lg bg-muted-foreground/40"></div>
              <div class="relative aspect-[3/4] rounded-lg shadow-lg"><WfBookCover :title="series.name" :author="series.author" size="sm" /></div>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-xs font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5"><Layers class="w-3.5 h-3.5" /> Bộ sách · 3 tập</p>
              <h1 class="mt-1 text-2xl font-semibold tracking-tight">{{ series.name }}</h1>
              <p class="text-sm text-muted-foreground">{{ series.author }} · 4 giờ 10 phút · {{ series.cat }}</p>
              <div class="mt-4 flex items-center gap-2">
                <Button><Play class="w-4 h-4" /> Nghe tiếp tập 2</Button>
                <Button variant="outline" size="icon" @click="setMode('manage'); tab = 'series'"><Pencil class="w-4 h-4" /></Button>
              </div>
            </div>
          </div>

          <div class="mt-6 rounded-lg border border-border divide-y divide-border">
            <div v-for="v in series.vols" :key="v.n" class="flex items-center gap-4 px-4 py-3 hover:bg-muted/40">
              <span class="w-14 text-sm font-semibold text-muted-foreground">Tập {{ v.n }}</span>
              <div class="h-12 w-9 shrink-0 rounded shadow-sm"><WfBookCover :title="v.title" :author="series.author" size="sm" /></div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium truncate">{{ v.title }}</p>
                <p class="text-xs text-muted-foreground">{{ v.dur }} · {{ v.progress === 100 ? 'Đã nghe xong' : v.progress === 0 ? 'Chưa nghe' : `Đã nghe ${v.progress}%` }}</p>
                <div class="mt-1.5 h-1 w-48 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: v.progress + '%' }"></div></div>
              </div>
              <span class="h-8 w-8 grid place-items-center rounded-full" :class="v.n === 2 ? 'bg-primary text-primary-foreground' : 'border border-border text-foreground'"><Play class="w-4 h-4 ml-0.5" /></span>
              <MoreHorizontal class="w-4 h-4 text-muted-foreground" />
            </div>
          </div>
          <p class="mt-3 text-xs text-muted-foreground">Các tập luôn xếp theo số tập. Thêm tập mới: khi tạo sách, chọn bộ này ở ô "Bộ sách". Nghe hết một tập thì tự chuyển sang tập sau.</p>
        </main>

        <!-- ─── Kệ sách (1, 3, 4, 5 dùng chung nền này) ─── -->
        <main v-else class="flex-1 min-w-0 flex flex-col">
          <div class="px-6 pt-6">
            <div class="flex items-center justify-between">
              <div>
                <h1 class="text-xl font-semibold tracking-tight">Thư viện</h1>
                <p class="text-sm text-muted-foreground">9 cuốn · 1 bộ sách · 12 giờ 40 phút · lưu ở ~/Sano/Sach</p>
              </div>
              <Button><FilePlus2 class="w-4 h-4" /> Tạo sách mới</Button>
            </div>

            <div class="mt-4 flex items-center gap-2">
              <div class="flex-1 flex items-center h-9 rounded-md border border-input bg-background px-3">
                <Search class="w-4 h-4 text-muted-foreground mr-2" />
                <input class="flex-1 bg-transparent outline-none text-sm" placeholder="Tìm theo tên sách, tác giả, bộ sách…" />
              </div>
              <div class="relative">
                <button class="h-9 px-3 rounded-md border bg-background text-sm flex items-center gap-1.5"
                  :class="manual ? 'border-primary text-primary' : 'border-input'" @click="sortOpen = !sortOpen">
                  <span :class="manual ? 'text-primary/70' : 'text-muted-foreground'">Sắp xếp:</span> {{ sort }} <ChevronDown class="w-4 h-4 opacity-60" />
                </button>
                <div v-if="sortOpen" class="absolute right-0 top-full mt-1 w-56 rounded-lg border border-border bg-popover shadow-lg py-1 z-20">
                  <button v-for="(s, i) in sorts" :key="s" class="w-full flex items-center justify-between px-3 py-1.5 text-sm text-left hover:bg-muted"
                    :class="[s === sort && 'text-primary font-medium', i === 0 && 'border-b border-border mb-1 pb-2']"
                    @click="sort = s; sortOpen = false; if (s === 'Tự sắp xếp') mode = 'manual'">
                    <span>{{ s }}<span v-if="i === 0" class="block text-[11px] font-normal text-muted-foreground">Kéo bìa sách để đổi chỗ</span></span>
                    <Check v-if="s === sort" class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Lọc danh mục + nút Quản lý -->
            <div class="mt-3 flex flex-wrap items-center gap-1.5">
              <button v-for="c in ([['all', 'Tất cả · 9'], ['listening', 'Đang nghe · 4'], ['Kinh doanh', 'Kinh doanh · 5'], ['Kỹ năng', 'Kỹ năng · 2'], ['Tài chính', 'Tài chính · 2'], ['Sức khoẻ', 'Sức khoẻ · 1'], ['__none', 'Chưa phân loại · 1']] as string[][])" :key="c[0]"
                class="h-7 px-3 rounded-full text-xs border"
                :class="filter === c[0] ? 'border-primary bg-primary/10 text-primary font-medium' : 'border-border text-muted-foreground hover:text-foreground'"
                @click="filter = c[0]">{{ c[1] }}</button>
              <button class="h-7 px-2.5 rounded-full text-xs text-muted-foreground hover:text-foreground flex items-center gap-1" @click="setMode('manage'); tab = 'cat'">
                <Settings2 class="w-3.5 h-3.5" /> Quản lý
              </button>
            </div>

            <!-- Dải nhắc khi đang tự sắp xếp -->
            <div v-if="manual" class="mt-3 flex items-center gap-2 rounded-lg bg-primary/5 border border-primary/20 px-3 py-2 text-sm">
              <GripVertical class="w-4 h-4 text-primary" />
              <span class="flex-1">Kéo bìa sách để đổi chỗ. Thứ tự được lưu lại, lần sau mở vẫn giữ nguyên.</span>
              <button class="text-xs text-muted-foreground hover:text-foreground" @click="items = [...initial]">Về thứ tự cũ</button>
            </div>
          </div>

          <div class="flex-1 overflow-auto px-6 pb-6">
            <div class="mt-4 grid grid-cols-5 gap-4">
              <div v-for="b in items" :key="b.id" class="text-left group relative"
                :draggable="manual" :class="[manual && 'cursor-grab', dragId === b.id && 'opacity-40', overId === b.id && dragId !== b.id && 'ring-2 ring-primary ring-offset-4 ring-offset-background rounded-lg']"
                @dragstart="dragId = b.id" @dragover.prevent="overId = b.id" @dragleave="overId = null" @drop="onDrop(b.id)" @dragend="dragId = overId = null">
                <!-- thẻ bộ sách: bìa xếp chồng -->
                <div class="relative" @click="b.kind === 'series' && !manual && setMode('series')">
                  <template v-if="b.kind === 'series'">
                    <div class="absolute inset-0 translate-x-2 -translate-y-1.5 rounded-lg bg-muted-foreground/25"></div>
                    <div class="absolute inset-0 translate-x-1 -translate-y-0.5 rounded-lg bg-muted-foreground/40"></div>
                  </template>
                  <div class="relative aspect-[3/4] rounded-lg shadow-md group-hover:shadow-xl transition" :class="!manual && 'cursor-pointer group-hover:-translate-y-0.5'">
                    <WfBookCover :title="b.title" :author="b.author" size="sm" />
                    <span v-if="b.kind === 'series'" class="absolute bottom-2 right-2 flex items-center gap-1 rounded-full bg-black/60 text-white text-[11px] font-medium px-2 py-0.5"><Layers class="w-3 h-3" /> 3 tập</span>
                    <span v-if="manual" class="absolute top-1.5 left-1.5 h-7 w-7 grid place-items-center rounded-md bg-black/50 text-white"><GripVertical class="w-4 h-4" /></span>
                  </div>
                </div>
                <p class="mt-2 text-sm font-medium truncate">{{ b.title }}</p>
                <p class="text-xs text-muted-foreground truncate">{{ b.author }} · {{ b.meta }}</p>
                <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: b.progress + '%' }"></div></div>
                <p class="mt-1 text-[11px] text-muted-foreground">{{ b.note }}</p>
              </div>
            </div>
            <p v-if="mode === 'shelf'" class="mt-4 text-xs text-muted-foreground">Bấm thẻ "Kinh doanh cho người mới" để mở bộ sách. Một bộ chiếm một chỗ trên kệ, nên cuốn khác không chen vào giữa các tập.</p>
          </div>
        </main>
      </div>

      <!-- ═══ 3. Sửa thông tin: thêm ô Bộ sách + Tập ═══ -->
      <div v-if="mode === 'edit'" class="absolute inset-0 bg-black/40 grid place-items-center z-30">
        <div class="w-[460px] rounded-xl border border-border bg-background shadow-2xl p-5">
          <div class="flex items-start justify-between">
            <h2 class="font-semibold">Sửa thông tin sách</h2>
            <button class="text-muted-foreground" @click="setMode('shelf')"><X class="w-4 h-4" /></button>
          </div>
          <div class="mt-4 space-y-3 text-sm">
            <label class="block">Tên sách<input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" value="Giữ chân khách cũ" /></label>
            <label class="block">Tác giả<input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" value="Nguyễn Văn A" /></label>
            <div>Danh mục <span class="text-muted-foreground">· không bắt buộc</span>
              <div class="mt-1 flex items-center h-9 rounded-md border border-input bg-background px-3"><Tag class="w-4 h-4 text-muted-foreground mr-2" /><span class="flex-1">Kinh doanh</span><ChevronDown class="w-4 h-4 text-muted-foreground" /></div>
            </div>
            <div class="flex gap-3">
              <div class="flex-1 relative">Bộ sách <span class="text-muted-foreground">· không bắt buộc</span>
                <button class="mt-1 w-full flex items-center h-9 rounded-md border bg-background px-3 text-left" :class="seriesOpen ? 'border-ring ring-2 ring-ring/20' : 'border-input'" @click="seriesOpen = !seriesOpen">
                  <Layers class="w-4 h-4 text-muted-foreground mr-2" />
                  <span class="flex-1 truncate" :class="!seriesValue && 'text-muted-foreground'">{{ seriesValue || 'Không thuộc bộ nào' }}</span>
                  <ChevronDown class="w-4 h-4 text-muted-foreground" />
                </button>
                <div v-if="seriesOpen" class="absolute left-0 right-0 top-full mt-1 rounded-lg border border-border bg-popover shadow-lg py-1 z-10">
                  <button class="w-full flex items-center justify-between px-3 py-1.5 text-left hover:bg-muted" @click="seriesValue = ''; seriesOpen = false">
                    <span class="text-muted-foreground">Không thuộc bộ nào</span><Check v-if="!seriesValue" class="w-4 h-4 text-primary" />
                  </button>
                  <button v-for="s in seriesList" :key="s.name" class="w-full flex items-center justify-between px-3 py-1.5 text-left hover:bg-muted" @click="seriesValue = s.name; volNo = s.n + 1; seriesOpen = false">
                    <span :class="seriesValue === s.name && 'text-primary font-medium'">{{ s.name }}</span>
                    <span class="flex items-center gap-2 text-xs text-muted-foreground">{{ s.n }} tập <Check v-if="seriesValue === s.name" class="w-4 h-4 text-primary" /></span>
                  </button>
                  <div class="border-t border-border mt-1 pt-1">
                    <button class="w-full flex items-center gap-2 px-3 py-1.5 text-left text-primary hover:bg-muted"><Plus class="w-4 h-4" /> Tạo bộ sách mới</button>
                  </div>
                </div>
              </div>
              <label class="w-24 block" :class="!seriesValue && 'opacity-40'">Tập số
                <input v-model="volNo" type="number" min="1" :disabled="!seriesValue" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" />
              </label>
            </div>
            <p v-if="seriesValue" class="text-xs text-muted-foreground">Bộ "{{ seriesValue }}" đang có tập 1, 2, 3. Trùng số tập với cuốn khác thì Sano báo để bạn chọn lại.</p>
            <p class="text-xs text-muted-foreground">Bước 1 khi tạo sách cũng có hai ô Bộ sách và Tập số này.</p>
          </div>
          <div class="mt-5 flex justify-end gap-2">
            <Button variant="outline" @click="setMode('shelf')">Huỷ</Button>
            <Button @click="setMode('shelf')">Lưu</Button>
          </div>
        </div>
      </div>

      <!-- ═══ 5. Quản lý danh mục và bộ sách ═══ -->
      <div v-if="mode === 'manage'" class="absolute inset-0 bg-black/40 grid place-items-center z-30">
        <div class="w-[520px] rounded-xl border border-border bg-background shadow-2xl">
          <div class="flex items-start justify-between p-5 pb-0">
            <div>
              <h2 class="font-semibold">Quản lý danh mục và bộ sách</h2>
              <p class="text-xs text-muted-foreground mt-0.5">Đổi tên áp cho mọi cuốn. Xoá thì sách vẫn giữ nguyên.</p>
            </div>
            <button class="text-muted-foreground" @click="setMode('shelf')"><X class="w-4 h-4" /></button>
          </div>
          <div class="px-5 mt-4 flex gap-1 border-b border-border">
            <button v-for="t in ([['cat', 'Danh mục · 4'], ['series', 'Bộ sách · 2']] as ['cat' | 'series', string][])" :key="t[0]"
              class="h-9 px-3 text-sm -mb-px border-b-2" :class="tab === t[0] ? 'border-primary text-foreground font-medium' : 'border-transparent text-muted-foreground'"
              @click="tab = t[0]; confirmDel = null">{{ t[1] }}</button>
          </div>
          <div class="p-3 max-h-[360px] overflow-auto">
            <div v-for="r in (tab === 'cat' ? cats : seriesList)" :key="r.name" class="rounded-md">
              <!-- đang đổi tên -->
              <div v-if="editing === r.name" class="flex items-center gap-2 px-2 py-1.5 bg-muted/40 rounded-md">
                <component :is="tab === 'cat' ? Tag : Layers" class="w-4 h-4 text-muted-foreground" />
                <input v-model="editValue" class="flex-1 h-8 rounded-md border border-ring ring-2 ring-ring/20 bg-background px-2 text-sm" />
                <Button size="sm" @click="r.name = editValue; editing = null">Lưu</Button>
                <Button size="sm" variant="ghost" @click="editing = null">Huỷ</Button>
              </div>
              <!-- xác nhận xoá -->
              <div v-else-if="confirmDel === r.name" class="px-3 py-2.5 rounded-md bg-destructive/5 border border-destructive/20 text-sm">
                <p>Xoá {{ tab === 'cat' ? 'danh mục' : 'bộ sách' }} "<b>{{ r.name }}</b>"?</p>
                <p class="text-xs text-muted-foreground mt-0.5">{{ r.n }} cuốn {{ tab === 'cat' ? 'sẽ về "Chưa phân loại"' : 'sẽ tách ra thành sách lẻ trên kệ' }}. Sách không bị xoá.</p>
                <div class="mt-2 flex justify-end gap-2">
                  <Button size="sm" variant="outline" @click="confirmDel = null">Huỷ</Button>
                  <Button size="sm" variant="destructive" @click="confirmDel = null">Xoá {{ tab === 'cat' ? 'danh mục' : 'bộ sách' }}</Button>
                </div>
              </div>
              <div v-else class="group flex items-center gap-3 px-3 h-11 rounded-md hover:bg-muted/50 text-sm">
                <component :is="tab === 'cat' ? Tag : Layers" class="w-4 h-4 text-muted-foreground" />
                <span class="flex-1">{{ r.name }}</span>
                <span class="text-xs text-muted-foreground w-16 text-right">{{ r.n }} {{ tab === 'cat' ? 'cuốn' : 'tập' }}</span>
                <button class="h-7 w-7 grid place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground" title="Đổi tên" @click="editing = r.name; editValue = r.name"><Pencil class="w-4 h-4" /></button>
                <button class="h-7 w-7 grid place-items-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive" title="Xoá" @click="confirmDel = r.name"><Trash2 class="w-4 h-4" /></button>
                <ChevronRight v-if="tab === 'series'" class="w-4 h-4 text-muted-foreground" />
              </div>
            </div>
          </div>
          <div class="px-5 py-3 border-t border-border flex items-center justify-between text-xs text-muted-foreground">
            <span>{{ tab === 'cat' ? 'Đổi tên trùng một danh mục đã có thì hai danh mục gộp làm một.' : 'Thứ tự các tập đổi ở ô "Tập số" của từng cuốn.' }}</span>
            <Button size="sm" variant="outline" @click="setMode('shelf')">Xong</Button>
          </div>
        </div>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D5 · Bộ sách, tự sắp xếp, quản lý danh mục — bổ sung cho D2. Bộ sách gom các tập thành một thẻ, trong bộ luôn xếp theo số tập. "Tự sắp xếp" là một kiểu trong menu Sắp xếp, kéo bìa để đổi chỗ (thử kéo được ở trạng thái 4). Quản lý mở từ nút "Quản lý" cạnh các nút lọc.
    </p>
  </div>
</template>
