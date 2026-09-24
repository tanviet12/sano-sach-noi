<script setup lang="ts">
// Thanh tiêu đề 36px như wireframe. macOS: nút đóng/thu nhỏ thật của hệ điều hành
// nằm đè bên trái (TitleBarHiddenInset). Windows/Linux dùng khung cửa sổ gốc → ẩn.
// Trình duyệt (vite dev): vẽ 3 chấm giả để so với wireframe.
defineProps<{ os: string }>()
</script>

<template>
  <div
    v-if="os === 'darwin' || os === 'browser'"
    class="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border bg-muted/50"
    style="--wails-draggable: drag"
  >
    <template v-if="os === 'browser'">
      <span class="h-3 w-3 rounded-full bg-rag-red"></span>
      <span class="h-3 w-3 rounded-full bg-rag-amber"></span>
      <span class="h-3 w-3 rounded-full bg-rag-green"></span>
    </template>
    <span class="flex-1 text-center text-xs text-muted-foreground">Sano</span>
  </div>
</template>
