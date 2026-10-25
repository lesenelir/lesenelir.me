'use client'

import { useAtomValue } from 'jotai'

import { cn } from '@/lib/utils'

import { photosAtom } from '@/atoms/photos'

import { LayoutSelector } from '@/components/photos/layout-selector'
import { PhotoItem } from '@/components/photos/photo-item'
import { PhotoModal } from '@/components/photos/photo-modal'

// Matches the widest grid row (lg:grid-cols-3), so any first-row photo that becomes LCP is preloaded
const PRELOAD_COUNT = 3

export function PhotoGallery() {
  const photos = useAtomValue(photosAtom)

  return (
    <div className={'w-full'}>
      <LayoutSelector />

      <div
        className={cn(
          'photos-single:flex photos-single:flex-col photos-single:items-center photos-single:space-y-8',
          'photos-row:flex photos-row:gap-6 photos-row:overflow-x-auto photos-row:scroll-smooth photos-row:pb-4',
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
