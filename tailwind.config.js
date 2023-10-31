const config = {
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
        textLight: '#313233',
        textDark: '#BBBBBB',
        // border line
        borderUnderline: '#7D7D7D'
      },
      typography: ({ theme }) => ({
        DEFAULT: {
          css: {
            a: {
              color: '#000000',
              fontSize: theme('fontSize.base'),
              fontWeight: 'normal',
              textDecoration: 'underline',
              textDecorationColor: theme('colors.navbarLight'),
              textDecorationThickness: '1px',
              textUnderlineOffset: '2px',

              transitionProperty: theme('transitionProperty.transition'),
              transitionDuration: theme('transitionDuration.100'),
              transitionTimingFunction: theme('transitionTimingFunction.ease-in'),

              '&:hover': {
                color: theme('colors.navbarLight'),
                opacity: 0.7,
              }
            }
          }
        },
        invert: {
          css: {
            a: {
              color: '#ffffff',
              textDecorationColor: theme('colors.navbarDark'),

              transitionProperty: theme('transitionProperty.transition'),
              transitionDuration: theme('transitionDuration.100'),
              transitionTimingFunction: theme('transitionTimingFunction.ease-in'),

              '&:hover': {
                color: theme('colors.navbarDark'),
                opacity: 0.7,
              }
            }
          }
        }
      })
    }
  },
  plugins: [require('@tailwindcss/typography')]
}

export default config
