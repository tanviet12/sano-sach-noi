// Một trình phát dùng chung cho các nút "Nghe" (giọng mẫu, nghe thử): bấm đoạn
// khác thì dừng đoạn đang phát.
import { onBeforeUnmount, ref } from 'vue'

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
    el.src = url
    playing.value = id
    try {
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

  onBeforeUnmount(stop)
  return { playing, error, toggle, stop }
}
