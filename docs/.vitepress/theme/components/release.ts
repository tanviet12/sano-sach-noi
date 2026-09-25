// Link tải thẳng file cài (không qua trang Releases). Lúc build: theo phiên bản trong file VERSION
// (releases/download/v<phiên bản>/… luôn tồn tại cho bản đó). Vào trình duyệt: hỏi GitHub bản mới
// nhất, có thì đổi link sang file của bản đó (trang tài liệu chưa build lại vẫn tải bản mới).
import { onMounted, ref } from 'vue'
import versionRaw from '../../../../VERSION?raw'

export const REPO = 'https://github.com/tanviet12/sano-sach-noi'
export const RELEASES = REPO + '/releases/latest'
const BUILD_VERSION = versionRaw.trim()

export type Asset = 'mac' | 'win' | 'win-zip' | 'linux' | 'm4b'
const names = (v: string): Record<Asset, string> => ({
  mac: `Sano-${v}-macos-universal.dmg`,
  win: `Sano-${v}-windows-amd64-setup.exe`,
  'win-zip': `Sano-${v}-windows-amd64-portable.zip`,
  linux: `Sano-${v}-linux-amd64.AppImage`,
  m4b: 'Ky-nang-mem-cho-nguoi-tre-sach-mau.m4b',
})

function linksFor(v: string): Record<Asset, { name: string; url: string }> {
  const n = names(v)
  const out = {} as Record<Asset, { name: string; url: string }>
  for (const k of Object.keys(n) as Asset[]) out[k] = { name: n[k], url: `${REPO}/releases/download/v${v}/${n[k]}` }
  return out
}

// Dùng chung giữa các component: chỉ hỏi GitHub một lần mỗi lần mở trang.
const version = ref(BUILD_VERSION)
const links = ref(linksFor(BUILD_VERSION))
let asked = false

async function askLatest() {
  if (asked) return
  asked = true
  try {
    const res = await fetch('https://api.github.com/repos/tanviet12/sano-sach-noi/releases/latest', {
      headers: { Accept: 'application/vnd.github+json' },
    })
    if (!res.ok) return
    const rel = (await res.json()) as { tag_name?: string; assets?: { name: string; browser_download_url: string }[] }
    const v = (rel.tag_name ?? '').replace(/^v/, '')
    if (!/^\d+\.\d+\.\d+$/.test(v) || !rel.assets?.length) return
    const byName = new Map(rel.assets.map((a) => [a.name, a.browser_download_url]))
    const next = linksFor(v)
    // Chỉ nhận link của đúng repo, đúng tên file mong đợi; thiếu file nào thì giữ link bản build.
    for (const k of Object.keys(next) as Asset[]) {
      const u = byName.get(next[k].name)
      if (u && u.startsWith(REPO + '/releases/download/')) next[k].url = u
      else next[k] = links.value[k]
    }
    version.value = v
    links.value = next
  } catch {
    // không có mạng / bị giới hạn lượt gọi: giữ link theo VERSION
  }
}

export function useRelease() {
  onMounted(askLatest)
  return { version, links }
}
