<script setup lang="ts">
// Trang demo riêng /demo (wireframe) — link để gửi người khác nghe thử trước khi cài:
// tủ sách 5 cuốn mẫu (mỗi cuốn một giọng), một đoạn đọc bằng 12 giọng, trước/sau chuẩn hoá,
// lời mời tải. Cuối trang: phần đổi ở trang chủ (dải giọng + nút Nghe thêm demo).
// Wireframe tĩnh, audio giả lập bằng bộ đếm, KHÔNG gọi API. Mở ở bản dev: ?wireframe=demo
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ArrowRight, AudioLines, BookOpen, Download, Headphones, Pause, Play } from 'lucide-vue-next'
import BookCover from '@/components/sano/BookCover.vue'
import WfSiteNav from './WfSiteNav.vue'
import WfSiteReviewBar from './WfSiteReviewBar.vue'

const q = new URLSearchParams(window.location.search)
const dark = ref(q.get('theme') === 'dark')
watch(dark, (v) => document.documentElement.classList.toggle('dark', v), { immediate: true })
const notes = ref(q.get('notes') !== '0')
const noteCls = 'inline-flex items-center gap-1 rounded border border-dashed border-amber-500/70 bg-amber-50 px-1.5 py-0.5 text-[11px] font-medium text-amber-800 dark:bg-amber-950/40 dark:text-amber-300'

// 5 cuốn mẫu tự viết (docs/demo-books), mỗi cuốn một giọng; mỗi cuốn 3 chương đầu.
const books = [
  { title: 'Kỹ năng mềm cho người trẻ', voice: 'Thiện Minh', vdesc: 'nam · Bắc · kể chuyện', chapters: [['Lắng nghe chủ động', 31], ['Giao tiếp hiệu quả', 22], ['Làm việc nhóm', 23]] },
  { title: 'Tâm lý tích cực', voice: 'Trúc Ly', vdesc: 'nữ · Bắc · tự nhiên', chapters: [['Tư duy phát triển', 34], ['Vượt qua khó khăn', 30], ['Lòng biết ơn', 28]] },
  { title: 'Khởi nghiệp từ số 0', voice: 'Thái Sơn', vdesc: 'nam · Nam · kể chuyện', chapters: [['Bắt đầu từ vấn đề', 33], ['Kiểm chứng ý tưởng', 29], ['Sản phẩm đầu tiên', 27]] },
  { title: 'Tài chính cá nhân cơ bản', voice: 'Thục Đoan', vdesc: 'nữ · Nam · kể chuyện', chapters: [['Quản lý chi tiêu', 30], ['Tiết kiệm', 26], ['Đầu tư', 31]] },
  { title: 'Lãnh đạo cho quản lý mới', voice: 'Ngọc Trân', vdesc: 'nữ · Trung · tự nhiên', chapters: [['Từ người làm đến người dẫn dắt', 32], ['Phản hồi', 27], ['Ra quyết định', 25]] },
] as const
const bookIdx = ref(0)
const chIdx = ref(0)
const book = computed(() => books[bookIdx.value])
const chapter = computed(() => book.value.chapters[chIdx.value])

// 12 giọng đọc cùng một đoạn, nhóm theo miền
const regions = [
  { name: 'Miền Bắc', voices: [['Thiện Minh', 'nam · kể chuyện'], ['Trúc Ly', 'nữ · tự nhiên'], ['Mai Anh', 'nữ · tin tức'], ['Hải Đăng', 'nam · tự nhiên'], ['Thiền Tâm Đức', 'nam · kể chuyện'], ['Ngọc Huyền', 'nữ · tự nhiên']] },
  { name: 'Miền Trung', voices: [['Quang Sơn', 'nam · tự nhiên'], ['Ngọc Trân', 'nữ · tự nhiên']] },
  { name: 'Miền Nam', voices: [['Thái Sơn', 'nam · kể chuyện'], ['Thùy Dung', 'nữ · tin tức'], ['Thục Đoan', 'nữ · kể chuyện'], ['Mỹ Duyên', 'nữ · đọc truyện']] },
]
const passage = 'Tư duy phát triển là niềm tin rằng năng lực của con người không cố định, mà có thể cải thiện qua nỗ lực, học hỏi và rèn luyện…'

