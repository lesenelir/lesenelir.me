import { ThemeProvider } from 'next-themes'
import type { ReactNode } from 'react'

import { ProgressBar } from '@/components/common/progress-bar'
import { SongsProvider } from '@/components/songs/songs-provider'

interface ProvidersProps {
  children: ReactNode
}

export function Providers({ children }: ProvidersProps) {
  // Without transitions while the theme switches, so colors change at once instead of element by element
  return (
    <ThemeProvider disableTransitionOnChange>
      <ProgressBar>
        <SongsProvider>{children}</SongsProvider>
      </ProgressBar>
    </ThemeProvider>
  )
}
