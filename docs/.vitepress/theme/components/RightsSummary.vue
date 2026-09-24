<script setup lang="ts">
// Bản quyền (wireframe WfLanding mục 6): 3 ý như màn đồng ý điều khoản trong app.
import { withBase } from 'vitepress'
import { ArrowRight, Ban, FileText, ShieldCheck } from 'lucide-vue-next'

const rights = [
  { icon: FileText, title: 'Tài liệu bạn có quyền dùng', desc: 'Tài liệu của bạn, sách hết bảo hộ, hoặc được tác giả cho phép.' },
  { icon: Ban, title: 'Không dùng sách còn bản quyền', desc: 'Trừ khi được tác giả hoặc chủ sở hữu cho phép.' },
  { icon: ShieldCheck, title: 'Chạy trên máy bạn', desc: 'Tài liệu không gửi đi đâu. Bạn tự chịu trách nhiệm nội dung.' },
]
</script>

<template>
  <section id="ban-quyen" class="sano-section">
    <div class="sano-container">
      <h2 class="sano-h2">Bản quyền</h2>
      <p class="sano-lead">Sano là công cụ cho tài liệu của chính bạn. Trước khi dùng, phần mềm yêu cầu đồng ý điều khoản sử dụng.</p>
      <div class="grid">
        <div v-for="r in rights" :key="r.title" class="item">
          <span class="sano-icon-tile sm"><component :is="r.icon" :size="16" aria-hidden="true" /></span>
          <div class="txt">
            <p class="t">{{ r.title }}</p>
            <p class="d sano-muted">{{ r.desc }}</p>
          </div>
        </div>
      </div>
      <a :href="withBase('/dieu-khoan-su-dung')" class="more">Đọc điều khoản sử dụng đầy đủ <ArrowRight :size="16" aria-hidden="true" /></a>
    </div>
  </section>
</template>

<style scoped>
.grid {
  margin-top: 24px;
  display: grid;
  gap: 12px;
}
@media (min-width: 768px) {
  .grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
.item {
  display: flex;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
}
.sm {
  width: 36px;
  height: 36px;
  border-radius: 6px;
}
.txt {
  min-width: 0;
}
.t {
  margin: 0;
  font-size: 14px;
  font-weight: 500;
}
.d {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
}
.more {
  margin-top: 16px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  font-weight: 500;
  color: var(--vp-c-brand-1);
  text-decoration: none;
}
.more:hover {
  text-decoration: underline;
}
</style>
