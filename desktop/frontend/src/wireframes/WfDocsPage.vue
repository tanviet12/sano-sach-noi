<script setup lang="ts">
// P4 — Trang con mẫu "Hướng dẫn" (wireframe), bám khung trang tài liệu mặc định của VitePress:
// thanh bên trái (sidebar theo nhóm) + nội dung markdown + cột "Trên trang này" (aside outline)
// + "Sửa trang này trên GitHub" / cập nhật lần cuối / trang trước–sau.
// Điện thoại: thanh "☰ Menu" + "Trên trang này" dưới navbar, bấm Menu mở thanh bên dạng ngăn kéo.
// Nội dung mẫu: rút gọn từ docs/nghe-tren-dien-thoai.md. Wireframe tĩnh, KHÔNG gọi API.
// Mở ở bản dev: ?wireframe=docs (thêm &theme=dark để chụp nhanh).
import { ref, watch } from 'vue'
import { ChevronDown, ChevronLeft, ChevronRight, Info, Lightbulb, Menu, SquarePen, TriangleAlert, X } from 'lucide-vue-next'
import WfSiteNav from './WfSiteNav.vue'
import WfSiteReviewBar from './WfSiteReviewBar.vue'
import bpLibrary from '../../../../docs/images/nghe-tren-dien-thoai/2-thu-vien.jpg'
import bpPlayer from '../../../../docs/images/nghe-tren-dien-thoai/3-trinh-phat.jpg'
import bpChapters from '../../../../docs/images/nghe-tren-dien-thoai/4-muc-luc-chuong.jpg'

const REPO = 'https://github.com/tanviet12/sano-sach-noi'

const q = new URLSearchParams(window.location.search)
const dark = ref(q.get('theme') === 'dark')
watch(dark, (v) => document.documentElement.classList.toggle('dark', v), { immediate: true })
const notes = ref(q.get('notes') !== '0')
const drawer = ref(false) // thanh bên dạng ngăn kéo trên điện thoại
const outlineOpen = ref(false)

// Sidebar VitePress: nhóm + mục. Trang đang xem: Nghe trên điện thoại.
const sidebar = [
  { text: 'Bắt đầu', items: ['Cài đặt', 'Tạo sách đầu tiên'] },
  { text: 'Dùng Sano', items: ['Làm mượt tài liệu', 'Nghe trên điện thoại'] },
  { text: 'Hỗ trợ', items: ['Câu hỏi thường gặp', 'Gỡ cài đặt', 'Điều khoản sử dụng', 'Nhật ký thay đổi'] },
]
const active = 'Nghe trên điện thoại'

const outline = [
  { id: 'xuat-m4b', text: 'Xuất file M4B', level: 2 },
  { id: 'iphone', text: 'iPhone, iPad', level: 2 },
  { id: 'bookplayer', text: 'Cách nhanh: BookPlayer', level: 3 },
  { id: 'apple-books', text: 'App Sách của Apple', level: 3 },
  { id: 'android', text: 'Android', level: 2 },
  { id: 'truc-trac', text: 'Nếu có trục trặc', level: 2 },
]
const activeOutline = ref('iphone')

const noteCls = 'inline-flex items-center gap-1 rounded border border-dashed border-amber-500/70 bg-amber-50 px-1.5 py-0.5 text-[11px] font-medium text-amber-800 dark:bg-amber-950/40 dark:text-amber-300'
const h2 = 'mt-10 scroll-mt-24 border-t border-border pt-6 text-xl font-semibold tracking-tight'
const h3 = 'mt-6 scroll-mt-24 text-base font-semibold'
</script>

