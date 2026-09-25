// Trình phát sách dùng chung cả app (wireframe D4): màn nghe và thanh nghe nhỏ
// cùng điều khiển một thẻ audio, nên rời màn nghe (về Thư viện, Tạo sách, Cài đặt)
// sách vẫn phát tiếp. Mở cuốn khác thì thay cuốn đang phát.
import { computed, reactive, watch } from 'vue'
import { book, errText, type BookDetail } from './backend'
import { beforeClipPlay, clearAudioSource, setAudioSource } from './audio'
import { loadPosition, savePosition } from './position'
import { state } from './store'

export const SPEEDS = [0.75, 1, 1.25, 1.5, 1.75, 2]
const SPEED_KEY = 'sano.speed'

// Nhớ tốc độ đã chọn giữa các lần nghe (tiện ích riêng của máy, mất cũng không sao).
function loadSpeed() {
  try {
    const v = Number(localStorage.getItem(SPEED_KEY))
    return SPEEDS.includes(v) ? v : 1
  } catch {
    return 1
  }
}

export const player = reactive({
  slug: '', // cuốn đang nạp (khác state.playerSlug trong lúc đang mở cuốn mới)
  detail: null as BookDetail | null,
  error: '',
  current: 0,
  playing: false,
  time: 0,
  duration: 0,
  speed: loadSpeed(),
})

const audio = new Audio()
audio.preload = 'auto'
let pendingSeek = 0
let loadSeq = 0
let loading: Promise<void> = Promise.resolve()
let played = false // chỉ mở rồi thoát, chưa phát → không tính là "nghe gần đây"

export const tracks = computed(() => player.detail?.tracks ?? [])
export const track = computed(() => tracks.value[player.current])
export const totalSec = computed(() => tracks.value.reduce((n, t) => n + t.durationSec, 0))
export const pctTrack = computed(() => (player.duration ? Math.min(100, (player.time / player.duration) * 100) : 0))

/** Nạp cuốn `slug` (giữ nguyên nếu đang là cuốn đó). */
async function open(slug: string, autoplay: boolean) {
  if (slug && slug === player.slug && player.detail) {
    if (autoplay && !player.playing) void play()
    return
  }
  remember()
  audio.pause()
  player.playing = false
  player.detail = null
  player.error = ''
  player.slug = slug
  played = false
  if (!slug) {
    clearAudioSource(audio)
    return
  }
  try {
    const d = await book(slug)
    if (player.slug !== slug) return // đã mở cuốn khác trong lúc chờ
    player.detail = d
    const pos = loadPosition(slug)
    load(Math.min(pos?.track ?? 0, Math.max(0, tracks.value.length - 1)), pos?.time ?? 0)
    if (autoplay) void play()
  } catch (e) {
    if (player.slug === slug) player.error = errText(e)
  }
}

watch(
  () => state.playerSlug,
  (slug) => {
    const autoplay = state.playerAutoplay // bấm phát ở hàng Nghe tiếp
    state.playerAutoplay = false
    void open(slug, autoplay)
  },
)

/** Mở lại cùng cuốn từ Thư viện kèm "phát ngay" (slug không đổi nên watch không chạy). */
watch(
  () => state.playerAutoplay,
  (on) => {
    if (!on || state.playerSlug !== player.slug) return
    state.playerAutoplay = false
    void open(player.slug, true)
  },
)

function load(i: number, at = 0) {
  const t = tracks.value[i]
  if (!t) return
  player.current = i
  player.time = at
  player.duration = t.durationSec
  pendingSeek = at
  const seq = ++loadSeq
  // Linux: nạp qua blob (bất đồng bộ, xem lib/audio.ts); play() đợi xong mới phát.
  loading = setAudioSource(audio, t.url)
    .then(() => {
      if (seq !== loadSeq) return // đã chọn tiểu mục khác
      audio.defaultPlaybackRate = player.speed // nạp file mới trình duyệt đặt lại tốc độ theo giá trị này
      audio.playbackRate = player.speed
    })
    .catch((e) => {
      if (seq === loadSeq) player.error = errText(e)
    })
}

