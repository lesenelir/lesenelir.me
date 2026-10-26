'use client'

import { useEffect, useRef, useState } from 'react'

import type { TPostHeading } from '@/types'

import { cn } from '@/lib/utils'

// A section counts as current once its heading is this far from the top of the window or higher
const ACTIVE_OFFSET = 120

interface TableOfContentsProps {
  headings: TPostHeading[]
  className?: string
}

// The post's sections, with the one being read marked as it scrolls by
export function TableOfContents({ headings, className }: TableOfContentsProps) {
  const [activeId, setActiveId] = useState<string | null>(null)
  const navRef = useRef<HTMLElement>(null)

  useEffect(() => {
    // The headings of this post: a post left behind stays in the document, hidden, and may share their ids
    const article = navRef.current?.closest('article')
    let frame = 0
    const update = () => {
      cancelAnimationFrame(frame)
      frame = requestAnimationFrame(() => {
        let current: string | null = null
        for (const heading of headings) {
          const el = article?.querySelector(`#${CSS.escape(heading.id)}`)
          if (el && el.getBoundingClientRect().top <= ACTIVE_OFFSET) current = heading.id
        }
        setActiveId(current)
      })
    }

    update()
    window.addEventListener('scroll', update, { passive: true })
    return () => {
      window.removeEventListener('scroll', update)
      cancelAnimationFrame(frame)
    }
  }, [headings])

  return (
    <nav ref={navRef} aria-label={'Table of contents'} className={className}>
      <ul lang={'zh-CN'} className={'space-y-2 text-sm'}>
        {headings.map(heading => (
          <li key={heading.id} className={cn(heading.level === 3 && 'pl-3')}>
            <a
              href={`#${heading.id}`}
              aria-current={heading.id === activeId ? 'location' : undefined}
              className={cn(
                'block text-text-foreground transition-colors duration-200 hover:text-text-primary',
                heading.id === activeId && 'text-text-primary'
              )}
            >
              {heading.text}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  )
}
