import { createApp, type Component } from 'vue'
import App from './App.vue'
import { openURL } from './lib/backend'
import './style.css'

// Sáng/tối theo hệ điều hành
const media = window.matchMedia('(prefers-color-scheme: dark)')
const applyTheme = () => document.documentElement.classList.toggle('dark', media.matches)
applyTheme()
media.addEventListener('change', applyTheme)

// Link ra ngoài (target=_blank) mở bằng trình duyệt mặc định, không mở trong cửa sổ app
document.addEventListener('click', (e) => {
  const a = (e.target as HTMLElement | null)?.closest('a[target="_blank"]') as HTMLAnchorElement | null
  if (!a || !a.href || a.getAttribute('href') === '#') return
  e.preventDefault()
  openURL(a.href)
})

// Bản dev: ?wireframe=desktop|library|terms mở wireframe đã duyệt thay cho app
// (đối chiếu giao diện); ?wireframe=landing|docs|demo là wireframe trang tài liệu (VitePress).
// Cả nhánh này bị loại khỏi bản build production.
async function mount() {
  if (import.meta.env.DEV) {
    const wireframes: Record<string, () => Promise<{ default: Component }>> = {
      desktop: () => import('./wireframes/WfDesktop.vue'),
      library: () => import('./wireframes/WfDesktopLibrary.vue'),
      terms: () => import('./wireframes/WfDesktopTerms.vue'),
      landing: () => import('./wireframes/WfLanding.vue'),
      docs: () => import('./wireframes/WfDocsPage.vue'),
      demo: () => import('./wireframes/WfDemo.vue'),
    }
    const load = wireframes[new URLSearchParams(window.location.search).get('wireframe') ?? '']
    if (load) {
      createApp((await load()).default).mount('#app')
      return
    }
  }
  createApp(App).mount('#app')
}
mount()
