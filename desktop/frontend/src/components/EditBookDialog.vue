<script setup lang="ts">
// Hộp "Sửa thông tin sách" (menu ⋯ trên bìa, wireframe D2 + D5): tên, tác giả, danh mục,
// bộ sách + số tập.
// Lưu → phần Go ghi metadata.json + manifest.json trong gói zip. Lời giới thiệu đã
// đọc giữ nguyên.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Loader2, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { errText, updateBookInfo, type LibraryBook } from '../lib/backend'
import { refreshInfo } from '../lib/player'
import CategoryPicker from './CategoryPicker.vue'
import { seriesKey } from '../lib/find'

// series: các bộ sách đang có kèm danh sách số tập (của các cuốn KHÁC cuốn đang sửa)
const props = defineProps<{ book: LibraryBook; categories: [string, number][]; series: [string, number[]][] }>()
const emit = defineEmits<{ close: []; saved: [book: LibraryBook] }>()

const title = ref(props.book.title)
const author = ref(props.book.author)
const category = ref(props.book.category)
const seriesName = ref(props.book.series ?? '')
const volume = ref<number>(props.book.volume || 0)
const seriesGroups = computed<[string, number][]>(() => props.series.map(([n, v]) => [n, v.length]))
const taken = computed(() => props.series.find(([n]) => seriesKey(n) === seriesKey(seriesName.value))?.[1] ?? [])
// Chọn bộ khác → gợi ý tập kế tiếp của bộ đó; bỏ bộ → xoá số tập.
watch(seriesName, (n, old) => {
  if (!n) volume.value = 0
  else if (seriesKey(n) !== seriesKey(old ?? '')) volume.value = seriesKey(n) === seriesKey(props.book.series ?? '') && props.book.volume ? props.book.volume : Math.max(0, ...taken.value) + 1
})
const saving = ref(false)
const error = ref('')
const titleInput = ref<HTMLInputElement | null>(null)

async function save() {
  if (!title.value.trim() || saving.value) return
  saving.value = true
  error.value = ''
  try {
    const b = await updateBookInfo(props.book.slug, {
      title: title.value.trim(), author: author.value.trim(), category: category.value,
      series: seriesName.value, volume: seriesName.value ? Math.max(0, Math.floor(Number(volume.value) || 0)) : 0,
    })
    void refreshInfo(props.book.slug) // đang phát cuốn này → cập nhật tên ở thanh nghe nhỏ
    emit('saved', b)
  } catch (e) {
    error.value = errText(e)
  } finally {
    saving.value = false
  }
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && !saving.value) emit('close')
}
onMounted(() => {
  document.addEventListener('keydown', onKey)
  titleInput.value?.focus()
})
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="fixed inset-0 bg-black/40 grid place-items-center z-40" @mousedown.self="!saving && emit('close')">
    <form role="dialog" aria-modal="true" aria-labelledby="edit-book-title" class="w-[460px] rounded-xl border border-border bg-background shadow-2xl p-5" @submit.prevent="save">
      <div class="flex items-start justify-between">
        <h2 id="edit-book-title" class="font-semibold">Sửa thông tin sách</h2>
        <button type="button" aria-label="Đóng" class="text-muted-foreground hover:text-foreground" @click="emit('close')"><X class="w-4 h-4" /></button>
      </div>
      <div class="mt-4 space-y-3 text-sm">
        <label class="block">Tên sách<input ref="titleInput" v-model="title" maxlength="200" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" /></label>
        <label class="block">Tác giả<input v-model="author" maxlength="200" placeholder="Không bắt buộc" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" /></label>
        <div>
          <span>Danh mục <span class="text-muted-foreground">· không bắt buộc</span></span>
          <CategoryPicker v-model="category" :categories="categories" class="mt-1" />
        </div>
        <div class="flex gap-3">
          <div class="flex-1 min-w-0">
            <span>Bộ sách <span class="text-muted-foreground">· không bắt buộc</span></span>
            <CategoryPicker v-model="seriesName" kind="series" :categories="seriesGroups" class="mt-1" />
          </div>
          <label class="w-24 block" :class="!seriesName && 'opacity-40'">Tập số
            <input :value="volume || ''" @input="volume = Math.floor(Number(($event.target as HTMLInputElement).value)) || 0" type="number" min="1" max="999" :disabled="!seriesName" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" />
          </label>
        </div>
        <p v-if="seriesName && taken.length" class="text-xs text-muted-foreground">Bộ "{{ seriesName }}" đang có tập {{ [...taken].sort((a, b) => a - b).join(', ') }}. Trùng số tập với cuốn khác thì Sano báo để bạn chọn lại.</p>
        <p class="text-xs text-muted-foreground">Chỉ đổi thông tin hiển thị và gói zip. Lời giới thiệu đầu sách đã đọc giữ nguyên, muốn đọc lại tên mới thì tạo lại sách.</p>
        <p v-if="error" class="text-xs text-destructive">{{ error }}</p>
      </div>
      <div class="mt-5 flex justify-end gap-2">
        <Button type="button" variant="outline" :disabled="saving" @click="emit('close')">Huỷ</Button>
        <Button type="submit" :disabled="saving || !title.trim()"><Loader2 v-if="saving" class="w-4 h-4 animate-spin" /> Lưu</Button>
      </div>
    </form>
  </div>
</template>
