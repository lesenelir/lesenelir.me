import { atom } from 'jotai'

import type { TSong } from '@/types'

function createGlobalAudio(): HTMLAudioElement | null {
  if (typeof window === 'undefined') return null
  return new Audio()
}

export const songsAtom = atom<TSong[]>([])

export const currentSongAtom = atom<TSong | null>(null)

/**
 * Audio States
 */
export const audioInstanceAtom = atom<HTMLAudioElement | null>(createGlobalAudio())

export const isPlayingAtom = atom<boolean>(false)

export const audioControlsAtom = atom<null, [{ type: 'play' | 'pause' }], void>(
  null,
  (get, set, action: { type: 'play' | 'pause' }) => {
    const audio = get(audioInstanceAtom)
    if (!audio) return

    switch (action.type) {
      case 'play': {
        audio.play().catch(console.error)
        set(isPlayingAtom, true)
        break
      }
      case 'pause': {
        audio.pause()
        set(isPlayingAtom, false)
        break
      }
    }
  }
)
