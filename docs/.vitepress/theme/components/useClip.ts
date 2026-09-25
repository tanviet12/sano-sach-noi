// Phát từng đoạn ngắn (một giọng mẫu) bằng một thẻ <audio> dùng chung trong component.
// Bấm lại đoạn đang phát thì dừng; phát trình phát khác trên trang thì tự dừng (audioBus).
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { withBase } from 'vitepress'
import { announcePlay, onOtherPlay } from './audioBus'

export function useClip(id: string) {
  const current = ref('') // src đang phát
  const time = ref(0)
  let audio: HTMLAudioElement | undefined

  function stop() {
    audio?.pause()
    current.value = ''
  }
  function toggle(src: string) {
    if (current.value === src && audio && !audio.paused) return stop()
    if (!audio) {
      audio = new Audio()
      audio.preload = 'none'
      audio.addEventListener('timeupdate', () => (time.value = audio?.currentTime ?? 0))
      // Không nghe sự kiện pause: đổi src khi đang phát cũng bắn pause, sẽ xoá nhầm đoạn mới
      audio.addEventListener('ended', () => (current.value = ''))
    }
    audio.pause()
    audio.src = withBase(src)
    time.value = 0
    current.value = src
    announcePlay(id)
    void audio.play().catch(() => (current.value = ''))
  }

  let off: (() => void) | undefined
  onMounted(() => (off = onOtherPlay(id, stop)))
  onBeforeUnmount(() => {
    off?.()
    stop()
  })
  return { current, time, toggle }
}
