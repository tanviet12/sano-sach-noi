<script setup lang="ts">
// Thanh điều hướng của trang tài liệu (wireframe) — bám đúng khung navbar mặc định của VitePress:
// logo + tên bên trái, nav bên phải (Trang chủ · Hướng dẫn ▾ · Tải về), nút sáng/tối, icon GitHub
// (socialLinks). Điện thoại: nút ☰ mở menu toàn màn như VitePress.
// Dùng chung cho WfLanding + WfDocsPage. Wireframe tĩnh, không gọi API.
import { ref } from 'vue'
import { ChevronDown, Github, Menu, Moon, Search, Sun, X } from 'lucide-vue-next'

defineProps<{ current: 'home' | 'guide'; dark: boolean }>()
defineEmits<{ 'toggle-dark': [] }>()

type GuideLink = { text: string; slug: string }
const guide: GuideLink[] = [
  { text: 'Cài đặt', slug: 'cai-dat' },
  { text: 'Tạo sách đầu tiên', slug: 'tao-sach-dau-tien' },
  { text: 'Làm mượt tài liệu', slug: 'lam-muot-tai-lieu' },
  { text: 'Nghe trên điện thoại', slug: 'nghe-tren-dien-thoai' },
  { text: 'Câu hỏi thường gặp', slug: 'cau-hoi-thuong-gap' },
  { text: 'Gỡ cài đặt', slug: 'go-cai-dat' },
]
const REPO = 'https://github.com/tanviet12/sano-sach-noi'

const open = ref(false) // menu điện thoại
const dropdown = ref(false) // menu Hướng dẫn trên máy tính
</script>

<template>
  <header class="sticky top-0 z-30 border-b border-border bg-background/90 backdrop-blur">
    <div class="mx-auto flex h-16 max-w-6xl items-center gap-4 px-4 sm:px-6">
      <a href="?wireframe=landing" class="flex items-center gap-2.5 font-semibold focus-ring rounded-md">
        <img src="@/assets/logo.svg" alt="" class="h-8 w-8 rounded-lg" />
        <span class="text-base">Sano</span>
      </a>

      <!-- Ô tìm kiếm (VitePress search provider: local) -->
      <button type="button" class="ml-2 hidden md:flex h-9 items-center gap-2 rounded-md border border-border bg-muted/50 px-3 text-sm text-muted-foreground hover:text-foreground focus-ring">
        <Search class="w-4 h-4" /> Tìm trong tài liệu <kbd class="ml-4 rounded border border-border px-1.5 text-[10px]">⌘K</kbd>
      </button>

      <nav class="ml-auto hidden md:flex items-center gap-1 text-sm" aria-label="Điều hướng chính">
        <a href="?wireframe=landing" class="px-3 py-2 rounded-md hover:text-primary focus-ring" :class="current === 'home' && 'text-primary font-medium'" :aria-current="current === 'home' ? 'page' : undefined">Trang chủ</a>
        <div class="relative" @mouseenter="dropdown = true" @mouseleave="dropdown = false">
          <button type="button" class="flex items-center gap-1 px-3 py-2 rounded-md hover:text-primary focus-ring" :class="current === 'guide' && 'text-primary font-medium'" :aria-expanded="dropdown" @click="dropdown = !dropdown">
            Hướng dẫn <ChevronDown class="w-3.5 h-3.5" />
          </button>
          <div v-if="dropdown" class="absolute right-0 top-full w-56 rounded-lg border border-border bg-popover py-1.5 shadow-lg">
            <a v-for="g in guide" :key="g.slug" href="?wireframe=docs" class="block px-4 py-1.5 text-sm hover:bg-muted hover:text-primary">{{ g.text }}</a>
          </div>
        </div>
        <a href="?wireframe=landing#tai-ve" class="px-3 py-2 rounded-md hover:text-primary focus-ring">Tải về</a>
        <span class="mx-2 h-5 w-px bg-border" />
        <button type="button" class="h-9 w-9 grid place-items-center rounded-md hover:bg-muted focus-ring" :aria-label="dark ? 'Chuyển sang giao diện sáng' : 'Chuyển sang giao diện tối'" @click="$emit('toggle-dark')">
          <component :is="dark ? Sun : Moon" class="w-4 h-4" />
        </button>
        <a :href="REPO" target="_blank" rel="noopener" class="h-9 w-9 grid place-items-center rounded-md hover:bg-muted focus-ring" aria-label="Mã nguồn trên GitHub"><Github class="w-4 h-4" /></a>
      </nav>

      <div class="ml-auto flex items-center gap-1 md:hidden">
        <button type="button" class="h-9 w-9 grid place-items-center rounded-md hover:bg-muted focus-ring" aria-label="Tìm trong tài liệu"><Search class="w-4 h-4" /></button>
        <button type="button" class="h-9 w-9 grid place-items-center rounded-md hover:bg-muted focus-ring" :aria-label="open ? 'Đóng menu' : 'Mở menu'" :aria-expanded="open" @click="open = !open">
          <component :is="open ? X : Menu" class="w-5 h-5" />
        </button>
      </div>
    </div>

    <!-- Menu điện thoại (VitePress: VPNavScreen) -->
    <div v-if="open" class="md:hidden border-t border-border bg-background px-4 pb-6 pt-2 text-sm">
      <a href="?wireframe=landing" class="block border-b border-border py-3" :class="current === 'home' && 'text-primary font-medium'">Trang chủ</a>
      <p class="pt-3 pb-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Hướng dẫn</p>
      <a v-for="g in guide" :key="g.slug" href="?wireframe=docs" class="block py-2 pl-3">{{ g.text }}</a>
      <a href="?wireframe=landing#tai-ve" class="mt-2 block border-t border-border py-3">Tải về</a>
      <div class="flex items-center justify-between rounded-lg bg-muted/60 px-3 py-2">
        <span>Giao diện</span>
        <button type="button" class="flex items-center gap-1.5 rounded-md px-2 py-1 hover:bg-background focus-ring" @click="$emit('toggle-dark')">
          <component :is="dark ? Sun : Moon" class="w-4 h-4" /> {{ dark ? 'Sáng' : 'Tối' }}
        </button>
      </div>
      <a :href="REPO" target="_blank" rel="noopener" class="mt-3 flex items-center justify-center gap-2 py-2 text-muted-foreground"><Github class="w-4 h-4" /> GitHub</a>
    </div>
  </header>
</template>
