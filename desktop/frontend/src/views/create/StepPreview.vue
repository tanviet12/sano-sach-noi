<script setup lang="ts">
// B5 Nghe thử (bắt buộc): render thật lời mở đầu + 2 tiểu mục đầu (mỗi đoạn
// ~500 ký tự đầu), phát trong app, hiện đúng lời đã đọc. Sửa lời đọc rồi render
// lại đoạn đó — bản cuối dùng lời đã sửa.
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Check, Loader2, Pause, Pencil, Play, RotateCcw } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useClipPlayer } from '../../lib/audio'
import { INTRO_LABEL, addPreview, editClip, ensurePreview, estRender, markHeard, selectedStems, state, stemTitle } from '../../lib/store'

const player = useClipPlayer(markHeard)
const drafts = reactive<Record<string, string>>({})
const more = ref('')

onMounted(() => void ensurePreview())

// Mỗi lần có đoạn mới render → bản nháp = đúng lời đã đọc.
watch(
  () => state.clips.map((c) => c.stem + '\u0000' + c.text + '\u0000' + c.url),
  () => {
    for (const c of state.clips) drafts[c.stem] = c.text
  },
  { immediate: true },
)

const others = computed(() => selectedStems.value.filter((s) => !state.clips.some((c) => c.stem === s)))

function fmtDur(sec: number) {
  return `${Math.floor(sec / 60)}:${String(sec % 60).padStart(2, '0')}`
}

async function addMore() {
  if (!more.value) return
  const stem = more.value
  more.value = ''
  await addPreview([stem])
}

const rightsOk = computed(() => !!state.rightsConfirmedAt)
function toggleRights(e: Event) {
  state.rightsConfirmedAt = (e.target as HTMLInputElement).checked ? new Date().toISOString() : ''
}
</script>

<template>
  <div class="max-w-3xl">
    <h1 class="text-xl font-semibold tracking-tight">Nghe thử trước khi render</h1>
    <p class="text-sm text-muted-foreground">
      Render cả cuốn mất khoảng {{ estRender }} phút, huỷ giữa chừng là phải làm lại từ đầu. Hãy nghe vài đoạn để chắc giọng và cách đọc đã ổn.
    </p>
    <p v-if="state.previewing" class="mt-4 text-sm text-muted-foreground flex items-center gap-2">
      <Loader2 class="w-4 h-4 animate-spin" /> Đang đọc thử{{ state.clips.length ? ' đoạn mới' : ' các đoạn đầu' }}… lần đầu mất khoảng nửa phút để nạp bộ đọc.
    </p>
    <p v-if="state.previewError" class="mt-4 text-sm text-destructive">{{ state.previewError }}</p>
    <p v-if="player.error.value" class="mt-2 text-sm text-destructive">{{ player.error.value }}</p>
    <div class="mt-5 space-y-3">
      <div v-for="s in state.clips" :key="s.stem" class="rounded-lg border border-border p-4">
        <div class="flex items-center gap-3">
          <button :aria-label="player.playing.value === s.stem ? 'Dừng' : 'Nghe'" class="h-9 w-9 grid place-items-center rounded-full bg-primary text-primary-foreground shrink-0" @click="player.toggle(s.stem, s.url)">
            <component :is="player.playing.value === s.stem ? Pause : Play" class="w-4 h-4" />
          </button>
          <span class="flex-1 font-medium text-sm">{{ s.stem === 'intro' ? INTRO_LABEL : s.title }}</span>
          <Check v-if="state.heard.includes(s.stem)" class="w-4 h-4 text-rag-green" aria-label="Đã nghe" />
          <Badge variant="secondary">{{ s.full ? 'Đã render' : 'Đã render đoạn đầu' }} · {{ fmtDur(s.durationSec) }}</Badge>
        </div>
        <div class="mt-3 text-sm">
          <p class="text-xs font-medium text-muted-foreground flex items-center gap-1.5"><Pencil class="w-3 h-3" /> Lời đọc (sửa được — sửa xong bấm render lại đoạn này)</p>
          <textarea v-model="drafts[s.stem]" class="mt-1 w-full h-20 rounded-md border border-input bg-background p-2 text-sm"></textarea>
          <Button v-if="drafts[s.stem] !== s.text" variant="outline" size="sm" class="mt-1" :disabled="state.previewing" @click="editClip(s.stem, drafts[s.stem])">
            <RotateCcw class="w-4 h-4" /> Render lại đoạn này
          </Button>
        </div>
      </div>
    </div>
    <div v-if="others.length && state.clips.length" class="mt-3 flex items-center gap-2">
      <select v-model="more" class="h-9 rounded-md border border-input bg-background px-2 text-sm max-w-md" :disabled="state.previewing" aria-label="Chọn thêm đoạn khác để nghe thử">
        <option value="">+ Chọn thêm đoạn khác để nghe thử</option>
        <option v-for="st in others" :key="st" :value="st">{{ stemTitle(st) }}</option>
      </select>
      <Button variant="outline" size="sm" :disabled="!more || state.previewing" @click="addMore">Nghe thử đoạn này</Button>
    </div>

    <!-- Xác nhận quyền dùng tài liệu cho từng cuốn — bắt buộc trước khi render cả cuốn -->
    <label v-if="state.clips.length" class="mt-6 flex cursor-pointer items-start gap-2.5 rounded-lg border border-border bg-muted/30 p-4 text-sm"
      :class="rightsOk && 'border-primary/40 bg-primary/5'">
      <input :checked="rightsOk" type="checkbox" class="mt-0.5 h-4 w-4 accent-[hsl(var(--primary))]" @change="toggleRights" />
      <span>
        Tôi xác nhận có quyền dùng tài liệu này để làm sách nói
        <span class="text-muted-foreground">(tài liệu của tôi, tác phẩm đã hết thời hạn bảo hộ, hoặc được tác giả cho phép).</span>
      </span>
    </label>
  </div>
</template>
