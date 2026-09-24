<script setup lang="ts">
// Tên nhóm thanh bên phía trên tiêu đề trang hướng dẫn (wireframe WfDocsPage: "Dùng Sano").
import { computed } from 'vue'
import { useData } from 'vitepress'

type Item = { text?: string; link?: string; items?: Item[] }
const { theme, page } = useData()

const group = computed(() => {
  const path = '/' + page.value.relativePath.replace(/(index)?\.md$/, '')
  const groups = (theme.value.sidebar ?? []) as Item[]
  return groups.find((g) => g.items?.some((i) => i.link === path))?.text ?? ''
})
</script>

<template>
  <p v-if="group" class="sano-eyebrow">{{ group }}</p>
</template>
