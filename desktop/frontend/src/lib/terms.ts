// Điều khoản sử dụng — một nguồn: docs/dieu-khoan-su-dung.md (nhúng lúc build).
// Phiên bản lấy từ dòng "Phiên bản N" trong file; tăng số đó là app hỏi lại người dùng.
import raw from '../../../../docs/dieu-khoan-su-dung.md?raw'

export const TERMS_VERSION = Number(/Phiên bản (\d+)/.exec(raw)?.[1] ?? 1)

export type TermsBlock = { kind: 'h2' | 'li' | 'p' | 'hr'; html: string }

const escape = (s: string) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
const inline = (s: string) => escape(s).replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')

/** Markdown đơn giản của file điều khoản → khối hiển thị (bỏ tiêu đề # đầu file). */
export const TERMS_BLOCKS: TermsBlock[] = raw
  .split('\n')
  .filter((l) => l.trim() && !l.startsWith('# '))
  .map((l) => {
    if (l.startsWith('## ')) return { kind: 'h2', html: inline(l.slice(3)) }
    if (l.startsWith('- ')) return { kind: 'li', html: inline(l.slice(2)) }
    if (l.trim() === '---') return { kind: 'hr', html: '' }
    return { kind: 'p', html: inline(l) }
  })
