import { ThemeProvider } from 'next-themes'
import type { ReactNode } from 'react'

interface ProvidersProps {
  children: ReactNode
}

export function Providers({ children }: ProvidersProps) {
  // Without transitions while the theme switches, so colors change at once instead of element by element
  return <ThemeProvider disableTransitionOnChange>{children}</ThemeProvider>
}
