'use client'

import { useAtomValue } from 'jotai'

import { layoutModeAtom, photosAtom } from '@/atoms/photos'
import { LayoutSelector } from '@/components/photos/layout-selector'
import { PhotoItem } from '@/components/photos/photo-item'
import { PhotoModal } from '@/components/photos/photo-modal'

export function PhotoGallery() {
  const layoutMode = useAtomValue(layoutModeAtom)
  const photos = useAtomValue(photosAtom)

  const getContainerClassName = () => {
    switch (layoutMode) {
      case 'grid':
        return 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6'
      case 'row':
        return 'flex gap-6 overflow-x-auto scroll-smooth pb-4 scrollbar-thin scrollbar-thumb-text-foreground/20 scrollbar-track-transparent'
      case 'single':
        return 'flex flex-col items-center space-y-8'
      default:
        return 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6'
    }
  }

  return (
    <div className={'w-full'}>
      <LayoutSelector />

      <div className={getContainerClassName()}>
        {photos.map((photo) => (
          <PhotoItem key={photo.id} photo={photo} />
        ))}
      </div>

      <PhotoModal />
    </div>
  )
}
