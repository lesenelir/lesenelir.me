import { atom } from 'jotai'

import type { TSong } from '@/types'

export const songsAtom = atom<TSong[]>([])

export const currentSongAtom = atom<TSong | null>(null)

export const isPlayingAtom = atom<boolean>(false)