<template>
  <div class="min-h-dvh bg-background text-foreground">
    <WfSiteReviewBar page="docs" :dark="dark" :notes="notes" @toggle-dark="dark = !dark" @toggle-notes="notes = !notes" />
    <WfSiteNav current="guide" :dark="dark" @toggle-dark="dark = !dark" />

    <!-- Thanh phụ trên điện thoại (VitePress: VPLocalNav) -->
    <div class="sticky top-16 z-20 flex items-center justify-between border-b border-border bg-background/95 px-4 py-2 text-sm lg:hidden">
      <button type="button" class="flex items-center gap-1.5 rounded-md px-1 py-1 focus-ring" @click="drawer = true"><Menu class="w-4 h-4" /> Menu</button>
      <div class="relative">
        <button type="button" class="flex items-center gap-1 rounded-md px-1 py-1 text-muted-foreground focus-ring" :aria-expanded="outlineOpen" @click="outlineOpen = !outlineOpen">Trên trang này <ChevronDown class="w-3.5 h-3.5" /></button>
        <div v-if="outlineOpen" class="absolute right-0 top-full mt-1 w-60 rounded-lg border border-border bg-popover p-2 shadow-lg">
          <a v-for="o in outline" :key="o.id" :href="'#' + o.id" class="block rounded px-2 py-1.5 text-sm hover:bg-muted" :class="o.level === 3 && 'pl-5 text-muted-foreground'" @click="outlineOpen = false">{{ o.text }}</a>
        </div>
      </div>
    </div>

    <!-- Ngăn kéo thanh bên trên điện thoại -->
    <div v-if="drawer" class="fixed inset-0 z-40 lg:hidden">
      <div class="absolute inset-0 bg-black/40" @click="drawer = false" />
      <aside class="absolute inset-y-0 left-0 w-72 overflow-auto border-r border-border bg-background p-5">
        <div class="mb-4 flex items-center justify-between">
          <span class="flex items-center gap-2 font-semibold"><img src="@/assets/logo.svg" alt="" class="h-7 w-7 rounded-lg" /> Sano</span>
          <button type="button" class="grid h-8 w-8 place-items-center rounded-md hover:bg-muted focus-ring" aria-label="Đóng menu" @click="drawer = false"><X class="w-4 h-4" /></button>
        </div>
        <nav v-for="g in sidebar" :key="g.text" class="mb-5">
          <p class="mb-1.5 text-sm font-semibold">{{ g.text }}</p>
          <a v-for="it in g.items" :key="it" href="#" class="block py-1.5 text-sm" :class="it === active ? 'font-medium text-primary' : 'text-muted-foreground'" :aria-current="it === active ? 'page' : undefined">{{ it }}</a>
        </nav>
      </aside>
    </div>

    <div class="mx-auto flex max-w-7xl">
      <!-- Thanh bên trái (VitePress: themeConfig.sidebar) -->
      <aside class="sticky top-16 hidden h-[calc(100dvh-4rem)] w-64 shrink-0 overflow-auto border-r border-border bg-muted/30 px-6 py-8 lg:block">
        <p v-if="notes" :class="noteCls" class="mb-4">VitePress có sẵn: themeConfig.sidebar</p>
        <nav v-for="g in sidebar" :key="g.text" class="mb-6" :aria-label="g.text">
          <p class="mb-2 text-sm font-semibold">{{ g.text }}</p>
          <a v-for="it in g.items" :key="it" href="#"
            class="block border-l-2 py-1.5 pl-3 text-sm transition-colors"
            :class="it === active ? 'border-primary font-medium text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
            :aria-current="it === active ? 'page' : undefined">{{ it }}</a>
        </nav>
      </aside>

      <!-- Nội dung -->
      <main class="min-w-0 flex-1 px-4 py-8 sm:px-8 lg:px-12">
        <article class="mx-auto max-w-3xl text-[15px] leading-7">
          <p v-if="notes" :class="noteCls" class="mb-4">Nội dung: file markdown docs/huong-dan/nghe-tren-dien-thoai.md (chuyển từ docs/ hiện có)</p>
          <p class="text-sm font-medium text-primary">Dùng Sano</p>
          <h1 class="mt-1 text-3xl font-semibold tracking-tight">Nghe trên điện thoại, trên xe</h1>
          <p class="mt-4 text-muted-foreground">
            Sano xuất cả cuốn thành <strong class="text-foreground">một file <code class="rounded bg-muted px-1.5 py-0.5 text-[13px]">.m4b</code></strong>: có mục lục chương, tên sách, tác giả và ảnh bìa. Chép file này vào điện thoại là nghe được, không cần mạng.
          </p>

          <h2 id="xuat-m4b" :class="h2">Xuất file M4B</h2>
          <p class="mt-3">Mở cuốn sách trong <strong>Thư viện</strong> → <strong>Xuất M4B</strong> → chọn nơi lưu. Xong, Sano mở thư mục và chọn sẵn file.</p>
          <!-- Khối tip (VitePress: ::: tip) -->
          <div class="mt-4 rounded-lg border border-primary/20 bg-primary/5 px-4 py-3">
            <p class="flex items-center gap-1.5 text-sm font-semibold text-primary"><Lightbulb class="w-4 h-4" /> Mẹo</p>
            <p class="mt-1 text-sm">Dung lượng khoảng 29 MB cho mỗi giờ nghe, đủ rõ cho giọng đọc.</p>
          </div>
          <p class="mt-4">Dùng dòng lệnh:</p>
          <pre class="mt-2 overflow-x-auto rounded-lg bg-muted px-4 py-3 text-[13px] leading-6"><code>sano-docx2tts --m4b-from-dir ~/Sano/Sach/&lt;tên-sách&gt; --m4b ~/Downloads/sach.m4b</code></pre>

          <h2 id="iphone" :class="h2">iPhone, iPad</h2>
          <!-- Khối cảnh báo (VitePress: ::: warning) -->
          <div class="mt-4 rounded-lg border border-amber-500/30 bg-amber-50 px-4 py-3 dark:bg-amber-950/30">
            <p class="flex items-center gap-1.5 text-sm font-semibold text-amber-700 dark:text-amber-300"><TriangleAlert class="w-4 h-4" /> Lưu ý</p>
            <p class="mt-1 text-sm">App Sách trên iPhone không nhận file <code>.m4b</code> gửi qua AirDrop hay app Tệp. Dùng một trong hai cách dưới đây.</p>
          </div>
          <h3 id="bookplayer" :class="h3">Cách nhanh: BookPlayer</h3>
          <ol class="mt-2 list-decimal space-y-1 pl-6">
            <li>Cài <strong>BookPlayer</strong> từ App Store (miễn phí, mã nguồn mở).</li>
            <li>Đưa file <code>.m4b</code> vào iPhone: AirDrop từ Mac, hoặc iCloud Drive / Google Drive.</li>
            <li>Mở BookPlayer → bấm <strong>+</strong> → <strong>Import files</strong> → chọn file.</li>
          </ol>
          <div class="mt-5 grid grid-cols-3 gap-3 sm:max-w-md">
            <img v-for="(img, i) in [bpLibrary, bpPlayer, bpChapters]" :key="img" :src="img" :alt="['Thư viện', 'Đang nghe', 'Mục lục chương'][i]" class="w-full rounded-lg border border-border" />
          </div>
          <p class="mt-2 text-sm text-muted-foreground">Thư viện · Đang nghe · Mục lục chương lấy từ file M4B do Sano xuất. <span v-if="notes" :class="noteCls">Bấm ảnh để phóng to (medium-zoom như khuôn tài liệu cũ)</span></p>
          <h3 id="apple-books" :class="h3">App Sách của Apple</h3>
          <p class="mt-2">Cắm iPhone vào Mac → Finder → chọn iPhone → tab <strong>Sách nói</strong> → kéo file vào → <strong>Đồng bộ</strong>. Windows dùng app Apple Devices.</p>

          <h2 id="android" :class="h2">Android</h2>
          <p class="mt-3">Chép file vào thư mục riêng, ví dụ <code>Audiobooks/Tên sách/</code>, rồi mở bằng app nghe sách nói đọc được M4B. App tự nhận mục lục chương và bìa.</p>

          <h2 id="truc-trac" :class="h2">Nếu có trục trặc</h2>
          <!-- Khối info (VitePress: ::: info) -->
          <div class="mt-4 rounded-lg border border-border bg-muted/50 px-4 py-3">
            <p class="flex items-center gap-1.5 text-sm font-semibold"><Info class="w-4 h-4" /> Không thấy mục lục chương?</p>
            <p class="mt-1 text-sm text-muted-foreground">Một số trình phát nhạc thông thường không đọc mục lục sách nói. Dùng app sách nói như ở trên.</p>
          </div>

          <!-- Chân bài (VitePress: editLink + lastUpdated + prev/next) -->
          <div class="mt-12 flex flex-wrap items-center justify-between gap-2 text-sm text-muted-foreground">
            <a :href="REPO + '/edit/main/docs/huong-dan/nghe-tren-dien-thoai.md'" target="_blank" rel="noopener" class="flex items-center gap-1.5 text-primary hover:underline"><SquarePen class="w-4 h-4" /> Sửa trang này trên GitHub</a>
            <span>Cập nhật lần cuối: 24/09/2026</span>
          </div>
          <div class="mt-6 grid gap-3 border-t border-border pt-6 sm:grid-cols-2">
            <a href="#" class="rounded-lg border border-border px-4 py-3 hover:border-primary focus-ring">
              <span class="flex items-center gap-1 text-xs text-muted-foreground"><ChevronLeft class="w-3.5 h-3.5" /> Trang trước</span>
              <span class="font-medium text-primary">Làm mượt tài liệu</span>
            </a>
            <a href="#" class="rounded-lg border border-border px-4 py-3 text-right hover:border-primary focus-ring">
              <span class="flex items-center justify-end gap-1 text-xs text-muted-foreground">Trang sau <ChevronRight class="w-3.5 h-3.5" /></span>
              <span class="font-medium text-primary">Câu hỏi thường gặp</span>
            </a>
          </div>
        </article>
      </main>

      <!-- Cột "Trên trang này" (VitePress: aside outline, chỉ màn rộng) -->
      <aside class="sticky top-16 hidden h-[calc(100dvh-4rem)] w-56 shrink-0 overflow-auto py-8 pr-6 xl:block">
        <p class="mb-2 text-xs font-semibold">Trên trang này</p>
        <nav class="border-l border-border text-sm">
          <a v-for="o in outline" :key="o.id" :href="'#' + o.id"
            class="-ml-px block border-l-2 py-1 transition-colors"
            :class="[o.level === 3 ? 'pl-6' : 'pl-3', activeOutline === o.id ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground']"
            @click="activeOutline = o.id">{{ o.text }}</a>
        </nav>
      </aside>
    </div>
  </div>
</template>
