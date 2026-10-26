import { atom } from 'jotai'

import type { TPlayMode, TSong } from '@/types'

export const DEFAULT_VOLUME = 0.6

export const songsAtom = atom<TSong[]>([])

export const currentSongAtom = atom<TSong | null>(null)

/**
 * Audio States
 */
// SongsProvider creates the element when the page mounts and tears it down when it unmounts
export const audioInstanceAtom = atom<HTMLAudioElement | null>(null)

// Picks the next song when one ends, following the play mode
export const songEndedAtom = atom(null, (get, set) => {
  const audio = get(audioInstanceAtom)
  const songs = get(songsAtom)
  const currentSong = get(currentSongAtom)
  const playMode = get(playModeAtom)

  if (!audio || !currentSong || songs.length === 0) return

  const currentIndex = songs.findIndex(song => song.id === currentSong.id)
  let nextSong: TSong | null = null

  switch (playMode) {
    case 'forward': {
      // Stop at the end of playlists
      if (currentIndex >= songs.length - 1) {
        set(isPlayingAtom, false)
        return
      }
      nextSong = songs[currentIndex + 1]
      break
    }

    case 'repeat-all': {
      // Loop to first song after last song
      nextSong = currentIndex >= songs.length - 1 ? songs[0] : songs[currentIndex + 1]
      break
    }

    case 'repeat-one': {
      // Replay the same song
      nextSong = currentSong
      break
    }

    case 'shuffle': {
      // Play random next song (avoid current song if possible)
      if (songs.length === 1) {
        nextSong = currentSong
      } else {
        const availableSongs = songs.filter((_, index) => index !== currentIndex)
        const randomIndex = Math.floor(Math.random() * availableSongs.length)
        nextSong = availableSongs[randomIndex]
      }
      break
    }
  }

  if (nextSong) {
    set(currentSongAtom, nextSong)
    audio.src = nextSong.src
    audio
      .play()
      .then(() => set(isPlayingAtom, true))
      .catch(() => set(isPlayingAtom, false))
  }
})

export const playModeAtom = atom<TPlayMode>('forward')

export const isPlayingAtom = atom<boolean>(false)

export const currentTimeAtom = atom<number>(0)

export const durationAtom = atom<number>(0)

export const volumeAtom = atom<number>(DEFAULT_VOLUME)

export const isMutedAtom = atom<boolean>(false)

export const audioControlsAtom = atom<
  null,
  [
    | { type: 'play'; song: TSong }
    | { type: 'continue' }
    | { type: 'pause' }
    | { type: 'previous' }
    | { type: 'next' }
    | { type: 'seek'; time: number }
    | { type: 'setVolume'; volume: number }
    | { type: 'toggleMute' }
    | { type: 'togglePlayMode' },
  ],
  void
>(null, (get, set, action) => {
  const audio = get(audioInstanceAtom)
  if (!audio) return

  const songs = get(songsAtom)
  const currentSong = get(currentSongAtom)

  const playAudio = () => {
    // A song that failed to load (e.g. a dropped connection) stays failed: play() only rejects again until the
    // source is reloaded. A new src already reloads, so this matters when playing the same song again.
    if (audio.error) audio.load()
    audio
      .play()
      .then(() => set(isPlayingAtom, true))
      .catch(() => set(isPlayingAtom, false))
  }

  switch (action.type) {
    case 'play': {
      if (!action.song) return

      if (currentSong?.id !== action.song.id) {
        set(currentSongAtom, action.song)
        audio.src = action.song.src
      }
      playAudio()
      break
    }

    case 'continue': {
      if (!currentSong) return
      playAudio()
      break
    }

    case 'pause': {
      audio.pause()
      set(isPlayingAtom, false)
      break
    }

    case 'previous': {
      if (!currentSong || songs.length === 0) return

      const playMode = get(playModeAtom)
      const currentIndex = songs.findIndex(song => song.id === currentSong.id)
      let previousSong: TSong

      if (playMode === 'shuffle') {
        // In shuffle mode, select random previous song (avoid current if possible)
        if (songs.length === 1) {
          previousSong = currentSong
        } else {
          const availableSongs = songs.filter((_, index) => index !== currentIndex)
          const randomIndex = Math.floor(Math.random() * availableSongs.length)
          previousSong = availableSongs[randomIndex]
        }
      } else {
        // For other modes, use normal previous logic
        previousSong = currentIndex <= 0 ? songs[songs.length - 1] : songs[currentIndex - 1]
      }

      set(currentSongAtom, previousSong)
      audio.src = previousSong.src
      playAudio()
      break
    }

    case 'next': {
      if (!currentSong || songs.length === 0) return

      const playMode = get(playModeAtom)
      const currentIndex = songs.findIndex(song => song.id === currentSong.id)
      let nextSong: TSong

      if (playMode === 'shuffle') {
        // In shuffle mode, select random next song (avoid current if possible)
        if (songs.length === 1) {
          nextSong = currentSong
        } else {
          const availableSongs = songs.filter((_, index) => index !== currentIndex)
          const randomIndex = Math.floor(Math.random() * availableSongs.length)
          nextSong = availableSongs[randomIndex]
        }
      } else {
        // For other modes, use normal next logic
        nextSong = currentIndex >= songs.length - 1 ? songs[0] : songs[currentIndex + 1]
      }

      set(currentSongAtom, nextSong)
      audio.src = nextSong.src
      playAudio()
      break
    }

    case 'seek': {
      if (!currentSong) return
      const seekTime = Math.max(0, Math.min(action.time, audio.duration || 0))
      audio.currentTime = seekTime
      set(currentTimeAtom, seekTime)
      break
    }

    case 'setVolume': {
      const volume = Math.max(0, Math.min(1, action.volume))
      audio.volume = volume
      set(volumeAtom, volume)
      if (volume > 0) {
        set(isMutedAtom, false)
      } else {
        set(isMutedAtom, true)
      }
      break
    }

    case 'toggleMute': {
      const currentVolume = get(volumeAtom)
      const isMuted = get(isMutedAtom)

      if (isMuted) {
        audio.volume = currentVolume
        set(isMutedAtom, false)
      } else {
        audio.volume = 0
        set(isMutedAtom, true)
      }
      break
    }

    case 'togglePlayMode': {
      const currentMode = get(playModeAtom)
      let nextMode: TPlayMode

      switch (currentMode) {
        case 'forward':
          nextMode = 'repeat-all'
          break
        case 'repeat-all':
          nextMode = 'repeat-one'
          break
        case 'repeat-one':
          nextMode = 'shuffle'
          break
        case 'shuffle':
          nextMode = 'forward'
          break
      }

      set(playModeAtom, nextMode)
      break
    }
  }
})
