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

export const currentTimeAtom = atom<number>(0)

export const durationAtom = atom<number>(0)

export const audioControlsAtom = atom<
  null,
  [
    | { type: 'play'; song: TSong }
    | { type: 'continue' }
    | { type: 'pause' }
    | { type: 'previous' }
    | { type: 'next' }
    | { type: 'seek'; time: number }
  ],
  void
>(null, (get, set, action) => {
  const audio = get(audioInstanceAtom)
  if (!audio) return

  const currentSong = get(currentSongAtom)
  const songs = get(songsAtom)

  const setupAudioEventListeners = () => {
    audio.addEventListener('timeupdate', () => {
      set(currentTimeAtom, audio.currentTime)
    })

    audio.addEventListener('loadedmetadata', () => {
      set(durationAtom, audio.duration || 0)
    })

    audio.addEventListener('durationchange', () => {
      set(durationAtom, audio.duration || 0)
    })
  }

  const playAudio = () => {
    audio
      .play()
      .then(() => set(isPlayingAtom, true))
      .catch(() => set(isPlayingAtom, false))
  }

  const findAndPlaySong = (song: TSong) => {
    set(currentSongAtom, song)
    audio.src = song.src
    setupAudioEventListeners()
    playAudio()
  }

  switch (action.type) {
    case 'play':
      if (!action.song) return

      if (currentSong?.id !== action.song.id) {
        set(currentSongAtom, action.song)
        audio.src = action.song.src
        setupAudioEventListeners()
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

    case 'previous': {
      if (!currentSong || songs.length === 0) return
      const currentIndex = songs.findIndex((song) => song.id === currentSong.id)
      const previousSong = currentIndex <= 0 ? songs[songs.length - 1] : songs[currentIndex - 1]
      findAndPlaySong(previousSong)
      break
    }

    case 'next': {
      if (!currentSong || songs.length === 0) return
      const currentIndex = songs.findIndex((song) => song.id === currentSong.id)
      const nextSong = currentIndex >= songs.length - 1 ? songs[0] : songs[currentIndex + 1]
      findAndPlaySong(nextSong)
      break
    }

    case 'seek': {
      if (!currentSong) return
      const seekTime = Math.max(0, Math.min(action.time, audio.duration || 0))
      audio.currentTime = seekTime
      set(currentTimeAtom, seekTime)
      break
    }
  }
})
