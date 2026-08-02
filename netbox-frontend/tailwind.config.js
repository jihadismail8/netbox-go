/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#00857D',
          dark: '#006B64',
          light: '#4CA8A1',
        },
        secondary: '#6b7280',
        rich: {
          black: '#001423',
          'black-light': '#081B2A',
          'black-lighter': '#0D202E',
        },
        bright: {
          teal: '#00F2D4',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['Roboto Mono', 'monospace'],
      },
    },
  },
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')],
}
