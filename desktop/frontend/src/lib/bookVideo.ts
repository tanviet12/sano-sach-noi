// Tạo video cả cuốn (wireframe D15): dựng dòng thời gian (trích đoạn → màn tựa → các
// tiểu mục, thẻ chương, quãng nghỉ như khi nghe → màn kết), tính giờ từng câu (ước lượng
// rồi bám khoảng lặng thật như màn nghe), vẽ từng khung hình gửi sang Go, kèm thumbnail,
// phụ đề .srt, mô tả YouTube có mốc chương. Chạy ở module (không ở hộp thoại) nên đóng
// hộp thoại giữa chừng vẫn tạo tiếp.
import { reactive } from 'vue'
import {
  bookTexts, bookVideoBegin, bookVideoCancel, bookVideoFinish, bookVideoFrame, bookVideoState, errText, onEvent,
  type BookDetail, type BookVideoOverlay, type BookVideoSeg, type BookVideoStatus, type Track,
} from './backend'
import { buildLyrics, decodeSilences, sentenceBounds, snapToSilences, type Lyrics } from './lyrics'
import { gapsFor } from './pause'
import { audioBars, canvasPNG, loadImage, prepareCard } from './shareCard'
import { drawBrightBar, drawBrightWave, drawScene, introLines, sceneSize, type BVRatio, type Rect, type Scene, type SceneCommon } from './bookVideoCard'

export const TITLE_SEC = 3
export const CHAPTER_SEC = 2.5
export const END_SEC = 5
const MAX_DECODE_SEC = 20 * 60

export interface BookVideoOptions {
  from: number // chỉ số tiểu mục đầu (gồm)
  to: number // chỉ số tiểu mục cuối (gồm)
  intro: { track: number; from: number; to: number } | null // câu from..to (gồm) của tiểu mục
  ratio: BVRatio
  bg: number
  extras: { thumb: boolean; srt: boolean; desc: boolean }
  opening?: Opening // D16: lời giới thiệu có giọng đọc, nhạc hiệu, chuông sang chương
}

/** Một file âm thanh thêm đã tạo (BookVideoExtras). */
export interface ExtraClip {
  id: string
  url: string
  dur: number
}
export interface Opening {
  brand?: ExtraClip // "Bạn đang nghe sách nói, tạo bằng Sano."
  info?: ExtraClip // tên sách, tác giả, dịch giả, NXB
  end?: ExtraClip // "Tạo sách nói của bạn tại sanobook.com."
  music?: ExtraClip // nhạc hiệu
  chime?: ExtraClip // chuông sang chương
  skipIntroTrack: boolean // bỏ tiểu mục mở đầu tự có ("Bạn đang nghe sách nói. Cuốn sách: …")
}

/** Tác giả · Dịch giả · Nhà xuất bản (trường trống thì bỏ). */
export function creditsOf(d: BookDetail) {
  return [d.author && `Tác giả ${d.author}`, d.translator && `Dịch giả ${d.translator}`, d.publisher && `NXB ${d.publisher}`].filter(Boolean).join(' · ')
}
/** Tiểu mục đầu là lời mở đầu tự có của Sano ("Cuốn sách: …")? */
export function isIntroTrack(d: BookDetail, texts: { text: string }[], i: number) {
  return i === 0 && /(^|\/)ch01-sec01\.mp3$/.test(d.tracks[0]?.file ?? '') && /^Cuốn sách:/m.test(texts[0]?.text ?? '')
}

type Step = 'idle' | 'prep' | 'frames' | 'audio' | 'video' | 'files' | 'done' | 'error'
export const bv = reactive({
  open: false,
  slug: '',
  step: 'idle' as Step,
  done: 0, // tiểu mục đã dò / khung hình đã vẽ
  total: 0,
  pct: 0,
  error: '',
  dir: '',
  bytes: 0,
  durSec: 0,
  desc: '',
  title: '',
  video: '', // đường dẫn file video vừa tạo
})
export const bvBusy = () => ['prep', 'frames', 'audio', 'video', 'files'].includes(bv.step)

