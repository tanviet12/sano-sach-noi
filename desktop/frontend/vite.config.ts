import { writeFileSync } from 'node:fs'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'

const r = (p: string) => fileURLToPath(new URL(p, import.meta.url))

// Vite xoá sạch dist/ mỗi lần build → ghi lại file giữ chỗ để `go build` của
// desktop/ (go:embed all:frontend/dist) luôn có ít nhất một file và git không báo xoá.
function keepDistPlaceholder(): Plugin {
  return {
    name: 'sano-keep-dist-placeholder',
    apply: 'build',
    closeBundle() {
      writeFileSync(r('./dist/gitkeep'), '')
    },
  }
}

// Chỉ khi chạy dev: CSP trong index.html cho thêm websocket/HTTP localhost (HMR của
// Vite, IPC websocket của `wails dev` khi mở bằng trình duyệt). Bản build giữ nguyên.
function devCSP(): Plugin {
  return {
    name: 'sano-dev-csp',
    apply: 'serve',
    transformIndexHtml(html) {
      return html.replace("connect-src 'self'", "connect-src 'self' ws://localhost:* http://localhost:*")
    },
  }
}

export default defineConfig({
  plugins: [vue(), keepDistPlaceholder(), devCSP()],
  resolve: {
    alias: { '@': r('./src') },
  },
  server: {
    port: 5390, // port dev riêng của phần mềm
    strictPort: true,
    // lib/terms.ts + lib/prompt.ts đọc điều khoản và lời nhắc mẫu từ docs/ ở gốc repo (?raw)
    fs: { allow: [r('../..')] },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
