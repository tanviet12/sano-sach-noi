// Trạng thái xuất M4B dùng chung cho Trình phát và màn render xong. Mỗi lúc
// một cuốn (Go từ chối cuốn thứ hai); tiến độ nhận qua sự kiện m4b:*.
import { reactive } from 'vue'
import { cancelM4B as goCancel, errText, exportM4B as goExport, m4bStatus, onEvent, revealM4B, type M4BStatus } from './backend'

export const m4b = reactive({
  status: null as M4BStatus | null,
  asking: false, // đang mở hộp lưu file
  error: '', // lỗi khi bấm (vd thiếu ffmpeg) — khác lỗi trong lúc xuất
  errorSlug: '',
})

let listening = false
function listen() {
  if (listening) return
  listening = true
  onEvent<M4BStatus>('m4b:progress', (st) => { m4b.status = st })
  onEvent<M4BStatus>('m4b:finished', (st) => { m4b.status = st })
  void m4bStatus().then((st) => { if (st && !m4b.status) m4b.status = st }).catch(() => {})
}

/** Gọi trong setup của component dùng trạng thái M4B. */
export function useM4B() {
  listen()
  return m4b
}

/** Đang xuất (hoặc đang hỏi nơi lưu) — khoá nút xuất ở mọi màn. */
export function m4bBusy(): boolean {
  return m4b.asking || !!m4b.status?.running
}

export async function startM4B(slug: string) {
  if (m4bBusy()) return
  m4b.error = ''
  m4b.errorSlug = slug
  m4b.asking = true
  try {
    const st = await goExport(slug)
    if (st) m4b.status = st
  } catch (e) {
    m4b.error = errText(e)
  } finally {
    m4b.asking = false
  }
}

export async function cancelM4B() {
  await goCancel()
}

export async function showM4B() {
  m4b.error = ''
  try {
    await revealM4B()
  } catch (e) {
    m4b.error = errText(e)
    m4b.errorSlug = m4b.status?.slug ?? ''
  }
}

/** "12,4 MB" */
export function fmtSize(bytes: number): string {
  if (bytes >= 1 << 30) return (bytes / (1 << 30)).toLocaleString('vi-VN', { maximumFractionDigits: 1 }) + ' GB'
  return (bytes / (1 << 20)).toLocaleString('vi-VN', { maximumFractionDigits: 1 }) + ' MB'
}
