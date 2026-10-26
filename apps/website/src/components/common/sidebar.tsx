'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'

import { cn } from '@/lib/utils'

import { NowPlaying } from '@/components/songs/now-playing'

import { sidebarData } from '@/constants'

export function Sidebar() {
  const pathname = usePathname()

  return (
    <nav
      aria-label={'Main'}
      className={cn(
        'font-cursive text-text-foreground',
        'z-50 max-sm:absolute max-sm:top-6 max-sm:right-6 sm:fixed sm:top-16 sm:left-16',
        'flex h-auto gap-3 sm:flex-col'
      )}
    >
      {sidebarData.map(item => {
        // A section stays active on its sub-pages, e.g. posts on /posts/jotai
        const isActive = pathname === item.href || (item.href !== '/' && pathname.startsWith(`${item.href}/`))

        return (
          <div key={item.id} className={'flex items-center gap-1.5'}>
            <Link
              href={item.href}
              aria-current={isActive ? 'page' : undefined}
              className={cn(
                'hover:text-text-primary/85',
                isActive && 'font-semibold text-text-primary hover:text-text-primary'
              )}
            >
              {item.title}
            </Link>
            {/* The player lives in the root layout, so this shows and controls the song on every page */}
            {item.href === '/songs' && <NowPlaying />}
          </div>
        )
      })}
    </nav>
  )
}
