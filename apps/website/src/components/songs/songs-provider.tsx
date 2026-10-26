'use client'

import { createStore, Provider, useSetAtom } from 'jotai'
import { useHydrateAtoms } from 'jotai/utils'
import type { ReactNode } from 'react'
import { useEffect, useState } from 'react'

import type { TSong } from '@/types'

import {
  audioControlsAtom,
  audioInstanceAtom,
  currentSongAtom,
  currentTimeAtom,
  DEFAULT_VOLUME,
  durationAtom,
  isBufferingAtom,
  isPlayingAtom,
  loadErrorAtom,
  songEndedAtom,
  songsAtom,
  type TAudioAction,
} from '@/atoms/songs'

type Store = ReturnType<typeof createStore>

// The system's media controls (lock screen, media keys, headphones) show the song and its position and drive the
// player, where the browser supports it
const hasMediaSession = () => typeof navigator !== 'undefined' && 'mediaSession' in navigator

const updatePositionState = (audio: HTMLAudioElement) => {
  if (!hasMediaSession() || !Number.isFinite(audio.duration)) return
  navigator.mediaSession.setPositionState({
    duration: audio.duration,
    playbackRate: audio.playbackRate,
    position: Math.min(audio.currentTime, audio.duration),
  })
}

// The audio element and its listeners live as long as the root layout, so a song keeps playing from page to page
function AudioEffect({ store }: { store: Store }) {
  useEffect(() => {
    const audio = new Audio()
    audio.volume = DEFAULT_VOLUME
    store.set(audioInstanceAtom, audio)

    const listeners: Record<string, () => void> = {
      timeupdate: () => store.set(currentTimeAtom, audio.currentTime),
      loadedmetadata: () => {
        store.set(durationAtom, audio.duration || 0)
        updatePositionState(audio)
      },
      durationchange: () => store.set(durationAtom, audio.duration || 0),
      // Playing and paused come from the element itself, so pauses from outside the page count too
      playing: () => {
        store.set(isPlayingAtom, true)
        store.set(isBufferingAtom, false)
        store.set(loadErrorAtom, false)
        updatePositionState(audio)
      },
      pause: () => {
        store.set(isPlayingAtom, false)
        store.set(isBufferingAtom, false)
        updatePositionState(audio)
      },
      waiting: () => store.set(isBufferingAtom, true),
      seeked: () => updatePositionState(audio),
      error: () => {
        store.set(loadErrorAtom, true)
        store.set(isBufferingAtom, false)
        store.set(isPlayingAtom, false)
      },
      ended: () => store.set(songEndedAtom),
    }
    for (const [event, listener] of Object.entries(listeners)) audio.addEventListener(event, listener)

    return () => {
      for (const [event, listener] of Object.entries(listeners)) audio.removeEventListener(event, listener)

      audio.pause()
      // Not `audio.src = ''`, which resolves to the page URL and requests it
      audio.removeAttribute('src')
      audio.load()

      // Reset all audio states (the store outlives this effect in Strict Mode and Fast Refresh)
      store.set(audioInstanceAtom, null)
      store.set(isPlayingAtom, false)
      store.set(isBufferingAtom, false)
      store.set(loadErrorAtom, false)
      store.set(currentSongAtom, null)
      store.set(currentTimeAtom, 0)
      store.set(durationAtom, 0)
    }
  }, [store])

  return null
}

// Keeps the system's media controls in step with the player. Subscribes to the store directly, so it never renders.
function MediaSessionEffect({ store }: { store: Store }) {
  useEffect(() => {
    if (!hasMediaSession()) return
    const session = navigator.mediaSession
    const control = (action: TAudioAction) => store.set(audioControlsAtom, action)

    const handlers: [MediaSessionAction, MediaSessionActionHandler][] = [
      ['play', () => control({ type: 'continue' })],
      ['pause', () => control({ type: 'pause' })],
      ['previoustrack', () => control({ type: 'previous' })],
      ['nexttrack', () => control({ type: 'next' })],
      ['seekto', details => details.seekTime !== undefined && control({ type: 'seek', time: details.seekTime })],
    ]
    for (const [action, handler] of handlers) session.setActionHandler(action, handler)

    const syncMetadata = () => {
      const song = store.get(currentSongAtom)
      session.metadata = song ? new MediaMetadata({ title: song.title, artist: song.artist }) : null
    }
    const syncPlaybackState = () => {
      session.playbackState = store.get(currentSongAtom) ? (store.get(isPlayingAtom) ? 'playing' : 'paused') : 'none'
    }
    syncMetadata()
    syncPlaybackState()
    const unsubscribe = [
      store.sub(currentSongAtom, syncMetadata),
      store.sub(currentSongAtom, syncPlaybackState),
      store.sub(isPlayingAtom, syncPlaybackState),
    ]

    return () => {
      for (const stop of unsubscribe) stop()
      for (const [action] of handlers) session.setActionHandler(action, null)
      session.metadata = null
      session.playbackState = 'none'
    }
  }, [store])

  return null
}

// In the root layout, so the player outlives the /songs page. Its own store, which the photos page's Provider
// doesn't reach: the sidebar and the songs page sit outside it.
export function SongsProvider({ children }: { children: ReactNode }) {
  const [store] = useState(createStore)

  return (
    <Provider store={store}>
      {children}
      <AudioEffect store={store} />
      <MediaSessionEffect store={store} />
    </Provider>
  )
}

interface HydrateSongsProps {
  songs: TSong[]
  children: ReactNode
}

// Puts the /songs page's list into the player. useHydrateAtoms fills it while rendering, so the server HTML lists the
// songs, but only once per store; the effect brings in a newer list (the page is revalidated hourly) on later visits.
export function HydrateSongs({ songs, children }: HydrateSongsProps) {
  useHydrateAtoms([[songsAtom, songs]])

  const setSongs = useSetAtom(songsAtom)
  useEffect(() => {
    setSongs(songs)
  }, [songs, setSongs])

  return <>{children}</>
}
