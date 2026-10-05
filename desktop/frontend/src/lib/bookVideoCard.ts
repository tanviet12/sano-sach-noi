// Vẽ các cảnh của video cả cuốn (wireframe D15) bằng canvas: trích đoạn, màn tựa,
// thẻ chương, trong sách (câu đang đọc chữ to, câu kế nhạt, thanh tiến độ cả cuốn có
// vạch chương), màn kết, thumbnail. Toạ độ theo khung gốc 640×360 (dọc 360×640), nhân
// 3 ra 1920×1080; thumbnail luôn ngang 1280×720. Chữ chính tránh ~15% đáy (YouTube phủ
// thanh điều khiển, phụ đề). Khung dọc đăng Reels / TikTok / Shorts: app phủ ~14% đầu (giờ,
// thanh trên), ~28% đáy (tên người đăng, chú thích, ô bình luận) và cột nút thích / bình
// luận / chia sẻ bên phải, nên mọi chữ nằm trong TALL_SAFE.
import { cardLogo, clampLines, drawBackground, drawBars, drawBook, roundRect, SANS, wrap } from './shareCard'

export type BVRatio = 'wide' | 'tall'

export interface SceneCommon {
  ratio: BVRatio
  bg: number
  cover: HTMLImageElement | null
  title: string
  voice: string
  totalLabel: string // "27:00" — tổng thời lượng phần sách
}

export type Scene =
  | { kind: 'intro'; line: string; bars: number[]; reserve: number } // reserve: số dòng dành chỗ (câu dài nhất) để sóng âm đứng yên
  | { kind: 'title'; meta: string; badge: string; credits?: string } // badge: SÁCH NÓI ĐẦY ĐỦ / NGHE THỬ SÁCH NÓI; credits: tác giả, dịch giả, NXB
  | { kind: 'chapter'; chapter: string; ticks: number[] }
  | { kind: 'main'; chapter: string; section: string; line: string; next: string; ticks: number[] }
  | { kind: 'end' }
  | { kind: 'thumb'; meta: string; badge: string }

export interface Rect {
  x: number
  y: number
  w: number
  h: number
}
export interface SceneResult {
  wave: Rect | null // khung sóng âm (px ảnh) — trích đoạn
  bar: Rect | null // thanh tiến độ cả cuốn (px ảnh)
}

const K = 3
/** Vùng an toàn khung dọc (toạ độ gốc 360×640). */
const TALL_SAFE = { top: 90, bottom: 640 * 0.72, left: 28, right: 64 }
/** Bề ngang chữ trích đoạn — introLines() đo cùng bề ngang này. */
const introWidth = (ratio: BVRatio) => (ratio === 'wide' ? 540 : 360 - TALL_SAFE.left - TALL_SAFE.right)

// Nền (bìa làm mờ tốn thời gian) vẽ một lần cho mọi khung cùng cỡ / nền / bìa.
let bgCache: { key: string; canvas: HTMLCanvasElement } | null = null
function background(o: SceneCommon, W: number, H: number) {
  const key = `${W}x${H}|${o.bg}|${o.cover?.src ?? ''}|${o.title}`
  if (bgCache?.key !== key) {
    const c = document.createElement('canvas')
    c.width = W
    c.height = H
    drawBackground(c.getContext('2d')!, { bg: o.bg, cover: o.cover, title: o.title }, W, H)
    bgCache = { key, canvas: c }
  }
  return bgCache.canvas
}
export function sceneSize(ratio: BVRatio, kind: Scene['kind']) {
  if (kind === 'thumb') return { w: 1280, h: 720, bw: 640, bh: 360 }
  return ratio === 'wide' ? { w: 640 * K, h: 360 * K, bw: 640, bh: 360 } : { w: 360 * K, h: 640 * K, bw: 360, bh: 640 }
}

type Ctx = CanvasRenderingContext2D & { letterSpacing?: string }

function text(ctx: Ctx, s: string, x: number, y: number, font: string, color: string, spacing = 0) {
  ctx.font = font
  ctx.fillStyle = color
  ctx.letterSpacing = spacing + 'px'
  ctx.fillText(s, x, y)
  ctx.letterSpacing = '0px'
}

