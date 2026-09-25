// Chữ chạy theo lời đọc (wireframe D5, cách A): tách câu và ƯỚC LƯỢNG thời điểm
// bắt đầu từng câu từ thời lượng tiểu mục. Không có mốc thời gian thật trong gói
// sách, nên dựa vào cách bộ đọc VieNeu ngắt nghỉ: nói đều theo số ký tự, nghỉ
// ~0,3s ở dấu phẩy, ~0,5s hết câu, ~0,7s hết đoạn (scripts/tts/audio_gen.py).
//
// Hiện chữ gốc (như trong file Word) nhưng tính giờ trên lời đọc thật (script:
// có tên tiểu mục đọc ở đầu, số đã viết thành chữ) khi hai bên khớp số câu.

/** Tham số ước lượng (giây). `perSentence`: thời gian cố định mỗi câu ngoài phần
 *  theo ký tự — câu rất ngắn ("Một.", "Hai.") đọc lâu hơn số ký tự của nó. */
//  Chỉnh trên 163 tiểu mục thật (2 giọng): so với khoảng lặng đo từ file mp3, lệch
//  trung vị 0,35s, 90% câu dưới 1,4s. Nghỉ thật ngắn hơn cấu hình VieNeu vì phần
//  im lặng ở đầu/cuối mỗi đoạn audio đã bị cắt bớt khi ghép.
export const TIMING = { comma: 0.1, sentence: 0.3, paragraph: 0.5, perSentence: 0 }

export interface LyricSentence {
  text: string
  start: number // giây, tính từ đầu tiểu mục
  index: number // thứ tự câu trong tiểu mục
}
export interface Lyrics {
  paragraphs: LyricSentence[][]
  sentences: LyricSentence[]
}

interface Piece {
  text: string
  endOfParagraph: boolean
}

