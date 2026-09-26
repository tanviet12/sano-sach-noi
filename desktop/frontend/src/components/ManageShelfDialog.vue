<script setup lang="ts">
// Hộp "Quản lý danh mục và bộ sách" (wireframe D5, trạng thái 5): đổi tên áp cho mọi
// cuốn (danh mục trùng tên có sẵn thì gộp), xoá thì sách giữ nguyên — danh mục về
// "Chưa phân loại", các tập của bộ tách thành sách lẻ.
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { Layers, Loader2, Pencil, Tag, Trash2, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { deleteCategory, deleteSeries, errText, renameCategory, renameSeries } from '../lib/backend'
import { MAX_CATEGORY_LEN, MAX_SERIES_LEN } from '../lib/find'

type Tab = 'cat' | 'series'
const props = defineProps<{ categories: [string, number][]; series: [string, number][]; tab?: Tab }>()
const emit = defineEmits<{ close: []; changed: [] }>()

const tab = ref<Tab>(props.tab ?? 'cat')
const editing = ref<string | null>(null)
const draft = ref('')
const confirmDel = ref<string | null>(null)
const busy = ref(false)
const error = ref('')
const input = ref<HTMLInputElement[] | null>(null)

function setTab(t: Tab) {
  tab.value = t
  editing.value = confirmDel.value = null
  error.value = ''
}
async function startEdit(name: string) {
  confirmDel.value = null
  error.value = ''
  editing.value = name
  draft.value = name
  await nextTick()
  input.value?.[0]?.focus()
}
async function run(fn: () => Promise<number>) {
  busy.value = true
  error.value = ''
  try {
    await fn()
    editing.value = confirmDel.value = null
    emit('changed')
  } catch (e) {
    error.value = errText(e)
  } finally {
    busy.value = false
  }
}
function saveRename(old: string) {
  const name = draft.value.trim()
  if (!name || busy.value) return
  if (name === old) {
    editing.value = null
    return
  }
  void run(() => (tab.value === 'cat' ? renameCategory(old, name) : renameSeries(old, name)))
}
function remove(name: string) {
  if (busy.value) return
  void run(() => (tab.value === 'cat' ? deleteCategory(name) : deleteSeries(name)))
}

function onKey(e: KeyboardEvent) {
  if (e.key !== 'Escape' || busy.value) return
  if (editing.value || confirmDel.value) editing.value = confirmDel.value = null
  else emit('close')
}
onMounted(() => document.addEventListener('keydown', onKey))
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="fixed inset-0 bg-black/40 grid place-items-center z-40" @mousedown.self="!busy && emit('close')">
    <div role="dialog" aria-modal="true" aria-labelledby="manage-shelf-title" class="w-[520px] rounded-xl border border-border bg-background shadow-2xl">
      <div class="flex items-start justify-between p-5 pb-0">
        <div>
          <h2 id="manage-shelf-title" class="font-semibold">Quản lý danh mục và bộ sách</h2>
          <p class="text-xs text-muted-foreground mt-0.5">Đổi tên áp cho mọi cuốn. Xoá thì sách vẫn giữ nguyên.</p>
        </div>
        <button aria-label="Đóng" class="text-muted-foreground hover:text-foreground" :disabled="busy" @click="emit('close')"><X class="w-4 h-4" /></button>
      </div>
      <div role="tablist" class="px-5 mt-4 flex gap-1 border-b border-border">
        <button v-for="t in ([['cat', `Danh mục · ${categories.length}`], ['series', `Bộ sách · ${series.length}`]] as [Tab, string][])" :key="t[0]"
          role="tab" :aria-selected="tab === t[0]" class="h-9 px-3 text-sm -mb-px border-b-2"
          :class="tab === t[0] ? 'border-primary text-foreground font-medium' : 'border-transparent text-muted-foreground hover:text-foreground'"
          @click="setTab(t[0])">{{ t[1] }}</button>
      </div>

      <div class="p-3 max-h-[360px] overflow-auto">
        <p v-if="!(tab === 'cat' ? categories : series).length" class="px-3 py-6 text-center text-sm text-muted-foreground">
          {{ tab === 'cat' ? 'Chưa có danh mục nào. Đặt danh mục ở "Sửa thông tin" của từng cuốn.' : 'Chưa có bộ sách nào. Đặt bộ sách và số tập ở "Sửa thông tin" của từng cuốn.' }}
        </p>
        <div v-for="[name, n] in (tab === 'cat' ? categories : series)" :key="name">
          <!-- đang đổi tên -->
          <form v-if="editing === name" class="flex items-center gap-2 px-2 py-1.5 bg-muted/40 rounded-md" @submit.prevent="saveRename(name)">
            <component :is="tab === 'cat' ? Tag : Layers" class="w-4 h-4 text-muted-foreground shrink-0" />
            <input ref="input" v-model="draft" :maxlength="tab === 'cat' ? MAX_CATEGORY_LEN : MAX_SERIES_LEN" :aria-label="`Tên mới cho ${name}`"
              class="flex-1 min-w-0 h-8 rounded-md border border-ring ring-2 ring-ring/20 bg-background px-2 text-sm" />
            <Button type="submit" size="sm" :disabled="busy || !draft.trim()"><Loader2 v-if="busy" class="w-4 h-4 animate-spin" /> Lưu</Button>
            <Button type="button" size="sm" variant="ghost" :disabled="busy" @click="editing = null">Huỷ</Button>
          </form>
          <!-- xác nhận xoá -->
          <div v-else-if="confirmDel === name" class="px-3 py-2.5 rounded-md bg-destructive/5 border border-destructive/20 text-sm">
            <p>Xoá {{ tab === 'cat' ? 'danh mục' : 'bộ sách' }} "<b>{{ name }}</b>"?</p>
            <p class="text-xs text-muted-foreground mt-0.5">{{ n }} cuốn {{ tab === 'cat' ? 'sẽ về "Chưa phân loại"' : 'sẽ tách ra thành sách lẻ trên kệ' }}. Sách không bị xoá.</p>
            <div class="mt-2 flex justify-end gap-2">
              <Button size="sm" variant="outline" :disabled="busy" @click="confirmDel = null">Huỷ</Button>
              <Button size="sm" variant="destructive" :disabled="busy" @click="remove(name)"><Loader2 v-if="busy" class="w-4 h-4 animate-spin" /> Xoá {{ tab === 'cat' ? 'danh mục' : 'bộ sách' }}</Button>
            </div>
          </div>
          <div v-else class="flex items-center gap-3 px-3 h-11 rounded-md hover:bg-muted/50 text-sm">
            <component :is="tab === 'cat' ? Tag : Layers" class="w-4 h-4 text-muted-foreground shrink-0" />
            <span class="flex-1 truncate" :title="name">{{ name }}</span>
            <span class="text-xs text-muted-foreground w-16 text-right shrink-0">{{ n }} {{ tab === 'cat' ? 'cuốn' : 'tập' }}</span>
            <button class="h-7 w-7 grid place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground" :aria-label="`Đổi tên ${name}`" title="Đổi tên" :disabled="busy" @click="startEdit(name)"><Pencil class="w-4 h-4" /></button>
            <button class="h-7 w-7 grid place-items-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive" :aria-label="`Xoá ${name}`" title="Xoá" :disabled="busy" @click="confirmDel = name; editing = null"><Trash2 class="w-4 h-4" /></button>
          </div>
        </div>
        <p v-if="error" class="px-3 pt-2 text-xs text-destructive">{{ error }}</p>
      </div>

      <div class="px-5 py-3 border-t border-border flex items-center justify-between gap-3 text-xs text-muted-foreground">
        <span>{{ tab === 'cat' ? 'Đổi tên trùng một danh mục đã có thì hai danh mục gộp làm một.' : 'Thứ tự các tập đổi ở ô "Tập số" của từng cuốn.' }}</span>
        <Button size="sm" variant="outline" :disabled="busy" @click="emit('close')">Xong</Button>
      </div>
    </div>
  </div>
</template>
