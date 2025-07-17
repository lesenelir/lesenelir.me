'use client'

import { useState } from 'react'

import { LayoutSelector } from '@/components/photos/layout-selector'
import { PhotoItem } from '@/components/photos/photo-item'
import type { LayoutMode, Photo } from '@/types'

interface PhotoGalleryProps {
  photos: Photo[]
}

export function PhotoGallery({ photos }: PhotoGalleryProps) {
  const [layoutMode, setLayoutMode] = useState<LayoutMode>('single')

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
      <LayoutSelector layoutMode={layoutMode} setLayoutMode={setLayoutMode} />

      <div className={getContainerClassName()}>
        {photos.map((photo) => (
          <PhotoItem key={photo.id} photo={photo} layoutMode={layoutMode} />
        ))}
      </div>
    </div>
  )
}