/** Khối chữ tự thu cỡ cho vừa (maxLines, cao tối đa maxH); trả chiều cao đã vẽ. */
function block(ctx: Ctx, s: string, x: number, y: number, w: number, size: number, opts: { weight?: number; color?: string; lh?: number; maxLines?: number; maxH?: number; min?: number; family?: string; align?: 'left' | 'center' } = {}) {
  const { weight = 700, color = '#fff', lh = 1.3, maxLines = 5, maxH = 1e9, min = 10, family = SANS(), align = 'left' } = opts
  let lines: string[] = []
  for (; size >= min; size--) {
    ctx.font = `${weight} ${size}px ${family}`
    lines = wrap(ctx, s, w)
    if (lines.length <= maxLines && lines.length * size * lh <= maxH) break
  }
  ctx.font = `${weight} ${size}px ${family}`
  lines = clampLines(ctx, lines, maxLines, w)
  ctx.fillStyle = color
  ctx.textAlign = align
  lines.forEach((l, i) => ctx.fillText(l, align === 'center' ? x + w / 2 : x, y + size * 1.02 + i * size * lh))
  ctx.textAlign = 'left'
  return lines.length * size * lh
}

function brand(ctx: Ctx, bw: number, top: number, right: number) {
  const logo = cardLogo()
  const sans = SANS()
  ctx.font = `700 12px ${sans}`
  const w1 = ctx.measureText('Sano · ').width
  ctx.font = `400 12px ${sans}`
  const w2 = ctx.measureText('Tự tạo sách nói').width
  const tw = Math.max(w1 + w2, 70)
  const x = bw - right - tw
  if (logo) {
    ctx.save()
    ctx.shadowColor = 'rgba(0,0,0,0.3)'
    ctx.shadowBlur = 6
    ctx.drawImage(logo, x - 30, top, 25, 25)
    ctx.restore()
  }
  text(ctx, 'Sano · ', x, top + 11, `700 12px ${sans}`, '#fff')
  text(ctx, 'Tự tạo sách nói', x + w1, top + 11, `400 12px ${sans}`, 'rgba(255,255,255,0.8)')
  text(ctx, 'sanobook.com', x, top + 24, `600 10.5px ${sans}`, 'rgba(255,255,255,0.88)', 0.2)
}

/** Thanh tiến độ mờ + vạch chương (phần sáng do ffmpeg phủ dần / bản xem trước tô `progress`). */
function progressBar(ctx: Ctx, x: number, y: number, w: number, ticks: number[], totalLabel: string, progress?: number) {
  ctx.fillStyle = 'rgba(255,255,255,0.25)'
  roundRect(ctx, x, y, w, 4, 2)
  ctx.fill()
  if (progress) {
    ctx.fillStyle = '#fff'
    roundRect(ctx, x, y, Math.max(4, w * progress), 4, 2)
    ctx.fill()
  }
  ctx.fillStyle = 'rgba(255,255,255,0.6)'
  for (const t of ticks) if (t > 0.002 && t < 0.998) ctx.fillRect(x + w * t - 0.5, y - 2.5, 1, 9)
  ctx.textAlign = 'right'
  text(ctx, totalLabel, x + w, y + 17, `500 10px ${SANS()}`, 'rgba(255,255,255,0.7)')
  ctx.textAlign = 'left'
}

/** Một dòng, dài quá bề ngang thì cắt kèm "…" (đo theo font / giãn chữ đang đặt trên ctx). */
function oneLine(ctx: Ctx, s: string, w: number) {
  return clampLines(ctx, wrap(ctx, s, w), 1, w)[0] ?? ''
}

/** Thanh tiến độ cả cuốn: ngang sát đáy; dọc nằm trên vùng chú thích của app, chừa cột nút bên phải. */
function barRect(wide: boolean, bw: number, bh: number): Rect {
  if (wide) return { x: 30, y: bh - 30, w: bw - 60, h: 4 }
  return { x: TALL_SAFE.left, y: TALL_SAFE.bottom - 22, w: bw - TALL_SAFE.left - TALL_SAFE.right, h: 4 }
}

function coverShadow(ctx: Ctx, o: SceneCommon, x: number, y: number, w: number, h: number, r = 8) {
  ctx.save()
  ctx.shadowColor = 'rgba(0,0,0,0.5)'
  ctx.shadowBlur = 26
  ctx.shadowOffsetY = 10
  ctx.fillStyle = 'rgba(0,0,0,0.35)'
  roundRect(ctx, x, y, w, h, r)
  ctx.fill()
  ctx.restore()
  drawBook(ctx, o, x, y, w, h, r)
}

