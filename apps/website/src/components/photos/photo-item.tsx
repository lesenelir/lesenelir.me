import Image from 'next/image'

import type { LayoutMode, Photo } from '@/types'

interface PhotoItemProps {
  photo: Photo
  layoutMode: LayoutMode
}

export function PhotoItem({ photo, layoutMode }: PhotoItemProps) {
  const getImageClassName = () => {
    switch (layoutMode) {
      case 'single':
        return 'w-full max-w-2xl mx-auto object-contain'
      case 'row':
        return 'h-80 w-auto max-w-sm shrink-0 object-contain'
      case 'grid':
        return 'size-full object-cover'
      default:
        return 'size-full object-cover'
    }
  }

  const getContainerClassName = () => {
    switch (layoutMode) {
      case 'single':
        return 'flex justify-center bg-red-100'
      case 'row':
        return 'shrink-0'
      case 'grid':
        return 'relative aspect-square bg-red-100'
      default:
        return 'relative aspect-square'
    }
  }

  return (
    <div className={getContainerClassName()}>
      <Image
        src={photo.src}
        alt={photo.alt}
        width={layoutMode === 'single' ? 800 : layoutMode === 'row' ? 480 : 400}
        height={layoutMode === 'single' ? 600 : layoutMode === 'row' ? 320 : 400}
        className={getImageClassName()}
        sizes={
          layoutMode === 'single'
            ? '(max-width: 640px) 100vw, (max-width: 768px) 672px, 672px'
            : layoutMode === 'row'
              ? '(max-width: 640px) 384px, 480px'
              : '(max-width: 640px) 100vw, (max-width: 768px) 50vw, 33vw'
        }
        priority={false}
        loading={'lazy'}
      />
    </div>
  )
}