export async function play() {
  player.error = ''
  try {
    const seq = loadSeq
    await loading
    if (seq !== loadSeq || player.error) return
    await audio.play()
  } catch (e) {
    player.error = 'Không phát được: ' + errText(e)
  }
}

export function pause() {
  audio.pause()
}

export function toggle() {
  if (!track.value) return
  if (player.playing) audio.pause()
  else void play()
}

export function pick(i: number) {
  load(i)
  void play()
}

export function skip(delta: number) {
  const i = player.current + delta
  if (i < 0 || i >= tracks.value.length) return
  load(i)
  void play()
}

export function seek(delta: number) {
  audio.currentTime = Math.min(Math.max(0, audio.currentTime + delta), audio.duration || player.duration)
}

/** Tua tới vị trí `frac` (0..1) của tiểu mục đang phát. */
export function seekFrac(frac: number) {
  const d = audio.duration || player.duration
  if (d) audio.currentTime = Math.min(Math.max(0, frac), 1) * d
}

export function setSpeed(v: number) {
  player.speed = v
  audio.defaultPlaybackRate = v
  audio.playbackRate = v
  try {
    localStorage.setItem(SPEED_KEY, String(v))
  } catch {
    // không lưu được thì thôi
  }
}

/** Dừng hẳn và bỏ cuốn đang phát (nút ✕ ở thanh nghe nhỏ, xoá sách). */
export function closePlayer() {
  remember()
  audio.pause()
  clearAudioSource(audio)
  loadSeq++
  player.slug = ''
  player.detail = null
  player.playing = false
  player.error = ''
  state.playerSlug = ''
}

/** Cuốn `slug` vừa bị xoá: đang phát thì dừng. */
export function forgetBook(slug: string) {
  if (slug && (slug === player.slug || slug === state.playerSlug)) closePlayer()
}

/** Thông tin sách vừa sửa (tên, tác giả, bìa): cập nhật cuốn đang phát, không nạp lại âm thanh. */
export async function refreshInfo(slug: string) {
  if (slug !== player.slug || !player.detail) return
  try {
    const d = await book(slug)
    if (player.slug === slug && player.detail) player.detail = { ...d, tracks: player.detail.tracks }
  } catch {
    // giữ thông tin cũ
  }
}

function remember() {
  const slug = player.slug
  if (!slug || !tracks.value.length || !played) return
  const before = tracks.value.slice(0, player.current).reduce((n, t) => n + t.durationSec, 0)
  const pct = totalSec.value ? Math.min(100, ((before + player.time) / totalSec.value) * 100) : 0
  savePosition(slug, { track: player.current, time: player.time, pct })
}

// Nghe mẫu giọng / nghe thử ở Tạo sách → tạm dừng sách để không chồng tiếng.
beforeClipPlay(() => audio.pause())

audio.addEventListener('loadedmetadata', () => {
  if (audio.duration && isFinite(audio.duration)) player.duration = audio.duration
  if (pendingSeek) audio.currentTime = pendingSeek
  pendingSeek = 0
})
audio.addEventListener('play', () => {
  player.playing = true
  played = true
})
audio.addEventListener('pause', () => {
  player.playing = false
  remember()
})
let lastSave = 0
audio.addEventListener('timeupdate', () => {
  player.time = audio.currentTime
  if (Date.now() - lastSave > 5000) {
    lastSave = Date.now()
    remember()
  }
})
audio.addEventListener('ended', () => {
  if (player.current < tracks.value.length - 1) {
    load(player.current + 1)
    void play()
  } else {
    player.time = player.duration
    remember()
  }
})
audio.addEventListener('error', () => {
  if (audio.src) player.error = 'Không đọc được file âm thanh của tiểu mục này'
})
window.addEventListener('beforeunload', remember)
