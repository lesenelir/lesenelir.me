'use client'

import { useAtom, useAtomValue } from 'jotai'

import { currentSongAtom, isPlayingAtom, songsAtom } from '@/atoms/songs'
import { SongItem } from '@/components/songs/song-item'
import type { TSong } from '@/types/songs'

export function SongLists() {
  const songs = useAtomValue(songsAtom)
  const [currentSong, setCurrentSong] = useAtom(currentSongAtom)
  const [isPlaying, setIsPlaying] = useAtom(isPlayingAtom)

  const handlePlay = (song: TSong) => {
    if (currentSong?.id === song.id) {
      setIsPlaying(!isPlaying)
    } else {
      setCurrentSong(song)
      setIsPlaying(true)
    }
  }

  if (!songs.length) {
    return (
      <div className={'py-12 text-center'}>
        <div className={'text-text-foreground text-sm'}>No songs available</div>
      </div>
    )
  }

  return (
    <div className={'w-full space-y-1'}>
      {songs.map((song) => (
        <SongItem key={song.id} song={song} onPlay={handlePlay} />
      ))}
    </div>
  )
}
