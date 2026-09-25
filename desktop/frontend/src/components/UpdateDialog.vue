<script setup lang="ts">
// Hộp cập nhật (bám wireframe WfDesktop, P3-13 + bổ sung P3-7): có gì mới →
// Cập nhật ngay (tải + kiểm chữ ký ed25519 + SHA256 ở phần Go) → khởi động lại.
// Đang render: hẹn cập nhật khi render xong. Máy không tự thay được (chạy từ
// .dmg, thư mục không ghi được...): nút Mở trang tải như trước. Tải hỏng / chữ
// ký sai: báo rõ, bản đang dùng giữ nguyên.
import { computed, ref } from 'vue'
import { AlertTriangle, ArrowUpCircle, Check, ExternalLink, Info, Loader2, RefreshCw, RotateCcw, ShieldAlert, ShieldCheck, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { openURL } from '../lib/backend'
import { applyUpdate, applyUpdateLater, cancelUpdate, closeUpdate, rendering, startUpdate, state } from '../lib/store'

const info = computed(() => state.updateInfo)
const upd = computed(() => state.upd)
const published = computed(() => {
  const d = info.value?.published ? new Date(info.value.published) : null
  return d && !isNaN(d.getTime()) ? d.toLocaleDateString('vi-VN') : ''
})
const mb = (n: number) => `${Math.max(1, Math.round(n / (1 << 20)))} MB`

// Màn đang hiện: lỗi (tải/kiểm/thay hỏng) > đang tải > sẵn sàng > có gì mới.
const screen = computed(() => {
  if (upd.value.phase === 'error' || state.updError) return 'error'
  if (upd.value.phase === 'downloading') return 'downloading'
  if (upd.value.phase === 'ready') return 'ready'
  return 'info'
})
const errText = computed(() => state.updError || upd.value.error)
const errTitle = computed(() => {
  const e = errText.value.toLowerCase()
  if (e.includes('chữ ký')) return 'Chữ ký bản phát hành không hợp lệ'
  if (e.includes('sha256')) return 'File tải về không khớp mã đã ký'
  if (e.includes('render') || e.includes('đang cài') || e.includes('đang xuất')) return 'Chưa cập nhật được lúc này'
  return 'Tải bản mới không thành công'
})
// Lỗi đã biết: câu giải thích gọn (như wireframe); lỗi khác: nguyên văn từ phần Go.
const errBody = computed(() => {
  const e = errText.value.toLowerCase()
  if (e.includes('chữ ký bản phát hành không hợp lệ')) return 'Đây không phải bản chính thức của Sano nên đã huỷ cập nhật và xoá file tải về.'
  if (e.includes('không khớp mã sha256')) return 'File có thể bị hỏng hoặc bị tráo trên đường tải nên đã xoá và huỷ cập nhật.'
  return errText.value
})
// Lỗi do an toàn (chữ ký / mã sai): dùng biểu tượng khiên.
const errShield = computed(() => /chữ ký|sha256/i.test(errText.value))
const pct = computed(() => (upd.value.total > 0 ? Math.min(100, Math.round((upd.value.done / upd.value.total) * 100)) : 0))
const version = computed(() => upd.value.version || info.value?.version || '')
const agreeAfterRender = ref(state.updateAfterRender)
const restarting = ref(false)

function close() {
  if (screen.value !== 'downloading') closeUpdate()
}

function download() {
  if (info.value) openURL(info.value.url)
  closeUpdate()
}

function scheduleAfterRender() {
  state.updateAfterRender = true
  closeUpdate()
}

async function restart() {
  restarting.value = true
  await applyUpdate() // thành công thì Sano thoát; lỗi thì hiện màn lỗi
  restarting.value = false
}

function retry() {
  state.updError = ''
  void startUpdate()
}
</script>

<template>
  <div v-if="info" class="absolute inset-0 bg-background/70 backdrop-blur-sm grid place-items-center z-20" @keydown.esc="close">
    <div role="dialog" aria-modal="true" aria-labelledby="upd-title" class="w-[460px] rounded-xl border border-border bg-card text-card-foreground shadow-2xl p-6">
      <div class="flex items-start justify-between">
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 rounded-full bg-primary/10 grid place-items-center"><ArrowUpCircle class="w-5 h-5 text-primary" /></div>
          <div>
            <h2 id="upd-title" class="font-semibold">{{ screen === 'ready' ? 'Sẵn sàng cập nhật' : screen === 'error' ? 'Không cập nhật được' : `Có bản mới ${info.version}` }}</h2>
            <p class="text-xs text-muted-foreground">Đang dùng {{ state.version }}<template v-if="published"> · phát hành {{ published }}</template><template v-if="info.size"> · {{ mb(info.size) }}</template></p>
          </div>
        </div>
        <button v-if="screen !== 'downloading'" aria-label="Đóng" class="h-8 w-8 grid place-items-center rounded-md hover:bg-muted" @click="close"><X class="w-4 h-4" /></button>
      </div>

      <template v-if="screen === 'info'">
        <template v-if="info.notes.length">
          <p class="mt-5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Có gì mới</p>
          <ul class="mt-2 space-y-1.5 text-sm list-disc pl-5">
            <li v-for="(n, i) in info.notes" :key="i">{{ n }}</li>
          </ul>
        </template>
        <a class="mt-2 inline-block text-xs text-primary hover:underline" :href="info.url" target="_blank" rel="noopener">Xem đầy đủ thay đổi</a>

        <p v-if="!info.autoUpdate" class="mt-4 rounded-lg border border-border bg-muted/40 p-3 text-sm text-foreground/80 flex gap-2">
          <Info class="w-4 h-4 mt-0.5 shrink-0 text-muted-foreground" />
          <span><template v-if="info.manual">{{ info.manual }}. </template>Tải bản cài mới rồi cài đè — sách, tiến độ nghe và bộ đọc giữ nguyên.</span>
        </p>
        <div v-else-if="rendering" class="mt-4 rounded-lg border border-rag-amber/40 bg-rag-amber/10 p-3 text-sm">
          <p class="font-medium text-rag-amber flex items-center gap-1.5"><AlertTriangle class="w-4 h-4" /> Đang render "{{ state.render?.title }}"</p>
          <p class="mt-1 text-foreground/80">Cập nhật cần khởi động lại Sano. Chọn cập nhật khi render xong để không mất tiến độ.</p>
          <label class="mt-2 flex items-center gap-2"><input v-model="agreeAfterRender" type="checkbox" class="h-4 w-4 accent-[hsl(var(--primary))]" /> Tự cập nhật khi render xong</label>
        </div>

        <div class="mt-6 flex justify-end gap-2">
          <Button variant="outline" @click="close">Để sau</Button>
          <Button v-if="!info.autoUpdate" @click="download"><ExternalLink class="w-4 h-4" /> Mở trang tải</Button>
          <Button v-else-if="rendering" :disabled="!agreeAfterRender" @click="scheduleAfterRender">Hẹn cập nhật</Button>
          <Button v-else @click="startUpdate()">Cập nhật ngay</Button>
        </div>
      </template>

      <template v-else-if="screen === 'downloading'">
        <div class="mt-6 space-y-3 text-sm">
          <div>
            <div class="flex justify-between text-xs mb-1">
              <span class="flex items-center gap-1.5"><Loader2 class="w-3.5 h-3.5 animate-spin text-primary" />Đang tải bản {{ version }}</span>
              <span class="tabular-nums text-muted-foreground"><template v-if="upd.total">{{ mb(upd.done) }} / {{ mb(upd.total) }}</template></span>
            </div>
            <div class="h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary rounded-full transition-[width]" :style="{ width: pct + '%' }"></div></div>
          </div>
          <p class="flex items-center gap-1.5 text-xs text-muted-foreground"><span class="w-3.5 h-3.5 rounded-full border border-border inline-block"></span>Kiểm tra chữ ký bản phát hành</p>
        </div>
        <div class="mt-6 flex justify-between">
          <Button variant="ghost" @click="cancelUpdate()">Huỷ</Button>
        </div>
      </template>

      <template v-else-if="screen === 'error'">
        <div class="mt-5 rounded-lg border border-rag-red/40 bg-rag-red/10 p-3 text-sm">
          <p class="font-medium text-rag-red flex items-center gap-1.5"><component :is="errShield ? ShieldAlert : AlertTriangle" class="w-4 h-4 shrink-0" /> {{ errTitle }}</p>
          <p class="mt-1 text-foreground/80 first-letter:uppercase">{{ errBody }}</p>
        </div>
        <p class="mt-3 text-xs text-muted-foreground">Bản đang dùng giữ nguyên. Có thể thử lại sau, hoặc tải bản cài ở trang phát hành chính thức.</p>
        <div class="mt-6 flex justify-end gap-2">
          <Button variant="outline" @click="close">Đóng</Button>
          <Button variant="outline" @click="download"><ExternalLink class="w-4 h-4" /> Mở trang tải</Button>
          <Button v-if="upd.phase === 'error'" @click="retry"><RotateCcw class="w-4 h-4" /> Thử lại</Button>
        </div>
      </template>

      <template v-else>
        <div class="mt-5 rounded-lg border border-border p-3 text-sm space-y-2">
          <p class="flex items-center gap-2"><Check class="w-4 h-4 text-rag-green" /> Đã tải bản {{ version }}</p>
          <p class="flex items-center gap-2"><ShieldCheck class="w-4 h-4 text-rag-green" /> Chữ ký hợp lệ — bản phát hành chính thức của Sano</p>
        </div>
        <p class="mt-3 text-xs text-muted-foreground">Sách, tiến độ nghe và bộ đọc giữ nguyên. Sano sẽ tự mở lại sau vài giây.</p>
        <div class="mt-6 flex justify-end gap-2">
          <Button variant="outline" :disabled="restarting" @click="applyUpdateLater()">Khởi động lại sau</Button>
          <Button :disabled="restarting" @click="restart"><Loader2 v-if="restarting" class="w-4 h-4 animate-spin" /><RefreshCw v-else class="w-4 h-4" /> Khởi động lại</Button>
        </div>
      </template>
    </div>
  </div>
</template>
