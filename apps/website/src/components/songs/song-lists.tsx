'use client'

import { useAtom, useAtomValue, useSetAtom } from 'jotai'

import { audioControlsAtom, audioInstanceAtom, currentSongAtom, songsAtom } from '@/atoms/songs'
import { SongItem } from '@/components/songs/song-item'
import type { TSong } from '@/types/songs'

export function SongLists() {
  const songs = useAtomValue(songsAtom)
  const [currentSong, setCurrentSong] = useAtom(currentSongAtom)
  const audioInstance = useAtomValue(audioInstanceAtom)
  const setAudioControls = useSetAtom(audioControlsAtom)

  const handlePlay = (song: TSong) => {
    if (currentSong?.id === song.id) {
      setAudioControls({ type: 'pause' })
    } else {
      if (audioInstance) {
        audioInstance.src = song.src
      }
      setCurrentSong(song)
      setAudioControls({ type: 'play' })
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
