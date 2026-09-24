<script setup lang="ts">
// Giấy phép bên thứ ba (ghi công VieNeu-TTS + mô hình, uv, Python, ffmpeg...).
// Dùng trong màn Giới thiệu. Toàn văn lấy từ phần Go (desktop/licenses/, nhúng
// vào app); xem bằng trình duyệt thì dẫn tới file trong mã nguồn.
import { ref } from 'vue'
import { ChevronDown, ExternalLink, Scale } from 'lucide-vue-next'
import { isDesktop, openURL, thirdPartyNotices } from '../lib/backend'
import { REPO } from '../lib/mock'

const items = [
  { name: 'VieNeu-TTS', license: 'Apache-2.0', by: 'Phạm Nguyễn Ngọc Bảo', url: 'https://github.com/pnnbao97/VieNeu-TTS' },
  { name: 'Mô hình VieNeu-TTS v3 Turbo', license: 'Apache-2.0', by: 'pnnbao-ump', url: 'https://huggingface.co/pnnbao-ump/VieNeu-TTS-v3-Turbo' },
  { name: 'MOSS Audio Tokenizer Nano', license: 'Apache-2.0', by: 'OpenMOSS Team', url: 'https://huggingface.co/OpenMOSS-Team/MOSS-Audio-Tokenizer-Nano-ONNX' },
  { name: 'uv', license: 'MIT / Apache-2.0', by: 'Astral', url: 'https://github.com/astral-sh/uv' },
  { name: 'Python', license: 'PSF', by: 'Python Software Foundation', url: 'https://www.python.org' },
  { name: 'ffmpeg (chỉ tải khi máy chưa có)', license: 'GPLv3', by: 'FFmpeg', url: 'https://ffmpeg.org' },
  { name: 'Wails, Vue, Tailwind CSS', license: 'MIT', by: '', url: 'https://wails.io' },
]

const open = ref(false)
const full = ref('')

async function toggle() {
  open.value = !open.value
  if (open.value && !full.value && isDesktop()) full.value = await thirdPartyNotices()
}
</script>

<template>
  <div class="text-sm" data-testid="third-party-licenses">
    <button type="button" class="w-full flex items-center gap-2.5 px-4 py-3 hover:bg-muted/50 text-left" @click="toggle">
      <Scale class="w-4 h-4 text-muted-foreground" />
      <span class="flex-1">Giấy phép bên thứ ba</span>
      <span class="text-xs text-muted-foreground">VieNeu-TTS (Apache-2.0) · ffmpeg · …</span>
      <ChevronDown class="w-4 h-4 text-muted-foreground transition-transform" :class="open && 'rotate-180'" />
    </button>
    <div v-if="open" class="px-4 pb-4 space-y-3">
      <p class="text-xs text-muted-foreground">
        Giọng đọc của Sano là VieNeu-TTS (Apache-2.0) — cảm ơn tác giả đã mở mã nguồn và mô hình. Bộ đọc được tải từ nguồn gốc, đúng phiên bản ghim, không chỉnh sửa.
      </p>
      <ul class="rounded-md border border-border divide-y divide-border text-xs">
        <li v-for="it in items" :key="it.name" class="flex items-center gap-2 px-3 py-2">
          <span class="flex-1 min-w-0"><span class="font-medium">{{ it.name }}</span><span v-if="it.by" class="text-muted-foreground"> · {{ it.by }}</span></span>
          <span class="text-muted-foreground shrink-0">{{ it.license }}</span>
          <a :href="it.url" class="text-muted-foreground hover:text-foreground shrink-0" @click.prevent="openURL(it.url)"><ExternalLink class="w-3.5 h-3.5" /></a>
        </li>
      </ul>
      <pre v-if="full" class="max-h-72 overflow-auto rounded-md bg-muted/50 p-3 text-[11px] leading-relaxed whitespace-pre-wrap select-text">{{ full }}</pre>
      <a v-else :href="REPO + '/blob/main/desktop/licenses/THIRD-PARTY-NOTICES.md'" class="text-xs text-primary hover:underline" @click.prevent="openURL(REPO + '/blob/main/desktop/licenses/THIRD-PARTY-NOTICES.md')">Xem toàn văn giấy phép</a>
    </div>
  </div>
</template>
