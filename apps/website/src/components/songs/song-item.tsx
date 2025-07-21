import { useAtomValue } from 'jotai'

import { currentSongAtom, isPlayingAtom } from '@/atoms/songs'
import type { TSong } from '@/types/songs'

interface SongItemProps {
  song: TSong
  onPlay?: (song: TSong) => void
}

export function SongItem({ song, onPlay }: SongItemProps) {
  const currentSong = useAtomValue(currentSongAtom)
  const isPlaying = useAtomValue(isPlayingAtom)

  const isCurrentSong = currentSong?.id === song.id
  const isCurrentlyPlaying = isCurrentSong && isPlaying
  const handlePlay = () => {
    onPlay?.(song)
  }

  return (
    <button
      className={
        'hover:bg-foreground flex w-full cursor-pointer items-center gap-3 rounded-lg px-1.5 py-3 transition-colors duration-200'
      }
      aria-label={isCurrentlyPlaying ? 'Pause' : 'Play'}
      onClick={handlePlay}
    >
      <div className={'bg-foreground flex size-8 items-center justify-center rounded-full'}>
        {isCurrentlyPlaying ? (
          <span className={'i-mingcute-pause-fill size-4'} />
        ) : (
          <span className={'i-mingcute-play-fill size-4'} />
        )}
      </div>

      <div className={'flex min-w-0 flex-1 flex-col items-start'}>
        <p className={'text-text-primary truncate text-sm font-medium'}>{song.title}</p>
        <p className={'text-text-foreground/85 truncate text-xs'}>{song.artist}</p>
      </div>
    </button>
  )
}
