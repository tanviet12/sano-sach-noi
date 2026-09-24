<script setup lang="ts">
// Kế hoạch tiếp theo (wireframe WfLanding mục 7). Không ghi ngày.
import { Server } from 'lucide-vue-next'
</script>

<template>
  <section class="sano-section">
    <div class="sano-container">
      <h2 class="sano-h2">Kế hoạch tiếp theo</h2>
      <div class="box">
        <span class="tile"><Server :size="20" aria-hidden="true" /></span>
        <div>
          <div class="head">
            <p class="t">Máy chủ nghe sách riêng</p>
            <span class="tag">Giai đoạn sau</span>
          </div>
          <p class="d sano-muted">
            Dành cho gia đình và nhóm nhỏ: tự dựng trên máy chủ của bạn, nghe qua web và app điện thoại, nhớ vị trí nghe giữa các
            máy. Bản hiện tại là phần mềm máy tính.
          </p>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.box {
  margin-top: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px;
  border: 1px dashed var(--vp-c-divider);
  border-radius: 12px;
}
@media (min-width: 640px) {
  .box {
    flex-direction: row;
    align-items: flex-start;
  }
}
.tile {
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  border-radius: 8px;
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-text-2);
}
.head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.t {
  margin: 0;
  font-weight: 500;
}
.tag {
  border-radius: 9999px;
  background: var(--vp-c-bg-soft);
  padding: 2px 10px;
  font-size: 12px;
  font-weight: 600;
}
.d {
  margin: 4px 0 0;
  font-size: 14px;
  line-height: 1.6;
}
</style>
