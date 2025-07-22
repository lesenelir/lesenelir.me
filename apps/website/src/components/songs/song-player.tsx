'use client'

import { useAtomValue, useSetAtom } from 'jotai'
import type { MouseEvent } from 'react'

import {
  audioControlsAtom,
  currentSongAtom,
  currentTimeAtom,
  durationAtom,
  isMutedAtom,
  isPlayingAtom,
  playModeAtom,
  volumeAtom
} from '@/atoms/songs'
import { cn, formatTime } from '@/lib/utils'

const PlayModeIcon = () => {
  const playMode = useAtomValue(playModeAtom)

  switch (playMode) {
    case 'forward': {
      return <span className={'i-mingcute-forward-fill size-5'} />
    }
    case 'repeat-all': {
      return <span className={'i-mingcute-repeat-fill size-5'} />
    }
    case 'repeat-one': {
      return <span className={'i-mingcute-repeat-one-fill size-5'} />
    }
    case 'shuffle': {
      return <span className={'i-mingcute-shuffle-2-fill size-5'} />
    }
  }
}

export function SongPlayer() {
  const currentSong = useAtomValue(currentSongAtom)
  const isPlaying = useAtomValue(isPlayingAtom)
  const currentTime = useAtomValue(currentTimeAtom)
  const duration = useAtomValue(durationAtom)
  const volume = useAtomValue(volumeAtom)
  const isMuted = useAtomValue(isMutedAtom)
  const playMode = useAtomValue(playModeAtom)
  const setAudioControls = useSetAtom(audioControlsAtom)

  const progressPercentage = duration > 0 ? (currentTime / duration) * 100 : 0

  const baseButtonClass =
    'text-text-foreground/50 hover:text-text-foreground flex cursor-pointer items-center justify-center rounded-full transition-colors duration-200'

  const handleProgressClick = (e: MouseEvent<HTMLButtonElement>) => {
    if (duration === 0) return

    const rect = e.currentTarget.getBoundingClientRect()
    const clickX = e.clientX - rect.left
    const percentage = Math.max(0, Math.min(1, clickX / rect.width))
    const newTime = percentage * duration

    setAudioControls({ type: 'seek', time: newTime })
  }

  const handleVolumeClick = (e: MouseEvent<HTMLButtonElement>) => {
    const rect = e.currentTarget.getBoundingClientRect()
    const clickX = e.clientX - rect.left
    const percentage = Math.max(0, Math.min(1, clickX / rect.width))

    setAudioControls({ type: 'setVolume', volume: percentage })
  }

  return (
    <div className={'bg-background mb-4 w-full space-y-4 px-2 py-4'}>
      {/* Song Info Row */}
      <div className={'flex flex-wrap items-center gap-3'}>
        <div className={'bg-foreground flex size-12 items-center justify-center rounded-lg'}>
          <span className={'i-mingcute-music-2-fill text-text-foreground size-6'} />
        </div>
        <div className={'min-w-0 flex-1 space-y-0.5'}>
          <p className={'text-text-primary truncate text-sm font-medium'}>{currentSong?.title}</p>
          <p className={'text-text-foreground/85 truncate text-xs'}>{currentSong?.artist}</p>
        </div>
      </div>

      {/* Progress Section */}
      <div className={'flex flex-wrap items-center gap-3'}>
        <span className={'text-text-foreground/85 text-xs'}>{formatTime(currentTime)}</span>
        <div className={'flex flex-1 items-center'}>
          <button
            className={'bg-text-foreground/10 h-1 w-full cursor-pointer rounded-full'}
            onClick={handleProgressClick}
          >
            <div
              className={'bg-text-foreground/50 h-1 rounded-full'}
              style={{ width: `${progressPercentage}%` }}
            />
          </button>
        </div>
        <span className={'text-text-foreground/85 text-xs'}>{formatTime(duration)}</span>
      </div>

      {/* Player Controls */}
      <div className={'flex flex-wrap items-center justify-between'}>
        {/* play control groups */}
        <div className={'flex flex-wrap items-center gap-2'}>
          <button
            className={cn(baseButtonClass, 'size-10')}
            aria-label={'Previous'}
            onClick={() => setAudioControls({ type: 'previous' })}
          >
            <span className={'i-mingcute-skip-previous-fill size-6'} />
          </button>

          <button
            className={cn(baseButtonClass, 'size-12')}
            aria-label={isPlaying ? 'Pause' : 'Play'}
            onClick={() => setAudioControls({ type: isPlaying ? 'pause' : 'continue' })}
          >
            {isPlaying ? (
              <span className={'i-mingcute-pause-fill size-8'} />
            ) : (
              <span className={'i-mingcute-play-fill size-8'} />
            )}
          </button>

          <button
            className={cn(baseButtonClass, 'size-10')}
            aria-label={'Next'}
            onClick={() => setAudioControls({ type: 'next' })}
          >
            <span className={'i-mingcute-skip-forward-fill size-6'} />
          </button>
        </div>

        {/* repeat and volume control groups */}
        <div className={'flex flex-wrap items-center gap-2'}>
          <button
            className={cn(baseButtonClass, 'size-10')}
            aria-label={playMode}
            onClick={() => setAudioControls({ type: 'togglePlayMode' })}
          >
            <PlayModeIcon />
          </button>

          <button
            className={cn(baseButtonClass, 'size-10')}
            aria-label={'Volume'}
            onClick={() => setAudioControls({ type: 'toggleMute' })}
          >
            {isMuted || volume === 0 ? (
              <span className={'i-mingcute-volume-mute-fill size-5'} />
            ) : (
              <span className={'i-mingcute-volume-fill size-5'} />
            )}
          </button>
          <button
            className={'bg-text-foreground/10 h-1 w-16 cursor-pointer rounded-full'}
            aria-label={'Volume Progress'}
            onClick={handleVolumeClick}
          >
            <div
              className={'bg-text-foreground/50 h-1 rounded-full'}
              style={{ width: `${(isMuted ? 0 : volume) * 100}%` }}
            />
          </button>
        </div>
      </div>
    </div>
  )
}
