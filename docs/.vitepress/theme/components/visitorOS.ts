// Nhận diện hệ điều hành người xem để chọn sẵn file cài (nút tải trang chủ, trang Tải về).
// Lúc build (SSR) chưa biết máy → null; vào trình duyệt mới nhận diện.
import { onMounted, ref } from 'vue'

export type OS = 'mac' | 'win' | 'linux'

export function detectOS(): OS | 'mobile' | null {
  const nav = navigator as Navigator & { userAgentData?: { platform?: string; mobile?: boolean } }
  const platform = nav.userAgentData?.platform || nav.platform || ''
  const ua = nav.userAgent
  const touchMac = /Mac/.test(platform) && nav.maxTouchPoints > 1 // iPadOS báo là Mac
  if (nav.userAgentData?.mobile || /Android|iPhone|iPad|iPod/i.test(ua) || touchMac) return 'mobile'
  if (/Mac/i.test(platform) || /Macintosh/.test(ua)) return 'mac'
  if (/Win/i.test(platform) || /Windows/.test(ua)) return 'win'
  if (/Linux|X11|CrOS/i.test(platform + ua)) return 'linux'
  return null
}

/** null = chưa biết (SSR / máy lạ) · 'mobile' = điện thoại, máy tính bảng (Sano chỉ cài trên máy tính). */
export function useVisitorOS() {
  const os = ref<OS | 'mobile' | null>(null)
  onMounted(() => (os.value = detectOS()))
  return os
}