function applyStatus(st: BookVideoStatus) {
  if (!st) return
  bv.slug = st.slug
  bv.title = st.title
  if (st.running) {
    bv.step = st.phase
    bv.pct = st.pct
  } else if (st.done) {
    bv.step = 'done'
    bv.pct = 100
    bv.dir = st.dir
    bv.video = st.video
    bv.bytes = st.bytes
    bv.durSec = st.durSec
  } else if (st.error) {
    bv.step = st.error === 'đã huỷ' ? 'idle' : 'error'
    bv.error = st.error === 'đã huỷ' ? '' : st.error
  }
}
onEvent<BookVideoStatus>('bookvideo:progress', (st) => st.phase !== 'frames' && applyStatus(st))
onEvent<BookVideoStatus>('bookvideo:finished', applyStatus)

export function openBookVideo(slug: string) {
  if (bv.slug !== slug && !bvBusy()) {
    bv.step = 'idle'
    bv.error = ''
  }
  bv.slug = bvBusy() ? bv.slug : slug
  bv.open = true
  void bookVideoState().then((st) => st && st.slug === bv.slug && (st.running || bv.step === 'idle') && st.phase !== 'frames' && applyStatus(st))
}

let cancelled = false
export async function cancelBookVideo() {
  cancelled = true
  await bookVideoCancel()
  if (bv.step === 'prep' || bv.step === 'frames') bv.step = 'idle'
}

// ── Lời từng tiểu mục (giờ câu bám khoảng lặng thật như màn nghe) ──
export type Silences = Map<number, { end: number; len: number }[] | null>
async function trackLyrics(t: Track, text: string, script: string, sils: Silences, i: number): Promise<Lyrics | null> {
  if (!text) return null
  const l = buildLyrics(text, script, t.durationSec)
  if (t.durationSec > MAX_DECODE_SEC) return l
  const sil = await decodeSilences(t.url)
  sils.set(i, sil)
  return sil ? snapToSilences(l, sil) : l
}

/** Chương mới bắt đầu ở tiểu mục i (hiện thẻ chương / mốc mô tả)? */
export function chapterAt(d: BookDetail, i: number, from: number): string {
  const t = d.tracks[i]
  if (!t?.chapter || t.chapter === d.title) return ''
  return t.chapterStart || i === from ? t.chapter : ''
}

export const clock = (sec: number) => {
  const s = Math.max(0, Math.floor(sec))
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const ss = String(s % 60).padStart(2, '0')
  return h ? `${h}:${String(m).padStart(2, '0')}:${ss}` : `${m}:${ss}`
}
function srtTime(sec: number) {
  const ms = Math.max(0, Math.round(sec * 1000))
  const h = Math.floor(ms / 3600000)
  const m = Math.floor((ms % 3600000) / 60000)
  const s = Math.floor((ms % 60000) / 1000)
  const p = (n: number, w = 2) => String(n).padStart(w, '0')
  return `${p(h)}:${p(m)}:${p(s)},${p(ms % 1000, 3)}`
}

export interface Frame {
  scene: Scene
  sec: number
  at: number // giây bắt đầu trong video
}

export interface Timeline {
  segs: BookVideoSeg[]
  frames: Frame[]
  srt: { a: number; b: number; t: string }[]
  marks: { t: number; label: string }[]
  introDur: number
  bookStart: number
  bookEnd: number
  total: number
  totalLabel: string
  minutes: number
  badge: string
  outName: string
  partial: boolean
  part: string
}

/**
 * Trích đoạn: đầu, độ dài và điểm cắt sau từng câu (giây tính từ đầu trích đoạn). Cắt ở giữa
 * khoảng lặng thật giữa hai câu (sentenceBounds) để không dính chữ đầu câu kế.
 */
export function introSpan(d: BookDetail, o: BookVideoOptions, lyr: Map<number, Lyrics | null>, sils?: Silences) {
  if (!o.intro) return null
  const it = d.tracks[o.intro.track]
  const l = lyr.get(o.intro.track)
  if (!l?.sentences.length) return null
  const b = sentenceBounds(l, sils?.get(o.intro.track) ?? null, it.durationSec)
  const from = Math.min(o.intro.from, b.length - 1)
  const to = Math.min(Math.max(o.intro.to, from), b.length - 1)
  const a = b[from].start
  const cuts = b.slice(from, to + 1).map((x) => x.end - a)
  return { track: it, start: a, dur: Math.max(1, cuts[cuts.length - 1]), cuts, from, to }
}

