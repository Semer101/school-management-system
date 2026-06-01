import flowbite from 'flowbite/plugin'

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
    './node_modules/flowbite-react/lib/esm/**/*.js',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        void: 'var(--color-void)',
        surface: {
          DEFAULT: 'var(--color-surface)',
          elevated: 'var(--color-surface-elevated)',
          border: 'var(--color-surface-border)',
        },
        foreground: 'var(--color-foreground)',
        muted: 'var(--color-muted)',
        accent: {
          DEFAULT: 'var(--color-accent)',
          hover: 'var(--color-accent-hover)',
          muted: 'var(--color-accent-muted)',
        },
        danger: '#ef4444',
        success: '#10b981',
        warning: '#f59e0b',
        kpi: {
          students: '#3b82f6',
          teachers: '#8b5cf6',
          classes: '#f59e0b',
          subjects: '#10b981',
          attendance: '#06b6d4',
          finance: '#ec4899',
        },
      },
      fontFamily: {
        sans: ['Outfit', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'ui-monospace', 'monospace'],
      },
      boxShadow: {
        glass: 'var(--shadow-glass)',
        'glow-sm': 'var(--shadow-glow-sm)',
        'glow-md': 'var(--shadow-glow-md)',
      },
    },
  },
  plugins: [flowbite],
}
