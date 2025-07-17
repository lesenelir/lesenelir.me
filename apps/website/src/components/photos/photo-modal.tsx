import { useEffect } from 'react'

import { useAtomValue, useSetAtom } from 'jotai'
import Image from 'next/image'
import { createPortal } from 'react-dom'

import { closePhotoModalAtom, isPhotoModalOpenAtom, selectedPhotoAtom } from '@/atoms/photos'

export function PhotoModal() {
  const isOpen = useAtomValue(isPhotoModalOpenAtom)
  const selectedPhoto = useAtomValue(selectedPhotoAtom)
  const closeModal = useSetAtom(closePhotoModalAtom)

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        closeModal()
      }
    }

    if (isOpen) {
      document.addEventListener('keydown', handleKeyDown)
      return () => document.removeEventListener('keydown', handleKeyDown)
    }
  }, [isOpen, closeModal])

  if (!isOpen || !selectedPhoto) return null

  return createPortal(
    <div
      onClick={closeModal}
      className={'fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm'}
    >
      {/* Photo content */}
      <div className={'relative flex size-full items-center justify-center'}>
        <Image
          src={selectedPhoto.src}
          alt={selectedPhoto.alt}
          width={1200}
          height={800}
          className={'h-auto max-h-full w-full max-w-full object-contain lg:h-full lg:w-auto'}
          sizes={'(max-width: 1024px) 100vw, 100vh'}
          priority
        />
      </div>
    </div>,
    document.body
  )
}
