'use client'

import { useHydrateAtoms } from 'jotai/utils'
import type { ReactNode } from 'react'

import type { TPhoto } from '@/types'

import { photosAtom } from '@/atoms/photos'

interface PhotoProviderProps {
  photos: TPhoto[]
  children: ReactNode
}

export function PhotoProvider({ photos, children }: PhotoProviderProps) {
  useHydrateAtoms([[photosAtom, photos]])

  return <>{children}</>
}
