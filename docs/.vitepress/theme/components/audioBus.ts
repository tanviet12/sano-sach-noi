// Chỉ một trình phát chạy tại một thời điểm trên trang: phát cái này thì dừng cái kia.
const EVENT = 'sano-audio-play'

export function announcePlay(id: string) {
  window.dispatchEvent(new CustomEvent(EVENT, { detail: id }))
}

export function onOtherPlay(id: string, stop: () => void) {
  const handler = (e: Event) => {
    if ((e as CustomEvent<string>).detail !== id) stop()
  }
  window.addEventListener(EVENT, handler)
  return () => window.removeEventListener(EVENT, handler)
}

export const fmt = (s: number) => {
  const t = Math.max(0, Math.floor(s || 0))
  return `${Math.floor(t / 60)}:${String(t % 60).padStart(2, '0')}`
}
