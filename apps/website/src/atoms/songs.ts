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

export const audioControlsAtom = atom<
  null,
  [{ type: 'play'; song: TSong } | { type: 'continue' } | { type: 'pause' }],
  void
>(null, (get, set, action) => {
  const audio = get(audioInstanceAtom)
  if (!audio) return

  const playAudio = () => {
    audio
      .play()
      .then(() => set(isPlayingAtom, true))
      .catch(() => set(isPlayingAtom, false))
  }

  const currentSong = get(currentSongAtom)

  switch (action.type) {
    case 'play':
      if (!action.song) return

      if (currentSong?.id !== action.song.id) {
        set(currentSongAtom, action.song)
        audio.src = action.song.src
      }
      playAudio()
      break

    case 'continue':
      if (!currentSong) return
      playAudio()
      break

    case 'pause':
      audio.pause()
      set(isPlayingAtom, false)
      break
  }
})
