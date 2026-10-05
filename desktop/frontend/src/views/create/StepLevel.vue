<script setup lang="ts">
// B1 Cách đọc: màn A chọn 1 trong 3 cấp (wireframe D8), màn B nhờ AI (wireframe D8b):
// chọn AI trước (không chọn sẵn), rồi 3 bước mỗi bước một nút. Gemini bản miễn phí
// không tạo được file Word nên prompt của Gemini đòi khối mã để dán vào Sano (đã thử 27/09).
import { computed, onBeforeUnmount, ref } from 'vue'
import { Check, CheckCircle2, ChevronLeft, ClipboardPaste, Clock, Copy, Download, ExternalLink, FileText, Lightbulb, Pause, Play, Sparkles, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { copyText, errText, openURL, saveAIGuide } from '../../lib/backend'
import { AI_TOOLS, CHATGPT_PLANS, SKILL_STEPS, SKILL_ZIP, promptFor, shortInstruction, usePhrases, type ChatGPTPlan } from '../../lib/prompt'
import { saveAITool, saveLevel, state } from '../../lib/store'

const options = [
  { n: 1, title: 'Đọc nguyên văn', does: 'Sano đọc đúng từng chữ trong file của bạn: Word, EPUB hoặc PDF.', fit: 'bài viết, ghi chép đã dễ đọc; hoặc bạn muốn nghe y nguyên', time: 'Không cần chuẩn bị', best: false },
  { n: 2, title: 'Làm mượt', does: 'Nhờ AI đổi bảng, hình, danh sách, chữ viết tắt thành lời. Giữ nguyên ý và giọng tác giả.', fit: 'tài liệu nhiều bảng biểu, gạch đầu dòng', time: 'Thêm khoảng 5 phút với AI', best: false },
  { n: 3, title: 'Viết lại thành văn sách nói', does: 'Nhờ AI viết lại như người kể: chuyện trước, lý thuyết sau, chương ngắn, cuối chương có ba ý cần nhớ.', fit: 'sách, giáo trình, tài liệu dài muốn nghe cuốn như sách nói', time: 'Thêm khoảng 10–15 phút với AI', best: true },
]

// Nghe mẫu: cùng một đoạn, ba cách viết, render sẵn bằng giọng mặc định. Thiếu
// file thì ẩn nút (bản dev chưa render mẫu).
const samples = import.meta.glob<string>('../../assets/levels/cap-*.mp3', { eager: true, import: 'default', query: '?url' })
const sampleFor = (n: number) => samples[`../../assets/levels/cap-${n}.mp3`]
const playing = ref(0)
// Thời lượng từng đoạn đọc từ chính file (render lại mẫu không phải sửa số).
const durs = ref<Record<number, string>>({})
for (const n of [1, 2, 3]) {
  const src = sampleFor(n)
  if (!src) continue
  const a = new Audio()
  a.preload = 'metadata'
  a.onloadedmetadata = () => {
    const sec = Math.round(a.duration)
    if (Number.isFinite(sec)) durs.value = { ...durs.value, [n]: `${Math.floor(sec / 60)}:${String(sec % 60).padStart(2, '0')}` }
  }
  a.src = src
}
let audio: HTMLAudioElement | null = null
function stop() {
  audio?.pause()
  audio = null
  playing.value = 0
}
function toggle(n: number) {
  const on = playing.value === n
  stop()
  const src = sampleFor(n)
  if (on || !src) return
  audio = new Audio(src)
  audio.onended = stop
  playing.value = n
  void audio.play().catch(stop)
}
onBeforeUnmount(stop)

function pick(n: number) {
  saveLevel(n)
}

const tool = computed(() => AI_TOOLS.find((a) => a.k === state.aiTool))
const gemini = computed(() => state.aiTool === 'gemini')
const copied = ref(0)
async function copy(which: 1 | 3) {
  if (!state.aiTool) return
  const text = which === 1 ? promptFor(state.level, state.aiTool) : shortInstruction()
  if (await copyText(text)) {
    copied.value = which
    setTimeout(() => (copied.value = 0), 1500)
  }
}

// Hộp "Nạp skill" (wireframe D8b): theo AI đã chọn, mỗi bước có nút làm ngay tại dòng.
// ChatGPT chọn gói trước (không chọn sẵn): có mục Skills → zip như Claude; chưa có →
// Dự án + file hướng dẫn. Gemini không có hộp (Google có thể bỏ Gem từ 13/10/2026).
const skillOpen = ref(false)
const plan = ref<'' | ChatGPTPlan>('')
const saved = ref('')
const saveError = ref('')
const steps = computed(() => (state.aiTool === 'claude' ? SKILL_STEPS.claude : plan.value ? SKILL_STEPS[plan.value] : []))
const copiedSay = ref(-1)
async function copySay(i: number, text: string) {
  if (await copyText(text)) {
    copiedSay.value = i
    setTimeout(() => (copiedSay.value = -1), 1500)
  }
}
function openSkill() {
  plan.value = ''
  saved.value = ''
  saveError.value = ''
  skillOpen.value = true
}
function pickPlan(p: ChatGPTPlan) {
  plan.value = p
  saved.value = ''
  saveError.value = ''
}
async function download(kind: 'claude' | 'chatgpt') {
  saveError.value = ''
  try {
    const p = await saveAIGuide(kind)
    if (p) saved.value = p.split(/[\\/]/).slice(-2).join('/')
  } catch (e) {
    saveError.value = errText(e)
  }
}
</script>

<template>
  <!-- ═══ Màn A: chọn cách làm ═══ -->
  <div v-if="state.levelScreen === 'choose'" class="max-w-3xl">
    <h1 class="text-xl font-semibold tracking-tight">Chọn cách làm sách nói</h1>
    <p class="text-sm text-muted-foreground">Văn viết để đọc bằng mắt, đọc to lên thường nghe chán.<template v-if="sampleFor(1)"> Bấm <b class="font-medium text-foreground">Nghe mẫu</b> ở từng cách để nghe khác biệt, rồi chọn một cách.</template></p>
    <div class="mt-4 space-y-2.5" role="radiogroup" aria-label="Cách làm sách nói">
      <div v-for="o in options" :key="o.n" role="radio" :aria-checked="state.level === o.n" tabindex="0"
        class="relative flex items-start gap-4 rounded-xl border p-4 cursor-pointer transition"
        :class="state.level === o.n ? 'border-primary ring-2 ring-primary/15 bg-primary/5' : 'border-border hover:border-primary/40 hover:bg-muted/30'"
        @click="pick(o.n)" @keydown.enter.prevent="pick(o.n)" @keydown.space.prevent="pick(o.n)">
        <span class="mt-0.5 h-5 w-5 rounded-full border-2 grid place-items-center shrink-0" :class="state.level === o.n ? 'border-primary' : 'border-muted-foreground/40'">
          <span v-if="state.level === o.n" class="h-2.5 w-2.5 rounded-full bg-primary"></span>
        </span>
        <div class="flex-1 min-w-0">
          <p class="font-medium flex items-center gap-2">
            <span class="text-xs font-normal text-muted-foreground">Cấp {{ o.n }}</span> {{ o.title }}
            <span v-if="o.best" class="rounded-full bg-chart/15 text-chart text-[11px] font-semibold px-2 py-0.5">Nghe hấp dẫn nhất</span>
          </p>
          <p class="mt-0.5 text-sm text-muted-foreground">{{ o.does }}</p>
          <p class="mt-1.5 text-xs text-muted-foreground flex flex-wrap gap-x-4 gap-y-1">
            <span><span class="text-foreground">Hợp khi:</span> {{ o.fit }}</span>
            <span class="flex items-center gap-1"><Clock class="w-3 h-3" /> {{ o.time }}</span>
          </p>
        </div>
        <button v-if="sampleFor(o.n)" class="shrink-0 h-8 pl-2.5 pr-3 rounded-full border text-xs flex items-center gap-1.5"
          :class="playing === o.n ? 'border-chart bg-chart text-white' : 'border-border bg-background hover:bg-muted'"
          @click.stop="toggle(o.n)">
          <component :is="playing === o.n ? Pause : Play" class="w-3.5 h-3.5" /> {{ playing === o.n ? 'Dừng' : 'Nghe mẫu' }} <span v-if="durs[o.n]" :class="playing === o.n ? 'opacity-80' : 'text-muted-foreground'">· {{ durs[o.n] }}</span>
        </button>
      </div>
    </div>
    <p class="mt-4 text-sm text-muted-foreground flex items-start gap-2">
      <Lightbulb class="w-4 h-4 mt-0.5 text-rag-amber shrink-0" />
      <span>Không chắc? Sách hay tài liệu dài thì chọn <b class="font-medium text-foreground">cấp 3</b>. Bài ngắn, đọc đã trôi chảy thì chọn <b class="font-medium text-foreground">cấp 1</b>. Lần sau Sano nhớ lựa chọn của bạn.</span>
    </p>
  </div>

  <!-- ═══ Màn B: nhờ AI (wireframe D8b) ═══ -->
  <div v-else class="max-w-3xl">
    <button class="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1" @click="state.levelScreen = 'choose'"><ChevronLeft class="w-4 h-4" /> Đổi cách làm</button>
    <h1 class="mt-2 text-xl font-semibold tracking-tight">Nhờ AI {{ state.level === 3 ? 'viết lại thành văn sách nói' : 'làm mượt tài liệu' }}</h1>
    <p class="text-sm text-muted-foreground">AI bản miễn phí cũng được. {{ state.level === 3 ? 'Mất khoảng 10–15 phút.' : 'Mất khoảng 5 phút.' }}</p>

    <!-- Chọn AI: không chọn sẵn, chọn rồi Sano nhớ -->
    <p class="mt-5 text-sm font-medium">Bạn dùng AI nào?</p>
    <div class="mt-2 grid grid-cols-3 gap-3" role="radiogroup" aria-label="AI đang dùng">
      <button v-for="x in AI_TOOLS" :key="x.k" role="radio" :aria-checked="state.aiTool === x.k"
        class="relative rounded-xl border p-3.5 text-left transition"
        :class="state.aiTool === x.k ? 'border-primary ring-2 ring-primary/15 bg-primary/5' : 'border-border hover:border-primary/40 hover:bg-muted/30'"
        @click="saveAITool(x.k)">
        <span class="absolute top-3 right-3 h-5 w-5 rounded-full border-2 grid place-items-center" :class="state.aiTool === x.k ? 'border-primary bg-primary' : 'border-muted-foreground/30'">
          <Check v-if="state.aiTool === x.k" class="w-3 h-3 text-primary-foreground" />
        </span>
        <span class="block font-semibold">{{ x.name }}</span>
        <span class="mt-0.5 flex items-center gap-1.5 text-xs text-muted-foreground">
          <component :is="x.k === 'gemini' ? ClipboardPaste : FileText" class="w-3.5 h-3.5" /> {{ x.k === 'gemini' ? 'Trả về văn bản để dán' : 'Trả về file Word' }}
        </span>
      </button>
    </div>

    <div v-if="!tool" class="mt-4 rounded-xl border border-dashed border-border py-10 text-center text-sm text-muted-foreground">
      Chọn AI bạn dùng để Sano đưa đúng prompt và cách làm.
    </div>

    <template v-else>
      <ol class="mt-4 rounded-xl border border-border divide-y divide-border">
        <li class="flex items-center gap-4 px-4 py-3.5">
          <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">1</span>
          <p class="flex-1 font-medium">Sao chép prompt <span class="font-normal text-muted-foreground">(câu lệnh gửi cho AI)</span></p>
          <Button class="shrink-0 w-60" @click="copy(1)"><component :is="copied === 1 ? Check : Copy" class="w-4 h-4" /> {{ copied === 1 ? 'Đã sao chép' : `Sao chép prompt cho ${tool.name}` }}</Button>
        </li>
        <li class="flex items-center gap-4 px-4 py-3.5">
          <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">2</span>
          <p class="flex-1 font-medium">Mở {{ tool.name }}, đính kèm tài liệu của bạn (Word hoặc PDF), dán prompt rồi gửi</p>
          <Button variant="outline" class="shrink-0 w-60" @click="openURL(tool.url)">Mở {{ tool.name }} <ExternalLink class="w-3.5 h-3.5" /></Button>
        </li>
        <li class="flex items-center gap-4 px-4 py-3.5">
          <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">3</span>
          <p class="flex-1 font-medium">
            <template v-if="gemini">Gemini viết xong: bấm nút sao chép ở góc khung văn bản, rồi bấm <span class="text-primary">Tiếp</span></template>
            <template v-else>{{ tool.name }} viết xong: tải file Word về máy, rồi bấm <span class="text-primary">Tiếp</span></template>
          </p>
        </li>
      </ol>
      <p class="mt-3 text-xs text-muted-foreground flex items-start gap-1.5">
        <Lightbulb class="w-3.5 h-3.5 mt-px text-rag-amber shrink-0" />
        <span>Tài liệu dài thì gõ "tiếp" để AI làm phần sau. AI từ chối hoặc dừng giữa chừng thì bấm tạo lại câu trả lời. Đọc lại bản AI viết, nhất là số liệu, tên riêng.</span>
      </p>
      <div v-if="!gemini" class="mt-4 flex items-center gap-4 rounded-xl border border-chart/30 bg-chart/10 px-4 py-3.5">
        <span class="h-9 w-9 rounded-full bg-chart/15 text-chart grid place-items-center shrink-0"><Sparkles class="w-4 h-4" /></span>
        <div class="flex-1 min-w-0">
          <p class="font-medium">Làm sách thường xuyên? Nạp skill cho {{ tool.name }}</p>
          <p class="text-sm text-muted-foreground">Nạp một lần, lần sau chỉ cần gửi file và gõ một câu.</p>
        </div>
        <Button variant="outline" class="shrink-0 bg-background" @click="openSkill">Nạp skill cho {{ tool.name }}</Button>
      </div>
      <p v-else class="mt-5 flex items-center gap-2 text-sm text-muted-foreground">
        <Sparkles class="w-4 h-4 text-chart shrink-0" /> Với Gemini, mỗi lần làm sách chỉ cần dán prompt, Gemini chưa hỗ trợ skill.
      </p>
    </template>

    <!-- ═══ Hộp nạp skill: theo AI đã chọn, mỗi bước có nút tại dòng ═══ -->
    <div v-if="skillOpen && tool && !gemini" class="fixed inset-0 bg-black/40 grid place-items-center z-50" @click.self="skillOpen = false">
      <div class="w-[640px] max-w-[calc(100vw-2rem)] max-h-[calc(100vh-2rem)] overflow-auto rounded-xl border border-border bg-background shadow-2xl" role="dialog" aria-modal="true" :aria-label="`Nạp skill làm sách nói cho ${tool.name}`">
        <div class="flex items-start justify-between p-5 pb-0">
          <div>
            <h2 class="font-semibold">Nạp skill làm sách nói cho {{ tool.name }}</h2>
            <p class="text-xs text-muted-foreground mt-0.5">Nạp một lần. Sau đó mỗi lần làm sách chỉ cần đính kèm file và gõ một câu có sẵn.</p>
          </div>
          <button class="text-muted-foreground hover:text-foreground" aria-label="Đóng" @click="skillOpen = false"><X class="w-4 h-4" /></button>
        </div>
        <div class="p-5">
          <template v-if="state.aiTool === 'chatgpt'">
            <p class="text-sm font-medium">ChatGPT của bạn có mục Skills không?</p>
            <div class="mt-2 grid grid-cols-2 gap-3" role="radiogroup" aria-label="Gói ChatGPT">
              <button v-for="p in CHATGPT_PLANS" :key="p.k" role="radio" :aria-checked="plan === p.k" class="relative rounded-xl border p-3 text-left transition"
                :class="plan === p.k ? 'border-primary ring-2 ring-primary/15 bg-primary/5' : 'border-border hover:border-primary/40'" @click="pickPlan(p.k)">
                <span class="absolute top-3 right-3 h-4 w-4 rounded-full border-2 grid place-items-center" :class="plan === p.k ? 'border-primary bg-primary' : 'border-muted-foreground/30'"><Check v-if="plan === p.k" class="w-2.5 h-2.5 text-primary-foreground" /></span>
                <span class="block text-sm font-semibold">{{ p.title }}</span>
                <span class="mt-0.5 block text-xs text-muted-foreground">{{ p.sub }}</span>
              </button>
            </div>
          </template>

          <ol v-if="steps.length" class="rounded-xl border border-border divide-y divide-border text-sm" :class="state.aiTool === 'chatgpt' && 'mt-4'">
            <li v-for="(st, i) in steps" :key="i" class="flex items-center gap-3 px-4 py-3">
              <span class="h-6 w-6 rounded-full bg-primary text-primary-foreground grid place-items-center text-xs font-semibold shrink-0">{{ i + 1 }}</span>
              <span class="flex-1">{{ st.text }}</span>
              <Button v-if="st.act === 'file'" size="sm" class="shrink-0 w-48" @click="download(st.file === SKILL_ZIP ? 'claude' : 'chatgpt')">
                <component :is="saved ? CheckCircle2 : Download" class="w-4 h-4" /> {{ saved ? 'Đã lưu' : `Tải ${st.file}` }}
              </Button>
              <Button v-else-if="st.act === 'open'" size="sm" variant="outline" class="shrink-0 w-48" @click="openURL(st.url!)">{{ st.label }} <ExternalLink class="w-3.5 h-3.5" /></Button>
              <Button v-else-if="st.act === 'copy'" size="sm" variant="outline" class="shrink-0 w-48" @click="copy(3)"><component :is="copied === 3 ? Check : Copy" class="w-4 h-4" /> {{ copied === 3 ? 'Đã sao chép' : 'Sao chép câu dán' }}</Button>
            </li>
          </ol>
          <p v-if="saved" class="mt-2 text-xs text-muted-foreground">Đã lưu {{ saved }}</p>
          <p v-if="saveError" class="mt-2 text-xs text-destructive">{{ saveError }}</p>
          <!-- Từ nay làm sách: câu gõ mẫu có tên skill và cấp độ (wireframe D8b) -->
          <div v-if="steps.length" class="mt-4">
            <p class="text-sm font-medium">Từ nay làm sách: mở cuộc trò chuyện mới{{ plan === 'project' ? ' trong dự án "Sano – sách nói"' : '' }}, đính kèm tài liệu (Word hoặc PDF) rồi gõ:</p>
            <ul class="mt-2 space-y-2">
              <li v-for="(ph, i) in usePhrases(plan === 'project')" :key="i" class="rounded-lg border border-border px-3 py-2.5">
                <p class="text-xs font-medium text-muted-foreground">{{ ph.what }}</p>
                <div class="mt-1.5 flex items-center gap-2">
                  <span class="block flex-1 min-w-0 rounded-md bg-muted px-2.5 py-1.5 text-sm font-medium text-foreground">{{ ph.say }}</span>
                  <button class="h-8 px-2.5 rounded-md border border-border bg-background text-xs flex items-center gap-1 shrink-0 hover:bg-muted" @click="copySay(i, ph.say)">
                    <component :is="copiedSay === i ? Check : Copy" class="w-3.5 h-3.5" /> {{ copiedSay === i ? 'Đã sao chép' : 'Sao chép' }}
                  </button>
                </div>
              </li>
            </ul>
          </div>
          <p class="mt-3 text-xs text-muted-foreground">
            <template v-if="state.aiTool === 'claude'">Mọi gói Claude, kể cả miễn phí, đều nạp được skill. Skill cũng dùng được trong Claude Code.</template>
            <template v-else-if="plan === 'skills'">Mục Skills của ChatGPT dùng cùng định dạng skill với Claude, nên dùng chung một file.</template>
            <template v-else-if="plan === 'project'">Ô Hướng dẫn của Dự án tối đa 8.000 ký tự, nên hướng dẫn đầy đủ nằm trong file đính kèm.</template>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
