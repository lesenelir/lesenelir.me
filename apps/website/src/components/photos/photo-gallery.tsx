'use client'

import { useAtomValue } from 'jotai'
import type { CSSProperties } from 'react'
import { useCallback, useEffect, useRef, useState } from 'react'

import { cn } from '@/lib/utils'

import { photosAtom } from '@/atoms/photos'

import { LayoutSelector } from '@/components/photos/layout-selector'
import { PhotoItem } from '@/components/photos/photo-item'
import { PhotoModal } from '@/components/photos/photo-modal'

// Matches the widest grid row (lg:grid-cols-3), so any first-row photo that becomes LCP is preloaded
const PRELOAD_COUNT = 3

// How far the row fades out at an edge with more photos past it
const FADE_WIDTH = '64px'

// A scroll button moves the row by this much of its visible width, keeping a little of what was in view
const SCROLL_STEP = 0.8

const scrollButtonClassName = cn(
  // Centered on the photos, which are h-96 in the row layout
  'absolute top-48 z-10 hidden size-9 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full',
  'border border-dividing bg-background/80 text-text-primary backdrop-blur-sm',
  'transition-colors duration-200 hover:bg-background'
)

export function PhotoGallery() {
  const photos = useAtomValue(photosAtom)

  // In the row layout, an edge with more photos past it fades out, as the hint that the row scrolls. Elsewhere
  // nothing overflows, and the fade is only styled for the row layout anyway.
  const scrollerRef = useRef<HTMLDivElement>(null)
  const [overflow, setOverflow] = useState({ start: false, end: false })

  const updateOverflow = useCallback(() => {
    const el = scrollerRef.current
    if (!el) return
    const start = el.scrollLeft > 1
    const end = el.scrollLeft + el.clientWidth < el.scrollWidth - 1
    setOverflow(prev => (prev.start === start && prev.end === end ? prev : { start, end }))
  }, [])

  useEffect(() => {
    const el = scrollerRef.current
    if (!el) return

    // Also fires when the layout switches, since the container changes size
    const observer = new ResizeObserver(updateOverflow)
    observer.observe(el)
    return () => observer.disconnect()
  }, [updateOverflow])

  // A mouse wheel only scrolls up and down, so the row gets buttons for left and right
  const scrollRow = (direction: -1 | 1) => {
    const el = scrollerRef.current
    if (!el) return
    el.scrollBy({ left: direction * el.clientWidth * SCROLL_STEP, behavior: 'smooth' })
  }

  return (
    <div className={'w-full'}>
      <LayoutSelector />

      {/* The buttons sit outside the scrolling row, which the edge fade masks */}
      <div className={'relative'}>
        {/* Each button shows only in the row layout, and only on a side with more photos past it */}
        <button
          type='button'
          aria-label={'Scroll photos left'}
          onClick={() => scrollRow(-1)}
          className={cn(scrollButtonClassName, 'left-2', overflow.start && 'photos-row:flex')}
        >
          <span className={'i-mingcute-left-line size-5'} />
        </button>
        <button
          type='button'
          aria-label={'Scroll photos right'}
          onClick={() => scrollRow(1)}
          className={cn(scrollButtonClassName, 'right-2', overflow.end && 'photos-row:flex')}
        >
          <span className={'i-mingcute-right-line size-5'} />
        </button>

        <div
          ref={scrollerRef}
          onScroll={updateOverflow}
          style={
            {
              '--fade-start': overflow.start ? FADE_WIDTH : '0px',
              '--fade-end': overflow.end ? FADE_WIDTH : '0px',
            } as CSSProperties
          }
          className={cn(
            'photos-single:flex photos-single:flex-col photos-single:items-center photos-single:space-y-8',
            'photos-row:flex photos-row:gap-6 photos-row:overflow-x-auto photos-row:scroll-smooth photos-row:pb-4',
            'photos-row:[mask-image:linear-gradient(to_right,transparent,#000_var(--fade-start),#000_calc(100%-var(--fade-end)),transparent)]',
            'photos-grid:grid photos-grid:grid-cols-1 photos-grid:gap-6 photos-grid:sm:grid-cols-2 photos-grid:lg:grid-cols-3'
          )}
        >
          {photos.map((photo, index) => (
            <PhotoItem key={photo.id} photo={photo} preload={index < PRELOAD_COUNT} />
          ))}
        </div>
      </div>

      <PhotoModal />
    </div>
  )
}
