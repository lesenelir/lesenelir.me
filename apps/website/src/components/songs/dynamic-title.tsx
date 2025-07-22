'use client'

import { useEffect } from 'react'

import { useAtomValue } from 'jotai'

import { currentSongAtom, isPlayingAtom } from '@/atoms/songs'

export function DynamicTitle() {
  const currentSong = useAtomValue(currentSongAtom)
  const isPlaying = useAtomValue(isPlayingAtom)

  useEffect(() => {
    if (currentSong && isPlaying) {
      document.title = `${currentSong.title} - ${currentSong.artist}`
    }
  }, [currentSong, isPlaying])

  return null
}
