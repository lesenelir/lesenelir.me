import type { Getter, Setter } from 'jotai'
import { atom } from 'jotai'

import type { TPlayMode, TSong } from '@/types'

export const DEFAULT_VOLUME = 0.6

// "Previous" restarts the song once it has played longer than this, as music players do
const RESTART_AFTER = 3

export const songsAtom = atom<TSong[]>([])

export const currentSongAtom = atom<TSong | null>(null)

/**
 * Audio States
 */
// SongsProvider creates the element when the page mounts and tears it down when it unmounts
export const audioInstanceAtom = atom<HTMLAudioElement | null>(null)

export const playModeAtom = atom<TPlayMode>('forward')

// Set from the audio element's own events (SongsProvider), so pauses from outside the page (media keys, unplugged
// headphones) show too
export const isPlayingAtom = atom<boolean>(false)

// Waiting for data: from asking to play until the audio actually plays
export const isBufferingAtom = atom<boolean>(false)

// The current song failed to load; playing it again retries
export const loadErrorAtom = atom<boolean>(false)

export const currentTimeAtom = atom<number>(0)

export const durationAtom = atom<number>(0)

export const volumeAtom = atom<number>(DEFAULT_VOLUME)

export const isMutedAtom = atom<boolean>(false)

// Shuffle plays the songs in a shuffled order rather than one random pick at a time: none repeats until all have
// played, and "previous" goes back to the song actually played before. An order starts at the song it's made for.
const shuffleOrderAtom = atom<string[]>([])

const makeShuffleOrder = (songs: TSong[], first: TSong | null) => {
  const rest = songs.filter(song => song.id !== first?.id).map(song => song.id)
  for (let i = rest.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[rest[i], rest[j]] = [rest[j], rest[i]]
  }
  return first ? [first.id, ...rest] : rest
}

// The song before or after the current one, in list order or the shuffle order. Null in shuffle when there is
// nothing before the first song played, or nothing after the only song.
const songInDirection = (get: Getter, set: Setter, direction: 1 | -1): TSong | null => {
  const songs = get(songsAtom)
  const currentSong = get(currentSongAtom)
  if (!currentSong || songs.length === 0) return null

  if (get(playModeAtom) !== 'shuffle') {
    const index = songs.findIndex(song => song.id === currentSong.id)
    return songs[(index + direction + songs.length) % songs.length]
  }

  let order = get(shuffleOrderAtom)
  let index = order.indexOf(currentSong.id)
  // No order for this song yet, or every song has played: shuffle again, starting from this one
  if (index === -1 || (direction === 1 && index === order.length - 1)) {
    order = makeShuffleOrder(songs, currentSong)
    set(shuffleOrderAtom, order)
    index = 0
  }
  return songs.find(song => song.id === order[index + direction]) ?? null
}

const playAudio = (set: Setter, audio: HTMLAudioElement) => {
  // A song that failed to load (e.g. a dropped connection) stays failed: play() only rejects again until the
  // source is reloaded. A new src already reloads, so this matters when playing the same song again.
  if (audio.error) audio.load()
  set(loadErrorAtom, false)
  set(isBufferingAtom, true)
  // isPlaying follows the element's playing/pause events; this only covers a play() that never starts
  audio.play().catch(() => {
    set(isBufferingAtom, false)
    set(isPlayingAtom, false)
  })
}

const startSong = (set: Setter, audio: HTMLAudioElement, song: TSong) => {
  set(currentSongAtom, song)
  // A new source resets the element's duration without a durationchange event, so the last song's length and
  // position would show until the new one's metadata loads
  set(currentTimeAtom, 0)
  set(durationAtom, 0)
  audio.src = song.src
  playAudio(set, audio)
}

const restartSong = (set: Setter, audio: HTMLAudioElement) => {
  audio.currentTime = 0
  set(currentTimeAtom, 0)
}

// Picks the next song when one ends, following the play mode
export const songEndedAtom = atom(null, (get, set) => {
  const audio = get(audioInstanceAtom)
  const songs = get(songsAtom)
  const currentSong = get(currentSongAtom)
  if (!audio || !currentSong || songs.length === 0) return

  switch (get(playModeAtom)) {
    case 'forward': {
      // Stop at the end of the list
      const isLast = songs[songs.length - 1].id === currentSong.id
      if (isLast) return
      startSong(set, audio, songInDirection(get, set, 1) ?? currentSong)
      break
    }

    case 'repeat-one': {
      restartSong(set, audio)
      playAudio(set, audio)
      break
    }

    // repeat-all wraps to the first song; shuffle moves along its order and reshuffles at the end
    default: {
      const next = songInDirection(get, set, 1)
      if (next && next.id !== currentSong.id) {
        startSong(set, audio, next)
      } else {
        restartSong(set, audio)
        playAudio(set, audio)
      }
    }
  }
})

export type TAudioAction =
  | { type: 'play'; song: TSong }
  | { type: 'continue' }
  | { type: 'pause' }
  | { type: 'previous' }
  | { type: 'next' }
  | { type: 'seek'; time: number }
  | { type: 'setVolume'; volume: number }
  | { type: 'toggleMute' }
  | { type: 'togglePlayMode' }

export const audioControlsAtom = atom<null, [TAudioAction], void>(null, (get, set, action) => {
  const audio = get(audioInstanceAtom)
  if (!audio) return

  const songs = get(songsAtom)
  const currentSong = get(currentSongAtom)

  switch (action.type) {
    case 'play': {
      if (currentSong?.id === action.song.id) {
        playAudio(set, audio)
        break
      }
      // A song picked by hand starts a new shuffle order from it
      if (get(playModeAtom) === 'shuffle') set(shuffleOrderAtom, makeShuffleOrder(songs, action.song))
      startSong(set, audio, action.song)
      break
    }

    case 'continue': {
      // With nothing picked yet, start the list
      if (currentSong) {
        playAudio(set, audio)
      } else if (songs.length > 0) {
        startSong(set, audio, songs[0])
      }
      break
    }

    case 'pause': {
      audio.pause()
      break
    }

    case 'previous': {
      if (!currentSong) return
      const previous = audio.currentTime > RESTART_AFTER ? null : songInDirection(get, set, -1)
      if (previous) {
        startSong(set, audio, previous)
      } else {
        restartSong(set, audio)
      }
      break
    }

    case 'next': {
      if (!currentSong) return
      const next = songInDirection(get, set, 1)
      if (next) startSong(set, audio, next)
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
      set(isMutedAtom, volume === 0)
      break
    }

    case 'toggleMute': {
      const isMuted = get(isMutedAtom)
      audio.volume = isMuted ? get(volumeAtom) : 0
      set(isMutedAtom, !isMuted)
      break
    }

    case 'togglePlayMode': {
      const nextMode: Record<TPlayMode, TPlayMode> = {
        forward: 'repeat-all',
        'repeat-all': 'repeat-one',
        'repeat-one': 'shuffle',
        shuffle: 'forward',
      }
      const mode = nextMode[get(playModeAtom)]
      // Shuffle starts its order from the song playing now
      if (mode === 'shuffle') set(shuffleOrderAtom, makeShuffleOrder(songs, currentSong))
      set(playModeAtom, mode)
      break
    }
  }
})
