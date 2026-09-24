<script setup lang="ts">
// Toàn văn điều khoản sử dụng (dùng ở màn đồng ý và hộp xem lại trong Giới thiệu).
import { TERMS_BLOCKS } from '../lib/terms'
</script>

<template>
  <div class="text-sm leading-relaxed">
    <template v-for="(b, i) in TERMS_BLOCKS" :key="i">
      <h2 v-if="b.kind === 'h2'" class="mt-4 first:mt-0 font-semibold" v-html="b.html"></h2>
      <p v-else-if="b.kind === 'li'" class="relative mt-1 pl-4 before:absolute before:left-0 before:text-muted-foreground before:content-['•']" v-html="b.html"></p>
      <hr v-else-if="b.kind === 'hr'" class="my-3 border-border" />
      <p v-else class="mt-2 text-muted-foreground" v-html="b.html"></p>
    </template>
  </div>
</template>
