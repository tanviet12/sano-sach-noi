// Hằng số giao diện + dữ liệu giả chỉ dùng khi mở bằng trình duyệt thường
// (npm run dev, không có phần Go). Trong app, mọi dữ liệu Tạo sách / Thư viện
// đều đọc thật qua backend.ts.
import type { LibraryInfo, Outline, Voice } from './backend'

export const DOCS = 'https://tanviet12.github.io/sano-sach-noi'
export const REPO = 'https://github.com/tanviet12/sano-sach-noi'
export const AUTHOR_FB = 'https://www.facebook.com/buitanviet'

export const steps = [
  { n: 1, label: 'Nạp file' },
  { n: 2, label: 'Mục lục' },
  { n: 3, label: 'Giọng đọc' },
  { n: 4, label: 'Lời mở đầu' },
  { n: 5, label: 'Nghe thử' },
  { n: 6, label: 'Render' },
]

const sec = (stem: string, title: string, chars: number, toc = false) => ({ stem, title, chars, images: 0, toc, tocReason: toc ? 'tiêu đề' : '' })

/** Mục lục giả (bám wireframe) khi không có phần Go. */
export function mockOutline(): Outline {
  return {
    title: 'Kỹ năng giao tiếp',
    fileTitle: 'ky nang giao tiep',
    sections: 8,
    chars: 24900,
    sampleSentence: 'Bước thứ nhất, dừng việc đang làm và nhìn người nói.',
    chapters: [
      { title: 'Mục lục', sections: [sec('ch01-sec01', 'Mục lục', 640, true)] },
      { title: 'Chương 1. Lắng nghe chủ động', sections: [sec('ch02-sec01', '1.1 Vì sao lắng nghe khó', 2900), sec('ch02-sec02', '1.2 Ba bước lắng nghe', 3100), sec('ch02-sec03', '1.3 Bài tập mỗi ngày', 2200)] },
      { title: 'Chương 2. Giao tiếp rõ ràng', sections: [sec('ch03-sec01', '2.1 Nói ngắn, ý rõ', 4200), sec('ch03-sec02', '2.2 Hỏi để hiểu', 4900)] },
      { title: 'Chương 3. Quản lý thời gian', sections: [sec('ch04-sec01', '3.1 Việc quan trọng trước', 3800), sec('ch04-sec02', '3.2 Nói không đúng lúc', 3800)] },
    ],
    warnings: { images: 5, tables: 2, fakeHeadings: ['Ghi nhớ'], unknownAcronyms: [{ word: 'OKR', count: 3 }] },
  }
}

export function mockVoices(): Voice[] {
  // 25 giọng thật của VieNeu v3 Turbo (bản dev chạy trình duyệt không có bộ đọc)
  const v = (name: string, desc: string, featured = false) => ({ name, desc, featured })
  return [
    v('Adam bựa', 'Nam · Bắc · Phong cách tự nhiên', true), v('Trúc Ly', 'Nữ · Bắc · Phong cách tự nhiên', true),
    v('Thiện Minh', 'Nam · Bắc · Phong cách kể chuyện', true), v('Mai Anh', 'Nữ · Bắc · Phong cách tin tức', true),
    v('Hải Đăng', 'Nam · Bắc · Phong cách tự nhiên', true), v('Thùy Dung', 'Nữ · Nam · Phong cách tin tức', true),
    v('Thiền Tâm Đức', 'Nam · Bắc · Phong cách kể chuyện', true), v('Ngọc Huyền', 'Nữ · Bắc · Giọng đọc tự nhiên', true),
    v('Quang Sơn', 'Nam · Trung · Phong cách tự nhiên', true), v('Ngọc Trân', 'Nữ · Trung · Phong cách tự nhiên', true),
    v('Minh Đức', 'Nam · Bắc · Phong cách tin tức'), v('Phạm Tuyên', 'Nam · Bắc · Phong cách tự nhiên'),
    v('Thái Sơn', 'Nam · Nam · Phong cách kể chuyện'), v('Xuân Vĩnh', 'Nam · Bắc · Phong cách tự nhiên'),
    v('Thanh Bình', 'Nam · Bắc · Phong cách kể chuyện'), v('Ngọc Linh', 'Nữ · Bắc · Phong cách kể chuyện'),
    v('Đoan Trang', 'Nữ · Bắc · Phong cách tự nhiên'), v('Thục Đoan', 'Nữ · Nam · Phong cách kể chuyện'),
    v('Minh Triết', 'Nam · Nam · Phong cách tin tức'), v('Mỹ Duyên', 'Nữ · Nam · Phong cách đọc truyện'),
    v('Quỳnh Anh', 'Nữ · Bắc · Phong cách đọc truyện'), v('Đức Trí', 'Nam · Nam · Phong cách đọc truyện'),
    v('Kim Thanh', 'Nữ · Nam · Phong cách đọc truyện'), v('Adam', 'Nam · Nam · Giọng đọc tự nhiên'),
    v('Quốc Tuấn', 'Nam · Bắc · Phong cách tự nhiên'),
  ]
}

export function mockLibrary(): LibraryInfo {
  const voices = ['Hải Đăng', 'Ngọc Huyền', 'Thiện Minh', 'Hải Đăng', 'Thùy Dung', '']
  const b = (i: number, title: string, author: string, durationSec: number, category = '') => ({
    slug: `sach-${i}`, title, author, category, cover: '', coverUrl: '', zip: '', chapters: 4, sections: 8, durationSec,
    voice: voices[(i - 1) % voices.length],
    createdAt: new Date(Date.UTC(2026, 8, 20 - i)).toISOString(),
  })
  return {
    dir: '~/Sano/Sach',
    books: [
      b(1, 'Kỹ năng mềm cho người trẻ', 'Nguyễn Văn A', 4320, 'Kỹ năng'),
      b(2, 'Lãnh đạo cho quản lý mới', 'Trần Thị B', 3480, 'Kinh doanh'),
      b(3, 'Tài chính cá nhân cơ bản', 'Lê Văn C', 3900, 'Tài chính'),
      b(4, 'Khởi nghiệp từ số 0', 'Phạm Thị D', 2820, 'Kinh doanh'),
      b(5, 'Quản lý thời gian hiệu quả', 'Nguyễn Văn A', 7800, 'Kỹ năng'),
      b(6, 'Ghi chép cuộc họp', 'Phạm Thị D', 2040),
    ],
  }
}
