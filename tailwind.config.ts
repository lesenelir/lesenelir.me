import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  darkMode: 'class',
  safelist: [],
  theme: {
    extend: {
      fontFamily: {
        'comic': ['Comic Sans MS', 'sans-serif'],
        'icon': ['iconfont', 'sans-serif'],
      },
      colors: {
        // navbar right side
        navbarLight: '#374151',
        navbarDark: '#E5E7EB',
        // main text color
        textLight: '#555555',
        textDark: '#BBBBBB',
        // border line
        borderUnderline: '#7D7D7D',
      }
    }
  },
  plugins: [],
}

export default config
