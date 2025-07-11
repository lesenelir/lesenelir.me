'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'

import { sidebarData } from '@/constants'
import { cn } from '@/lib/utils'

export function Sidebar() {
  const pathname = usePathname()

  return (
    <aside
      className={cn(
        'text-text-foreground font-cursive',
        'fixed z-50 max-sm:right-6 max-sm:top-6 sm:left-16 sm:top-16',
        'flex h-auto gap-3 sm:flex-col'
      )}
    >
      {sidebarData.map((item) => (
        <Link
          key={item.id}
          href={item.href}
          className={cn(
            'hover:text-text-primary/85',
            (pathname === item.href ||
              (item.href === '/posts' && pathname.startsWith('/posts/'))) &&
              'text-text-primary hover:text-text-primary font-semibold'
          )}
        >
          {item.title}
        </Link>
      ))}
    </aside>
  )
}
