<script setup lang="ts">
// B3 Giọng đọc: danh sách giọng thật của bộ đọc; "Nghe mẫu" đọc câu mẫu bằng giọng đó.
import { computed, onMounted, ref } from 'vue'
import { Loader2, Pause, Play } from 'lucide-vue-next'
import { errText, speakSample, type Voice } from '../../lib/backend'
import { useClipPlayer } from '../../lib/audio'
import { loadVoices, state } from '../../lib/store'

const SHOWN = 6

const player = useClipPlayer()
const loadingVoice = ref<string | null>(null)
const sampleError = ref('')
const showAll = ref(false)
const urls = new Map<string, { text: string; url: string }>()

onMounted(() => {
  void loadVoices()
  if (!state.sampleSentence) state.sampleSentence = 'Bước thứ nhất, dừng việc đang làm và nhìn người nói.'
})

// Giọng bộ đọc đánh dấu nổi bật lên đầu. Không gắn nhãn khuyến nghị giọng nào:
// chưa có đánh giá giọng nào hợp sách nói nhất, người dùng tự nghe mẫu rồi chọn.
const ordered = computed(() => {
  const v = [...state.voices]
  v.sort((a, b) => Number(b.featured) - Number(a.featured))
  return v
})
const visible = computed(() => {
  if (showAll.value) return ordered.value
  const top = ordered.value.slice(0, SHOWN)
  const cur = ordered.value.find((v) => v.name === state.voice)
  return cur && !top.includes(cur) ? [...top, cur] : top
})

/** "Nữ · Bắc · Phong cách tự nhiên" → "Nữ · miền Bắc · tự nhiên". */
function desc(v: Voice) {
  const [gender, region, style] = v.desc.split(' · ')
  return [gender, region && `miền ${region}`, style?.replace(/^(Phong cách|Giọng đọc)\s+/i, '')].filter(Boolean).join(' · ')
}

async function sample(v: Voice) {
  sampleError.value = ''
  const text = state.sampleSentence.trim()
  const hit = urls.get(v.name)
  if (hit && hit.text === text) return player.toggle(v.name, hit.url)
  loadingVoice.value = v.name
  try {
    const url = await speakSample(v.name, text)
    urls.set(v.name, { text, url })
    await player.toggle(v.name, url)
  } catch (e) {
    sampleError.value = errText(e)
  } finally {
    loadingVoice.value = null
  }
}
</script>

<template>
  <div class="max-w-2xl">
    <h1 class="text-xl font-semibold tracking-tight">Chọn giọng đọc</h1>
    <p class="text-sm text-muted-foreground">Bấm nghe để thử từng giọng với một câu trong sách của bạn.</p>
    <p v-if="state.voicesError" class="mt-4 text-sm text-destructive">{{ state.voicesError }}</p>
    <p v-else-if="!state.voices.length" class="mt-5 text-sm text-muted-foreground flex items-center gap-2"><Loader2 class="w-4 h-4 animate-spin" /> Đang lấy danh sách giọng…</p>
    <div class="mt-5 grid gap-2">
      <label v-for="v in visible" :key="v.name" class="flex items-center gap-3 rounded-lg border px-4 py-3 cursor-pointer"
        :class="state.voice === v.name ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/50'">
        <input v-model="state.voice" type="radio" :value="v.name" class="h-4 w-4 accent-[hsl(var(--primary))]" />
        <span class="flex-1">
          <span class="font-medium text-sm">{{ v.name }}</span>
          <span class="block text-xs text-muted-foreground">{{ desc(v) }}</span>
        </span>
        <button class="h-8 px-3 rounded-full border border-border text-xs flex items-center gap-1.5 hover:bg-muted disabled:opacity-50"
          :disabled="loadingVoice !== null && loadingVoice !== v.name" @click.prevent="sample(v)">
          <Loader2 v-if="loadingVoice === v.name" class="w-3.5 h-3.5 animate-spin" />
          <component :is="player.playing.value === v.name ? Pause : Play" v-else class="w-3.5 h-3.5" /> Nghe mẫu
        </button>
      </label>
    </div>
    <button v-if="!showAll && ordered.length > SHOWN" class="mt-2 text-sm text-primary hover:underline" @click="showAll = true">
      Xem thêm {{ ordered.length - SHOWN }} giọng
    </button>
    <p v-if="sampleError || player.error.value" class="mt-3 text-sm text-destructive">{{ sampleError || player.error.value }}</p>
    <div class="mt-5 text-sm">
      <p class="font-medium">Câu nghe mẫu</p>
      <input v-model="state.sampleSentence" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" />
    </div>
  </div>
</template>
