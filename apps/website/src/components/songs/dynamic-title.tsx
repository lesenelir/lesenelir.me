'use client'

import { useAtomValue } from 'jotai'
import { useEffect } from 'react'

import { currentSongAtom, isPlayingAtom } from '@/atoms/songs'

export function DynamicTitle() {
  const currentSong = useAtomValue(currentSongAtom)
  const isPlaying = useAtomValue(isPlayingAtom)

  useEffect(() => {
    if (!currentSong || !isPlaying) return

    const songTitle = `${currentSong.title} - ${currentSong.artist}`
    const apply = () => {
      // Only on this page: leaving it may set the next page's title before this component unmounts
      if (window.location.pathname === '/songs' && document.title !== songTitle) document.title = songTitle
    }
    apply()

    // Coming back to /songs while a song plays, Next sets the page's own title after this effect has run
    const observer = new MutationObserver(apply)
    observer.observe(document.head, { childList: true, subtree: true, characterData: true })
    return () => observer.disconnect()
  }, [currentSong, isPlaying])

  return null
}
