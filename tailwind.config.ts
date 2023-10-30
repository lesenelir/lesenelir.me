import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}"
  ],
  darkMode: ['class', '[data-theme="dark"]'],
  safelist: [],
  theme: {
    extend: {
      fontFamily: {
        'comic': ['Comic Sans MS', 'sans-serif'],
        'icon': ['iconfont', 'sans-serif']
      },
      colors: {
        // background color
        bgDark: '#1D1F24',
        // navbar right side and footer RSS icon
        navbarLight: '#374151',
        navbarDark: '#E5E7EB',
        // main text color
        textLight: '#555555',
        textDark: '#BBBBBB',
        // border line
        borderUnderline: '#7D7D7D'
      }
    }
  },
  plugins: []
}

export default config
