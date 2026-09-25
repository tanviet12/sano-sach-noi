// Dữ liệu trang demo (/demo) + dải giọng ở trang chủ. Audio render bằng Sano từ sách mẫu tự viết
// (docs/demo-books), mỗi cuốn 3 chương đầu, mỗi cuốn một giọng. Thời lượng (giây) để hiện trước
// khi tải audio; trình phát đọc lại độ dài thật khi phát.

export type DemoBook = {
  slug: string
  title: string
  voice: string
  voiceDesc: string
  chapters: { title: string; sec: number }[]
}

export const demoBooks: DemoBook[] = [
  {
    slug: 'ky-nang-mem-cho-nguoi-tre', title: 'Kỹ năng mềm cho người trẻ', voice: 'Hải Đăng', voiceDesc: 'nam · Bắc · tự nhiên',
    chapters: [{ title: 'Lắng nghe chủ động', sec: 62.1 }, { title: 'Giao tiếp hiệu quả', sec: 61.6 }, { title: 'Làm việc nhóm', sec: 61.6 }],
  },
  {
    slug: 'tam-ly-tich-cuc', title: 'Tâm lý tích cực', voice: 'Trúc Ly', voiceDesc: 'nữ · Bắc · tự nhiên',
    chapters: [{ title: 'Tư duy phát triển', sec: 62.1 }, { title: 'Đứng dậy sau thất bại', sec: 58.5 }, { title: 'Lòng biết ơn', sec: 61.2 }],
  },
  {
    slug: 'khoi-nghiep-tu-so-0', title: 'Khởi nghiệp từ số 0', voice: 'Thái Sơn', voiceDesc: 'nam · Nam · kể chuyện',
    chapters: [{ title: 'Bắt đầu từ vấn đề', sec: 74.3 }, { title: 'Kiểm chứng ý tưởng', sec: 75.0 }, { title: 'Sản phẩm đầu tiên', sec: 75.8 }],
  },
  {
    slug: 'tai-chinh-ca-nhan-co-ban', title: 'Tài chính cá nhân cơ bản', voice: 'Thục Đoan', voiceDesc: 'nữ · Nam · kể chuyện',
    chapters: [{ title: 'Quản lý chi tiêu', sec: 76.2 }, { title: 'Tiết kiệm', sec: 72.9 }, { title: 'Đầu tư', sec: 73.9 }],
  },
  {
    slug: 'lanh-dao-cho-quan-ly-moi', title: 'Lãnh đạo cho quản lý mới', voice: 'Ngọc Trân', voiceDesc: 'nữ · Trung · tự nhiên',
    chapters: [{ title: 'Từ người làm đến người dẫn dắt', sec: 82.9 }, { title: 'Phản hồi', sec: 80.6 }, { title: 'Xây dựng đội ngũ', sec: 78.9 }],
  },
]

export const bookAudio = (b: DemoBook, i: number) => `/audio/demo/${b.slug}-${i + 1}.mp3`
export const bookCover = (b: DemoBook) => `/audio/demo/${b.slug}.jpg`

export type DemoVoice = { name: string; file: string; desc: string; sec: number }

// 12 giọng đọc cùng một đoạn (trích "Tâm lý tích cực"), nhóm theo miền
export const voiceRegions: { name: string; voices: DemoVoice[] }[] = [
  {
    name: 'Miền Bắc',
    voices: [
      { name: 'Thiện Minh', file: 'thien-minh', desc: 'nam · kể chuyện', sec: 16.7 },
      { name: 'Trúc Ly', file: 'truc-ly', desc: 'nữ · tự nhiên', sec: 15.4 },
      { name: 'Mai Anh', file: 'mai-anh', desc: 'nữ · tin tức', sec: 18.4 },
      { name: 'Hải Đăng', file: 'hai-dang', desc: 'nam · tự nhiên', sec: 15.8 },
      { name: 'Thiền Tâm Đức', file: 'thien-tam-duc', desc: 'nam · kể chuyện', sec: 22.2 },
      { name: 'Ngọc Huyền', file: 'ngoc-huyen', desc: 'nữ · tự nhiên', sec: 16.7 },
    ],
  },
  {
    name: 'Miền Trung',
    voices: [
      { name: 'Quang Sơn', file: 'quang-son', desc: 'nam · tự nhiên', sec: 19.6 },
      { name: 'Ngọc Trân', file: 'ngoc-tran', desc: 'nữ · tự nhiên', sec: 20.7 },
    ],
  },
  {
    name: 'Miền Nam',
    voices: [
      { name: 'Thái Sơn', file: 'thai-son', desc: 'nam · kể chuyện', sec: 19.6 },
      { name: 'Thùy Dung', file: 'thuy-dung', desc: 'nữ · tin tức', sec: 15.2 },
      { name: 'Thục Đoan', file: 'thuc-doan', desc: 'nữ · kể chuyện', sec: 18.9 },
      { name: 'Mỹ Duyên', file: 'my-duyen', desc: 'nữ · đọc truyện', sec: 21.9 },
    ],
  },
]
export const allVoices = voiceRegions.flatMap((r) => r.voices)
export const voiceAudio = (v: DemoVoice) => `/audio/giong/${v.file}.mp3`

export const voicePassage =
  'Tư duy phát triển là niềm tin rằng năng lực của con người không cố định, mà có thể cải thiện qua nỗ lực, học hỏi và rèn luyện. Đây là một trong những thái độ quan trọng nhất quyết định bạn tiến xa đến đâu. Người có tư duy cố định thường nghĩ rằng mình sinh ra đã giỏi hoặc đã kém ở một lĩnh vực, và khó thay đổi.'
