import { useAtomValue, useSetAtom } from 'jotai'
import Image from 'next/image'

import { layoutModeAtom, openPhotoModalAtom } from '@/atoms/photos'
import { cn } from '@/lib/utils'
import type { Photo } from '@/types'

interface PhotoItemProps {
  photo: Photo
}

export function PhotoItem({ photo }: PhotoItemProps) {
  const layoutMode = useAtomValue(layoutModeAtom)
  const openModal = useSetAtom(openPhotoModalAtom)

  const getImageClassName = () => {
    switch (layoutMode) {
      case 'single':
        return 'w-full max-w-2xl mx-auto object-contain'
      case 'row':
        return 'h-96 w-auto max-w-md shrink-0 object-contain'
      case 'grid':
        return 'w-full h-96 object-cover'
    }
  }

  return (
    <div
      className={cn(
        'cursor-pointer',
        layoutMode === 'row' && 'shrink-0',
        layoutMode === 'grid' && 'relative'
      )}
    >
      <Image
        src={photo.src}
        alt={photo.alt}
        width={layoutMode === 'single' ? 800 : layoutMode === 'row' ? 600 : 400}
        height={layoutMode === 'single' ? 600 : layoutMode === 'row' ? 384 : 288}
        className={getImageClassName()}
        onClick={() => openModal(photo)}
        sizes={
          layoutMode === 'single'
            ? '(max-width: 640px) 100vw, (max-width: 768px) 672px, 672px'
            : layoutMode === 'row'
              ? '(max-width: 640px) 448px, 600px'
              : '(max-width: 640px) 100vw, (max-width: 768px) 50vw, 33vw'
        }
        priority={false}
        loading={'lazy'}
      />
    </div>
  )
}
