import { atom } from 'jotai'

import type { TSong } from '@/types'

export const songsAtom = atom<TSong[]>([])
