'use client'

import { usePathname } from 'next/navigation'
import type { ReactNode } from 'react'
import { useState } from 'react'

import { cn } from '@/lib/utils'

interface PageTransitionProps {
  children: ReactNode
}

// Enter animation only: the new page renders at once instead of waiting for the old one to animate out.
// Keying on the path replays the CSS animation on every navigation, but not on the first load.
export function PageTransition({ children }: PageTransitionProps) {
  const pathname = usePathname()
  const [initialPathname] = useState(pathname)
  const [hasNavigated, setHasNavigated] = useState(false)

  if (!hasNavigated && pathname !== initialPathname) {
    setHasNavigated(true)
  }

  return (
    <div key={pathname} className={cn(hasNavigated && 'motion-safe:animate-page-in')}>
      {children}
    </div>
  )
}
