'use client'

import { useAtomValue } from 'jotai'

import { songsAtom } from '@/atoms/songs'
import { SongItem } from '@/components/songs/song-item'

export function SongLists() {
  const songs = useAtomValue(songsAtom)

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
        <SongItem key={song.id} song={song} />
      ))}
    </div>
  )
}
