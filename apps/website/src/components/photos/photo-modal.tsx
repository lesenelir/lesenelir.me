'use client'

import { useAtomValue, useSetAtom } from 'jotai'
import Image from 'next/image'
import type { PointerEvent } from 'react'
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

// A horizontal swipe at least this long, and clearly more horizontal than vertical, turns the photo
const SWIPE_DISTANCE = 50

// Dark, translucent controls that read on any photo and either theme's backdrop
const controlClassName = cn(
  'pointer-events-auto flex size-10 cursor-pointer items-center justify-center rounded-full',
  'bg-black/50 text-white backdrop-blur-sm transition-colors duration-200 hover:bg-black/70',
  'focus-visible:outline-2 focus-visible:outline-white'
)

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

  // While open, focus moves into the modal and the page behind it is inert, out of reach of Tab and screen readers,
  // and doesn't scroll (the scrollbar's gutter stays, so the page doesn't shift); closing gives focus back to the
  // photo that opened it
  const closeButtonRef = useRef<HTMLButtonElement>(null)
  useEffect(() => {
    if (!isOpen) return

    const opener = document.activeElement instanceof HTMLElement ? document.activeElement : null
    const modal = closeButtonRef.current?.closest('[role="dialog"]')
    const background = [...document.body.children].filter(
      (el): el is HTMLElement => el instanceof HTMLElement && el !== modal
    )
    const root = document.documentElement.style
    const { overflow, scrollbarGutter } = root

    for (const el of background) el.inert = true
    root.overflow = 'hidden'
    root.scrollbarGutter = 'stable'
    closeButtonRef.current?.focus()

    return () => {
      for (const el of background) el.inert = false
      root.overflow = overflow
      root.scrollbarGutter = scrollbarGutter
      opener?.focus({ preventScroll: true })
    }
  }, [isOpen])

  // Swipes on touch screens turn the photo. The swipe also ends with a click on the backdrop, which mustn't close the
  // modal, so it is marked as swiped and that click ignored.
  const swipeStartRef = useRef<{ x: number; y: number } | null>(null)
  const swipedRef = useRef(false)

  const handlePointerDown = (e: PointerEvent<HTMLDivElement>) => {
    if (e.pointerType === 'mouse') return
    swipeStartRef.current = { x: e.clientX, y: e.clientY }
    swipedRef.current = false
  }

  const handlePointerUp = (e: PointerEvent<HTMLDivElement>) => {
    const start = swipeStartRef.current
    swipeStartRef.current = null
    if (!start) return

    const dx = e.clientX - start.x
    const dy = e.clientY - start.y
    if (Math.abs(dx) >= SWIPE_DISTANCE && Math.abs(dx) > Math.abs(dy) * 1.5) {
      swipedRef.current = true
      navigatePhoto(dx < 0 ? 'next' : 'prev')
    }
  }

  const handleBackdropClick = () => {
    if (swipedRef.current) {
      swipedRef.current = false
      return
    }
    closeModal()
  }

  // Keyboard navigation
  useHotkeys('esc', closeModal, { enabled: isOpen })
  useHotkeys('arrowleft', () => navigatePhoto('prev'), { enabled: isOpen })
  useHotkeys('arrowright', () => navigatePhoto('next'), { enabled: isOpen })

  if (!mounted || !shownPhoto) return null

  const position = `${photos.findIndex(photo => photo.id === shownPhoto.id) + 1} / ${photos.length}`
  const canNavigate = photos.length > 1

  return createPortal(
    <div
      role={'dialog'}
      aria-modal={'true'}
      aria-label={shownPhoto.alt}
      onPointerDown={handlePointerDown}
      onPointerUp={handlePointerUp}
      onPointerCancel={() => {
        swipeStartRef.current = null
      }}
      className={cn(
        'fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm',
        // Horizontal swipes are ours; vertical panning and pinch-zooming stay the browser's
        'touch-pan-y touch-pinch-zoom',
        'transition-[opacity,visibility] duration-300 ease-out motion-reduce:transition-none',
        // starting: animates the first open too, when the element mounts already open
        isOpen ? 'visible opacity-100 starting:opacity-0' : 'invisible opacity-0'
      )}
    >
      {/* Backdrop: a click anywhere outside the controls closes the modal, the photo lets clicks through to it. The
          close button does the same for the keyboard and screen readers. */}
      <button
        type='button'
        tabIndex={-1}
        aria-hidden={'true'}
        className={'absolute inset-0'}
        onClick={handleBackdropClick}
      />

      {/* Photo content, inset so the controls sit around it rather than on it */}
      <div
        className={cn(
          'pointer-events-none relative flex size-full items-center justify-center px-4 py-16 sm:px-20',
          'transition-[opacity,scale] duration-300 ease-out motion-reduce:transition-none',
          isOpen ? 'scale-100 opacity-100 starting:scale-80 starting:opacity-0' : 'scale-80 opacity-0'
        )}
      >
        {/* A new <img> per photo: reusing one keeps showing the previous photo until the next has loaded, under the
            next one's caption and position */}
        <Image
          key={shownPhoto.id}
          {...getModalImageProps(shownPhoto)}
          alt={shownPhoto.alt}
          className={'h-auto max-h-full w-full max-w-full object-contain lg:h-full lg:w-auto'}
          loading={'eager'}
          fetchPriority={'high'}
        />
      </div>

      {/* Position and close */}
      <div className={'pointer-events-none absolute inset-x-0 top-0 flex items-center justify-between p-4'}>
        <span className={'rounded-full bg-black/50 px-3 py-1 text-xs text-white tabular-nums backdrop-blur-sm'}>
          {position}
        </span>
        <button
          ref={closeButtonRef}
          type='button'
          aria-label={'Close'}
          className={controlClassName}
          onClick={closeModal}
        >
          <span className={'i-mingcute-close-line size-5'} />
        </button>
      </div>

      {canNavigate && (
        <>
          <button
            type='button'
            aria-label={'Previous photo'}
            className={cn(controlClassName, 'absolute top-1/2 left-2 -translate-y-1/2 sm:left-4')}
            onClick={() => navigatePhoto('prev')}
          >
            <span className={'i-mingcute-left-line size-5'} />
          </button>
          <button
            type='button'
            aria-label={'Next photo'}
            className={cn(controlClassName, 'absolute top-1/2 right-2 -translate-y-1/2 sm:right-4')}
            onClick={() => navigatePhoto('next')}
          >
            <span className={'i-mingcute-right-line size-5'} />
          </button>
        </>
      )}

      {shownPhoto.desc && (
        <div className={'pointer-events-none absolute inset-x-0 bottom-0 flex justify-center p-4'}>
          <p className={'max-w-[90%] rounded-lg bg-black/50 px-3 py-1 text-center text-sm text-white backdrop-blur-sm'}>
            {shownPhoto.desc}
          </p>
        </div>
      )}

      {/* Tells screen readers which photo the arrows and swipes moved to */}
      <p className={'sr-only'} aria-live={'polite'}>
        {isOpen ? `${position}: ${shownPhoto.alt}` : ''}
      </p>
    </div>,
    document.body
  )
}
