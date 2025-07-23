'use client'

import { useHydrateAtoms } from 'jotai/utils'
import type { ReactNode } from 'react'

import { photosAtom } from '@/atoms/photos'
import type { TPhoto } from '@/types'

interface PhotoProviderProps {
  photos: TPhoto[]
  children: ReactNode
}

export function PhotoProvider({ photos, children }: PhotoProviderProps) {
  useHydrateAtoms([[photosAtom, photos]])

  return <>{children}</>
}