/**
 * Dòng thời gian của video: đoạn tiếng + khung hình (dùng chung cho Tạo video và Nghe thử
 * như video trong hộp thoại, nên nghe thử khớp đúng video sẽ tạo).
 */
export function buildTimeline(d: BookDetail, o: BookVideoOptions, lyr: Map<number, Lyrics | null>, introBars: number[], sils?: Silences): Timeline {
  const op = o.opening
  // Có lời giới thiệu mới thì bỏ tiểu mục mở đầu tự có của sách (không đọc trùng).
  const first = op?.skipIntroTrack && o.from === 0 && o.to > 0 ? 1 : o.from
  const range = d.tracks.slice(first, o.to + 1)
  const segs: BookVideoSeg[] = []
  const frames: Frame[] = []
  const extra = (c: ExtraClip) => segs.push({ file: '', extra: c.id, start: 0, dur: c.dur, silence: false })
  const srt: Timeline['srt'] = []
  const marks: Timeline['marks'] = []
  let t = 0
  let ft = 0 // giờ bắt đầu khung hình kế
  const frame = (scene: Scene, sec: number) => {
    frames.push({ scene, sec, at: ft })
    ft += sec
  }
  const silence = (dur: number) => segs.push({ file: '', start: 0, dur, silence: true })
  let introDur = 0
  const sp = introSpan(d, o, lyr, sils)
  if (o.intro && sp) {
    const ss = lyr.get(o.intro.track)?.sentences ?? []
    introDur = sp.dur
    const pick = ss.slice(sp.from, sp.to + 1)
    const reserve = Math.max(1, ...pick.map((x) => introLines(o.ratio, x.text)))
    segs.push({ file: sp.track.file, start: sp.start, dur: introDur, silence: false })
    pick.forEach((x, k) => {
      const sec = sp.cuts[k] - (k ? sp.cuts[k - 1] : 0)
      frame({ kind: 'intro', line: x.text, bars: introBars, reserve }, sec)
      srt.push({ a: t, b: t + sec, t: x.text })
      t += sec
    })
    t = introDur
    ft = introDur
    marks.push({ t: 0, label: 'Trích đoạn' })
  }
  const bookSec = range.reduce((n, x) => n + x.durationSec, 0)
  const minutes = Math.max(1, Math.round(bookSec / 60))
  const chapters = range.filter((_, k) => chapterAt(d, first + k, first)).length
  const meta = `Giọng ${d.voice} · ${minutes} phút${chapters ? ` · ${chapters} chương` : ''}`
  // Vài chương: ghi "Nghe thử", tên file / thư mục kèm tên chương để không đè video cả cuốn.
  const partial = o.from > 0 || o.to < d.tracks.length - 1
  const badge = partial ? 'NGHE THỬ SÁCH NÓI' : 'SÁCH NÓI ĐẦY ĐỦ'
  const names = range.map((_, k) => chapterAt(d, first + k, first)).filter(Boolean)
  const part = names.length > 1 ? `${names[0]} đến ${names[names.length - 1]}` : names[0] || range[0]?.title || ''
  const outName = partial ? `${d.title} - ${part}` : d.title
  // Màn tựa: giọng đọc thương hiệu → nhạc hiệu → giọng đọc giới thiệu sách (D16, kiểu Fonos);
  // tắt hết thì lặng TITLE_SEC như cũ.
  const titleParts: [ExtraClip | undefined, number][] = [[op?.brand, 0.25], [op?.music, 0.15], [op?.info, 0.6]]
  let titleSec = 0
  for (const [c, pad] of titleParts) {
    if (!c) continue
    extra(c)
    silence(pad)
    titleSec += c.dur + pad
  }
  if (!titleSec) {
    silence(TITLE_SEC)
    titleSec = TITLE_SEC
  }
  frame({ kind: 'title', meta, badge, credits: creditsOf(d) }, titleSec)
  t += titleSec
  const bookStart = t
  // Thẻ chương: chuông ngắn + lặng cho đủ ~1,5 giây; không chuông thì lặng CHAPTER_SEC.
  const chimePad = op?.chime ? Math.max(0.3, 1.5 - op.chime.dur) : 0
  const chapSec = op?.chime ? op.chime.dur + chimePad : CHAPTER_SEC

  // vạch chương: tính trước theo cùng cách cộng thời gian
  const gaps = gapsFor(d.slug)
  const ticks: number[] = []
  const plan: { i: number; chapter: string; gap: number }[] = []
  {
    let x = 0
    range.forEach((tr, k) => {
      const i = first + k
      const ch = chapterAt(d, i, first)
      if (ch) {
        ticks.push(x)
        x += chapSec
      }
      x += tr.durationSec
      const nextCh = k + 1 < range.length && !!chapterAt(d, i + 1, first)
      const gap = k + 1 < range.length ? (nextCh ? gaps.chapter : gaps.section) : 0
      x += gap
      plan.push({ i, chapter: ch, gap })
    })
    for (let n = 0; n < ticks.length; n++) ticks[n] = x ? ticks[n] / x : 0
  }
  let curChapter = ''
  for (const { i, chapter, gap } of plan) {
    const tr = d.tracks[i]
    if (chapter) {
      curChapter = chapter
      frame({ kind: 'chapter', chapter, ticks }, chapSec)
      if (op?.chime) {
        extra(op.chime)
        silence(chimePad)
      } else silence(CHAPTER_SEC)
      marks.push({ t, label: chapter })
      t += chapSec
    } else if (i === first && !marks.length) marks.push({ t, label: tr.title })
    segs.push({ file: tr.file, start: 0, dur: tr.durationSec, silence: false })
    const ss = lyr.get(i)?.sentences ?? []
    const chLabel = curChapter || tr.chapter || d.title
    if (!ss.length) frame({ kind: 'main', chapter: chLabel, section: tr.title, line: tr.title, next: '', ticks }, tr.durationSec + gap)
    ss.forEach((x, k) => {
      const a0 = k === 0 ? 0 : x.start
      const b0 = k + 1 < ss.length ? ss[k + 1].start : tr.durationSec
      frame({ kind: 'main', chapter: chLabel, section: tr.title, line: x.text, next: ss[k + 1]?.text ?? '', ticks }, Math.max(0.05, b0 - a0) + (k === ss.length - 1 ? gap : 0))
      srt.push({ a: t + x.start, b: t + b0, t: x.text })
    })
    t += tr.durationSec
    if (gap) {
      silence(gap)
      t += gap
    }
    ft = t // khung hình bám đúng giờ tiếng (tránh cộng dồn sai số)
  }
  const bookEnd = t
  // Màn kết: giọng đọc lời kết → nhạc hiệu tắt dần; tắt hết thì lặng END_SEC.
  let endSec = 0
  if (op?.end || op?.music) {
    silence(0.5)
    endSec = 0.5
    for (const c of [op?.end, op?.music]) {
      if (!c) continue
      extra(c)
      endSec += c.dur
    }
    silence(0.5)
    endSec += 0.5
  } else {
    silence(END_SEC)
    endSec = END_SEC
  }
  frame({ kind: 'end' }, endSec)
  t += endSec
  return { segs, frames, srt, marks, introDur, bookStart, bookEnd, total: t, totalLabel: clock(bookSec), minutes, badge, outName, partial, part }
}

