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
        void: '#050508',
        surface: {
          DEFAULT: '#0c0c12',
          elevated: '#12121c',
          border: 'rgba(148, 163, 184, 0.12)',
        },
        foreground: '#f1f5f9',
        muted: '#94a3b8',
        accent: {
          DEFAULT: '#22d3ee',
          hover: '#67e8f9',
          muted: 'rgba(34, 211, 238, 0.15)',
        },
        danger: '#f87171',
        success: '#34d399',
        warning: '#fbbf24',
      },
      fontFamily: {
        sans: ['Outfit', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'ui-monospace', 'monospace'],
      },
      boxShadow: {
        glass: '0 8px 32px rgba(0, 0, 0, 0.4)',
        'glow-sm': '0 0 20px rgba(34, 211, 238, 0.15)',
        'glow-md': '0 0 40px rgba(34, 211, 238, 0.2)',
      },
      backgroundImage: {
        grid: `linear-gradient(rgba(34, 211, 238, 0.03) 1px, transparent 1px),
          linear-gradient(90deg, rgba(34, 211, 238, 0.03) 1px, transparent 1px)`,
      },
      backgroundSize: {
        grid: '48px 48px',
      },
    },
  },
  plugins: [flowbite],
}
