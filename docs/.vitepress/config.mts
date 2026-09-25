import { fileURLToPath } from 'node:url'
import { defineConfig, type HeadConfig } from 'vitepress'

// Trang tài liệu + trang giới thiệu Sano (VitePress, GitHub Pages).
// Bám khuôn Chat-Quality-Agent/docs; giao diện theo wireframe đã duyệt
// desktop/frontend/src/wireframes/WfLanding.vue + WfDocsPage.vue.

const SITE = 'https://tanviet12.github.io/sano-sach-noi/'
const REPO = 'https://github.com/tanviet12/sano-sach-noi'
const SITE_TITLE = 'Sano – Tạo sách nói bằng AI từ file Word'
const SITE_DESC =
  'Sano giúp bạn tự làm sách nói bằng AI từ file Word của chính mình: giọng đọc tiếng Việt chạy trên máy, có mục lục chương, xuất M4B nghe trên điện thoại. Miễn phí, mã nguồn mở.'

// Huy hiệu vbsec ở chân trang: CHỈ bật khi lượt quét vbsec trên bản phát hành cho kết quả ĐẠT.
// Bật: passed: true + date: 'dd/mm/yyyy' (ngày quét).
const vbsec = { passed: true, date: '24/09/2026' }

const vbsecBadge = vbsec.passed
  ? `<br><a class="vbsec-badge" href="https://github.com/tanviet12/vbsec" target="_blank" rel="noopener"><span class="vbsec-name">vbsec</span><span class="vbsec-state">đã quét bảo mật · đạt${vbsec.date ? ' · ' + vbsec.date : ''}</span></a>`
  : ''

// Trang có sẵn trong docs/ nhưng không đặt frontmatter được (app nhúng nguyên văn, hoặc là
// tài liệu kỹ thuật giữ nguyên nội dung): gán title + description ở đây.
const extraMeta: Record<string, { title: string; description: string }> = {
  'dieu-khoan-su-dung.md': {
    title: 'Điều khoản sử dụng',
    description:
      'Điều khoản sử dụng Sano, phần mềm tạo sách nói bằng AI từ file Word: chỉ dùng tài liệu bạn có quyền sử dụng, không làm sách nói từ tác phẩm còn bản quyền, dữ liệu nằm trên máy bạn.',
  },
  'tts-build-guide.md': {
    title: 'Đọc giọng bằng VieNeu-TTS (dòng lệnh)',
    description:
      'Hướng dẫn cài VieNeu-TTS v3 và tạo sách nói bằng AI từ dòng lệnh với sano-docx2tts: giọng đọc, chuẩn hoá văn nói, xuất M4B.',
  },
  'book-zip-format.md': {
    title: 'Định dạng gói zip sách nói',
    description: 'Cấu trúc gói book-<slug>.zip mà Sano tạo để sao lưu hoặc chuyển sách nói sang máy khác.',
  },
  'vieneu-tts-patch.md': {
    title: 'Ghi chú VieNeu-TTS v3',
    description: 'Ghi chú kỹ thuật về cách Sano gọi VieNeu-TTS v3 Turbo để đọc sách nói tiếng Việt.',
  },
  'design-system.md': {
    title: 'Design system',
    description: 'Token màu, khoảng cách, chữ và component dùng trong giao diện phần mềm Sano.',
  },
}

const guide = [
  { text: 'Cài đặt Sano', link: '/cai-dat' },
  { text: 'Tạo sách đầu tiên', link: '/tao-sach-dau-tien' },
  { text: 'Làm mượt tài liệu', link: '/lam-muot-tai-lieu' },
  { text: 'Nghe trên điện thoại', link: '/nghe-tren-dien-thoai' },
  { text: 'Nghe khi lái xe ô tô', link: '/nghe-khi-lai-xe' },
  { text: 'Câu hỏi thường gặp', link: '/cau-hoi-thuong-gap' },
  { text: 'Gỡ cài đặt', link: '/go-cai-dat' },
]

