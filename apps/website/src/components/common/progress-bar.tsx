'use client'

import { ProgressProvider } from '@bprogress/next/app'
import type { ReactNode } from 'react'

interface ProgressBarProps {
  children: ReactNode
}

// A bar along the top of the window while a link's page loads, in a light gray of the text color. Only after 150 ms: a
// prefetched page opens sooner, and the bar would only flash.
export function ProgressBar({ children }: ProgressBarProps) {
  return (
    <ProgressProvider
      color={'color-mix(in oklch, var(--color-text-primary) 40%, transparent)'}
      delay={150}
      options={{ showSpinner: false }}
    >
      {children}
    </ProgressProvider>
  )
}
