'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'

import { cn } from '@/lib/utils'

import { sidebarData } from '@/constants'

export function Sidebar() {
  const pathname = usePathname()

  return (
    <aside
      className={cn(
        'font-cursive text-text-foreground',
        'z-50 max-sm:absolute max-sm:top-6 max-sm:right-6 sm:fixed sm:top-16 sm:left-16',
        'flex h-auto gap-3 sm:flex-col'
      )}
    >
      {sidebarData.map(item => (
        <Link
          key={item.id}
          href={item.href}
          className={cn(
            'hover:text-text-primary/85',
            (pathname === item.href || (item.href === '/posts' && pathname.startsWith('/posts/'))) &&
              'font-semibold text-text-primary hover:text-text-primary'
          )}
        >
          {item.title}
        </Link>
      ))}
    </aside>
  )
}
