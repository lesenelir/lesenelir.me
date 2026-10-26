'use client'

import { useTheme } from 'next-themes'
import { useEffect, useState } from 'react'

import { cn } from '@/lib/utils'

const themes = [
  { value: 'system', label: 'System', icon: 'i-mingcute-computer-line' },
  { value: 'light', label: 'Light', icon: 'i-mingcute-sun-line' },
  { value: 'dark', label: 'Dark', icon: 'i-mingcute-moon-line' },
] as const

// Cycles the theme through system, light and dark. Under the sidebar on wider screens; in the top-left corner on
// phones, across from the sidebar's links.
export function ThemeToggle() {
  const { theme, setTheme } = useTheme()

  // The chosen theme is only known in the browser, so the server renders an empty button of the same size
  const [mounted, setMounted] = useState(false)
  useEffect(() => {
    setMounted(true)
  }, [])

  const index = Math.max(
    0,
    themes.findIndex(t => t.value === theme)
  )
  const current = themes[index]
  const next = themes[(index + 1) % themes.length]

  return (
    <button
      type='button'
      onClick={() => setTheme(next.value)}
      aria-label={mounted ? `Theme: ${current.label}. Switch to ${next.label}` : 'Theme'}
      title={mounted ? `Theme: ${current.label}` : undefined}
      className={cn(
        'z-50 flex size-6 cursor-pointer items-center justify-center text-text-foreground transition-colors duration-200 hover:text-text-primary',
        'max-sm:absolute max-sm:top-6 max-sm:left-6 sm:fixed sm:bottom-16 sm:left-16'
      )}
    >
      {mounted && <span className={cn(current.icon, 'size-4.5')} />}
    </button>
  )
}
