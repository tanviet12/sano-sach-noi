<script setup lang="ts">
// D6 — Nhập sách từ gói zip (người khác gửi / sao lưu). Bám khung D1 (1100×720).
// - Thư viện: nút "Nhập sách" cạnh "Tạo sách mới"; kéo thả file .zip vào cửa sổ cũng được.
// - Chọn file xong: hộp xem trước (bìa, tên, tác giả, giọng, số chương, thời lượng, dung
//   lượng) + nhắc quyền dùng → "Nhập vào thư viện".
// - Trùng sách đã có (cùng mã sách): Thay thế / Giữ cả hai / Huỷ.
// - Đang nhập: tiến độ (giải nén mp3). Xong: báo + "Nghe ngay".
// - Gói lỗi: nói rõ lý do, không nhập gì.
// Kiểm gói trước khi giải nén: đúng cấu trúc (manifest.json, chapters.json, mp3), không có
// đường dẫn ra ngoài thư mục sách, tổng dung lượng giải nén có giới hạn.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API.
import { ref } from 'vue'
import {
  Library, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Search, ChevronDown, Mic, X, Download, Upload,
  Check, AlertTriangle, Loader2, Play, FileArchive, Sun, Moon, Copy, Replace,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

type Mode = 'library' | 'drag' | 'preview' | 'duplicate' | 'importing' | 'done' | 'error' | 'empty'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'preview')
const dark = ref(false)

const books = [
  { title: 'Kỹ năng mềm cho người trẻ', author: 'Nguyễn Văn A', dur: '1 giờ 12 phút', voice: 'Hải Đăng' },
  { title: 'Lãnh đạo cho quản lý mới', author: 'Trần Thị B', dur: '58 phút', voice: 'Ngọc Huyền' },
  { title: 'Tài chính cá nhân cơ bản', author: 'Lê Văn C', dur: '1 giờ 05 phút', voice: 'Thiện Minh' },
  { title: 'Quản lý thời gian hiệu quả', author: 'Nguyễn Văn A', dur: '2 giờ 10 phút', voice: 'Thùy Dung' },
]
const incoming = { title: 'Khởi nghiệp từ số 0', author: 'Phạm Thị D', voice: 'Hải Đăng', chapters: 6, sections: 24, dur: '47 phút', size: '43 MB', file: 'book-khoi-nghiep-tu-so-0.zip' }

