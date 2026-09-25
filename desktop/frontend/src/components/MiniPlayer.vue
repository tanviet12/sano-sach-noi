<script setup lang="ts">
// Thanh nghe nhỏ ở đáy cửa sổ (wireframe D4): hiện khi có cuốn đang nạp và đang ở
// màn khác màn nghe. Bấm tên sách → mở lại màn nghe; ✕ → dừng hẳn.
import { Loader2, Pause, Play, RotateCcw, RotateCw, X } from 'lucide-vue-next'
import BookCover from '@/components/sano/BookCover.vue'
import { fmtClock } from '../lib/position'
import { openBook } from '../lib/store'
import { closePlayer, pctTrack, player, seek, toggle, track } from '../lib/player'
</script>

<template>
  <div class="shrink-0 border-t border-border bg-background" role="region" aria-label="Đang nghe">
    <div class="h-0.5 bg-muted"><div class="h-full bg-primary" :style="{ width: pctTrack + '%' }"></div></div>
    <div class="h-16 flex items-center gap-3 px-4">
      <template v-if="player.detail">
        <button class="flex items-center gap-3 min-w-0 flex-1 text-left" title="Mở màn nghe" @click="openBook(player.slug)">
          <span class="h-11 w-8 shrink-0 rounded overflow-hidden shadow">
            <img v-if="player.detail.coverUrl" :src="player.detail.coverUrl" alt="" class="h-full w-full object-cover" />
            <BookCover v-else :title="player.detail.title" class="h-full w-full rounded shadow-none" />
          </span>
          <span class="min-w-0">
            <span class="block text-sm font-medium truncate">{{ player.detail.title }}</span>
            <span class="block text-xs truncate" :class="player.error ? 'text-destructive' : 'text-muted-foreground'">
              <template v-if="player.error">{{ player.error }}</template>
              <template v-else>{{ track?.title }}<template v-if="player.detail.voice"> · Giọng {{ player.detail.voice }}</template></template>
            </span>
          </span>
        </button>
        <span class="text-xs text-muted-foreground tabular-nums">{{ fmtClock(player.time) }} / {{ fmtClock(player.duration) }}</span>
        <button aria-label="Lùi 15 giây" title="Lùi 15 giây" class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" @click="seek(-15)"><RotateCcw class="w-4 h-4" /></button>
        <button :aria-label="player.playing ? 'Dừng' : 'Phát'" class="h-10 w-10 grid place-items-center rounded-full bg-primary text-primary-foreground" @click="toggle">
          <Pause v-if="player.playing" class="w-5 h-5" /><Play v-else class="w-5 h-5 ml-0.5" />
        </button>
        <button aria-label="Tới 30 giây" title="Tới 30 giây" class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" @click="seek(30)"><RotateCw class="w-4 h-4" /></button>
      </template>
      <span v-else class="flex-1 flex items-center gap-2 text-sm text-muted-foreground">
        <template v-if="player.error"><span class="text-destructive">{{ player.error }}</span></template>
        <template v-else><Loader2 class="w-4 h-4 animate-spin" /> Đang mở sách…</template>
      </span>
      <button aria-label="Dừng nghe" title="Dừng nghe" class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted text-muted-foreground" @click="closePlayer"><X class="w-4 h-4" /></button>
    </div>
  </div>
</template>
