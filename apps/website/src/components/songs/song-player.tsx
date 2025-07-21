'use client'

import { useAtomValue, useSetAtom } from 'jotai'
import type { MouseEvent } from 'react'

import {
  audioControlsAtom,
  currentSongAtom,
  currentTimeAtom,
  durationAtom,
  isPlayingAtom
} from '@/atoms/songs'
import { formatTime } from '@/lib/utils'

export function SongPlayer() {
  const currentSong = useAtomValue(currentSongAtom)
  const isPlaying = useAtomValue(isPlayingAtom)
  const currentTime = useAtomValue(currentTimeAtom)
  const duration = useAtomValue(durationAtom)
  const setAudioControls = useSetAtom(audioControlsAtom)

  const progressPercentage = duration > 0 ? (currentTime / duration) * 100 : 0

  const handlePlayPause = () => {
    setAudioControls({ type: isPlaying ? 'pause' : 'continue' })
  }

  const handlePrevious = () => {
    setAudioControls({ type: 'previous' })
  }

  const handleNext = () => {
    setAudioControls({ type: 'next' })
  }

  const handleProgressClick = (e: MouseEvent<HTMLButtonElement>) => {
    if (duration === 0) return

    const rect = e.currentTarget.getBoundingClientRect()
    const clickX = e.clientX - rect.left
    const percentage = Math.max(0, Math.min(1, clickX / rect.width))
    const newTime = percentage * duration

    setAudioControls({ type: 'seek', time: newTime })
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
        <div className={'flex-1'}>
          <button
            className={'bg-foreground/70 h-1 w-full cursor-pointer rounded-full'}
            onClick={handleProgressClick}
          >
            <div
              className={'bg-dividing h-1 rounded-full'}
              style={{ width: `${progressPercentage}%` }}
            />
          </button>
        </div>
        <span className={'text-text-foreground/85 text-xs'}>{formatTime(duration)}</span>
      </div>

      {/* Player Controls */}
      <div className={'flex flex-wrap items-center justify-between'}>
        <button
          className={
            'text-text-foreground/70 hover:text-text-foreground flex size-10 cursor-pointer items-center justify-center rounded-full transition-colors'
          }
          aria-label={'Previous'}
          onClick={handlePrevious}
        >
          <span className={'i-mingcute-skip-previous-fill size-6'} />
        </button>

        <button
          className={
            'text-text-foreground/70 hover:text-text-foreground flex size-12 cursor-pointer items-center justify-center rounded-full transition-colors'
          }
          aria-label={isPlaying ? 'Pause' : 'Play'}
          onClick={handlePlayPause}
        >
          {isPlaying ? (
            <span className={'i-mingcute-pause-fill size-8'} />
          ) : (
            <span className={'i-mingcute-play-fill size-8'} />
          )}
        </button>

        <button
          className={
            'text-text-foreground/70 hover:text-text-foreground flex size-10 cursor-pointer items-center justify-center rounded-full transition-colors'
          }
          aria-label={'Next'}
          onClick={handleNext}
        >
          <span className={'i-mingcute-skip-forward-fill size-6'} />
        </button>

        <button
          className={
            'text-text-foreground/70 hover:text-text-foreground flex size-10 cursor-pointer items-center justify-center rounded-full transition-colors'
          }
          aria-label={'Volume'}
        >
          <span className={'i-mingcute-volume-fill size-5'} />
        </button>
      </div>
    </div>
  )
}
