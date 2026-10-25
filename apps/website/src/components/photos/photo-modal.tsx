'use client'

import { useAtomValue, useSetAtom } from 'jotai'
import Image from 'next/image'
import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useHotkeys } from 'react-hotkeys-hook'

import { getModalImageProps, preloadModalImage } from '@/lib/photo-modal-image'
import { cn } from '@/lib/utils'

import {
  closePhotoModalAtom,
  isPhotoModalOpenAtom,
  navigatePhotoAtom,
  photosAtom,
  selectedPhotoAtom,
} from '@/atoms/photos'

export function PhotoModal() {
  const [mounted, setMounted] = useState(false)
  const photos = useAtomValue(photosAtom)
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

  // Fetch the neighbors of the open photo, so the arrow keys show them without waiting
  useEffect(() => {
    if (!isOpen || !selectedPhoto) return

    const index = photos.findIndex(photo => photo.id === selectedPhoto.id)
    if (index === -1) return

    preloadModalImage(photos[(index + 1) % photos.length])
    preloadModalImage(photos[(index - 1 + photos.length) % photos.length])
  }, [isOpen, selectedPhoto, photos])

  // While open, focus moves into the modal and the page behind it is inert, out of reach of Tab and screen readers;
  // closing gives focus back to the photo that opened it
  const closeButtonRef = useRef<HTMLButtonElement>(null)
  useEffect(() => {
    if (!isOpen) return

    const opener = document.activeElement instanceof HTMLElement ? document.activeElement : null
    const modal = closeButtonRef.current?.closest('[role="dialog"]')
    const background = [...document.body.children].filter(
      (el): el is HTMLElement => el instanceof HTMLElement && el !== modal
    )

    for (const el of background) el.inert = true
    closeButtonRef.current?.focus()

    return () => {
      for (const el of background) el.inert = false
      opener?.focus({ preventScroll: true })
    }
  }, [isOpen])

  // Keyboard navigation
  useHotkeys('esc', closeModal, { enabled: isOpen })
  useHotkeys('arrowleft', () => navigatePhoto('prev'), { enabled: isOpen })
  useHotkeys('arrowright', () => navigatePhoto('next'), { enabled: isOpen })

  if (!mounted || !shownPhoto) return null

  return createPortal(
    <div
      role={'dialog'}
      aria-modal={'true'}
      aria-label={shownPhoto.alt}
      className={cn(
        'fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm',
        'transition-[opacity,visibility] duration-300 ease-out motion-reduce:transition-none',
        // starting: animates the first open too, when the element mounts already open
        isOpen ? 'visible opacity-100 starting:opacity-0' : 'invisible opacity-0'
      )}
    >
      {/* Backdrop: a click anywhere closes the modal, the photo lets clicks through to it */}
      <button
        ref={closeButtonRef}
        type='button'
        aria-label={'Close'}
        className={'absolute inset-0'}
        onClick={closeModal}
      />

      {/* Photo content */}
      <div
        className={cn(
          'pointer-events-none relative flex size-full items-center justify-center',
          'transition-[opacity,scale] duration-300 ease-out motion-reduce:transition-none',
          isOpen ? 'scale-100 opacity-100 starting:scale-80 starting:opacity-0' : 'scale-80 opacity-0'
        )}
      >
        <Image
          {...getModalImageProps(shownPhoto)}
          alt={shownPhoto.alt}
          className={'h-auto max-h-full w-full max-w-full object-contain lg:h-full lg:w-auto'}
          loading={'eager'}
          fetchPriority={'high'}
        />
      </div>
    </div>,
    document.body
  )
}
