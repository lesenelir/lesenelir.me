import { atom, useAtomValue, useSetAtom } from 'jotai'
import { useMemo } from 'react'

import type { TSong } from '@/types/songs'

import { cn } from '@/lib/utils'

import { audioControlsAtom, currentSongAtom, isPlayingAtom } from '@/atoms/songs'

interface SongItemProps {
  song: TSong
}

export function SongItem({ song }: SongItemProps) {
  // This song's state: a row re-renders only when its own state changes, not on every play, pause or switch to
  // another song
  const stateAtom = useMemo(
    () =>
      atom(get => {
        if (get(currentSongAtom)?.id !== song.id) return 'idle'
        return get(isPlayingAtom) ? 'playing' : 'paused'
      }),
    [song.id]
  )
  const state = useAtomValue(stateAtom)
  const setAudioControls = useSetAtom(audioControlsAtom)

  const isPlaying = state === 'playing'
  const isCurrent = state !== 'idle'

  return (
    <button
      type='button'
      // The current song stays highlighted while paused too
      aria-current={isCurrent ? 'true' : undefined}
      className={cn(
        'flex w-full cursor-pointer items-center gap-3 rounded-lg px-1.5 py-3 transition-colors duration-200 hover:bg-foreground',
        isCurrent && 'bg-foreground'
      )}
      // Names the song as well as the action; "Play" alone made every row sound the same to a screen reader
      aria-label={`${isPlaying ? 'Pause' : 'Play'} ${song.title} by ${song.artist}`}
      onClick={() => setAudioControls(isPlaying ? { type: 'pause' } : { type: 'play', song })}
    >
      <div
        className={cn(
          'flex size-8 shrink-0 items-center justify-center rounded-full',
          isCurrent ? 'bg-background' : 'bg-foreground'
        )}
      >
        {isPlaying ? (
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
