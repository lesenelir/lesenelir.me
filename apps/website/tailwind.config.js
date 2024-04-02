const config = {
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}"
  ],
  darkMode: ['class', '[data-theme="dark"]'],
  theme: {
    extend: {
      fontFamily: {
        'comic': ['Comic Sans MS', 'sans-serif'],
        'icon': ['iconfont', 'sans-serif']
      },
      colors: {
        // background color
        bgDark: '#030712',
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
              color: theme('colors.black'),
              fontSize: theme('fontSize.base'),
              // fontWeight: 'normal',
              textDecoration: 'underline',
              textDecorationColor: theme('colors.gray.300'),
              textDecorationThickness: '1px',
              textUnderlineOffset: '4px',

              transitionProperty: theme('transitionProperty.transition'),
              transitionDuration: theme('transitionDuration.300'),
              transitionTimingFunction: theme('transitionTimingFunction.ease-in-out'),

              '&:hover': {
                textDecorationColor: '#000',
              }
            },
            pre: {
              backgroundColor: `${theme('colors.gray.100')} !important`,
            },
            '.hljs-params, .hljs-variable': {
              color: `${theme('colors.gray.700')} !important`,
            },
            'pre code': {
              borderRadius: '4px',
              color: `${theme('colors.gray.700')} !important`,
              backgroundColor: `${theme('colors.gray.100')} !important`,
            }
          }
        },
        invert: {
          css: {
            a: {
              color: theme('colors.white'),
              textDecorationColor: theme('colors.zinc.600'),
              textUnderlineOffset: '4px',

              transitionProperty: theme('transitionProperty.transition'),
              transitionDuration: theme('transitionDuration.300'),
              transitionTimingFunction: theme('transitionTimingFunction.ease-in-out'),

              '&:hover': {
                textDecorationColor: '#ffffff',
              }
            },
            pre: {
              backgroundColor: `${theme('colors.bgDark')} !important`,
            },
            '.hljs-params, .hljs-variable': {
              color: `${theme('colors.gray.200')} !important`,
            },
            'pre code': {
              borderRadius: '4px',
              color: `${theme('colors.gray.200')} !important`,
              backgroundColor: `${theme('colors.bgDark')} !important`,
            }
          }
        }
      })
    }
  },
  plugins: [
    require('@tailwindcss/typography'),
  ]
}

export default config
