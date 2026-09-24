import type { Config } from 'tailwindcss'
import animate from 'tailwindcss-animate'
import { theme } from './tailwind.theme'

// Wireframe chỉ xem ở bản dev → bản build không sinh class riêng của wireframe.
const prod = process.env.NODE_ENV === 'production'

export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{vue,ts}', ...(prod ? ['!./src/wireframes/**'] : [])],
  theme,
  plugins: [animate],
} satisfies Config
