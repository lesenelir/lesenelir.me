'use client'

import { createStore, Provider } from 'jotai'
import { useHydrateAtoms } from 'jotai/utils'
import type { ReactNode } from 'react'
import { useEffect, useRef } from 'react'

import type { TSong } from '@/types'

import {
  audioInstanceAtom,
  currentSongAtom,
  currentTimeAtom,
  DEFAULT_VOLUME,
  durationAtom,
  isPlayingAtom,
  songEndedAtom,
  songsAtom,
} from '@/atoms/songs'

interface SongsProviderProps {
  songs: TSong[]
  children: ReactNode
}

function HydrateAtoms({ songs, children }: SongsProviderProps) {
  useHydrateAtoms([[songsAtom, songs]])

  return <>{children}</>
}

// The audio element and its listeners live exactly as long as the page, so nothing piles up across visits
function AudioEffect({ store }: { store: ReturnType<typeof createStore> }) {
  useEffect(() => {
    const audio = new Audio()
    audio.volume = DEFAULT_VOLUME
    store.set(audioInstanceAtom, audio)

    const onTimeUpdate = () => store.set(currentTimeAtom, audio.currentTime)
    const onDurationChange = () => store.set(durationAtom, audio.duration || 0)
    const onEnded = () => store.set(songEndedAtom)

    audio.addEventListener('timeupdate', onTimeUpdate)
    audio.addEventListener('loadedmetadata', onDurationChange)
    audio.addEventListener('durationchange', onDurationChange)
    audio.addEventListener('ended', onEnded)

    return () => {
      audio.removeEventListener('timeupdate', onTimeUpdate)
      audio.removeEventListener('loadedmetadata', onDurationChange)
      audio.removeEventListener('durationchange', onDurationChange)
      audio.removeEventListener('ended', onEnded)

      audio.pause()
      // Not `audio.src = ''`, which resolves to the page URL and requests it
      audio.removeAttribute('src')
      audio.load()

      // Reset all audio states (the store outlives this effect in Strict Mode and Fast Refresh)
      store.set(audioInstanceAtom, null)
      store.set(isPlayingAtom, false)
      store.set(currentSongAtom, null)
      store.set(currentTimeAtom, 0)
      store.set(durationAtom, 0)
    }
  }, [store])

  return null
}

export function SongsProvider({ songs, children }: SongsProviderProps) {
  // create a page scope store
  const storeRef = useRef<ReturnType<typeof createStore> | null>(null)

  if (!storeRef.current) {
    storeRef.current = createStore()
  }

  return (
    <Provider store={storeRef.current}>
      <HydrateAtoms songs={songs}>{children}</HydrateAtoms>
      <AudioEffect store={storeRef.current} />
    </Provider>
  )
}
