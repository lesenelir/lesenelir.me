import { useSetAtom } from 'jotai'
import Image from 'next/image'
import { useRef } from 'react'

import type { TPhoto } from '@/types'

import { preloadModalImage } from '@/lib/photo-modal-image'
import { cn } from '@/lib/utils'

import { openPhotoModalAtom } from '@/atoms/photos'

interface PhotoItemProps {
  photo: TPhoto
  // Preload the LCP image; next/image lazy-loads the rest by default
  preload?: boolean
}

// Long enough to skip photos the pointer only passes over, e.g. while scrolling
const HOVER_PRELOAD_DELAY = 100

export function PhotoItem({ photo, preload = false }: PhotoItemProps) {
  const openModal = useSetAtom(openPhotoModalAtom)
  const hoverTimer = useRef<ReturnType<typeof setTimeout>>(undefined)

  return (
    // Full width in the single column: the column centers its items, which would shrink this to the photo, and a
    // photo that hasn't loaded yet would then be only as wide as its alt text
    <div className={'cursor-pointer photos-single:w-full photos-row:shrink-0 photos-grid:relative'}>
      <Image
        src={photo.src}
        alt={photo.alt}
        // The real dimensions, so the photo holds its place before it loads
        width={photo.width}
        height={photo.height}
        className={cn(
          'photos-single:mx-auto photos-single:w-full photos-single:max-w-2xl photos-single:object-contain',
          'photos-row:h-96 photos-row:w-auto photos-row:max-w-md photos-row:shrink-0 photos-row:object-contain',
          'photos-grid:h-96 photos-grid:w-full photos-grid:object-cover'
        )}
        onClick={() => openModal(photo)}
        // Fetch the modal's larger image once the pointer rests on the photo or presses it, so it is ready, or
        // well under way, when the modal opens
        onPointerEnter={() => {
          hoverTimer.current = setTimeout(() => preloadModalImage(photo), HOVER_PRELOAD_DELAY)
        }}
        onPointerLeave={() => clearTimeout(hoverTimer.current)}
        onPointerDown={() => preloadModalImage(photo)}
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
