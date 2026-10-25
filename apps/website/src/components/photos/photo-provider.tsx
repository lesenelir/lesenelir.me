'use client'

import { createStore, Provider } from 'jotai'
import { useHydrateAtoms } from 'jotai/utils'
import type { ReactNode } from 'react'
import { useState } from 'react'

import type { TPhoto } from '@/types'

import { photosAtom } from '@/atoms/photos'

interface PhotoProviderProps {
  photos: TPhoto[]
  children: ReactNode
}

export function PhotoProvider({ photos, children }: PhotoProviderProps) {
  // A store per page: Jotai's default store is a module singleton that every request shares on the server, and
  // useHydrateAtoms fills an atom only once per store, so the server kept rendering the first request's photos
  const [store] = useState(createStore)
  useHydrateAtoms([[photosAtom, photos]], { store })

  return <Provider store={store}>{children}</Provider>
}
