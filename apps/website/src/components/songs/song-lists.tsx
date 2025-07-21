'use client'

import { useAtomValue } from 'jotai'

import { songsAtom } from '@/atoms/songs'

export function SongLists() {
  const songs = useAtomValue(songsAtom)
  console.log('songs:', songs)

  return <>song lists</>
}
