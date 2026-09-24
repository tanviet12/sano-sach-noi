<script setup lang="ts">
// Khối "Bộ đọc" trong Cài đặt: trạng thái bộ đọc, nguồn (app tự cài / cài tay /
// biến môi trường), nút cài khi chưa có, kiểm tra lại, gỡ bộ đọc (hỏi xác nhận
// bằng hộp thoại của hệ điều hành, hiện dung lượng thật sẽ giải phóng).
import { computed, onMounted, ref } from 'vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Download, Loader2, RefreshCw, Trash2 } from 'lucide-vue-next'
import { errText, mockSetupStatus, uninstallTTS } from '../lib/backend'
import { refreshSetupInfo, refreshTTS, state } from '../lib/store'

onMounted(() => {
  refreshSetupInfo()
})

const source = computed(() => {
  switch (state.tts?.source) {
    case 'app': return `Sano tự cài · ${state.tts.dataDir ?? ''}`
    case 'legacy': return 'Cài tay theo hướng dẫn (~/VieNeu-TTS-v3)'
    case 'env': return 'Biến môi trường SANO_TTS_PYTHON'
    default: return '—'
  }
})

// Có bộ đọc do Sano cài (thư mục bộ đọc > 1 MB — chỉ có script thì không tính).
const removable = computed(() => (state.setupInfo?.usedBytes ?? 0) > 2 ** 20)
const busy = ref(false)
const msg = ref('')
const msgErr = ref(false)

function humanSize(b: number) {
  return b >= 2 ** 30 ? (b / 2 ** 30).toFixed(1).replace('.', ',') + ' GB' : Math.round(b / 2 ** 20) + ' MB'
}

async function uninstall() {
  busy.value = true
  msg.value = ''
  msgErr.value = false
  try {
    const res = await uninstallTTS()
    if (!res.cancelled) {
      msg.value = `Đã gỡ bộ đọc, giải phóng ${humanSize(res.freedBytes)}.`
      state.voices = []
      state.setup = mockSetupStatus() // màn cài quay về trạng thái chưa cài
      await Promise.all([refreshTTS(), refreshSetupInfo()])
    }
  } catch (e) {
    msgErr.value = true
    msg.value = errText(e)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="rounded-lg border border-border divide-y divide-border text-sm" data-testid="tts-status">
    <div class="flex items-center justify-between gap-3 px-4 py-3">
      <span>VieNeu-TTS v3 Turbo · {{ state.tts?.modelsOk ? 'mô hình đã tải' : 'chưa sẵn sàng' }}</span>
      <Badge v-if="state.ttsChecking" variant="secondary"><Loader2 class="w-3 h-3 mr-1 animate-spin" /> Đang kiểm tra</Badge>
      <Badge v-else-if="state.tts?.ready" class="bg-rag-green/15 text-rag-green border-0">{{ state.tts.message }}</Badge>
      <Badge v-else-if="state.tts" class="bg-rag-amber/15 text-rag-amber border-0">{{ state.tts.message }}</Badge>
    </div>
    <div v-if="state.tts" class="px-4 py-3 text-xs text-muted-foreground space-y-1 select-text">
      <p>Nguồn: {{ source }}</p>
      <p>Python: <span class="font-mono">{{ state.tts.python || '—' }}</span><template v-if="state.tts.pythonVersion"> · {{ state.tts.pythonVersion }}</template></p>
      <p>Script đọc giọng: <span class="font-mono">{{ state.tts.scriptsDir || 'không thấy' }}</span></p>
      <p>ffmpeg: <span class="font-mono">{{ state.tts.ffmpeg || 'không thấy' }}</span></p>
      <p v-if="state.tts.detail" class="whitespace-pre-line">{{ state.tts.detail }}</p>
    </div>
    <div v-if="state.tts && !state.tts.ready && !state.ttsChecking" class="flex items-center justify-between px-4 py-3">
      <span>Cài bộ đọc (tự tải về, vài phút)</span>
      <Button size="sm" data-testid="tts-install" @click="state.view = 'setup'"><Download class="w-4 h-4" /> Cài bộ đọc</Button>
    </div>
    <div class="flex items-center justify-between px-4 py-3"><span>Giọng mặc định</span><span class="text-muted-foreground">Thiện Minh · Nam · miền Bắc · kể chuyện</span></div>
    <div class="flex items-center justify-between px-4 py-3"><span>Kiểm tra bộ đọc</span><Button variant="outline" size="sm" :disabled="state.ttsChecking" @click="refreshTTS"><RefreshCw class="w-4 h-4" :class="state.ttsChecking && 'animate-spin'" /> Chạy kiểm tra</Button></div>
    <div class="flex items-center justify-between gap-3 px-4 py-3" data-testid="tts-uninstall">
      <span class="text-muted-foreground">
        <template v-if="removable">Gỡ bộ đọc và mô hình (giải phóng {{ humanSize(state.setupInfo?.usedBytes ?? 0) }})</template>
        <template v-else>Gỡ bộ đọc và mô hình — chưa có bộ đọc do Sano cài</template>
        <span v-if="msg" class="block text-xs mt-0.5" :class="msgErr ? 'text-destructive' : 'text-rag-green'">{{ msg }}</span>
      </span>
      <Button variant="ghost" size="sm" class="text-destructive shrink-0" :disabled="!removable || busy" @click="uninstall">
        <Loader2 v-if="busy" class="w-4 h-4 animate-spin" /><Trash2 v-else class="w-4 h-4" /> Gỡ
      </Button>
    </div>
  </div>
</template>
