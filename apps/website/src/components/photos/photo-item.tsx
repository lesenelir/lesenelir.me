import { useSetAtom } from 'jotai'
import Image from 'next/image'

import type { TPhoto } from '@/types'

import { cn } from '@/lib/utils'

import { openPhotoModalAtom } from '@/atoms/photos'

interface PhotoItemProps {
  photo: TPhoto
  // Preload the LCP image; next/image lazy-loads the rest by default
  preload?: boolean
}

export function PhotoItem({ photo, preload = false }: PhotoItemProps) {
  const openModal = useSetAtom(openPhotoModalAtom)

  return (
    <div className={'cursor-pointer photos-row:shrink-0 photos-grid:relative'}>
      <Image
        src={photo.src}
        alt={photo.alt}
        width={800}
        height={600}
        className={cn(
          'photos-single:mx-auto photos-single:w-full photos-single:max-w-2xl photos-single:object-contain',
          'photos-row:h-96 photos-row:w-auto photos-row:max-w-md photos-row:shrink-0 photos-row:object-contain',
          'photos-grid:h-96 photos-grid:w-full photos-grid:object-cover'
        )}
        onClick={() => openModal(photo)}
        // One value for every layout, since the layout is only known to CSS: the widest a photo shows is 672px
        // (single column); grid and row photos are narrower and fetch a slightly larger file than they need
        sizes={'(max-width: 640px) 100vw, 672px'}
        preload={preload}
      />
      {photo.desc && (
        <p
          className={cn(
            'mt-2 text-center text-sm text-text-primary/85',
            'photos-single:mx-auto photos-single:max-w-2xl photos-row:max-w-md photos-grid:max-w-full'
          )}
        >
          {photo.desc}
        </p>
      )}
    </div>
  )
}
