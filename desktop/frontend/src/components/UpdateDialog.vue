<script setup lang="ts">
// Hộp "Có bản mới": ghi chú phát hành thật từ GitHub Releases + nút mở trang tải.
// Người dùng tự tải bản cài; sách, tiến độ nghe và bộ đọc giữ nguyên khi cài đè.
// Tự tải và thay (cần chữ ký số) làm ở bản sau.
import { computed } from 'vue'
import { ArrowUpCircle, ExternalLink, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { openURL } from '../lib/backend'
import { state } from '../lib/store'

const info = computed(() => state.updateInfo)
const published = computed(() => {
  const d = info.value?.published ? new Date(info.value.published) : null
  return d && !isNaN(d.getTime()) ? d.toLocaleDateString('vi-VN') : ''
})

function close() {
  state.update = 'closed'
}

function download() {
  if (info.value) openURL(info.value.url)
  close()
}
</script>

<template>
  <div v-if="info" class="absolute inset-0 bg-background/70 backdrop-blur-sm grid place-items-center z-20" @keydown.esc="close">
    <div role="dialog" aria-modal="true" aria-labelledby="upd-title" class="w-[460px] rounded-xl border border-border bg-card text-card-foreground shadow-2xl p-6">
      <div class="flex items-start justify-between">
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 rounded-full bg-primary/10 grid place-items-center"><ArrowUpCircle class="w-5 h-5 text-primary" /></div>
          <div>
            <h2 id="upd-title" class="font-semibold">Có bản mới {{ info.version }}</h2>
            <p class="text-xs text-muted-foreground">Đang dùng {{ state.version }}<template v-if="published"> · phát hành {{ published }}</template></p>
          </div>
        </div>
        <button aria-label="Đóng" class="h-8 w-8 grid place-items-center rounded-md hover:bg-muted" @click="close"><X class="w-4 h-4" /></button>
      </div>

      <template v-if="info.notes.length">
        <p class="mt-5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Có gì mới</p>
        <ul class="mt-2 space-y-1.5 text-sm list-disc pl-5">
          <li v-for="(n, i) in info.notes" :key="i">{{ n }}</li>
        </ul>
      </template>

      <p class="mt-4 text-sm text-muted-foreground">
        Tải bản cài mới rồi cài đè lên bản đang dùng. Sách, tiến độ nghe và bộ đọc giữ nguyên.
      </p>

      <div class="mt-6 flex justify-end gap-2">
        <Button variant="outline" @click="close">Để sau</Button>
        <Button @click="download"><ExternalLink class="w-4 h-4" /> Mở trang tải</Button>
      </div>
    </div>
  </div>
</template>
