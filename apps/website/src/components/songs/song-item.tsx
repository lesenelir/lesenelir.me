import { atom, useAtomValue, useSetAtom } from 'jotai'
import { useMemo } from 'react'

import type { TSong } from '@/types/songs'

import { audioControlsAtom, currentSongAtom, isPlayingAtom } from '@/atoms/songs'

interface SongItemProps {
  song: TSong
}

export function SongItem({ song }: SongItemProps) {
  // Whether this song is the one playing: a row re-renders only when that changes, not on every play, pause or
  // switch to another song
  const isCurrentlyPlayingAtom = useMemo(
    () => atom(get => get(isPlayingAtom) && get(currentSongAtom)?.id === song.id),
    [song.id]
  )
  const isCurrentlyPlaying = useAtomValue(isCurrentlyPlayingAtom)
  const setAudioControls = useSetAtom(audioControlsAtom)

  const handlePlay = () => {
    if (isCurrentlyPlaying) {
      setAudioControls({ type: 'pause' })
    } else {
      setAudioControls({ type: 'play', song })
    }
  }

  return (
    <button
      type='button'
      className={
        'flex w-full cursor-pointer items-center gap-3 rounded-lg px-1.5 py-3 transition-colors duration-200 hover:bg-foreground'
      }
      // Names the song as well as the action; "Play" alone made every row sound the same to a screen reader
      aria-label={`${isCurrentlyPlaying ? 'Pause' : 'Play'} ${song.title} by ${song.artist}`}
      onClick={handlePlay}
    >
      <div className={'flex size-8 shrink-0 items-center justify-center rounded-full bg-foreground'}>
        {isCurrentlyPlaying ? (
          <span className={'i-mingcute-pause-fill size-4'} />
        ) : (
          <span className={'i-mingcute-play-fill size-4'} />
        )}
      </div>

      <div className={'min-w-0 flex-1 text-left'}>
        <p className={'truncate text-sm font-medium text-text-primary'}>{song.title}</p>
        <p className={'truncate text-xs text-text-foreground'}>{song.artist}</p>
      </div>
    </button>
  )
}
