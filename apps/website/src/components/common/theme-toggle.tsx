'use client'

import { useTheme } from 'next-themes'
import { useEffect } from 'react'

import { setThemeChoice } from '@/lib/theme-choice'
import { cn } from '@/lib/utils'

const nextTheme: Record<string, string> = { system: 'light', light: 'dark', dark: 'system' }

// Cycles the theme through system, light and dark. Under the sidebar on wider screens; in the top-left corner on
// phones, across from the sidebar's links.
export function ThemeToggle() {
  const { theme, setTheme } = useTheme()

  // Also follows a theme changed in another tab, which next-themes picks up
  useEffect(() => {
    setThemeChoice(theme)
  }, [theme])

  // The icon shown is picked by CSS from <html data-theme-choice>, so it is right from the first paint. Hidden icons
  // are left out of the button's accessible name, so it names the theme in use.
  const iconClass = 'hidden size-4.5'

  return (
    // /80 is the lightest this icon can be and still stand out 3:1 from the background, in both themes
    <button
      type='button'
      onClick={() => setTheme(nextTheme[theme ?? 'system'] ?? 'system')}
      title={'Switch theme'}
      className={cn(
        'z-50 flex size-6 cursor-pointer items-center justify-center text-text-foreground/80 transition-colors duration-200 hover:text-text-primary',
        'max-sm:absolute max-sm:top-6 max-sm:left-6 sm:fixed sm:bottom-16 sm:left-16'
      )}
    >
      <span
        role='img'
        aria-label={'Theme: system'}
        className={cn('i-mingcute-computer-line theme-system:block', iconClass)}
      />
      <span role='img' aria-label={'Theme: light'} className={cn('i-mingcute-sun-line theme-light:block', iconClass)} />
      <span role='img' aria-label={'Theme: dark'} className={cn('i-mingcute-moon-line theme-dark:block', iconClass)} />
    </button>
  )
}
