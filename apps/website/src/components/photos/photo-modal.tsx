'use client'

import { useAtomValue, useSetAtom } from 'jotai'
import Image from 'next/image'
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useHotkeys } from 'react-hotkeys-hook'

import { cn } from '@/lib/utils'

import { closePhotoModalAtom, isPhotoModalOpenAtom, navigatePhotoAtom, selectedPhotoAtom } from '@/atoms/photos'

export function PhotoModal() {
  const [mounted, setMounted] = useState(false)
  const isOpen = useAtomValue(isPhotoModalOpenAtom)
  const selectedPhoto = useAtomValue(selectedPhotoAtom)
  const closeModal = useSetAtom(closePhotoModalAtom)
  const navigatePhoto = useSetAtom(navigatePhotoAtom)

  // Closing clears the selected photo at once, so keep showing the last one while the modal fades out
  const [shownPhoto, setShownPhoto] = useState(selectedPhoto)
  if (selectedPhoto && selectedPhoto !== shownPhoto) {
    setShownPhoto(selectedPhoto)
  }

  useEffect(() => {
    setMounted(true)
  }, [])

  // Keyboard navigation
  useHotkeys('esc', closeModal, { enabled: isOpen })
  useHotkeys('arrowleft', () => navigatePhoto('prev'), { enabled: isOpen })
  useHotkeys('arrowright', () => navigatePhoto('next'), { enabled: isOpen })

  if (!mounted || !shownPhoto) return null

  return createPortal(
    <div
      className={cn(
        'fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm',
        'transition-[opacity,visibility] duration-300 ease-out motion-reduce:transition-none',
        // starting: animates the first open too, when the element mounts already open
        isOpen ? 'visible opacity-100 starting:opacity-0' : 'invisible opacity-0'
      )}
    >
      {/* Backdrop: a click anywhere closes the modal, the photo lets clicks through to it */}
      <button type='button' aria-label={'Close'} className={'absolute inset-0'} onClick={closeModal} />

      {/* Photo content */}
      <div
        className={cn(
          'pointer-events-none relative flex size-full items-center justify-center',
          'transition-[opacity,scale] duration-300 ease-out motion-reduce:transition-none',
          isOpen ? 'scale-100 opacity-100 starting:scale-80 starting:opacity-0' : 'scale-80 opacity-0'
        )}
      >
        <Image
          src={shownPhoto.src}
          alt={shownPhoto.alt}
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
