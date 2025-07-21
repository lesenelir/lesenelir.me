'use client'

import { Provider } from 'jotai'
import { useHydrateAtoms } from 'jotai/utils'
import type { ReactNode } from 'react'

import { songsAtom } from '@/atoms/songs'
import type { TSong } from '@/types'

interface SongsProviderProps {
  songs: TSong[]
  children: ReactNode
}

function HydrateAtoms({ songs, children }: SongsProviderProps) {
  useHydrateAtoms([[songsAtom, songs]])

  return <>{children}</>
}

export function SongsProvider({ songs, children }: SongsProviderProps) {
  return (
    <Provider>
      <HydrateAtoms songs={songs}>{children}</HydrateAtoms>
    </Provider>
  )
}
