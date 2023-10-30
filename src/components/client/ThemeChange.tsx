'use client'

import { useTheme } from 'next-themes'

export default function ThemeChange() {
  const { theme, setTheme } = useTheme()

  const handlerTheme = () => {
    if (theme === 'light') {
      setTheme('dark')
    } else {
      setTheme('light')
    }
  }

  return (
    <div onClick={handlerTheme}>
      Theme
    </div>
  )
}
