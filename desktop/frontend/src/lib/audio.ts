// Một trình phát dùng chung cho các nút "Nghe" (giọng mẫu, nghe thử): bấm đoạn
// khác thì dừng đoạn đang phát.
import { onBeforeUnmount, ref } from 'vue'
import { platform } from './backend'

// Linux (WebKitGTK): trình phát âm thanh không tải được file qua scheme riêng
// của Wails (NotSupportedError, mã 4) — fetch() thì được. Nên trên Linux đọc
// file bằng fetch rồi phát qua URL blob:. macOS/Windows phát thẳng (tua theo
// Range, không nạp cả file vào bộ nhớ). Đã thử trên Ubuntu 22.04/24.04, Debian 12
// (scripts/ci/webkit_audio_probe.py).
let blobMode: Promise<boolean> | null = null
// ?blobaudio=1 lúc dev: thử cách blob trên máy không phải Linux.
const forceBlob = import.meta.env.DEV && new URLSearchParams(window.location.search).get('blobaudio') === '1'
const useBlob = () => (blobMode ??= platform().then((p) => p === 'linux' || forceBlob))
const blobs = new WeakMap<HTMLMediaElement, string>()

/** Gán nguồn cho thẻ audio (Linux: qua blob). Lỗi tải thì ném lỗi. */
export async function setAudioSource(el: HTMLMediaElement, url: string): Promise<void> {
  const old = blobs.get(el)
  blobs.delete(el)
  if (!(await useBlob())) {
    el.src = url
  } else {
    const res = await fetch(url)
    if (!res.ok) throw new Error(`Không đọc được file âm thanh (mã ${res.status})`)
    const data = await res.blob()
    const obj = URL.createObjectURL(data.type ? data : new Blob([data], { type: 'audio/mpeg' }))
    blobs.set(el, obj)
    el.src = obj
  }
  if (old) URL.revokeObjectURL(old)
}

/** Bỏ nguồn và giải phóng blob (nếu có). */
export function clearAudioSource(el: HTMLMediaElement) {
  el.removeAttribute('src')
  const old = blobs.get(el)
  if (old) URL.revokeObjectURL(old)
  blobs.delete(el)
}

const clipHooks: (() => void)[] = []

/** Gọi trước khi một đoạn nghe mẫu/nghe thử bắt đầu phát (trình phát sách tạm dừng). */
export function beforeClipPlay(fn: () => void) {
  clipHooks.push(fn)
}

export function useClipPlayer(onPlay?: (id: string) => void) {
  const playing = ref<string | null>(null)
  const error = ref('')
  const el = new Audio()
  el.preload = 'auto'
  el.addEventListener('ended', () => (playing.value = null))
  // Báo "đã nghe" khi âm thanh thật sự chạy (không dựa vào promise của play()).
  el.addEventListener('playing', () => {
    if (playing.value) onPlay?.(playing.value)
  })
  el.addEventListener('error', () => {
    if (playing.value) error.value = 'Không phát được đoạn này'
    playing.value = null
  })

  async function toggle(id: string, url: string) {
    error.value = ''
    if (playing.value === id) {
      el.pause()
      playing.value = null
      return
    }
    el.pause()
    playing.value = id
    clipHooks.forEach((fn) => fn())
    try {
      await setAudioSource(el, url)
      if (playing.value !== id) return // bấm đoạn khác trong lúc đang tải
      await el.play()
    } catch (e) {
      if ((e as DOMException)?.name === 'AbortError') return // bấm đoạn khác khi đoạn này đang tải
      playing.value = null
      error.value = 'Không phát được: ' + String(e)
    }
  }

  function stop() {
    el.pause()
    playing.value = null
  }

  onBeforeUnmount(() => {
    stop()
    clearAudioSource(el)
  })
  return { playing, error, toggle, stop }
}
