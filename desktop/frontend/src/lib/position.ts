// Nhớ vị trí nghe từng cuốn (theo máy, trong bộ nhớ của cửa sổ app).
export interface Position {
  track: number
  time: number
  pct: number // % cả cuốn đã nghe, để hiện ở Thư viện
  at?: number // lúc nghe gần nhất (ms), để sắp "Nghe gần đây" / hàng "Nghe tiếp"
}

const key = (slug: string) => `sano:pos:${slug}`

export function loadPosition(slug: string): Position | null {
  try {
    const raw = localStorage.getItem(key(slug))
    if (!raw) return null
    const p = JSON.parse(raw) as Partial<Position>
    return { track: Number(p.track) || 0, time: Number(p.time) || 0, pct: Number(p.pct) || 0, at: Number(p.at) || 0 }
  } catch {
    return null
  }
}

export function savePosition(slug: string, p: Position) {
  try {
    localStorage.setItem(key(slug), JSON.stringify({ ...p, at: p.at || Date.now() }))
  } catch {
    // bộ nhớ trình duyệt bị chặn: bỏ qua, chỉ mất tính năng nhớ vị trí
  }
}

/** "1 giờ 12 phút" / "58 phút" / "40 giây". */
export function fmtLong(sec: number) {
  if (sec < 60) return `${Math.max(0, Math.round(sec))} giây`
  const m = Math.round(sec / 60)
  if (m < 60) return `${m} phút`
  return `${Math.floor(m / 60)} giờ ${String(m % 60).padStart(2, '0')} phút`
}

/** "3:05" / "1:02:07". */
export function fmtClock(sec: number) {
  const s = Math.max(0, Math.floor(sec))
  const h = Math.floor(s / 3600)
  const mm = Math.floor((s % 3600) / 60)
  const ss = String(s % 60).padStart(2, '0')
  return h ? `${h}:${String(mm).padStart(2, '0')}:${ss}` : `${mm}:${ss}`
}
