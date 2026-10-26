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

  return (
    <div className={'w-full'}>
      <LayoutSelector />

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

      <PhotoModal />
    </div>
  )
}
