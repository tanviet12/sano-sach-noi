import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

/** Merge Tailwind class names safely (twMerge handles conflicts). */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Format số giây → "Xg Yp" (giờ/phút) cho thời lượng sách, hoặc "Yp" nếu < 1 giờ.
 *  Dùng cho tổng thời lượng sách + section. Trả "0p" cho giá trị <= 0 / không hợp lệ. */
export function formatDuration(sec: number | null | undefined): string {
  if (!sec || sec <= 0 || Number.isNaN(sec)) return '0p'
  const total = Math.floor(sec)
  const hours = Math.floor(total / 3600)
  const minutes = Math.floor((total % 3600) / 60)
  if (hours > 0) return minutes > 0 ? `${hours}g ${minutes}p` : `${hours}g`
  if (minutes > 0) return `${minutes}p`
  return `${total}s`
}

/** Format số giây → "M:SS" cho mốc thời gian phát (vd 90 → "1:30"). */
export function formatTimestamp(sec: number | null | undefined): string {
  const total = Math.max(0, Math.floor(sec ?? 0))
  const m = Math.floor(total / 60)
  const s = total % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

/** Format ISO timestamp → thời gian tương đối tiếng Việt ("vừa xong", "5 phút trước",
 *  "3 giờ trước", "2 ngày trước"). Quá 7 ngày → ngày dạng DD/MM/YYYY. Dùng cho thông báo. */
export function formatRelativeTime(iso: string | Date | null | undefined): string {
  if (!iso) return ''
  const d = typeof iso === 'string' ? new Date(iso) : iso
  if (Number.isNaN(d.getTime())) return ''
  const diffSec = Math.floor((Date.now() - d.getTime()) / 1000)
  if (diffSec < 0) return 'vừa xong'
  if (diffSec < 60) return 'vừa xong'
  const min = Math.floor(diffSec / 60)
  if (min < 60) return `${min} phút trước`
  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour} giờ trước`
  const day = Math.floor(hour / 24)
  if (day <= 7) return `${day} ngày trước`
  return formatDateVN(d).split(' ')[0]
}

/** Format ISO timestamp → "DD/MM/YYYY HH:mm" (Asia/Ho_Chi_Minh locale).
 *  Returns empty string for null/undefined/invalid input. */
export function formatDateVN(iso: string | Date | null | undefined): string {
  if (!iso) return ''
  const d = typeof iso === 'string' ? new Date(iso) : iso
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${pad(d.getDate())}/${pad(d.getMonth() + 1)}/${d.getFullYear()} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
