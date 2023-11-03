const config = {
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}"
  ],
  darkMode: ['class', '[data-theme="dark"]'],
  safelist: [
    {
      pattern: /hljs+/,
    }
  ],
  theme: {
    hljs: {
      theme: 'nord'
    },
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
              color: theme('colors.navbarLight'),
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
                color: '#000',
              }
            },
            pre: {
              backgroundColor: theme('colors.gray.50'),
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
              color: theme('colors.navbarDark'),
              textDecorationColor: theme('colors.navbarDark'),

              transitionProperty: theme('transitionProperty.transition'),
              transitionDuration: theme('transitionDuration.100'),
              transitionTimingFunction: theme('transitionTimingFunction.ease-in'),

              '&:hover': {
                color: '#ffffff',
              }
            },
            pre: {
              backgroundColor: theme('colors.bgDark'),
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
    require('tailwind-highlightjs'),
  ]
}

export default config
