/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  darkMode: 'media',
  theme: {
    extend: {
      colors: {
        accent: 'var(--accent)',
      },
      fontFamily: {
        sans: ['var(--sans)'],
        heading: ['var(--heading)'],
        mono: ['var(--mono)'],
      },
    },
  },
  plugins: [],
}