/** Lời ước lượng (không dò khoảng lặng) — đủ nhanh cho Nghe thử như video. */
export function quickLyrics(d: BookDetail, texts: { text: string; script: string }[]) {
  const m = new Map<number, Lyrics | null>()
  d.tracks.forEach((t, i) => m.set(i, texts[i]?.text ? buildLyrics(texts[i].text, texts[i].script, t.durationSec) : null))
  return m
}

/** Chạy cả lượt: dò lời → vẽ khung hình → Go ghép + mã hoá (tiến độ qua sự kiện). */
export async function startBookVideo(d: BookDetail, o: BookVideoOptions) {
  cancelled = false
  bv.slug = d.slug
  bv.title = d.title
  bv.error = ''
  bv.step = 'prep'
  bv.done = 0
  bv.pct = 0
  const range = d.tracks.slice(o.from, o.to + 1)
  bv.total = range.length + (o.intro ? 1 : 0)
  try {
    const texts = await bookTexts(d.slug)
    await prepareCard()
    const cover = d.coverUrl ? await loadImage(d.coverUrl) : null
    const common: SceneCommon = { ratio: o.ratio, bg: o.bg, cover, title: d.title, voice: d.voice, totalLabel: '' }

    // 1. Lời + giờ câu của từng tiểu mục trong phần chọn.
    const lyr = new Map<number, Lyrics | null>()
    const sils: Silences = new Map()
    const need = [...new Set([...(o.intro ? [o.intro.track] : []), ...range.map((_, k) => o.from + k)])]
    for (const i of need) {
      if (cancelled) return
      lyr.set(i, await trackLyrics(d.tracks[i], texts[i]?.text ?? '', texts[i]?.script ?? '', sils, i))
      bv.done++
    }

    // 2. Dòng thời gian.
    const sp = introSpan(d, o, lyr, sils)
    const introBars = sp ? await audioBars(sp.track.url, sp.start, sp.start + sp.dur, 48).catch(() => Array.from({ length: 48 }, () => 0.35)) : []
    const tl = buildTimeline(d, o, lyr, introBars, sils)
    const { segs, frames, srt, marks, introDur, bookStart, bookEnd, minutes, badge, outName, partial, part } = tl
    common.totalLabel = tl.totalLabel

    // 3. Vẽ + gửi khung hình.
    if (cancelled) return
    const id = await bookVideoBegin(d.slug, d.title)
    bv.step = 'frames'
    bv.done = 0
    bv.total = frames.length
    const c = document.createElement('canvas')
    let wave: Rect | null = null
    let bar: Rect | null = null
    for (let n = 0; n < frames.length; n++) {
      if (cancelled) return
      const r = drawScene(c, common, frames[n].scene)
      wave ??= r.wave
      bar ??= r.bar
      await bookVideoFrame(id, n, c.toDataURL('image/jpeg', 0.9).split(',')[1])
      bv.done = n + 1
      if (n % 8 === 0) await new Promise((res) => setTimeout(res)) // nhường giao diện
    }
    const overlays: BookVideoOverlay[] = []
    if (wave && introDur) {
      drawBrightWave(c, wave, introBars)
      overlays.push({ png: (await canvasPNG(c)).b64, ...wave, t0: 0, t1: introDur })
    }
    if (bar && bookEnd > bookStart) {
      drawBrightBar(c, bar)
      overlays.push({ png: (await canvasPNG(c)).b64, ...bar, t0: bookStart, t1: bookEnd })
    }
    let thumb = ''
    if (o.extras.thumb) {
      drawScene(c, common, { kind: 'thumb', meta: `${minutes} phút · Giọng ${d.voice}`, badge })
      thumb = (await canvasPNG(c)).b64
    }
    const desc = [
      partial ? `${d.title} — Nghe thử sách nói (${part}), giọng ${d.voice}.` : `${d.title} — Sách nói đầy đủ, giọng ${d.voice}.`,
      ...(creditsOf(d) ? [creditsOf(d) + '.'] : []),
      '',
      ...marks.map((m, k) => `${clock(k === 0 ? 0 : m.t)} ${m.label}`),
      '',
      'Sách nói tạo bằng Sano — tự tạo sách nói miễn phí từ file Word, EPUB, PDF: https://sanobook.com',
    ].join('\n')
    bv.desc = o.extras.desc ? desc : ''
    bv.title = outName
    const { w, h } = sceneSize(o.ratio, 'main')
    if (cancelled) return
    await bookVideoFinish(id, {
      frames: frames.map((f) => f.sec), segs, overlays, width: w, height: h, title: outName, thumb,
      srt: o.extras.srt ? srt.map((s, k) => `${k + 1}\n${srtTime(s.a)} --> ${srtTime(s.b)}\n${s.t}\n`).join('\n') : '',
      desc: o.extras.desc ? desc + '\n' : '',
    })
    bv.step = 'audio'
  } catch (e) {
    if (cancelled) return
    bv.step = 'error'
    bv.error = errText(e)
  }
}
