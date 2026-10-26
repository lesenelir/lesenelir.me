'use client'

import { usePathname } from 'next/navigation'
import type { ReactNode } from 'react'
import { useLayoutEffect, useRef, useState } from 'react'

import { cn } from '@/lib/utils'

interface PageTransitionProps {
  children: ReactNode
}

// Enter animation only: the new page renders at once instead of waiting for the old one to animate out. Not on the
// first load, and replayed on each navigation rather than by keying on the path: a key would remount every page, and
// the pages Next keeps hidden (cacheComponents) would lose their state.
export function PageTransition({ children }: PageTransitionProps) {
  const pathname = usePathname()
  const [initialPathname] = useState(pathname)
  const [hasNavigated, setHasNavigated] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const shownPathname = useRef(pathname)

  if (!hasNavigated && pathname !== initialPathname) {
    setHasNavigated(true)
  }

  useLayoutEffect(() => {
    const element = ref.current
    if (!element || pathname === shownPathname.current) return
    shownPathname.current = pathname
    // A finished animation is gone from getAnimations(), so it restarts by being taken off for a style recalc
    element.style.animation = 'none'
    void element.offsetWidth
    element.style.animation = ''
  }, [pathname])

  return (
    <div ref={ref} className={cn(hasNavigated && 'motion-safe:animate-page-in')}>
      {children}
    </div>
  )
}
