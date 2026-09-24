<script setup lang="ts">
// D3 — Điều khoản sử dụng (phần mềm desktop). Hiện MỘT lần sau khi cài xong bộ đọc,
// trước khi vào Thư viện; hỏi lại khi phiên bản điều khoản đổi. Xem lại được ở Giới thiệu.
// Nội dung một nguồn: docs/dieu-khoan-su-dung.md (app thật nhúng file này).
// Wireframe tĩnh: KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import { FileText, Ban, ShieldCheck, Check, Sun, Moon, Info, Library, FilePlus2, Settings, LifeBuoy, ExternalLink, ChevronRight, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import terms from '../../../../docs/dieu-khoan-su-dung.md?raw'

type Mode = 'first' | 'updated' | 'about'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'first')
const dark = ref(false)
const agreed = ref(false)
const scrolledEnd = ref(false)

// Hiện markdown đơn giản: ## tiêu đề, - gạch đầu dòng, **đậm**, đoạn thường.
type Block = { kind: 'h1' | 'h2' | 'li' | 'p' | 'hr'; html: string }
const bold = (s: string) => s.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
const blocks = computed<Block[]>(() =>
  terms.split('\n').filter((l) => l.trim()).map((l) => {
    if (l.startsWith('# ')) return { kind: 'h1', html: bold(l.slice(2)) }
    if (l.startsWith('## ')) return { kind: 'h2', html: bold(l.slice(3)) }
    if (l.startsWith('- ')) return { kind: 'li', html: bold(l.slice(2)) }
    if (l.trim() === '---') return { kind: 'hr', html: '' }
    return { kind: 'p', html: bold(l) }
  }),
)
const body = computed(() => blocks.value.filter((b) => b.kind !== 'h1'))

const summary = [
  { icon: FileText, title: 'Tài liệu bạn có quyền dùng', desc: 'Tài liệu của bạn, sách hết bảo hộ, hoặc được tác giả cho phép.' },
  { icon: Ban, title: 'Không dùng sách còn bản quyền', desc: 'Trừ khi được tác giả hoặc chủ sở hữu cho phép.' },
  { icon: ShieldCheck, title: 'Chạy trên máy bạn', desc: 'Tài liệu không gửi đi đâu. Bạn tự chịu trách nhiệm nội dung.' },
]

