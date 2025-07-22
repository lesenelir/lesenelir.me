'use client'

import { useEffect, useRef } from 'react'

import { createStore, Provider } from 'jotai'
import { useHydrateAtoms } from 'jotai/utils'
import type { ReactNode } from 'react'

import {
  audioInstanceAtom,
  currentSongAtom,
  currentTimeAtom,
  durationAtom,
  isPlayingAtom,
  songsAtom
} from '@/atoms/songs'
import type { TSong } from '@/types'

interface SongsProviderProps {
  songs: TSong[]
  children: ReactNode
}

function HydrateAtoms({ songs, children }: SongsProviderProps) {
  useHydrateAtoms([[songsAtom, songs]])

  return <>{children}</>
}

function CleanupEffect({ store }: { store: ReturnType<typeof createStore> }) {
  useEffect(() => {
    // Cleanup function that runs when the component unmounts
    return () => {
      const audio = store.get(audioInstanceAtom)

      if (audio) {
        // Pause audio and reset states
        audio.pause()
        audio.src = ''

        // Reset all audio states
        store.set(isPlayingAtom, false)
        store.set(currentSongAtom, null)
        store.set(currentTimeAtom, 0)
        store.set(durationAtom, 0)
      }
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
      <CleanupEffect store={storeRef.current} />
    </Provider>
  )
}
