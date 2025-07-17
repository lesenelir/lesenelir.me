import Image from 'next/image'

import { cn } from '@/lib/utils'
import type { LayoutMode, Photo } from '@/types'

interface PhotoItemProps {
  photo: Photo
  layoutMode: LayoutMode
}

export function PhotoItem({ photo, layoutMode }: PhotoItemProps) {
  const getImageProps = () => {
    switch (layoutMode) {
      case 'grid':
        return {
          width: 400,
          height: 300,
          className: 'w-full h-[300px] object-cover rounded-lg shadow-lg'
        }
      case 'single':
        return {
          width: photo.width,
          height: photo.height,
          className: 'w-full max-w-2xl mx-auto rounded-lg shadow-lg'
        }
      case 'row': {
        const aspectRatio = photo.width / photo.height
        const rowHeight = 300
        return {
          width: Math.round(rowHeight * aspectRatio),
          height: rowHeight,
          className: 'h-[300px] object-cover rounded-lg shadow-lg shrink-0'
        }
      }
      default:
        return {
          width: photo.width,
          height: photo.height,
          className: 'w-full rounded-lg shadow-lg'
        }
    }
  }

  const imageProps = getImageProps()

  return (
    <div
      className={cn(
        'overflow-hidden rounded-lg shadow-lg',
        layoutMode === 'single' && 'mb-8',
        layoutMode === 'row' && 'shrink-0'
      )}
    >
      <Image
        src={photo.src}
        alt={photo.alt}
        width={imageProps.width}
        height={imageProps.height}
        className={imageProps.className}
        sizes={
          layoutMode === 'single'
            ? '(max-width: 768px) 100vw, 672px'
            : layoutMode === 'row'
              ? '300px'
              : '(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw'
        }
      />
    </div>
  )
}