const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'create', label: 'Tạo sách mới', icon: FilePlus2 },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
  { key: 'about', label: 'Giới thiệu', icon: Info },
]
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center max-w-[1100px]">
      <button v-for="m in ([
        ['library', '1 · Thư viện: nút Nhập sách'], ['drag', '2 · Kéo thả file zip'], ['preview', '3 · Xem trước'],
        ['duplicate', '4 · Trùng sách'], ['importing', '5 · Đang nhập'], ['done', '6 · Xong'], ['error', '7 · Gói lỗi'], ['empty', '8 · Thư viện trống'],
      ] as [Mode, string][])" :key="m[0]" class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="mode = m[0]">{{ m[1] }}</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

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
            <span v-for="n in nav" :key="n.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
              :class="n.key === 'library' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </span>
          </nav>
          <div class="px-2 mt-2">
            <span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span>
          </div>
          <div class="flex-1"></div>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.8 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <main class="relative flex-1 min-w-0 overflow-hidden p-6">
          <!-- ─── Thư viện trống ─── -->
          <div v-if="mode === 'empty'" class="h-full grid place-items-center">
            <div class="text-center max-w-sm">
              <Library class="w-10 h-10 mx-auto text-muted-foreground" />
              <p class="mt-3 font-medium">Thư viện chưa có sách</p>
              <p class="mt-1 text-sm text-muted-foreground">Tạo sách nói từ file Word, hoặc nhập gói sách (.zip) người khác gửi cho bạn.</p>
              <div class="mt-4 flex justify-center gap-2">
                <Button><FilePlus2 class="w-4 h-4" /> Tạo sách mới</Button>
                <Button variant="outline" @click="mode = 'preview'"><Upload class="w-4 h-4" /> Nhập sách</Button>
              </div>
            </div>
          </div>

          <!-- ─── Thư viện (nền cho các hộp) ─── -->
          <template v-else>
            <div class="flex items-center justify-between">
              <div>
                <h1 class="text-xl font-semibold tracking-tight">Thư viện</h1>
                <p class="text-sm text-muted-foreground">{{ mode === 'done' ? 5 : 4 }} cuốn · 5 giờ 25 phút · lưu ở ~/Sano/Sach</p>
              </div>
              <div class="flex gap-2">
                <!-- MỚI -->
                <Button variant="outline" @click="mode = 'preview'"><Upload class="w-4 h-4" /> Nhập sách</Button>
                <Button><FilePlus2 class="w-4 h-4" /> Tạo sách mới</Button>
              </div>
            </div>
            <div class="mt-4 flex items-center gap-2">
              <div class="flex-1 flex items-center h-9 rounded-md border border-input bg-background px-3 text-sm text-muted-foreground"><Search class="w-4 h-4 mr-2" /> Tìm theo tên sách, tác giả hoặc giọng đọc…</div>
              <span class="h-9 px-3 rounded-md border border-input bg-background text-sm flex items-center gap-1.5"><span class="text-muted-foreground">Sắp xếp:</span> Mới tạo nhất <ChevronDown class="w-4 h-4 text-muted-foreground" /></span>
            </div>
            <div class="mt-5 grid grid-cols-5 gap-4">
              <div v-if="mode === 'done'" class="rounded-lg ring-2 ring-primary ring-offset-2 ring-offset-background">
                <div class="aspect-[3/4] rounded-lg overflow-hidden shadow-md"><WfBookCover :title="incoming.title" :author="incoming.author" /></div>
                <p class="mt-2 text-sm font-medium truncate">{{ incoming.title }}</p>
                <p class="text-xs text-muted-foreground">{{ incoming.author }} · {{ incoming.dur }}</p>
                <p class="text-xs text-muted-foreground flex items-center gap-1"><Mic class="w-3 h-3" /> Giọng {{ incoming.voice }}</p>
              </div>
              <div v-for="b in books" :key="b.title">
                <div class="aspect-[3/4] rounded-lg overflow-hidden shadow-md"><WfBookCover :title="b.title" :author="b.author" /></div>
                <p class="mt-2 text-sm font-medium truncate">{{ b.title }}</p>
                <p class="text-xs text-muted-foreground">{{ b.author }} · {{ b.dur }}</p>
                <p class="text-xs text-muted-foreground flex items-center gap-1"><Mic class="w-3 h-3" /> Giọng {{ b.voice }}</p>
              </div>
            </div>
          </template>

          <!-- ─── 2 · Kéo thả ─── -->
          <div v-if="mode === 'drag'" class="absolute inset-3 rounded-xl border-2 border-dashed border-primary bg-primary/5 backdrop-blur-[1px] grid place-items-center">
            <div class="text-center">
              <FileArchive class="w-10 h-10 mx-auto text-primary" />
              <p class="mt-3 font-medium">Thả gói sách (.zip) để nhập vào thư viện</p>
              <p class="mt-1 text-sm text-muted-foreground">File .docx thì vào Tạo sách mới</p>
            </div>
          </div>

          <!-- ─── 6 · Xong: thông báo ─── -->
          <div v-if="mode === 'done'" class="absolute bottom-5 left-1/2 -translate-x-1/2 flex items-center gap-3 rounded-lg border border-border bg-background shadow-lg pl-4 pr-2 py-2 text-sm">
            <Check class="w-4 h-4 text-rag-green" /> Đã nhập “{{ incoming.title }}”
            <Button size="sm"><Play class="w-4 h-4" /> Nghe ngay</Button>
            <button class="h-8 w-8 grid place-items-center rounded-md text-muted-foreground hover:bg-muted"><X class="w-4 h-4" /></button>
          </div>
        </main>
      </div>

      <!-- ─── Hộp thoại (3 · 4 · 5 · 7) ─── -->
      <div v-if="['preview', 'duplicate', 'importing', 'error'].includes(mode)" class="absolute inset-0 top-9 bg-black/40 grid place-items-center">
        <div class="w-[480px] rounded-xl border border-border bg-background shadow-2xl">
          <div class="flex items-center justify-between px-5 pt-4">
            <h2 class="font-semibold">{{ mode === 'duplicate' ? 'Thư viện đã có cuốn này' : mode === 'error' ? 'Không nhập được gói này' : 'Nhập sách' }}</h2>
            <button v-if="mode !== 'importing'" class="text-muted-foreground hover:text-foreground"><X class="w-4 h-4" /></button>
          </div>

          <!-- 7 · lỗi -->
          <div v-if="mode === 'error'" class="px-5 py-4 text-sm">
            <div class="flex gap-3 rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-3">
              <AlertTriangle class="w-5 h-5 text-destructive shrink-0" />
              <div>
                <p class="font-medium">{{ incoming.file }}</p>
                <p class="mt-1 text-muted-foreground">Gói thiếu file chapters.json nên không phải gói sách của Sano, hoặc file bị hỏng khi gửi. Nhờ người gửi xuất lại bằng nút “Xuất gói zip”.</p>
              </div>
            </div>
            <p class="mt-3 text-xs text-muted-foreground">Chưa có gì được thêm vào thư viện.</p>
          </div>

          <template v-else>
            <!-- xem trước sách -->
            <div class="px-5 pt-4 flex gap-4">
              <div class="w-20 shrink-0 aspect-[3/4] rounded-md overflow-hidden shadow"><WfBookCover :title="incoming.title" :author="incoming.author" size="sm" /></div>
              <div class="min-w-0 text-sm">
                <p class="font-medium text-base leading-snug">{{ incoming.title }}</p>
                <p class="text-muted-foreground">{{ incoming.author }}</p>
                <p class="mt-2 text-muted-foreground flex items-center gap-1"><Mic class="w-3.5 h-3.5" /> Giọng {{ incoming.voice }}</p>
                <p class="text-muted-foreground">{{ incoming.chapters }} chương · {{ incoming.sections }} mục · {{ incoming.dur }} · {{ incoming.size }}</p>
                <p class="mt-1 text-xs text-muted-foreground truncate flex items-center gap-1"><FileArchive class="w-3.5 h-3.5 shrink-0" /> {{ incoming.file }}</p>
              </div>
            </div>

            <!-- 4 · trùng -->
            <div v-if="mode === 'duplicate'" class="px-5 pt-4 space-y-2 text-sm">
              <p class="text-muted-foreground">“{{ incoming.title }}” đã có trong thư viện (tạo ngày 20/9). Bạn muốn làm gì?</p>
              <label class="flex gap-3 rounded-lg border border-primary bg-primary/5 px-3 py-2.5 cursor-pointer">
                <input type="radio" name="dup" checked class="mt-0.5 accent-[hsl(var(--primary))]" />
                <span><span class="font-medium flex items-center gap-1.5"><Copy class="w-3.5 h-3.5" /> Giữ cả hai</span><span class="block text-xs text-muted-foreground">Cuốn mới thêm vào với tên “{{ incoming.title }} (2)”.</span></span>
              </label>
              <label class="flex gap-3 rounded-lg border border-border px-3 py-2.5 cursor-pointer">
                <input type="radio" name="dup" class="mt-0.5 accent-[hsl(var(--primary))]" />
                <span><span class="font-medium flex items-center gap-1.5"><Replace class="w-3.5 h-3.5" /> Thay thế cuốn đang có</span><span class="block text-xs text-muted-foreground">Cuốn cũ chuyển vào Thùng rác (lấy lại được). Vị trí đang nghe giữ nguyên.</span></span>
              </label>
            </div>

            <!-- 5 · đang nhập -->
            <div v-if="mode === 'importing'" class="px-5 pt-5 text-sm">
              <div class="flex items-center justify-between text-muted-foreground"><span class="flex items-center gap-2"><Loader2 class="w-4 h-4 animate-spin" /> Đang giải nén âm thanh… 15/24 mục</span><span class="tabular-nums">62%</span></div>
              <div class="mt-2 h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full w-[62%] bg-primary rounded-full"></div></div>
            </div>

            <!-- 3 · nhắc quyền dùng -->
            <p v-if="mode === 'preview'" class="mx-5 mt-4 rounded-md bg-muted/60 px-3 py-2 text-xs text-muted-foreground">
              Chỉ nhập sách bạn có quyền nghe, ví dụ sách tự làm, sách được tác giả cho phép chia sẻ, hoặc tác phẩm đã hết bản quyền.
            </p>
          </template>

          <div class="flex justify-end gap-2 px-5 py-4">
            <template v-if="mode === 'error'"><Button @click="mode = 'library'">Đóng</Button></template>
            <template v-else-if="mode === 'importing'"><Button variant="outline">Huỷ</Button></template>
            <template v-else>
              <Button variant="outline" @click="mode = 'library'">Huỷ</Button>
              <Button @click="mode = mode === 'preview' ? 'importing' : 'importing'"><Download class="w-4 h-4" /> {{ mode === 'duplicate' ? 'Tiếp tục' : 'Nhập vào thư viện' }}</Button>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