function onScroll(e: Event) {
  const el = e.target as HTMLElement
  if (el.scrollTop + el.clientHeight >= el.scrollHeight - 8) scrolledEnd.value = true
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
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in ([['first', 'Lần đầu (sau khi cài bộ đọc)'], ['updated', 'Điều khoản có bản mới'], ['about', 'Xem lại trong Giới thiệu']] as [Mode, string][])" :key="m[0]"
        class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="mode = m[0]; agreed = false">{{ m[1] }}</button>
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

      <!-- ═══ Màn đồng ý điều khoản (toàn màn, như màn cài) ═══ -->
      <div v-if="mode !== 'about'" class="flex-1 min-h-0 overflow-auto">
        <div class="mx-auto max-w-2xl px-6 py-8 flex flex-col h-full">
          <div class="flex items-center gap-3">
            <img src="@/assets/logo.svg" alt="" class="h-10 w-10 rounded-xl" />
            <div>
              <h1 class="text-xl font-semibold tracking-tight">Điều khoản sử dụng</h1>
              <p class="text-sm text-muted-foreground">
                <template v-if="mode === 'first'">Bộ đọc đã sẵn sàng. Một bước cuối trước khi bắt đầu.</template>
                <template v-else>Điều khoản vừa được cập nhật lên phiên bản 2. Vui lòng đọc và đồng ý lại để tiếp tục.</template>
              </p>
            </div>
          </div>

          <!-- Tóm tắt 3 ý chính -->
          <div class="mt-5 grid grid-cols-3 gap-2.5">
            <div v-for="s in summary" :key="s.title" class="flex gap-3 rounded-lg border border-border p-3">
              <span class="h-8 w-8 shrink-0 rounded-md bg-primary/10 text-primary grid place-items-center"><component :is="s.icon" class="w-4 h-4" /></span>
              <div class="min-w-0">
                <p class="text-sm font-medium">{{ s.title }}</p>
                <p class="text-xs text-muted-foreground leading-relaxed">{{ s.desc }}</p>
              </div>
            </div>
          </div>

          <!-- Toàn văn, cuộn trong khung -->
          <div class="mt-4 flex-1 min-h-[180px] rounded-lg border border-border bg-muted/30 overflow-auto px-5 py-4 text-sm leading-relaxed" @scroll="onScroll">
            <template v-for="(b, i) in body" :key="i">
              <h2 v-if="b.kind === 'h2'" class="mt-4 first:mt-0 font-semibold" v-html="b.html"></h2>
              <p v-else-if="b.kind === 'li'" class="pl-4 relative before:content-['•'] before:absolute before:left-0 before:text-muted-foreground mt-1" v-html="b.html"></p>
              <hr v-else-if="b.kind === 'hr'" class="my-3 border-border" />
              <p v-else class="mt-2 text-muted-foreground" v-html="b.html"></p>
            </template>
          </div>

          <label class="mt-4 flex items-start gap-2.5 text-sm cursor-pointer">
            <input v-model="agreed" type="checkbox" class="mt-0.5 h-4 w-4 accent-[hsl(var(--primary))]" />
            <span>Tôi đã đọc và đồng ý với Điều khoản sử dụng. Tôi chỉ đưa vào Sano tài liệu tôi có quyền sử dụng và tự chịu trách nhiệm về nội dung mình tạo ra.</span>
          </label>
          <div class="mt-4 flex items-center justify-between">
            <Button variant="ghost" class="text-muted-foreground">Thoát Sano</Button>
            <Button :disabled="!agreed">Đồng ý và bắt đầu <ChevronRight class="w-4 h-4" /></Button>
          </div>
        </div>
      </div>

      <!-- ═══ Xem lại trong Giới thiệu ═══ -->
      <div v-else class="flex-1 flex min-h-0">
        <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
          <div class="px-4 py-4 flex items-center gap-2"><img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" /><span class="font-semibold tracking-tight">Sano</span></div>
          <nav class="px-2 space-y-0.5">
            <span v-for="n in nav" :key="n.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
              :class="n.key === 'about' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'"><component :is="n.icon" class="w-4 h-4" /> {{ n.label }}</span>
          </nav>
          <div class="px-2 mt-2"><span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span></div>
        </aside>
        <main class="flex-1 min-w-0 overflow-auto p-6 relative">
          <h1 class="text-xl font-semibold tracking-tight">Giới thiệu</h1>
          <p class="text-sm text-muted-foreground">(các phần khác giữ nguyên như hiện tại)</p>
          <h2 class="mt-6 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Thông tin</h2>
          <div class="mt-3 max-w-2xl rounded-lg border border-border divide-y divide-border text-sm">
            <div class="flex items-center gap-2.5 px-4 py-3 bg-primary/5"><FileText class="w-4 h-4 text-muted-foreground" /><span class="flex-1">Điều khoản sử dụng</span><span class="text-xs text-rag-green flex items-center gap-1"><Check class="w-3.5 h-3.5" /> Đã đồng ý ngày 24/09/2026 · phiên bản 1</span></div>
            <div class="flex items-center gap-2.5 px-4 py-3 text-muted-foreground">Hướng dẫn sử dụng · Báo lỗi · Mã nguồn · Giấy phép bên thứ ba …</div>
          </div>
          <!-- Hộp xem toàn văn (bấm vào dòng Điều khoản) -->
          <div class="absolute inset-0 bg-black/40 grid place-items-center">
            <div class="w-[640px] max-h-[600px] flex flex-col rounded-xl border border-border bg-background shadow-2xl">
              <div class="flex items-center justify-between px-5 py-4 border-b border-border">
                <h2 class="font-semibold">Điều khoản sử dụng · phiên bản 1</h2>
                <button class="text-muted-foreground" aria-label="Đóng"><X class="w-4 h-4" /></button>
              </div>
              <div class="flex-1 overflow-auto px-5 py-4 text-sm leading-relaxed">
                <template v-for="(b, i) in body" :key="i">
                  <h2 v-if="b.kind === 'h2'" class="mt-4 first:mt-0 font-semibold" v-html="b.html"></h2>
                  <p v-else-if="b.kind === 'li'" class="pl-4 relative before:content-['•'] before:absolute before:left-0 before:text-muted-foreground mt-1" v-html="b.html"></p>
                  <hr v-else-if="b.kind === 'hr'" class="my-3 border-border" />
                  <p v-else class="mt-2 text-muted-foreground" v-html="b.html"></p>
                </template>
              </div>
              <div class="px-5 py-3 border-t border-border flex justify-end"><Button variant="outline">Đóng</Button></div>
            </div>
          </div>
        </main>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D3 · Điều khoản sử dụng — hiện một lần sau khi cài xong bộ đọc, phải tick đồng ý mới vào được app; "Thoát Sano" đóng phần mềm. Lưu lần đồng ý (phiên bản + ngày) trong thư mục dữ liệu của app; điều khoản đổi phiên bản thì hỏi lại. Nội dung lấy từ docs/dieu-khoan-su-dung.md.
    </p>
  </div>
</template>
