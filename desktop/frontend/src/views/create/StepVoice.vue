<script setup lang="ts">
// B3 Giọng đọc: danh sách giọng thật của bộ đọc, gom theo miền (Bắc/Trung/Nam) + lọc
// Nam/Nữ (wireframe D4). "Nghe mẫu" đọc câu mẫu bằng giọng đó.
import { computed, onMounted, ref, watch } from 'vue'
import { Loader2, Pause, Play } from 'lucide-vue-next'
import { errText, speakSample, type Voice } from '../../lib/backend'
import { useClipPlayer } from '../../lib/audio'
import { lastVoice, loadVoices, state } from '../../lib/store'

const REGION_KEY = 'sano.voiceRegion'
const REGIONS = ['Bắc', 'Trung', 'Nam']

const player = useClipPlayer()
const loadingVoice = ref<string | null>(null)
const sampleError = ref('')
const urls = new Map<string, { text: string; url: string }>()
const used = lastVoice()

onMounted(() => {
  void loadVoices()
  if (!state.sampleSentence) state.sampleSentence = 'Bước thứ nhất, dừng việc đang làm và nhìn người nói.'
})

/** "Nữ · Bắc · Phong cách tự nhiên" → giới tính, miền, phong cách. */
function parts(v: Voice) {
  const [gender = '', region = '', style = ''] = v.desc.split(' · ')
  return { gender, region, style: style.replace(/^(Phong cách|Giọng đọc)\s+/i, '') }
}

// Miền: lựa chọn lần trước (nhớ trên máy) → miền của giọng đang chọn → Bắc.
function loadRegion() {
  try {
    return localStorage.getItem(REGION_KEY) || ''
  } catch {
    return ''
  }
}
const region = ref(loadRegion())
const gender = ref<'' | 'Nam' | 'Nữ'>('')
watch(
  () => state.voices.length,
  () => {
    const known = [...REGIONS, 'all']
    if (known.includes(region.value)) return
    const cur = state.voices.find((v) => v.name === state.voice)
    region.value = (cur && parts(cur).region) || 'Bắc'
  },
  { immediate: true },
)
function setRegion(r: string) {
  region.value = r
  try {
    localStorage.setItem(REGION_KEY, r)
  } catch {
    // không lưu được thì thôi
  }
}
const regionCount = (r: string) => state.voices.filter((v) => r === 'all' || parts(v).region === r).length
const regionTabs = computed(() => [...REGIONS.filter((r) => regionCount(r) > 0), 'all'])

// Giọng bộ đọc đánh dấu nổi bật lên đầu. Không gắn nhãn khuyến nghị giọng nào:
// chưa có đánh giá giọng nào hợp sách nói nhất, người dùng tự nghe mẫu rồi chọn.
const visible = computed(() => {
  const v = state.voices.filter((x) => {
    const p = parts(x)
    return (region.value === 'all' || p.region === region.value) && (!gender.value || p.gender === gender.value)
  })
  v.sort((a, b) => Number(b.featured) - Number(a.featured))
  return v
})

function desc(v: Voice) {
  const p = parts(v)
  return [p.gender, p.style, region.value === 'all' && p.region && `miền ${p.region}`].filter(Boolean).join(' · ')
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
    <p class="text-sm text-muted-foreground">
      Bấm nghe để thử từng giọng với một câu trong sách của bạn. Giọng đọc của
      <a href="https://github.com/pnnbao97/VieNeu-TTS" target="_blank" rel="noopener" class="text-primary hover:underline">VieNeu-TTS</a>, mã nguồn mở, chạy ngay trên máy.
    </p>
    <p v-if="state.voicesError" class="mt-4 text-sm text-destructive">{{ state.voicesError }}</p>
    <p v-else-if="!state.voices.length" class="mt-5 text-sm text-muted-foreground flex items-center gap-2"><Loader2 class="w-4 h-4 animate-spin" /> Đang lấy danh sách giọng…</p>
    <div v-if="state.voices.length" class="mt-5 flex flex-wrap items-center justify-between gap-3">
      <div class="inline-flex rounded-lg border border-border p-0.5 bg-muted/40" role="tablist" aria-label="Miền">
        <button v-for="r in regionTabs" :key="r" role="tab" :aria-selected="region === r" class="h-8 px-3.5 rounded-md text-sm whitespace-nowrap"
          :class="region === r ? 'bg-background shadow-sm font-medium' : 'text-muted-foreground hover:text-foreground'" @click="setRegion(r)">
          {{ r === 'all' ? 'Tất cả' : `Miền ${r}` }} <span class="text-xs text-muted-foreground tabular-nums">{{ regionCount(r) }}</span>
        </button>
      </div>
      <div class="flex items-center gap-1.5">
        <span class="text-xs text-muted-foreground mr-0.5">Giọng</span>
        <button v-for="g in (['', 'Nam', 'Nữ'] as const)" :key="g" class="h-8 px-3 rounded-full border text-xs whitespace-nowrap"
          :class="gender === g ? 'border-primary bg-primary/10 text-primary font-medium' : 'border-border text-muted-foreground hover:text-foreground'"
          :aria-pressed="gender === g" @click="gender = g">{{ g || 'Tất cả' }}</button>
      </div>
    </div>
    <div class="mt-3 grid gap-2">
      <label v-for="v in visible" :key="v.name" class="flex items-center gap-3 rounded-lg border px-4 py-2.5 cursor-pointer"
        :class="state.voice === v.name ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/50'">
        <input v-model="state.voice" type="radio" :value="v.name" class="h-4 w-4 accent-[hsl(var(--primary))]" />
        <span class="flex-1">
          <span class="font-medium text-sm">{{ v.name }}</span>
          <span v-if="v.name === used" class="ml-2 text-[11px] rounded-full bg-muted px-2 py-0.5 text-muted-foreground">Dùng lần trước</span>
          <span class="block text-xs text-muted-foreground">{{ desc(v) }}</span>
        </span>
        <button class="h-8 px-3 rounded-full border border-border text-xs flex items-center gap-1.5 hover:bg-muted disabled:opacity-50"
          :disabled="loadingVoice !== null && loadingVoice !== v.name" @click.prevent="sample(v)">
          <Loader2 v-if="loadingVoice === v.name" class="w-3.5 h-3.5 animate-spin" />
          <component :is="player.playing.value === v.name ? Pause : Play" v-else class="w-3.5 h-3.5" /> Nghe mẫu
        </button>
      </label>
      <p v-if="state.voices.length && !visible.length" class="text-sm text-muted-foreground">Không có giọng {{ gender.toLowerCase() }} ở miền này.</p>
    </div>
    <p v-if="sampleError || player.error.value" class="mt-3 text-sm text-destructive">{{ sampleError || player.error.value }}</p>
    <div class="mt-5 text-sm">
      <p class="font-medium">Câu nghe mẫu</p>
      <input v-model="state.sampleSentence" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" />
    </div>
  </div>
</template>
