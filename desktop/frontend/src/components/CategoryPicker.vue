<script setup lang="ts">
// Ô chọn danh mục (bước 1 tạo sách + hộp Sửa thông tin sách), bám wireframe D2:
// bấm là mở danh sách — "Không phân loại", các danh mục đã dùng kèm số cuốn,
// cuối danh sách luôn có "Tạo danh mục mới" (bấm → ô nhập + nút Thêm, Enter cũng thêm).
// kind="series": cùng ô đó cho bộ sách (wireframe D5) — "Không thuộc bộ nào", "Tạo bộ sách mới".
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
import { Check, ChevronDown, Layers, Plus, Tag } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { MAX_CATEGORY_LEN, MAX_SERIES_LEN, mergeCategory, seriesKey } from '../lib/find'

const props = defineProps<{ modelValue: string; categories: [string, number][]; kind?: 'category' | 'series' }>()
const series = computed(() => props.kind === 'series')
const t = computed(() =>
  series.value
    ? { none: 'Không thuộc bộ nào', empty: 'Chưa có bộ sách nào. Tạo bộ để gom các tập của cùng một bộ.', create: 'Tạo bộ sách mới', ph: 'Tên bộ sách, vd: Kinh doanh cho người mới', unit: 'tập', label: 'Bộ sách', max: MAX_SERIES_LEN }
    : { none: 'Không phân loại', empty: 'Chưa có danh mục nào. Tạo danh mục đầu tiên để dễ lọc sách sau này.', create: 'Tạo danh mục mới', ph: 'Tên danh mục, vd: Kỹ năng', unit: 'cuốn', label: 'Danh mục', max: MAX_CATEGORY_LEN },
)
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const open = ref(false)
const creating = ref(false)
const draft = ref('')
const root = ref<HTMLElement | null>(null)
const input = ref<HTMLInputElement | null>(null)

function toggle() {
  open.value = !open.value
  creating.value = false
  draft.value = ''
}

function pick(c: string) {
  emit('update:modelValue', c)
  open.value = false
  creating.value = false
}

async function startCreate() {
  creating.value = true
  await nextTick()
  input.value?.focus()
}

function add() {
  const names = props.categories.map(([k]) => k)
  const c = series.value
    ? (names.find((n) => seriesKey(n) === seriesKey(draft.value)) ?? [...draft.value.trim().split(/\s+/).join(' ')].slice(0, MAX_SERIES_LEN).join('').trim())
    : mergeCategory(draft.value, names)
  if (c) pick(c)
  draft.value = ''
}

// Bấm ra ngoài hoặc Esc thì đóng danh sách (Esc không đóng luôn hộp thoại bên ngoài).
function onDocDown(e: MouseEvent) {
  if (open.value && root.value && !root.value.contains(e.target as Node)) open.value = false
}
function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && open.value) {
    e.stopPropagation()
    open.value = false
  }
}
document.addEventListener('mousedown', onDocDown)
onBeforeUnmount(() => document.removeEventListener('mousedown', onDocDown))
</script>

<template>
  <div ref="root" class="relative" @keydown="onKey">
    <button type="button" class="w-full flex items-center h-9 rounded-md border bg-background px-3 text-left"
      :class="open ? 'border-ring ring-2 ring-ring/20' : 'border-input'" aria-haspopup="listbox" :aria-expanded="open" @click="toggle">
      <component :is="series ? Layers : Tag" class="w-4 h-4 text-muted-foreground mr-2 shrink-0" />
      <span class="flex-1 truncate" :class="!modelValue && 'text-muted-foreground'">{{ modelValue || (series ? t.none : 'Chọn danh mục') }}</span>
      <ChevronDown class="w-4 h-4 text-muted-foreground shrink-0" />
    </button>
    <div v-if="open" role="listbox" :aria-label="t.label" class="absolute left-0 right-0 top-full mt-1 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-30 max-h-72 overflow-auto">
      <button type="button" role="option" :aria-selected="!modelValue" class="w-full flex items-center justify-between px-3 py-1.5 text-left hover:bg-muted" @click="pick('')">
        <span class="text-muted-foreground">{{ t.none }}</span><Check v-if="!modelValue" class="w-4 h-4 text-primary" />
      </button>
      <template v-if="categories.length">
        <button v-for="[c, n] in categories" :key="c" type="button" role="option" :aria-selected="modelValue === c"
          class="w-full flex items-center justify-between gap-2 px-3 py-1.5 text-left hover:bg-muted" @click="pick(c)">
          <span class="truncate" :class="modelValue === c && 'text-primary font-medium'">{{ c }}</span>
          <span class="flex items-center gap-2 text-xs text-muted-foreground shrink-0">{{ n }} {{ t.unit }} <Check v-if="modelValue === c" class="w-4 h-4 text-primary" /></span>
        </button>
      </template>
      <p v-else class="px-3 py-1.5 text-xs text-muted-foreground">{{ t.empty }}</p>
      <div class="border-t border-border mt-1 pt-1">
        <div v-if="creating" class="flex items-center gap-2 px-2 py-1">
          <input ref="input" v-model="draft" :maxlength="t.max" class="flex-1 min-w-0 h-8 rounded-md border border-input bg-background px-2"
            :placeholder="t.ph" :aria-label="`${t.label} mới`" @keydown.enter.prevent="add" />
          <Button type="button" size="sm" :disabled="!draft.trim()" @click="add">Thêm</Button>
        </div>
        <button v-else type="button" class="w-full flex items-center gap-2 px-3 py-1.5 text-left text-primary hover:bg-muted" @click="startCreate">
          <Plus class="w-4 h-4" /> {{ t.create }}
        </button>
      </div>
    </div>
  </div>
</template>
