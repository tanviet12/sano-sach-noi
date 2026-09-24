<script setup lang="ts">
// B2 Mục lục: mục lục thật từ file Word; trang mục lục gốc bỏ tick sẵn.
import { estListen, estRender, reloadOutline, state, totalChars } from '../../lib/store'
</script>

<template>
  <div class="max-w-3xl">
    <div class="flex items-end justify-between">
      <div>
        <h1 class="text-xl font-semibold tracking-tight">Chọn phần sẽ đọc</h1>
        <p class="text-sm text-muted-foreground">Bỏ tick trang bìa, mục lục gốc hay phần không cần nghe.</p>
      </div>
      <div class="text-right text-sm">
        <p class="tabular-nums"><span class="font-medium">{{ totalChars.toLocaleString('vi-VN') }}</span> ký tự</p>
        <p class="text-xs text-muted-foreground">Nghe ~{{ estListen }} phút · render ~{{ estRender }} phút trên máy này</p>
      </div>
    </div>
    <label class="mt-4 flex items-center gap-2 text-sm">
      <input v-model="state.keepHeadingNumbers" type="checkbox" class="h-4 w-4 accent-[hsl(var(--primary))]" @change="reloadOutline" />
      Đọc cả số đầu tiêu đề (vd "1.2. Lắng nghe" đọc là "một chấm hai, Lắng nghe")
    </label>
    <div class="mt-4 rounded-lg border border-border divide-y divide-border">
      <div v-for="c in state.toc" :key="c.stems[0] ?? c.title">
        <label class="flex items-center gap-3 px-4 py-3 cursor-pointer" :class="!c.on && 'text-muted-foreground'">
          <input v-model="c.on" type="checkbox" class="h-4 w-4 accent-[hsl(var(--primary))]" />
          <span class="flex-1 font-medium text-sm">{{ c.title }}</span>
          <span v-if="c.note" class="text-xs text-rag-amber">{{ c.note }}</span>
          <span class="text-xs tabular-nums text-muted-foreground w-24 text-right">{{ c.chars.toLocaleString('vi-VN') }} ký tự</span>
        </label>
        <label v-for="s in c.sections" :key="s.stem" class="flex items-center gap-3 pl-11 pr-4 py-2 cursor-pointer text-sm" :class="(!c.on || !s.on) && 'text-muted-foreground'">
          <input v-model="s.on" type="checkbox" :disabled="!c.on" class="h-4 w-4 accent-[hsl(var(--primary))]" />
          <span class="flex-1">{{ s.title }}</span>
          <span v-if="s.note" class="text-xs text-rag-amber">{{ s.note }}</span>
          <span class="text-xs tabular-nums text-muted-foreground w-24 text-right">{{ s.chars.toLocaleString('vi-VN') }} ký tự</span>
        </label>
      </div>
    </div>
  </div>
</template>
