'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'

import { sidebarData } from '@/constants'
import { cn } from '@/lib/utils'

export function Sidebar() {
  const pathname = usePathname()

  return (
    <aside className={'text-text-foreground font-cursive mr-16 flex flex-col gap-3'}>
      {sidebarData.map((item) => (
        <Link
          key={item.id}
          href={item.href}
          className={cn(
            'hover:text-text-primary/85',
            pathname === item.href && 'text-text-primary hover:text-text-primary font-semibold'
          )}
        >
          {item.title}
        </Link>
      ))}
    </aside>
  )
}
