<script setup lang="ts">
// B6 Render thật: tiến độ từ sự kiện Go (tiểu mục x/y, ký tự, thời gian còn lại),
// huỷ được. Xong → sách đã nằm trong thư viện ~/Sano/Sach/<slug>/.
import { computed, ref, watch } from 'vue'
import { BookOpen, Check, ChevronRight, Download, FolderOpen, Loader2, Package } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import M4BProgress from '../../components/M4BProgress.vue'
import { book, errText, openBookFolder, revealBookZip, type BookDetail } from '../../lib/backend'
import { INTRO_LABEL, cancelRender, go, introText, openBook, remainMin, renderPct, state } from '../../lib/store'
import { m4bBusy, startM4B, useM4B } from '../../lib/m4b'

useM4B()

interface PlanChapter {
  title: string
  stems: string[] // stem đầu ra chNN-secNN
  chars: number[]
}

// Kế hoạch lấy theo lựa chọn lúc bấm render: chương còn tiểu mục được chọn,
// đánh số lại từ 1 (lời mở đầu chiếm ch01) — khớp cách bookmaker đặt tên file.
function buildPlan(): PlanChapter[] {
  const plan: PlanChapter[] = []
  const intro = introText() !== ''
  if (intro) plan.push({ title: INTRO_LABEL, stems: ['ch01-sec01'], chars: [introText().length] })
  const charsOf: Record<string, number> = {}
  for (const ch of state.outline?.chapters ?? []) for (const s of ch.sections) charsOf[s.stem] = s.chars
  for (const c of state.toc) {
    if (!c.on) continue
    const stems = c.sections.length ? c.sections.filter((s) => s.on).map((s) => s.stem) : c.stems
    if (!stems.length) continue
    const n = plan.length + 1
    plan.push({
      title: c.title,
      stems: stems.map((_, j) => `ch${String(n).padStart(2, '0')}-sec${String(j + 1).padStart(2, '0')}`),
      chars: stems.map((s) => charsOf[s] ?? 1),
    })
  }
  return plan
}
const plan = ref<PlanChapter[]>(buildPlan())
watch(() => state.render?.title, () => (plan.value = buildPlan()))

// Bộ đọc làm tuần tự theo tên file → "done" tiểu mục đầu tiên theo thứ tự tên là đã xong.
const doneStems = computed(() => {
  const all = plan.value.flatMap((c) => c.stems).sort()
  return new Set(all.slice(0, state.render?.progress?.done ?? 0))
})
const chapters = computed(() =>
  plan.value.map((c) => {
    const total = c.chars.reduce((a, b) => a + b, 0) || 1
    const done = c.stems.reduce((n, s, i) => n + (doneStems.value.has(s) ? c.chars[i] : 0), 0)
    return { title: c.title, pct: Math.round((done / total) * 100) }
  }),
)
// Chương đang đọc = chương đầu tiên chưa xong → luôn có vòng xoay, kể cả lúc còn 0%.
const activeIdx = computed(() => chapters.value.findIndex((c) => c.pct < 100))
const p = computed(() => state.render?.progress)
// Chưa có tiểu mục nào xong: bộ đọc đang khởi động/nạp mô hình (có thể mất vài chục giây).
const starting = computed(() => (p.value?.done ?? 0) === 0 && p.value?.phase !== 'package')

// Màn đã xong: đọc cuốn vừa tạo từ thư viện để hiện thời lượng thật.
const made = ref<BookDetail | null>(null)
const actionError = ref('')
watch(
  () => state.render?.slug,
  async (slug) => {
    made.value = null
    if (!slug) return
    try {
      made.value = await book(slug)
    } catch (e) {
      actionError.value = errText(e)
    }
  },
  { immediate: true },
)
const listenMin = computed(() => Math.max(1, Math.round((made.value?.durationSec ?? 0) / 60)))

async function act(fn: (slug: string) => Promise<void>) {
  actionError.value = ''
  try {
    if (state.render?.slug) await fn(state.render.slug)
  } catch (e) {
    actionError.value = errText(e)
  }
}
</script>