/** Tách đoạn (xuống dòng) rồi tách câu theo . ! ? … (giữ dấu, ngoặc/nháy đóng theo sau). */
export function splitSentences(text: string): Piece[][] {
  return text
    .split(/\n+/)
    .map((p) => p.trim())
    .filter(Boolean)
    .map((p) => {
      const parts = p.match(/[^.!?…]+(?:[.!?…]+["'”’»)\]]*)?|[.!?…]+/g) ?? [p]
      const out: string[] = []
      for (const raw of parts) {
        const s = raw.trim()
        if (!s) continue
        // Mẩu quá ngắn (vd "1." trong "Bước 1. …", "..." lẻ) gộp vào câu trước.
        if (out.length && (s.replace(/[^\p{L}\p{N}]/gu, '').length < 3 || /^[.!?…]/.test(s))) out[out.length - 1] += ' ' + s
        else out.push(s)
      }
      return out.map((s, i) => ({ text: s, endOfParagraph: i === out.length - 1 }))
    })
}

const commas = (s: string) => (s.match(/[,;:]/g) ?? []).length
const chars = (s: string) => s.replace(/\s+/g, ' ').length

/** Thời điểm bắt đầu từng mẩu khi đọc lần lượt trong `duration` giây. */
function startTimes(pieces: Piece[], duration: number): number[] {
  if (!pieces.length) return []
  const pauseAfter = pieces.map((p, i): number => (i === pieces.length - 1 ? 0 : p.endOfParagraph ? TIMING.paragraph : TIMING.sentence))
  const inner = pieces.map((p) => commas(p.text) * TIMING.comma + TIMING.perSentence)
  let pauses = pauseAfter.reduce((a, b) => a + b, 0) + inner.reduce((a, b) => a + b, 0)
  const totalChars = pieces.reduce((n, p) => n + chars(p.text), 0) || 1
  // Tiểu mục ngắn bất thường: nghỉ không được chiếm quá 40% thời lượng.
  const scale = pauses > duration * 0.4 ? (duration * 0.4) / pauses : 1
  pauses *= scale
  const perChar = Math.max(0, duration - pauses) / totalChars
  const out: number[] = []
  let t = 0
  pieces.forEach((p, i) => {
    out.push(t)
    t += chars(p.text) * perChar + (inner[i] + pauseAfter[i]) * scale
  })
  return out
}

/** Dựng lời hiển thị + thời điểm ước lượng cho một tiểu mục. */
export function buildLyrics(text: string, script: string, duration: number): Lyrics {
  const shown = splitSentences(text)
  const flatShown = shown.flat()
  let starts: number[] | null = null

  const spoken = splitSentences(script || text).flat()
  const lead = spoken.length - flatShown.length // câu đọc thêm ở đầu (tên tiểu mục)
  if (duration > 0 && lead >= 0 && lead <= 3) {
    starts = startTimes(spoken, duration).slice(lead)
  } else if (duration > 0) {
    // Lời đọc lệch số câu với chữ gốc: tính trên chữ gốc, thêm tên tiểu mục (đoạn
    // đầu của lời đọc) vào trước để chừa đúng phần mở đầu.
    const sp = splitSentences(script)
    const title = sp.length > shown.length ? sp[0] : []
    starts = startTimes([...title, ...flatShown], duration).slice(title.length)
  }

  let index = 0
  const paragraphs = shown.map((p) => p.map((s) => ({ text: s.text, start: starts?.[index] ?? 0, index: index++ })))
  return { paragraphs, sentences: paragraphs.flat() }
}

/**
 * Khoảng lặng trong âm thanh (giây kết thúc + độ dài), dò theo năng lượng từng ô
 * 20ms: dưới ngưỡng ~-35 dB so với đỉnh, kéo dài từ `minSec`.
 */
export function findSilences(samples: Float32Array, rate: number, minSec = 0.2): { end: number; len: number }[] {
  const win = Math.max(1, Math.round(rate * 0.02))
  const n = Math.floor(samples.length / win)
  const rms = new Float32Array(n)
  let peak = 0
  for (let w = 0; w < n; w++) {
    let sum = 0
    for (let i = w * win; i < (w + 1) * win; i++) sum += samples[i] * samples[i]
    rms[w] = Math.sqrt(sum / win)
    if (rms[w] > peak) peak = rms[w]
  }
  const thr = peak * 0.018 // ≈ -35 dB
  const out: { end: number; len: number }[] = []
  let start = -1
  for (let w = 0; w <= n; w++) {
    const quiet = w < n && rms[w] < thr
    if (quiet && start < 0) start = w
    if (!quiet && start >= 0) {
      const len = ((w - start) * win) / rate
      if (len >= minSec && start > 0 && w < n) out.push({ end: (w * win) / rate, len })
      start = -1
    }
  }
  return out
}

/**
 * Kéo thời điểm ước lượng về khoảng lặng thật gần nhất (trong ±`window` giây),
 * giữ đúng thứ tự câu. Khoảng lặng dài (hết câu) được ưu tiên hơn nghỉ dấu phẩy.
 */
export function snapToSilences(l: Lyrics, silences: { end: number; len: number }[], window = 1.5): Lyrics {
  let prev = 0
  let k = 0
  for (const s of l.sentences) {
    if (s.index === 0) continue
    let best = -1
    let bestScore = Infinity
    for (let j = k; j < silences.length; j++) {
      const d = silences[j].end - s.start
      if (d > window) break
      if (d < -window || silences[j].end <= prev) continue
      const score = Math.abs(d) - Math.min(silences[j].len, 0.8) * 0.8
      if (score < bestScore) {
        bestScore = score
        best = j
      }
    }
    if (best >= 0) {
      s.start = silences[best].end
      k = best + 1
    }
    s.start = Math.max(s.start, prev + 0.2)
    prev = s.start
  }
  return l
}

/** Câu đang đọc ở giây `t` (câu cuối cùng đã bắt đầu). */
export function sentenceAt(l: Lyrics, t: number): number {
  let cur = 0
  for (const s of l.sentences) {
    if (s.start <= t + 0.05) cur = s.index
    else break
  }
  return cur
}