/** Vẽ một cảnh; bản xem trước tô phần sáng: `progress` (sóng âm trích đoạn), `bar` (tiến độ cả cuốn), 0–1. */
export function drawScene(canvas: HTMLCanvasElement, o: SceneCommon, s: Scene, preview: { progress?: number; bar?: number } = {}): SceneResult {
  const { w: W, h: H, bw, bh } = sceneSize(o.ratio, s.kind)
  canvas.width = W
  canvas.height = H
  const ctx = canvas.getContext('2d')! as Ctx
  const k = W / bw
  const sans = SANS()
  const wide = s.kind === 'thumb' || o.ratio === 'wide'
  ctx.save()
  ctx.drawImage(background(o, W, H), 0, 0)
  ctx.scale(k, k)
  const res: SceneResult = { wave: null, bar: null }
  const px = (r: Rect): Rect => ({ x: Math.round(r.x * k), y: Math.round(r.y * k), w: Math.round(r.w * k), h: Math.round(r.h * k) })
  const safeBottom = wide ? bh * 0.85 : TALL_SAFE.bottom // YouTube phủ ~15% đáy; Reels / TikTok ~28%
  const S = TALL_SAFE
  const top = wide ? 44 : S.top + 40 // mép trên phần chữ, dưới logo

  if (s.kind !== 'end' && s.kind !== 'thumb') brand(ctx, bw, wide ? 16 : S.top, wide ? 18 : S.left)

  if (s.kind === 'intro') {
    const x = wide ? 52 : S.left
    const w = introWidth(o.ratio)
    const lineH = Math.min(wide ? 150 : 260, safeBottom - (wide ? 140 : top + 110))
    const size = wide ? 28 : 26
    // đo trước để canh giữa theo chiều dọc
    ctx.font = `700 ${size}px ${sans}`
    const est = Math.min(lineH, Math.min(5, Math.max(s.reserve, wrap(ctx, s.line, w).length)) * size * 1.3)
    const groupH = 22 + 14 + est + 16 + 26 + 16 + 14
    let y = wide ? Math.max(44, (safeBottom - groupH) / 2 + 10) : Math.max(top, (top + safeBottom - groupH) / 2)
    const tag = '✦  TRÍCH ĐOẠN'
    ctx.font = `700 10px ${sans}`
    ctx.letterSpacing = '1.6px'
    const tagW = ctx.measureText(tag).width + 20
    ctx.letterSpacing = '0px'
    ctx.fillStyle = 'rgba(255,255,255,0.2)'
    roundRect(ctx, x, y, tagW, 20, 10)
    ctx.fill()
    text(ctx, tag, x + 10, y + 14, `700 10px ${sans}`, '#fff', 1.6)
    y += 22 + 14
    block(ctx, s.line, x, y, w, size, { maxLines: 5, maxH: est, min: 16 })
    y += est + 16
    const wr = { x, y, w: Math.min(260, w), h: 26 }
    drawBars(ctx, wr.x, wr.y, wr.w, wr.h, s.bars, 'rgba(255,255,255,0.35)')
    if (preview.progress) {
      ctx.save()
      ctx.beginPath()
      ctx.rect(wr.x, wr.y, wr.w * preview.progress, wr.h)
      ctx.clip()
      drawBars(ctx, wr.x, wr.y, wr.w, wr.h, s.bars, 'rgba(255,255,255,0.95)')
      ctx.restore()
    }
    res.wave = px(wr)
    y += 26 + 16
    block(ctx, `${o.title} · Giọng ${o.voice}`, x, y, w, 12, { weight: 400, color: 'rgba(255,255,255,0.78)', maxLines: 2 })
  } else if (s.kind === 'title' || s.kind === 'thumb') {
    const thumb = s.kind === 'thumb'
    if (wide) {
      const cw = thumb ? 192 : 154
      const ch = (cw * 4) / 3
      const cx = thumb ? 44 : 62
      const cy = (thumb ? bh : safeBottom) / 2 - ch / 2 + (thumb ? 0 : 6)
      if (thumb) {
        ctx.save()
        ctx.translate(cx + cw / 2, cy + ch / 2)
        ctx.rotate(-0.035)
        coverShadow(ctx, o, -cw / 2, -ch / 2, cw, ch)
        ctx.restore()
      } else coverShadow(ctx, o, cx, cy, cw, ch)
      const tx = cx + cw + (thumb ? 34 : 30)
      const tw = bw - tx - 30
      let y = cy + (thumb ? 30 : 34)
      if (thumb) {
        ctx.font = `900 15px ${sans}`
        const bwid = ctx.measureText(s.badge).width + 16
        ctx.fillStyle = '#fff'
        roundRect(ctx, tx, y, bwid, 24, 4)
        ctx.fill()
        text(ctx, s.badge, tx + 8, y + 17.5, `900 15px ${sans}`, '#be123c', 0.4)
        y += 36
        y += block(ctx, o.title, tx, y, tw, 40, { weight: 900, lh: 1.08, maxLines: 4, maxH: 190, min: 22 }) + 12
        block(ctx, s.meta, tx, y, tw, 15, { weight: 600, color: 'rgba(255,255,255,0.92)', maxLines: 2 })
        brand(ctx, bw, bh - 44, 20)
      } else {
        text(ctx, s.badge, tx, y + 10, `700 11px ${sans}`, 'rgba(255,255,255,0.72)', 2.4)
        y += 20
        y += block(ctx, o.title, tx, y, tw, 34, { family: 'Georgia, serif', lh: 1.15, maxLines: 4, maxH: 170, min: 20 }) + 12
        if (s.kind === 'title' && s.credits) y += block(ctx, s.credits, tx, y, tw, 13, { weight: 500, color: 'rgba(255,255,255,0.92)', maxLines: 2 }) + 4
        block(ctx, s.meta, tx, y, tw, 13, { weight: 400, color: 'rgba(255,255,255,0.78)', maxLines: 2 })
      }
    } else {
      // chữ canh giữa: chừa đều hai bên bằng lề phải (cột nút của app)
      const cw = 126
      const ch = (cw * 4) / 3
      const cy = top - 4
      const tx = S.right - 16
      const tw = bw - tx * 2
      coverShadow(ctx, o, (bw - cw) / 2, cy, cw, ch)
      let y = cy + ch + 28
      ctx.textAlign = 'center'
      text(ctx, s.badge, bw / 2, y, `700 11px ${sans}`, 'rgba(255,255,255,0.72)', 2.4)
      ctx.textAlign = 'left'
      y += 12
      y += block(ctx, o.title, tx, y, tw, 26, { family: 'Georgia, serif', lh: 1.15, maxLines: 3, maxH: 92, min: 16, align: 'center' }) + 8
      if (s.kind === 'title' && s.credits) y += block(ctx, s.credits, tx, y, tw, 12, { weight: 500, color: 'rgba(255,255,255,0.92)', maxLines: 2, maxH: Math.max(14, safeBottom - y - 36), min: 10, align: 'center' }) + 4
      block(ctx, s.meta, tx, y, tw, 12, { weight: 400, color: 'rgba(255,255,255,0.78)', maxLines: 2, maxH: Math.max(14, safeBottom - y), min: 10, align: 'center' })
    }
  } else if (s.kind === 'chapter') {
    const x = wide ? 70 : S.right - 16
    const w = bw - x * 2
    const mid = wide ? safeBottom / 2 : (top + safeBottom) / 2 - 10
    ctx.textAlign = 'center'
    ctx.font = `700 11px ${sans}`
    ctx.letterSpacing = '1.8px'
    text(ctx, oneLine(ctx, o.title.toUpperCase().slice(0, 60), w), bw / 2, mid - 40, `700 11px ${sans}`, 'rgba(255,255,255,0.7)', 1.8)
    ctx.textAlign = 'left'
    block(ctx, s.chapter, x, mid - 24, w, wide ? 34 : 30, { maxLines: 3, maxH: 130, min: 18, align: 'center', lh: 1.2 })
    const bar = barRect(wide, bw, bh)
    progressBar(ctx, bar.x, bar.y, bar.w, s.ticks, o.totalLabel, preview.bar)
    res.bar = px(bar)
  } else if (s.kind === 'main') {
    if (wide) {
      const cw = 136
      const ch = (cw * 4) / 3
      const top = 62
      const bottom = safeBottom - 6
      const cy = top + (bottom - top - (ch + 44)) / 2
      coverShadow(ctx, o, 52, cy, cw, ch)
      let fy = cy + ch + 10
      fy += block(ctx, o.title, 52, fy, cw, 12.5, { weight: 600, maxLines: 2, lh: 1.2 })
      block(ctx, 'Giọng ' + o.voice, 52, fy + 1, cw, 10.5, { weight: 400, color: 'rgba(255,255,255,0.72)', maxLines: 1 })
      const tx = 52 + cw + 36
      const tw = bw - tx - 48
      let y = top + 6
      ctx.font = `700 11px ${sans}`
      text(ctx, oneLine(ctx, s.chapter.toUpperCase(), tw), tx, y + 9, `700 11px ${sans}`, 'rgba(255,255,255,0.72)', 1.4)
      y += 14
      ctx.font = `400 12px ${sans}`
      text(ctx, clampLines(ctx, [s.section], 1, tw)[0], tx, y + 11, `400 12px ${sans}`, 'rgba(255,255,255,0.62)')
      y += 26
      const avail = bottom - y
      const h1 = block(ctx, s.line, tx, y, tw, 27, { maxLines: 5, maxH: avail * 0.68, min: 15, lh: 1.28 })
      if (s.next) block(ctx, s.next, tx, y + h1 + 12, tw, 19, { color: 'rgba(255,255,255,0.35)', maxLines: 2, maxH: avail - h1 - 12, min: 12, lh: 1.28 })
    } else {
      // bìa nhỏ bên trái, tên sách + giọng bên phải; dưới là tên chương, câu đang đọc, câu kế
      const x = S.left
      const tw = bw - S.left - S.right
      const cw = 60
      const ch = (cw * 4) / 3
      const cy = top - 6
      coverShadow(ctx, o, x, cy, cw, ch, 6)
      const ix = x + cw + 14
      const iw = bw - ix - S.right
      const th = block(ctx, o.title, ix, cy + 14, iw, 13, { weight: 600, maxLines: 2, lh: 1.2 })
      block(ctx, 'Giọng ' + o.voice, ix, cy + 14 + th + 3, iw, 11, { weight: 400, color: 'rgba(255,255,255,0.72)', maxLines: 1 })
      let y = cy + ch + 22
      ctx.font = `700 11px ${sans}`
      text(ctx, oneLine(ctx, s.chapter.toUpperCase(), tw), x, y + 9, `700 11px ${sans}`, 'rgba(255,255,255,0.72)', 1.2)
      y += 22
      const bottom = barRect(wide, bw, bh).y - 16
      const h1 = block(ctx, s.line, x, y, tw, 25, { maxLines: 6, maxH: (bottom - y) * 0.72, min: 15, lh: 1.28 })
      if (s.next) block(ctx, s.next, x, y + h1 + 12, tw, 18, { color: 'rgba(255,255,255,0.35)', maxLines: 3, maxH: bottom - y - h1 - 12, min: 12, lh: 1.28 })
    }
    const bar = barRect(wide, bw, bh)
    progressBar(ctx, bar.x, bar.y, bar.w, s.ticks, o.totalLabel, preview.bar)
    res.bar = px(bar)
  } else {
    // Màn kết
    const logo = cardLogo()
    const cy = wide ? safeBottom / 2 - 10 : (S.top + safeBottom) / 2 + 30
    if (logo) {
      ctx.save()
      ctx.shadowColor = 'rgba(0,0,0,0.35)'
      ctx.shadowBlur = 18
      ctx.drawImage(logo, bw / 2 - 36, cy - 92, 72, 72)
      ctx.restore()
    }
    ctx.textAlign = 'center'
    text(ctx, 'Tạo sách nói của bạn', bw / 2, cy + 12, `700 ${wide ? 30 : 26}px ${sans}`, '#fff')
    text(ctx, 'Miễn phí · chạy ngay trên máy · từ Word, EPUB, PDF', bw / 2, cy + 36, `400 ${wide ? 14 : 12.5}px ${sans}`, 'rgba(255,255,255,0.82)')
    ctx.font = `700 18px ${sans}`
    const pw = ctx.measureText('sanobook.com').width + 40
    ctx.fillStyle = '#fff'
    roundRect(ctx, bw / 2 - pw / 2, cy + 56, pw, 38, 19)
    ctx.fill()
    text(ctx, 'sanobook.com', bw / 2, cy + 81, `700 18px ${sans}`, '#171717', 0.3)
    ctx.textAlign = 'left'
  }
  ctx.restore()
  return res
}

