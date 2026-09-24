// Tìm, sắp xếp, gom danh mục cho Thư viện (thuần, không gọi phần Go).
import type { LibraryBook } from './backend'

/** Độ dài tối đa tên danh mục — khớp library.MaxCategoryLen bên Go. */
export const MAX_CATEGORY_LEN = 40

/** Bỏ dấu tiếng Việt + chữ thường: "Kỹ Năng" → "ky nang". */
export function fold(s: string): string {
  return s
    .normalize('NFD')
    .replace(/\p{M}/gu, '')
    .replace(/đ/g, 'd')
    .replace(/Đ/g, 'D')
    .toLowerCase()
}

/** Khớp tên sách hoặc tác giả, không phân biệt hoa thường và dấu. */
export function matches(b: Pick<LibraryBook, 'title' | 'author'>, query: string): boolean {
  const q = fold(query.trim()).replace(/\s+/g, ' ')
  if (!q) return true
  return fold(b.title).includes(q) || fold(b.author).includes(q)
}

/** Chuẩn hoá tên danh mục: bỏ khoảng trắng thừa, cắt độ dài (khớp Go). */
export function normalizeCategory(s: string): string {
  return [...s.trim().split(/\s+/).join(' ')].slice(0, MAX_CATEGORY_LEN).join('').trim()
}

/** Trùng (không phân biệt hoa thường) một danh mục đã có → dùng cách viết đã có. */
export function mergeCategory(name: string, existing: string[]): string {
  const n = normalizeCategory(name)
  if (!n) return ''
  const hit = existing.find((e) => e.toLocaleLowerCase('vi') === n.toLocaleLowerCase('vi'))
  return hit ?? n
}

/**
 * Danh mục đang dùng kèm số cuốn, nhiều cuốn trước (cùng số thì theo tên). Cách viết
 * khác hoa thường (vd sách cũ tạo bằng dòng lệnh) gộp làm một, lấy cách viết nhiều cuốn nhất.
 */
export function categoryCounts(books: Pick<LibraryBook, 'category'>[]): [string, number][] {
  const groups = new Map<string, Map<string, number>>()
  for (const b of books) {
    const c = normalizeCategory(b.category ?? '')
    if (!c) continue
    const k = c.toLocaleLowerCase('vi')
    const g = groups.get(k) ?? new Map<string, number>()
    g.set(c, (g.get(c) ?? 0) + 1)
    groups.set(k, g)
  }
  const out: [string, number][] = []
  for (const g of groups.values()) {
    const spellings = [...g.entries()].sort((a, b) => b[1] - a[1])
    out.push([spellings[0][0], spellings.reduce((n, [, x]) => n + x, 0)])
  }
  return out.sort((a, b) => b[1] - a[1] || collator.compare(a[0], b[0]))
}

export const collator = new Intl.Collator('vi', { sensitivity: 'base', numeric: true })

export type SortKey = 'newest' | 'recent' | 'title' | 'author' | 'longest'
export const SORTS: { key: SortKey; label: string }[] = [
  { key: 'newest', label: 'Mới tạo nhất' },
  { key: 'recent', label: 'Nghe gần đây' },
  { key: 'title', label: 'Tên A–Z' },
  { key: 'author', label: 'Tác giả A–Z' },
  { key: 'longest', label: 'Dài nhất' },
]

const SORT_KEY = 'sano.librarySort'

/** Kiểu sắp xếp đã chọn lần trước (tiện ích riêng của máy, mất cũng không sao). */
export function loadSort(): SortKey {
  try {
    const v = localStorage.getItem(SORT_KEY) as SortKey | null
    return SORTS.some((s) => s.key === v) ? (v as SortKey) : 'newest'
  } catch {
    return 'newest'
  }
}

export function saveSort(k: SortKey) {
  try {
    localStorage.setItem(SORT_KEY, k)
  } catch {
    // bộ nhớ trình duyệt bị chặn: chỉ mất tính năng nhớ lựa chọn
  }
}

export interface ShelfBook extends LibraryBook {
  progress: number // % đã nghe (0–100)
  listenedAt: number // lúc nghe gần nhất (ms), 0 = chưa nghe / không rõ
}

export function sortBooks(list: ShelfBook[], key: SortKey): ShelfBook[] {
  const out = [...list]
  const newest = (a: ShelfBook, b: ShelfBook) => b.createdAt.localeCompare(a.createdAt)
  switch (key) {
    case 'recent':
      return out.sort((a, b) => b.listenedAt - a.listenedAt || newest(a, b))
    case 'title':
      return out.sort((a, b) => collator.compare(a.title, b.title))
    case 'author':
      // cuốn không có tác giả xếp cuối
      return out.sort((a, b) => (!a.author ? 1 : 0) - (!b.author ? 1 : 0) || collator.compare(a.author, b.author) || collator.compare(a.title, b.title))
    case 'longest':
      return out.sort((a, b) => b.durationSec - a.durationSec)
    default:
      return out.sort(newest)
  }
}

/** Đang nghe dở: đã nghe một phần, chưa xong. */
export function isListening(b: ShelfBook): boolean {
  return b.progress > 0 && b.progress < 99
}

/** "hôm nay" / "hôm qua" / "3 ngày trước" / "tuần trước" / "2 tuần trước" / "tháng trước". */
export function ago(at: number, now = Date.now()): string {
  if (!at) return ''
  const day = (t: number) => {
    const d = new Date(t)
    return Math.floor(new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime() / 86_400_000)
  }
  const days = Math.max(0, Math.round(day(now) - day(at)))
  if (days === 0) return 'hôm nay'
  if (days === 1) return 'hôm qua'
  if (days < 7) return `${days} ngày trước`
  if (days < 14) return 'tuần trước'
  if (days < 30) return `${Math.floor(days / 7)} tuần trước`
  if (days < 60) return 'tháng trước'
  return `${Math.floor(days / 30)} tháng trước`
}
