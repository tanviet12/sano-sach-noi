<script setup lang="ts">
// Chỗ trống cho ảnh chụp phần mềm chưa có. Dùng được trong markdown: <ShotPlaceholder label="Bước Nạp file" />
import { ImageIcon } from 'lucide-vue-next'

withDefaults(defineProps<{ label?: string; ratio?: string }>(), { label: '', ratio: '16 / 10' })
</script>

<template>
  <div class="sano-placeholder shot" :style="{ aspectRatio: ratio }" role="img" :aria-label="(label ? label + ' — ' : '') + 'ảnh chụp sẽ bổ sung'">
    <span class="in">
      <ImageIcon :size="20" aria-hidden="true" />
      <span v-if="label" class="label">{{ label }}</span>
      <span>Ảnh chụp sẽ bổ sung</span>
    </span>
  </div>
</template>

<style scoped>
.shot {
  width: 100%;
  margin: 16px 0;
}
.in {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  opacity: 0.85;
}
.label {
  font-weight: 500;
  color: var(--vp-c-text-1);
}
</style>