/** Thanh tiến độ sáng (nền trong suốt, cỡ đúng khung) để ffmpeg tô dần theo thời gian. */
export function drawBrightBar(canvas: HTMLCanvasElement, r: Rect) {
  canvas.width = r.w
  canvas.height = r.h
  const ctx = canvas.getContext('2d')!
  ctx.fillStyle = '#fff'
  roundRect(ctx, 0, 0, r.w, r.h, r.h / 2)
  ctx.fill()
}

/** Số dòng câu chiếm ở cảnh trích đoạn (để dành chỗ theo câu dài nhất). */
export function introLines(ratio: BVRatio, line: string) {
  const ctx = document.createElement('canvas').getContext('2d')!
  ctx.font = `700 ${ratio === 'wide' ? 28 : 26}px ${SANS()}`
  return Math.min(5, wrap(ctx, line, introWidth(ratio)).length)
}

/** Sóng âm sáng của trích đoạn (nền trong suốt, cỡ đúng khung px) để ffmpeg tô dần. */
export function drawBrightWave(canvas: HTMLCanvasElement, r: Rect, bars: number[]) {
  canvas.width = r.w
  canvas.height = r.h
  const ctx = canvas.getContext('2d')!
  ctx.scale(K, K)
  drawBars(ctx, 0, 0, r.w / K, r.h / K, bars, 'rgba(255,255,255,0.95)')
}