// Một trình phát chung cho cả trang: key = nguồn đang phát ('book' hoặc tên giọng)
const playingKey = ref('')
const time = ref(0)
let timer: number | undefined
function stop() {
  playingKey.value = ''
  clearInterval(timer)
}
function play(key: string, len: number) {
  if (playingKey.value === key) return stop()
  stop()
  playingKey.value = key
  time.value = 0
  timer = window.setInterval(() => {
    time.value += 0.5
    if (time.value >= len) stop()
  }, 500)
}
function pickBook(i: number) {
  bookIdx.value = i
  chIdx.value = 0
  play('book', book.value.chapters[0][1])
}
function pickChapter(i: number) {
  chIdx.value = i
  play('book', chapter.value[1])
}
onBeforeUnmount(stop)
const fmt = (s: number) => `${Math.floor(s / 60)}:${String(Math.floor(s % 60)).padStart(2, '0')}`
const pct = computed(() => (playingKey.value === 'book' ? (time.value / chapter.value[1]) * 100 : 0))
</script>

<template>
  <div class="min-h-dvh bg-background text-foreground">
    <WfSiteReviewBar page="demo" :dark="dark" :notes="notes" @toggle-dark="dark = !dark" @toggle-notes="notes = !notes" />
    <WfSiteNav current="home" :dark="dark" @toggle-dark="dark = !dark" />

    <main>
      <!-- 1. Mở đầu -->
      <section class="border-b border-border bg-gradient-to-b from-primary/5 to-transparent">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <p v-if="notes" :class="noteCls" class="mb-3">Trang mới docs/demo.md (layout: home) · link gửi đi: tanviet12.github.io/sano-sach-noi/demo</p>
          <p class="text-sm font-medium text-primary">Nghe thử Sano</p>
          <h1 class="mt-2 max-w-3xl text-3xl font-semibold tracking-tight sm:text-5xl">Sách nói tạo bằng AI, <span class="text-primary">ngay trên máy tính</span></h1>
          <p class="mt-4 max-w-2xl text-lg text-muted-foreground">
            5 cuốn sách mẫu, 12 giọng đọc tiếng Việt. Tất cả tạo từ file Word bằng Sano, không cần phòng thu, không gửi tài liệu lên mạng.
          </p>
          <div class="mt-6 flex flex-col gap-3 sm:flex-row">
            <a href="#tu-sach" class="inline-flex items-center justify-center gap-2 rounded-lg bg-primary px-5 py-3 font-medium text-primary-foreground shadow focus-ring"><Headphones class="w-5 h-5" /> Nghe ngay</a>
            <a href="#tai" class="inline-flex items-center justify-center gap-2 rounded-lg border border-border px-5 py-3 font-medium hover:bg-muted focus-ring"><Download class="w-4 h-4" /> Tải Sano miễn phí</a>
          </div>
        </div>
      </section>

      <!-- 2. Tủ sách nghe thử -->
      <section id="tu-sach" class="border-b border-border">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <p v-if="notes" :class="noteCls" class="mb-3">Component tự làm DemoShelf.vue · audio mp3 thật trong docs/public/audio/demo/ (3 chương đầu mỗi cuốn)</p>
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Tủ sách nghe thử</h2>
          <p class="mt-2 max-w-2xl text-muted-foreground">Chọn một cuốn, mỗi cuốn đọc bằng một giọng khác nhau. Bìa sách do Sano tự vẽ.</p>

          <!-- Chọn sách: 5 bìa một hàng (điện thoại: vuốt ngang) -->
          <div class="mt-8 -mx-4 flex snap-x gap-3 overflow-x-auto px-4 pb-2 sm:mx-0 sm:grid sm:grid-cols-5 sm:overflow-visible sm:px-0">
            <button v-for="(b, i) in books" :key="b.title" type="button" class="w-32 shrink-0 snap-start rounded-lg border p-2 text-left transition-colors focus-ring sm:w-auto"
              :class="i === bookIdx ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/60'" @click="pickBook(i)">
              <BookCover :title="b.title" author="Sách mẫu Sano" size="sm" class="aspect-[3/4] w-full" />
              <p class="mt-2 line-clamp-2 text-xs font-medium leading-snug">{{ b.title }}</p>
              <p class="flex items-center gap-1 text-[11px] text-muted-foreground"><AudioLines v-if="i === bookIdx && playingKey === 'book'" class="w-3 h-3 text-primary" />{{ b.voice }}</p>
            </button>
          </div>

          <!-- Trình phát cuốn đang chọn -->
          <div class="mt-6 rounded-xl border border-border bg-card p-5 shadow-sm">
            <div class="grid gap-5 md:grid-cols-[auto_1fr_1fr] md:items-start">
              <BookCover :title="book.title" author="Sách mẫu Sano" class="aspect-[3/4] w-28 shadow-lg sm:w-32" />
              <div class="min-w-0">
                <p class="text-lg font-semibold leading-snug">{{ book.title }}</p>
                <p class="mt-1 text-sm text-muted-foreground">Giọng <span class="font-medium text-foreground">{{ book.voice }}</span> · {{ book.vdesc }}</p>
                <p class="mt-4 truncate text-sm text-primary">Chương {{ chIdx + 1 }}. {{ chapter[0] }}</p>
                <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-muted"><div class="h-full rounded-full bg-primary transition-all duration-300" :style="{ width: pct + '%' }" /></div>
                <div class="mt-1 flex justify-between text-[11px] tabular-nums text-muted-foreground"><span>{{ fmt(playingKey === 'book' ? time : 0) }}</span><span>{{ fmt(chapter[1]) }}</span></div>
                <button type="button" :aria-label="playingKey === 'book' ? 'Dừng' : 'Phát'" class="mt-2 grid h-12 w-12 place-items-center rounded-full bg-primary text-primary-foreground shadow focus-ring" @click="play('book', chapter[1])">
                  <Pause v-if="playingKey === 'book'" class="w-5 h-5" /><Play v-else class="ml-0.5 w-5 h-5" />
                </button>
              </div>
              <div class="min-w-0 divide-y divide-border rounded-lg border border-border">
                <button v-for="(c, i) in book.chapters" :key="c[0]" type="button" class="flex w-full items-center justify-between gap-2 px-3 py-2.5 text-left text-sm hover:bg-muted/60 focus-ring"
                  :class="i === chIdx ? 'bg-primary/10 font-medium text-primary' : ''" @click="pickChapter(i)">
                  <span class="flex min-w-0 items-center gap-2">
                    <AudioLines v-if="i === chIdx" class="w-3.5 h-3.5 shrink-0" /><span v-else class="w-3.5 shrink-0 text-center text-xs text-muted-foreground">{{ i + 1 }}</span>
                    <span class="truncate">{{ c[0] }}</span>
                  </span>
                  <span class="shrink-0 text-xs tabular-nums text-muted-foreground">{{ fmt(c[1]) }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 3. Một đoạn, 12 giọng -->
      <section class="border-b border-border bg-muted/30">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <p v-if="notes" :class="noteCls" class="mb-3">Component tự làm VoiceGallery.vue · 12 mp3 ~15 giây, cùng một đoạn</p>
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Một đoạn, 12 giọng</h2>
          <p class="mt-2 max-w-2xl text-muted-foreground">Cùng một đoạn văn, bấm từng giọng để so sánh. Sano có 25 giọng, chọn được ngay khi tạo sách.</p>
          <blockquote class="mt-6 max-w-3xl border-l-4 border-primary/60 pl-4 text-muted-foreground italic">{{ passage }}</blockquote>
          <div class="mt-8 space-y-6">
            <div v-for="r in regions" :key="r.name">
              <p class="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">{{ r.name }}</p>
              <div class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
                <button v-for="v in r.voices" :key="v[0]" type="button" class="flex items-center gap-3 rounded-lg border bg-card px-3 py-2.5 text-left focus-ring transition-colors"
                  :class="playingKey === v[0] ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/60'" @click="play(v[0], 15)">
                  <span class="grid h-9 w-9 shrink-0 place-items-center rounded-full" :class="playingKey === v[0] ? 'bg-primary text-primary-foreground' : 'bg-primary/10 text-primary'">
                    <Pause v-if="playingKey === v[0]" class="w-4 h-4" /><Play v-else class="ml-0.5 w-4 h-4" />
                  </span>
                  <span class="min-w-0 flex-1"><span class="block font-medium">{{ v[0] }}</span><span class="block text-xs text-muted-foreground">{{ v[1] }}</span></span>
                  <span class="text-xs tabular-nums text-muted-foreground">{{ playingKey === v[0] ? fmt(time) : '0:15' }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 4. Trước / sau chuẩn hoá -->
      <section class="border-b border-border">
        <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
          <p v-if="notes" :class="noteCls" class="mb-3">Dùng lại NormalizeCompare.vue của trang chủ, không đổi</p>
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Đọc tự nhiên, không đọc máy móc</h2>
          <p class="mt-2 max-w-2xl text-muted-foreground">Số, chữ viết tắt, ký hiệu được đọc thành lời như người thật đọc.</p>
          <div class="mt-6 grid h-40 max-w-xl place-items-center rounded-xl border border-dashed border-border text-sm text-muted-foreground">[ Khối Trước / sau khi chuẩn hoá văn nói ]</div>
        </div>
      </section>

      <!-- 5. Mời tải -->
      <section id="tai" class="bg-primary/5">
        <div class="mx-auto max-w-6xl px-4 py-14 text-center sm:px-6">
          <h2 class="text-2xl font-semibold tracking-tight sm:text-3xl">Làm sách nói từ tài liệu của bạn</h2>
          <p class="mx-auto mt-2 max-w-xl text-muted-foreground">Miễn phí, mã nguồn mở, chạy trên Windows, macOS, Linux. Nghe trên máy tính, trên điện thoại và trên ô tô qua CarPlay, Android Auto.</p>
          <div class="mt-6 flex flex-col justify-center gap-3 sm:flex-row">
            <a href="#" class="inline-flex items-center justify-center gap-2 rounded-lg bg-primary px-5 py-3 font-medium text-primary-foreground shadow focus-ring"><Download class="w-5 h-5" /> Tải Sano</a>
            <a href="#" class="inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-background px-5 py-3 font-medium hover:bg-muted focus-ring"><BookOpen class="w-4 h-4" /> Xem hướng dẫn</a>
          </div>
        </div>
      </section>

      <!-- Phần đổi ở trang chủ -->
      <section class="border-t-4 border-dashed border-amber-500/60 bg-amber-50/40 dark:bg-amber-950/10">
        <div class="mx-auto max-w-6xl px-4 py-10 sm:px-6">
          <p :class="noteCls">Thay đổi ở TRANG CHỦ, mục "Nghe thử": thêm dải này ngay dưới trình phát hiện có</p>
          <div class="mt-4 rounded-xl border border-border bg-card p-4 shadow-sm">
            <div class="flex flex-wrap items-center gap-2">
              <span class="mr-1 text-sm font-medium">Nghe giọng khác:</span>
              <button v-for="v in ['Trúc Ly', 'Thái Sơn', 'Ngọc Trân', 'Thục Đoan', 'Hải Đăng']" :key="v" type="button"
                class="inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-sm focus-ring" :class="playingKey === v ? 'border-primary bg-primary/10 text-primary' : 'border-border hover:bg-muted'" @click="play(v, 15)">
                <Pause v-if="playingKey === v" class="w-3.5 h-3.5" /><Play v-else class="w-3.5 h-3.5" /> {{ v }}
              </button>
              <a href="?wireframe=demo" class="ml-auto inline-flex items-center gap-1 text-sm font-medium text-primary hover:underline">Nghe thêm 5 cuốn, 12 giọng <ArrowRight class="w-4 h-4" /></a>
            </div>
          </div>
          <p class="mt-3 text-xs text-muted-foreground">Menu trên cùng thêm mục "Nghe thử" dẫn tới /demo.</p>
        </div>
      </section>
    </main>
  </div>
</template>
