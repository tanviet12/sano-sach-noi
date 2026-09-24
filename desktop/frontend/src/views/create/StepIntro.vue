<script setup lang="ts">
// B4 Lời mở đầu: đọc đầu tiên, trước chương 1. Tự điền theo tên sách/tác giả
// cho tới khi người dùng tự sửa.
import { onMounted, ref } from 'vue'
import { Info, Loader2, Pause, Play } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { errText, speakSample } from '../../lib/backend'
import { useClipPlayer } from '../../lib/audio'
import { defaultIntro, state } from '../../lib/store'

const player = useClipPlayer()
const loading = ref(false)
const error = ref('')

onMounted(() => {
  if (!state.introTouched) state.introText = defaultIntro()
})

async function listen() {
  if (player.playing.value) return player.stop()
  error.value = ''
  loading.value = true
  try {
    await player.toggle('intro', await speakSample(state.voice, state.introText))
  } catch (e) {
    error.value = errText(e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl">
    <h1 class="text-xl font-semibold tracking-tight">Lời mở đầu</h1>
    <p class="text-sm text-muted-foreground">Đoạn đọc đầu tiên, trước chương 1. Mỗi dòng đọc cách nhau một nhịp nghỉ.</p>
    <label class="mt-5 flex items-center gap-2 text-sm"><input v-model="state.introEnabled" type="checkbox" class="h-4 w-4 accent-[hsl(var(--primary))]" /> Có lời mở đầu</label>
    <textarea v-model="state.introText" :disabled="!state.introEnabled" class="mt-3 w-full h-36 rounded-md border border-input bg-background p-3 text-sm leading-relaxed disabled:opacity-50" @input="state.introTouched = true"></textarea>
    <p class="mt-2 text-xs text-muted-foreground flex items-center gap-1.5"><Info class="w-3.5 h-3.5" /> Đặt chữ "Cuốn sách:" trước tên sách để bộ đọc không nuốt mất tên ở đầu câu.</p>
    <Button variant="outline" size="sm" class="mt-4" :disabled="!state.introEnabled || !state.introText.trim() || loading" @click="listen">
      <Loader2 v-if="loading" class="w-4 h-4 animate-spin" />
      <component :is="player.playing.value ? Pause : Play" v-else class="w-4 h-4" /> Nghe lời mở đầu
    </Button>
    <p v-if="error || player.error.value" class="mt-2 text-sm text-destructive">{{ error || player.error.value }}</p>
  </div>
</template>
