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
export function matches(b: Pick<LibraryBook, 'title' | 'author'> & { voice?: string }, query: string): boolean {
  const q = fold(query.trim()).replace(/\s+/g, ' ')
  if (!q) return true
  return fold(b.title).includes(q) || fold(b.author).includes(q) || fold(b.voice ?? '').includes(q)
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

export type SortKey = 'manual' | 'newest' | 'recent' | 'title' | 'author' | 'longest'
export const SORTS: { key: SortKey; label: string }[] = [
  { key: 'manual', label: 'Tự sắp xếp' },
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

// ── Kệ sách: bộ sách gom thành một thẻ (wireframe D5) ─────────────────────

export const MAX_SERIES_LEN = 80

/** Khoá bộ sách (không phân biệt hoa thường) — khớp library.SeriesKey bên Go. */
export function seriesKey(name: string): string {
  return [...name.trim().split(/\s+/).join(' ')].slice(0, MAX_SERIES_LEN).join('').trim().toLowerCase()
}

export type ShelfItem =
  | { kind: 'book'; key: string; book: ShelfBook }
  | {
      kind: 'series'
      key: string // "s:<tên bộ chữ thường>"
      name: string
      author: string
      vols: ShelfBook[] // theo số tập
      coverUrl: string
      title: string // tập 1 (vẽ bìa mặc định)
      durationSec: number
      progress: number // % đã nghe cả bộ (theo thời lượng)
      listenedAt: number
      createdAt: string // tập mới nhất
      current: ShelfBook // tập đang nghe dở / tập kế tiếp chưa nghe
    }

/** Tập nên nghe tiếp: tập đang dở, không thì tập đầu chưa nghe xong, không thì tập 1. */
export function nextVolume(vols: ShelfBook[]): ShelfBook {
  return vols.find(isListening) ?? vols.find((v) => v.progress < 99) ?? vols[0]
}

/** Gom sách thành thẻ trên kệ: sách lẻ giữ nguyên, các tập cùng bộ thành một thẻ. */
export function shelfItems(books: ShelfBook[]): ShelfItem[] {
  const out: ShelfItem[] = []
  const series = new Map<string, ShelfBook[]>()
  for (const b of books) {
    if (!b.series) {
      out.push({ kind: 'book', key: `b:${b.slug}`, book: b })
      continue
    }
    const k = seriesKey(b.series)
    if (!series.has(k)) series.set(k, [])
    series.get(k)!.push(b)
  }
  for (const [k, vols] of series) {
    vols.sort((a, b) => a.volume - b.volume || collator.compare(a.title, b.title))
    const dur = vols.reduce((n, v) => n + v.durationSec, 0)
    const heard = vols.reduce((n, v) => n + (v.durationSec * v.progress) / 100, 0)
    out.push({
      kind: 'series', key: `s:${k}`, name: vols[0].series, author: vols[0].author, vols,
      coverUrl: vols[0].coverUrl, title: vols[0].title, durationSec: dur,
      progress: dur ? Math.round((heard / dur) * 100) : 0,
      listenedAt: Math.max(...vols.map((v) => v.listenedAt)),
      createdAt: vols.map((v) => v.createdAt).sort().pop() ?? '',
      current: nextVolume(vols),
    })
  }
  return out
}

/** Sắp xếp thẻ trên kệ. "Tự sắp xếp": theo thứ tự đã lưu; thẻ chưa có trong thứ tự (sách mới) lên đầu, mới nhất trước. */
export function sortShelf(items: ShelfItem[], key: SortKey, order: string[] = []): ShelfItem[] {
  const f = (i: ShelfItem) => (i.kind === 'book' ? i.book : i)
  const name = (i: ShelfItem) => (i.kind === 'book' ? i.book.title : i.name)
  const newest = (a: ShelfItem, b: ShelfItem) => f(b).createdAt.localeCompare(f(a).createdAt)
  const out = [...items]
  switch (key) {
    case 'manual': {
      const pos = new Map(order.map((k, i) => [k, i]))
      return out.sort((a, b) => {
        const pa = pos.get(a.key) ?? -1
        const pb = pos.get(b.key) ?? -1
        if (pa < 0 || pb < 0) return pa < 0 && pb < 0 ? newest(a, b) : pa < 0 ? -1 : 1
        return pa - pb
      })
    }
    case 'recent':
      return out.sort((a, b) => f(b).listenedAt - f(a).listenedAt || newest(a, b))
    case 'title':
      return out.sort((a, b) => collator.compare(name(a), name(b)))
    case 'author':
      return out.sort((a, b) => (!f(a).author ? 1 : 0) - (!f(b).author ? 1 : 0) || collator.compare(f(a).author, f(b).author) || collator.compare(name(a), name(b)))
    case 'longest':
      return out.sort((a, b) => f(b).durationSec - f(a).durationSec)
    default:
      return out.sort(newest)
  }
}

/** Đưa thẻ `from` vào chỗ thẻ `to` (kéo thả), trả thứ tự khoá mới của cả kệ. */
export function moveItem(keys: string[], from: string, to: string): string[] {
  const list = keys.filter((k) => k !== from)
  const i = keys.indexOf(from)
  const j = keys.indexOf(to)
  if (i < 0 || j < 0 || i === j) return keys
  list.splice(list.indexOf(to) + (i < j ? 1 : 0), 0, from)
  return list
}