export default defineConfig({
  lang: 'vi-VN',
  title: SITE_TITLE,
  titleTemplate: ':title – Sano',
  description: SITE_DESC,
  base: '/sano-sach-noi/',
  cleanUrls: true,
  lastUpdated: true,
  // demo-books: văn bản thô để render audio mẫu; prompts: lời nhắc mẫu app nhúng (?raw)
  srcExclude: ['demo-books/**', 'prompts/**', 'README.md'],

  sitemap: { hostname: SITE },

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/sano-sach-noi/favicon.svg' }],
    ['link', { rel: 'apple-touch-icon', href: '/sano-sach-noi/og-image.png' }],
    ['meta', { name: 'theme-color', content: '#c60505' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:site_name', content: 'Sano' }],
    ['meta', { property: 'og:locale', content: 'vi_VN' }],
    ['meta', { property: 'og:image', content: SITE + 'og-image.png' }],
    ['meta', { property: 'og:image:width', content: '512' }],
    ['meta', { property: 'og:image:height', content: '512' }],
    ['meta', { property: 'og:image:alt', content: 'Logo Sano' }],
    ['meta', { name: 'twitter:card', content: 'summary' }],
  ],

  // Trang không có frontmatter riêng → gán title/description ở extraMeta
  transformPageData(pageData) {
    const m = extraMeta[pageData.relativePath]
    if (m) {
      pageData.title = m.title
      pageData.description = m.description
      pageData.frontmatter.title ??= m.title
      pageData.frontmatter.description ??= m.description
    }
  },

  // og:title / og:description / og:url / canonical riêng từng trang
  transformHead({ pageData, siteData }) {
    const path = pageData.relativePath.replace(/(^|\/)index\.md$/, '$1').replace(/\.md$/, '')
    const url = SITE + path
    const fm = pageData.frontmatter
    const title =
      fm.titleTemplate === false || !pageData.title ? pageData.title || siteData.title : `${pageData.title} – Sano`
    const desc = pageData.description || siteData.description
    const head: HeadConfig[] = [
      ['link', { rel: 'canonical', href: url }],
      ['meta', { property: 'og:url', content: url }],
      ['meta', { property: 'og:title', content: title }],
      ['meta', { property: 'og:description', content: desc }],
      ['meta', { name: 'twitter:title', content: title }],
      ['meta', { name: 'twitter:description', content: desc }],
    ]
    return head
  },

  markdown: {
    container: { tipLabel: 'Mẹo', warningLabel: 'Lưu ý', dangerLabel: 'Cảnh báo', infoLabel: 'Thông tin', detailsLabel: 'Chi tiết' },
    config(md) {
      // Link tương đối trỏ ra ngoài trang tài liệu (../desktop/README.md, prompts/*.txt…) →
      // đổi thành link GitHub để không gãy trên trang (file gốc vẫn đọc tốt trên GitHub).
      md.core.ruler.push('sano-repo-links', (state) => {
        const file = (state.env as { relativePath?: string }).relativePath ?? ''
        const dir = file.includes('/') ? file.slice(0, file.lastIndexOf('/') + 1) : ''
        for (const block of state.tokens) {
          for (const t of block.children ?? []) {
            if (t.type !== 'link_open') continue
            const href = t.attrGet('href')
            if (!href || /^([a-z]+:|#|\/)/i.test(href)) continue
            const [p, hash = ''] = href.split('#')
            const outside = p.startsWith('../')
            const asset = !/\.(md|html)$/.test(p) && /\.[a-z0-9]+$/i.test(p)
            if (!outside && !asset) continue
            const full = new URL(p, 'file:///docs/' + dir).pathname.slice(1)
            t.attrSet('href', `${REPO}/blob/main/${full}${hash ? '#' + hash : ''}`)
            t.attrSet('target', '_blank')
            t.attrSet('rel', 'noopener')
          }
        }
      })
    },
  },

  vite: {
    // Logo tài trợ + logo lấy thẳng từ desktop/frontend/src/assets (một nguồn với app)
    server: { fs: { allow: [fileURLToPath(new URL('../..', import.meta.url))] } },
  },

  themeConfig: {
    logo: { src: '/logo.svg', alt: '' },
    siteTitle: 'Sano',

    nav: [
      { text: 'Trang chủ', link: '/', activeMatch: '^/$' },
      { text: 'Nghe thử', link: '/demo', activeMatch: '^/demo' },
      {
        text: 'Hướng dẫn',
        activeMatch: '^/(?!$|demo|tai-ve)',
        items: guide,
      },
      { text: 'Tải về', link: '/tai-ve', activeMatch: '^/tai-ve' },
    ],

    sidebar: [
      {
        text: 'Bắt đầu',
        items: [
          { text: 'Cài đặt Sano', link: '/cai-dat' },
          { text: 'Tạo sách đầu tiên', link: '/tao-sach-dau-tien' },
        ],
      },
      {
        text: 'Dùng Sano',
        items: [
          { text: 'Làm mượt tài liệu', link: '/lam-muot-tai-lieu' },
          { text: 'Nghe trên điện thoại', link: '/nghe-tren-dien-thoai' },
          { text: 'Nghe khi lái xe ô tô', link: '/nghe-khi-lai-xe' },
          { text: 'Thư viện & danh mục', link: '/thu-vien' },
          { text: 'Xuất M4B / gói zip', link: '/xuat-m4b-goi-zip' },
        ],
      },
      {
        text: 'Hỗ trợ',
        items: [
          { text: 'Câu hỏi thường gặp', link: '/cau-hoi-thuong-gap' },
          { text: 'Mở app lần đầu', link: '/mo-app-lan-dau' },
          { text: 'Gỡ cài đặt', link: '/go-cai-dat' },
          { text: 'Điều khoản sử dụng', link: '/dieu-khoan-su-dung' },
          { text: 'Nhật ký thay đổi', link: '/nhat-ky-thay-doi' },
        ],
      },
      {
        text: 'Dành cho nhà phát triển',
        collapsed: true,
        items: [
          { text: 'Đọc giọng bằng dòng lệnh', link: '/tts-build-guide' },
          { text: 'Định dạng gói zip', link: '/book-zip-format' },
          { text: 'Ghi chú VieNeu-TTS', link: '/vieneu-tts-patch' },
          { text: 'Design system', link: '/design-system' },
        ],
      },
    ],

    socialLinks: [{ icon: 'github', link: REPO, ariaLabel: 'Mã nguồn trên GitHub' }],

    search: {
      provider: 'local',
      options: {
        translations: {
          button: { buttonText: 'Tìm trong tài liệu', buttonAriaLabel: 'Tìm trong tài liệu' },
          modal: {
            displayDetails: 'Hiện chi tiết',
            resetButtonTitle: 'Xoá',
            backButtonTitle: 'Đóng',
            noResultsText: 'Không tìm thấy kết quả cho',
            footer: { selectText: 'chọn', navigateText: 'di chuyển', closeText: 'đóng' },
          },
        },
      },
    },

    editLink: { pattern: `${REPO}/edit/main/docs/:path`, text: 'Sửa trang này trên GitHub' },
    lastUpdated: { text: 'Cập nhật lần cuối', formatOptions: { dateStyle: 'short' } },
    outline: { level: [2, 3], label: 'Trên trang này' },
    docFooter: { prev: 'Trang trước', next: 'Trang sau' },
    returnToTopLabel: 'Lên đầu trang',
    sidebarMenuLabel: 'Menu',
    darkModeSwitchLabel: 'Giao diện',
    lightModeSwitchTitle: 'Chuyển sang giao diện sáng',
    darkModeSwitchTitle: 'Chuyển sang giao diện tối',
    langMenuLabel: 'Ngôn ngữ',
    externalLinkIcon: false,
    notFound: {
      title: 'Không tìm thấy trang',
      quote: 'Trang này không có hoặc đã đổi địa chỉ.',
      linkText: 'Về trang chủ',
      linkLabel: 'Về trang chủ',
    },

    footer: {
      message: `<a href="${REPO}" target="_blank" rel="noopener">GitHub</a> · Phát hành theo giấy phép <a href="${REPO}/blob/main/LICENSE" target="_blank" rel="noopener">MIT</a> · Giọng đọc <a href="https://github.com/pnnbao97/VieNeu-TTS" target="_blank" rel="noopener">VieNeu-TTS</a> (Apache-2.0)${vbsecBadge}`,
      copyright: 'Copyright © 2026 Bùi Tấn Việt',
    },

    // Đọc được trong theme (useData().theme.value.vbsec) nếu cần hiện ở chỗ khác
    vbsec,
  },
})
