import { useAtomValue, useSetAtom } from 'jotai'

import type { TSong } from '@/types/songs'

import { audioControlsAtom, currentSongAtom, isPlayingAtom } from '@/atoms/songs'

interface SongItemProps {
  song: TSong
}

export function SongItem({ song }: SongItemProps) {
  const currentSong = useAtomValue(currentSongAtom)
  const isPlaying = useAtomValue(isPlayingAtom)
  const setAudioControls = useSetAtom(audioControlsAtom)

  const isCurrentSong = currentSong?.id === song.id
  const isCurrentlyPlaying = isCurrentSong && isPlaying

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
      aria-label={isCurrentlyPlaying ? 'Pause' : 'Play'}
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
        <p className={'truncate font-medium text-sm text-text-primary'}>{song.title}</p>
        <p className={'truncate text-text-foreground/85 text-xs'}>{song.artist}</p>
      </div>
    </button>
  )
}
