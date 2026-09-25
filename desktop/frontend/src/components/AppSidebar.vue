<script setup lang="ts">
import { Library, FilePlus2, Settings, Info, Loader2, ArrowUpCircle, LifeBuoy, ExternalLink, RefreshCw } from 'lucide-vue-next'
import faviconUrl from '@/assets/favicon.svg'
import { AUTHOR_FB, DOCS } from '../lib/mock'
import { go, remainMin, renderPct, rendering, state, type View } from '../lib/store'

const nav = [
  { key: 'library' as View, label: 'Thư viện', icon: Library },
  { key: 'create' as View, label: 'Tạo sách mới', icon: FilePlus2 },
  { key: 'settings' as View, label: 'Cài đặt', icon: Settings },
  { key: 'about' as View, label: 'Giới thiệu', icon: Info },
]
</script>

<template>
  <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
    <div class="px-4 py-4 flex items-center gap-2">
      <img :src="faviconUrl" alt="" class="h-7 w-7 rounded-md" />
      <span class="font-semibold tracking-tight">Sano</span>
    </div>
    <nav class="px-2 space-y-0.5">
      <button
        v-for="n in nav" :key="n.key"
        class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
        :class="state.view === n.key || (n.key === 'library' && state.view === 'player') ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
        @click="go(n.key)"
      >
        <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
      </button>
    </nav>

    <div class="px-2 mt-2">
      <a :href="DOCS" target="_blank" rel="noopener" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground hover:bg-muted hover:text-foreground">
        <LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" />
      </a>
    </div>

    <div class="flex-1"></div>

    <!-- Thẻ tiến độ render (hiện ở mọi màn khi đang render) -->
    <button v-if="rendering && state.view !== 'create'" class="m-3 rounded-lg border border-border bg-background p-3 text-left hover:border-primary/50" @click="go('create')">
      <div class="flex items-center gap-1.5 text-xs font-medium"><Loader2 class="w-3.5 h-3.5 animate-spin text-primary" /> Đang render</div>
      <p class="mt-1 text-xs text-muted-foreground truncate">{{ state.render?.title }}</p>
      <div class="mt-2 h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary rounded-full" :style="{ width: renderPct + '%' }"></div></div>
      <div class="mt-1 flex justify-between text-[11px] text-muted-foreground tabular-nums"><span>{{ renderPct }}%</span><span>còn ~{{ remainMin }} phút</span></div>
    </button>
    <button v-if="state.updateInfo" class="mx-3 mb-2 flex items-center gap-2 rounded-md px-2 h-8 text-xs text-primary hover:bg-primary/10" @click="state.update = 'info'">
      <template v-if="state.upd.applyOnQuit"><RefreshCw class="w-4 h-4" /> Khởi động lại để cập nhật</template>
      <template v-else><ArrowUpCircle class="w-4 h-4" /> Có bản mới {{ state.updateInfo.version }}</template>
    </button>
    <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">
      Phiên bản {{ state.version || '…' }} · mã nguồn mở<br />
      Tác giả <a :href="AUTHOR_FB" target="_blank" rel="noopener" class="hover:text-foreground underline-offset-2 hover:underline">Bùi Tấn Việt</a>
    </div>
  </aside>
</template>
