<script setup lang="ts">
// Màn cài bộ đọc lần đầu (bám wireframe WfDesktop, màn "setup"): báo trước dung
// lượng + thời gian, bấm Cài là app tự tải uv → Python → VieNeu-TTS → mô hình →
// ffmpeg (nếu thiếu) → đọc thử. Tiến độ từng bước từ phần Go (sự kiện
// setup:progress), huỷ được, lỗi thì báo rõ + thử lại; lần sau mở cài tiếp.
import { computed, onMounted } from 'vue'
import { AlertTriangle, Check, Clock, Cpu, HardDrive, Loader2, RefreshCw, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import logoUrl from '@/assets/logo.svg'
import { isDesktop, openURL } from '../lib/backend'
import { DOCS } from '../lib/mock'
import { cancelSetup, leaveSetup, refreshSetupInfo, refreshTTS, startSetup, state } from '../lib/store'

onMounted(() => {
  if (!state.setupInfo) refreshSetupInfo()
})

const gb = (b: number) => (b / 2 ** 30).toFixed(1).replace('.', ',') + ' GB'

const setup = computed(() => state.setup)
const running = computed(() => setup.value.running)
const ready = computed(() => !!state.tts?.ready)
const failed = computed(() => !running.value && !!setup.value.error)
// Đã làm dở (lượt này, hoặc lần mở app trước còn để lại > 50 MB) → nút "Cài tiếp".
const started = computed(
  () => setup.value.steps.some((s) => s.state !== 'pending') || (state.setupInfo?.usedBytes ?? 0) > 50 * 2 ** 20,
)

// Tải ~1 GB ở 50 Mbps (~6 MB/giây) + ~1 phút cài thư viện, đọc thử.
const minutes = computed(() => {
  const dl = state.setupInfo?.downloadBytes ?? 1000 * 2 ** 20
  return Math.max(1, Math.round(dl / (6.25 * 2 ** 20) / 60 + 1))
})

const machine = computed(() => {
  const i = state.setupInfo
  if (!i) return '…'
  const cpu = i.cpu ? i.cpu.replace(/\(R\)|\(TM\)|CPU|@.*$/g, '').replace(/\s+/g, ' ').trim() : `${i.os} ${i.arch}`
  const ram = i.ramBytes ? ` · ${Math.round(i.ramBytes / 2 ** 30)} GB RAM` : ''
  return `${cpu}${ram} · ${i.enough ? 'đủ' : 'thiếu dung lượng'}`
})

const eta = computed(() => {
  const s = setup.value.etaSec
  if (!running.value || s <= 0) return ''
  return s < 60 ? 'Còn dưới 1 phút' : `Còn khoảng ${Math.ceil(s / 60)} phút`
})

function goLibrary() {
  leaveSetup() // chưa đồng ý điều khoản thì sang màn điều khoản trước
}

async function recheck() {
  await refreshTTS()
  if (state.tts?.ready) goLibrary()
}
</script>

<template>
  <!-- flex + m-auto (không dùng grid place-items-center): cửa sổ thấp thì cuộn được, không bị cắt mất phần trên -->
  <div class="flex-1 min-h-0 flex overflow-auto p-6">
    <div class="w-full max-w-lg m-auto">
      <img :src="logoUrl" alt="" class="h-12 w-12 rounded-xl" />
      <h1 class="mt-5 text-2xl font-semibold tracking-tight">Chào mừng đến với Sano</h1>
      <p class="mt-2 text-sm text-muted-foreground">
        Để đọc sách thành giọng nói, Sano cần tải bộ đọc tiếng Việt về máy. Việc này chỉ làm một lần, sau đó dùng không cần mạng.
      </p>

      <div class="mt-6 rounded-lg border border-border divide-y divide-border text-sm" data-testid="setup-info">
        <div class="flex items-center justify-between px-4 py-3"><span class="flex items-center gap-2"><HardDrive class="w-4 h-4 text-muted-foreground" />Dung lượng cần</span><span class="font-medium tabular-nums">~{{ gb(state.setupInfo?.needBytes ?? 1500 * 2 ** 20) }}</span></div>
        <div class="flex items-center justify-between px-4 py-3"><span class="flex items-center gap-2"><Clock class="w-4 h-4 text-muted-foreground" />Thời gian tải (mạng 50 Mbps)</span><span class="font-medium tabular-nums">~{{ minutes }} phút</span></div>
        <div class="flex items-center justify-between px-4 py-3"><span class="flex items-center gap-2"><Cpu class="w-4 h-4 text-muted-foreground" />Máy của bạn</span><span class="font-medium" :class="state.setupInfo && !state.setupInfo.enough && 'text-destructive'">{{ machine }}</span></div>
      </div>
      <p v-if="state.setupInfo?.note" class="mt-2 text-xs text-muted-foreground flex items-center gap-1.5"><AlertTriangle class="w-3.5 h-3.5 text-rag-amber shrink-0" />{{ state.setupInfo.note }}</p>

      <div class="mt-6 space-y-3" data-testid="setup-steps">
        <div v-for="it in setup.steps" :key="it.key" :data-step="it.key" :data-state="it.state">
          <div class="flex justify-between text-xs mb-1 gap-3">
            <span class="flex items-center gap-1.5 min-w-0">
              <Check v-if="it.state === 'done' || it.state === 'skipped'" class="w-3.5 h-3.5 text-rag-green shrink-0" />
              <Loader2 v-else-if="it.state === 'running'" class="w-3.5 h-3.5 animate-spin text-primary shrink-0" />
              <X v-else-if="it.state === 'error'" class="w-3.5 h-3.5 text-destructive shrink-0" />
              <span v-else class="w-3.5 h-3.5 rounded-full border border-border inline-block shrink-0"></span>
              <span class="shrink-0">{{ it.label }}</span>
              <span v-if="it.detail" class="text-muted-foreground truncate">· {{ it.detail }}</span>
            </span>
            <span class="tabular-nums text-muted-foreground shrink-0">{{ Math.floor(it.pct) }}%</span>
          </div>
          <div class="h-1.5 rounded-full bg-muted overflow-hidden">
            <div class="h-full rounded-full transition-[width] duration-500" :class="it.state === 'error' ? 'bg-destructive' : 'bg-primary'" :style="{ width: it.pct + '%' }"></div>
          </div>
        </div>
      </div>

      <div v-if="failed || state.setupError" class="mt-4 rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-xs select-text" data-testid="setup-error">
        <p class="font-medium text-destructive whitespace-pre-line line-clamp-6">{{ state.setupError || setup.error }}</p>
        <p v-if="setup.hint && !state.setupError" class="mt-1 text-muted-foreground">{{ setup.hint }}</p>
        <p v-if="setup.logFile && !state.setupError" class="mt-1 text-muted-foreground">Nhật ký cài đặt: <span class="font-mono">{{ setup.logFile }}</span></p>
      </div>
      <p v-else-if="setup.cancelled && !running" class="mt-4 text-xs text-muted-foreground">Đã huỷ. Phần đã tải xong được giữ lại — bấm <span class="font-medium text-foreground">Cài tiếp</span> để làm nốt.</p>
      <p v-else-if="setup.done && ready" class="mt-4 text-xs text-rag-green font-medium">Cài xong bộ đọc. Bạn tạo sách được rồi.</p>
      <p v-else-if="setup.done && !ready && state.tts" class="mt-4 text-xs text-muted-foreground"><span class="font-medium text-foreground">{{ state.tts.message }}.</span> {{ state.tts.detail }}</p>

      <div class="mt-6 flex items-center justify-between gap-4">
        <span class="text-xs text-muted-foreground">
          <template v-if="running">{{ eta || 'Đang cài' }} · có thể đóng cửa sổ, lần sau mở sẽ tải tiếp</template>
          <template v-else-if="setup.done && ready">Gỡ bộ đọc được trong Cài đặt</template>
          <template v-else>Cài một lần, cần mạng · không cần mở Terminal</template>
          <br /><a :href="DOCS + '/cai-dat'" target="_blank" rel="noopener" class="text-primary hover:underline" @click.prevent="openURL(DOCS + '/cai-dat')">Cài không được? Xem hướng dẫn</a>
        </span>
        <div class="flex gap-2 shrink-0">
          <template v-if="running">
            <Button variant="outline" data-testid="setup-cancel" @click="cancelSetup"><X class="w-4 h-4" /> Huỷ</Button>
            <Button variant="ghost" @click="goLibrary">Bỏ qua</Button>
          </template>
          <template v-else-if="setup.done && ready">
            <Button data-testid="setup-finish" @click="goLibrary">Bắt đầu tạo sách</Button>
          </template>
          <template v-else-if="ready && !failed">
            <Button variant="outline" :disabled="state.ttsChecking" @click="recheck"><RefreshCw class="w-4 h-4" /> Kiểm tra lại</Button>
            <Button @click="goLibrary">Bộ đọc đã sẵn sàng</Button>
          </template>
          <template v-else>
            <Button variant="outline" @click="goLibrary">Bỏ qua</Button>
            <Button data-testid="setup-start" :disabled="!isDesktop() || (state.setupInfo !== null && !state.setupInfo.enough)" @click="startSetup">
              <template v-if="failed">Thử lại</template>
              <template v-else-if="started">Cài tiếp</template>
              <template v-else>Cài bộ đọc</template>
            </Button>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>