<template>
  <div class="max-w-2xl">
    <template v-if="state.render?.error">
      <h1 class="text-xl font-semibold tracking-tight">Render không thành công</h1>
      <pre class="mt-3 max-h-64 overflow-auto whitespace-pre-wrap rounded-md border border-destructive/40 bg-destructive/5 p-3 text-xs">{{ state.render.error }}</pre>
      <div class="mt-5 flex gap-2">
        <Button variant="outline" @click="state.render = null; state.step = 5">Quay lại nghe thử</Button>
      </div>
    </template>

    <template v-else-if="!state.render?.done">
      <h1 class="text-xl font-semibold tracking-tight">Đang render cả cuốn</h1>
      <p class="text-sm text-muted-foreground">Có thể dùng máy bình thường hoặc thu nhỏ cửa sổ. Nếu đóng Sano, bạn sẽ được hỏi có dừng render không.</p>
      <div class="mt-6">
        <div class="flex justify-between text-sm">
          <span class="font-medium flex items-center gap-1.5"><Loader2 class="w-3.5 h-3.5 animate-spin text-primary" />{{ renderPct }}%<span class="ml-2 font-normal text-muted-foreground">· tiểu mục {{ p?.done ?? 0 }}/{{ p?.total || '…' }}</span></span>
          <span class="text-muted-foreground tabular-nums">
            <template v-if="p?.phase === 'package'">đang đóng gói…</template>
            <template v-else-if="starting">đang khởi động bộ đọc…</template>
            <template v-else>còn ~{{ remainMin }} phút · {{ (p?.doneChars ?? 0).toLocaleString('vi-VN') }} / {{ (p?.totalChars ?? 0).toLocaleString('vi-VN') }} ký tự</template>
          </span>
        </div>
        <div class="mt-2 h-2.5 rounded-full bg-muted overflow-hidden">
          <div v-if="starting && renderPct === 0" class="h-full w-1/3 bg-primary/60 rounded-full animate-[sano-indeterminate_1.4s_ease-in-out_infinite]"></div>
          <div v-else class="h-full bg-primary rounded-full transition-[width]" :style="{ width: renderPct + '%' }"></div>
        </div>
      </div>
      <div class="mt-6 rounded-lg border border-border divide-y divide-border">
        <div v-for="(c, i) in chapters" :key="i" class="flex items-center gap-3 px-4 py-3 text-sm">
          <Check v-if="c.pct === 100" class="w-4 h-4 text-rag-green" />
          <Loader2 v-else-if="i === activeIdx" class="w-4 h-4 animate-spin text-primary" />
          <span v-else class="w-4 h-4 rounded-full border border-border"></span>
          <span class="flex-1" :class="c.pct === 0 && i !== activeIdx && 'text-muted-foreground'">{{ c.title }}</span>
          <span class="text-xs tabular-nums text-muted-foreground">{{ c.pct }}%</span>
        </div>
      </div>
      <div class="mt-5 flex gap-2">
        <Button variant="outline" class="text-destructive" @click="cancelRender">Huỷ render</Button>
      </div>
    </template>

    <template v-else>
      <div class="h-12 w-12 rounded-full bg-rag-green/15 grid place-items-center"><Check class="w-6 h-6 text-rag-green" /></div>
      <h1 class="mt-4 text-xl font-semibold tracking-tight">Đã tạo xong "{{ made?.title ?? state.render.title }}"</h1>
      <p class="text-sm text-muted-foreground">{{ listenMin }} phút nghe · {{ made?.chapters ?? '…' }} chương · đã thêm vào thư viện</p>
      <div class="mt-6 grid gap-3">
        <button class="flex items-center gap-3 rounded-lg border border-border p-4 text-left hover:bg-muted/50" @click="openBook(state.render.slug)">
          <BookOpen class="w-5 h-5 text-primary" /><span class="flex-1"><span class="block font-medium text-sm">Nghe ngay</span><span class="block text-xs text-muted-foreground">Mở trong thư viện</span></span><ChevronRight class="w-4 h-4 text-muted-foreground" />
        </button>
        <div class="rounded-lg border border-border">
          <button class="w-full flex items-center gap-3 p-4 text-left hover:bg-muted/50 disabled:opacity-60 disabled:cursor-not-allowed disabled:hover:bg-transparent" :disabled="m4bBusy()" @click="startM4B(state.render.slug)">
            <Download class="w-5 h-5 text-primary" /><span class="flex-1"><span class="block font-medium text-sm">Xuất file M4B</span><span class="block text-xs text-muted-foreground">Một file có mục lục chương, chép sang điện thoại nghe bằng app sách nói bất kỳ</span></span><ChevronRight class="w-4 h-4 text-muted-foreground" />
          </button>
          <div class="empty:hidden px-4 pb-3 flex flex-wrap gap-2"><M4BProgress :slug="state.render.slug" /></div>
        </div>
        <button class="flex items-center gap-3 rounded-lg border border-border p-4 text-left hover:bg-muted/50" @click="act(revealBookZip)">
          <Package class="w-5 h-5 text-primary" /><span class="flex-1"><span class="block font-medium text-sm">Xuất gói zip</span><span class="block text-xs text-muted-foreground">Sao lưu hoặc chuyển sách sang máy khác</span></span><ChevronRight class="w-4 h-4 text-muted-foreground" />
        </button>
      </div>
      <div class="mt-4 flex items-center gap-2">
        <Button variant="ghost" size="sm" @click="act(openBookFolder)"><FolderOpen class="w-4 h-4" /> Mở thư mục</Button>
        <span class="text-xs text-muted-foreground truncate" :title="made?.dir">{{ made?.dir }}</span>
      </div>
      <div class="mt-2">
        <Button variant="outline" size="sm" @click="go('create')">Tạo cuốn khác</Button>
      </div>
      <p v-if="actionError" class="mt-3 text-sm text-destructive">{{ actionError }}</p>
    </template>
  </div>
</template>

<style>
@keyframes sano-indeterminate {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(300%); }
}
</style>
