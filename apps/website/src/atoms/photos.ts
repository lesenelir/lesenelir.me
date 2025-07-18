import { atom } from 'jotai'
import { atomWithStorage } from 'jotai/utils'

import type { LayoutMode, Photo } from '@/types'

export const layoutModeAtom = atomWithStorage<LayoutMode>('photos-layout-mode', 'single')

export const photosAtom = atom<Photo[]>([])

export const selectedPhotoAtom = atom<Photo | null>(null)

export const isPhotoModalOpenAtom = atom(false)

export const openPhotoModalAtom = atom(null, (_get, set, photo: Photo) => {
  set(selectedPhotoAtom, photo)
  set(isPhotoModalOpenAtom, true)
})

export const closePhotoModalAtom = atom(null, (_get, set) => {
  set(isPhotoModalOpenAtom, false)
  set(selectedPhotoAtom, null)
})

export const navigatePhotoAtom = atom(null, (get, set, direction: 'prev' | 'next') => {
  const photos = get(photosAtom)
  const selectedPhoto = get(selectedPhotoAtom)

  if (!selectedPhoto || photos.length === 0) return

  const currentIndex = photos.findIndex((photo) => photo.id === selectedPhoto.id)
  if (currentIndex === -1) return

  let nextIndex: number
  if (direction === 'next') {
    nextIndex = currentIndex === photos.length - 1 ? 0 : currentIndex + 1
  } else {
    nextIndex = currentIndex === 0 ? photos.length - 1 : currentIndex - 1
  }

  set(selectedPhotoAtom, photos[nextIndex])
})
