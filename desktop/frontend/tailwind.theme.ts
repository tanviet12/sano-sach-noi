import type { Config } from 'tailwindcss'

// Theme của phần mềm (dùng trong tailwind.config.ts). Chỉ khai báo theme,
// plugin khai báo ở tailwind.config.ts.
// Đổi primary accent? Update HSL trong src/style.css (--primary + --ring).

export const theme: Config['theme'] = {
  container: {
    center: true,
    padding: '1rem',
    screens: { '2xl': '1400px' },
  },
  extend: {
    colors: {
      border: 'hsl(var(--border))',
      input: 'hsl(var(--input))',
      ring: 'hsl(var(--ring))',
      background: 'hsl(var(--background))',
      foreground: 'hsl(var(--foreground))',
      primary: {
        DEFAULT: 'hsl(var(--primary))',
        foreground: 'hsl(var(--primary-foreground))',
      },
      secondary: {
        DEFAULT: 'hsl(var(--secondary))',
        foreground: 'hsl(var(--secondary-foreground))',
      },
      destructive: {
        DEFAULT: 'hsl(var(--destructive))',
        foreground: 'hsl(var(--destructive-foreground))',
      },
      muted: {
        DEFAULT: 'hsl(var(--muted))',
        foreground: 'hsl(var(--muted-foreground))',
      },
      accent: {
        DEFAULT: 'hsl(var(--accent))',
        foreground: 'hsl(var(--accent-foreground))',
      },
      popover: {
        DEFAULT: 'hsl(var(--popover))',
        foreground: 'hsl(var(--popover-foreground))',
      },
      card: {
        DEFAULT: 'hsl(var(--card))',
        foreground: 'hsl(var(--card-foreground))',
      },
      // RAG semantic — nếu project không cần status RAG, có thể xoá block này.
      rag: {
        green: 'hsl(142 71% 45%)',
        amber: 'hsl(38 92% 50%)',
        red: 'hsl(0 72% 51%)',
      },
    },
    borderRadius: {
      lg: 'var(--radius)',
      md: 'calc(var(--radius) - 2px)',
      sm: 'calc(var(--radius) - 4px)',
    },
    fontFamily: {
      sans: ['Inter', 'IBM Plex Sans', 'system-ui', 'sans-serif'],
    },
    keyframes: {
      'accordion-down': {
        from: { height: '0' },
        to: { height: 'var(--reka-accordion-content-height)' },
      },
      'accordion-up': {
        from: { height: 'var(--reka-accordion-content-height)' },
        to: { height: '0' },
      },
    },
    animation: {
      'accordion-down': 'accordion-down 0.2s ease-out',
      'accordion-up': 'accordion-up 0.2s ease-out',
    },
  },
}
