import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'
import mediumZoom from 'medium-zoom'
import { h, nextTick, onMounted, watch } from 'vue'
import { useRoute } from 'vitepress'
import './style.css'

import DocEyebrow from './components/DocEyebrow.vue'
import HomeHero from './components/HomeHero.vue'
import SamplePlayer from './components/SamplePlayer.vue'
import NormalizeCompare from './components/NormalizeCompare.vue'
import HowItWorks from './components/HowItWorks.vue'
import PhoneSection from './components/PhoneSection.vue'
import Highlights from './components/Highlights.vue'
import RightsSummary from './components/RightsSummary.vue'
import Roadmap from './components/Roadmap.vue'
import Sponsors from './components/Sponsors.vue'
import Author from './components/Author.vue'
import ShotPlaceholder from './components/ShotPlaceholder.vue'
import DemoShelf from './components/DemoShelf.vue'
import VoiceGallery from './components/VoiceGallery.vue'
import VoiceStrip from './components/VoiceStrip.vue'
import VoiceCredit from './components/VoiceCredit.vue'

export default {
  extends: DefaultTheme,
  Layout: () =>
    h(DefaultTheme.Layout, null, {
      // Tên nhóm thanh bên phía trên tiêu đề trang (wireframe WfDocsPage: "Dùng Sano")
      'doc-before': () => h(DocEyebrow),
    }),
  enhanceApp({ app }) {
    app.component('HomeHero', HomeHero)
    app.component('SamplePlayer', SamplePlayer)
    app.component('NormalizeCompare', NormalizeCompare)
    app.component('HowItWorks', HowItWorks)
    app.component('PhoneSection', PhoneSection)
    app.component('Highlights', Highlights)
    app.component('RightsSummary', RightsSummary)
    app.component('Roadmap', Roadmap)
    app.component('Sponsors', Sponsors)
    app.component('Author', Author)
    app.component('ShotPlaceholder', ShotPlaceholder)
    app.component('DemoShelf', DemoShelf)
    app.component('VoiceGallery', VoiceGallery)
    app.component('VoiceStrip', VoiceStrip)
    app.component('VoiceCredit', VoiceCredit)
  },
  setup() {
    // Bấm ảnh trong bài để phóng to (như khuôn tài liệu cũ)
    const route = useRoute()
    const initZoom = () => mediumZoom('.vp-doc img:not(.no-zoom)', { background: 'var(--vp-c-bg)' })
    onMounted(initZoom)
    watch(
      () => route.path,
      () => nextTick(initZoom),
    )
  },
} satisfies Theme
