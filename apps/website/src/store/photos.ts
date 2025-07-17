import { atom } from 'jotai'
import { atomWithStorage } from 'jotai/utils'

import type { LayoutMode, Photo } from '@/types'

export const layoutModeAtom = atomWithStorage<LayoutMode>('photos-layout-mode', 'single')

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
