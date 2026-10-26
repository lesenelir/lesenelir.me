'use client'

import { useAtomValue, useSetAtom } from 'jotai'
import { useHotkeys } from 'react-hotkeys-hook'

import type { TPlayMode } from '@/types'

import { cn, formatTime } from '@/lib/utils'

import {
  audioControlsAtom,
  currentSongAtom,
  currentTimeAtom,
  durationAtom,
  isBufferingAtom,
  isMutedAtom,
  isPlayingAtom,
  loadErrorAtom,
  playModeAtom,
  volumeAtom,
} from '@/atoms/songs'

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

const playModeLabels: Record<TPlayMode, string> = {
  forward: 'Play in order',
  'repeat-all': 'Repeat all',
  'repeat-one': 'Repeat one',
  shuffle: 'Shuffle',
}

interface RangeBarProps {
  label: string
  value: number
  max: number
  step: number
  valueText: string
  onChange: (value: number) => void
  className?: string
}

// A thin bar with an invisible native range input laid over it: the input brings clicking, dragging, the arrow keys
// and a slider to screen readers, and the bar keeps the look. The taller input is also an easier target.
const RangeBar = ({ label, value, max, step, valueText, onChange, className }: RangeBarProps) => (
  <div
    className={cn(
      'relative h-1 rounded-full bg-text-foreground/10',
      'has-focus-visible:ring-2 has-focus-visible:ring-text-foreground/50 has-focus-visible:ring-offset-4 has-focus-visible:ring-offset-background',
      className
    )}
  >
    <div
      className={'h-1 rounded-full bg-text-foreground/80'}
      style={{ width: `${max > 0 ? (value / max) * 100 : 0}%` }}
    />
    <input
      type='range'
      aria-label={label}
      aria-valuetext={valueText}
      min={0}
      max={max}
      step={step}
      value={value}
      disabled={max === 0}
      onChange={e => onChange(Number(e.currentTarget.value))}
      className={'absolute inset-x-0 -top-2 h-5 w-full cursor-pointer opacity-0 disabled:cursor-default'}
    />
  </div>
)

// The only part that follows the playback time, which ticks about 4 times a second, so the rest of the player
// doesn't re-render with it
const SongProgress = () => {
  const currentTime = useAtomValue(currentTimeAtom)
  const duration = useAtomValue(durationAtom)
  const setAudioControls = useSetAtom(audioControlsAtom)

  return (
    <div className={'flex flex-wrap items-center gap-3'}>
      <span className={'text-xs text-text-foreground'}>{formatTime(currentTime)}</span>
      <RangeBar
        label={'Seek'}
        value={Math.min(currentTime, duration)}
        max={duration}
        step={1}
        valueText={`${formatTime(currentTime)} of ${formatTime(duration)}`}
        onChange={time => setAudioControls({ type: 'seek', time })}
        className={'flex-1'}
      />
      <span className={'text-xs text-text-foreground'}>{formatTime(duration)}</span>
    </div>
  )
}

export function SongPlayer() {
  const currentSong = useAtomValue(currentSongAtom)
  const isPlaying = useAtomValue(isPlayingAtom)
  const isBuffering = useAtomValue(isBufferingAtom)
  const loadError = useAtomValue(loadErrorAtom)
  const volume = useAtomValue(volumeAtom)
  const isMuted = useAtomValue(isMutedAtom)
  const playMode = useAtomValue(playModeAtom)
  const setAudioControls = useSetAtom(audioControlsAtom)

  // Loading counts as playing for the button: pressing it then stops the song from starting
  const isActive = isPlaying || isBuffering
  const togglePlay = () => setAudioControls({ type: isActive ? 'pause' : 'continue' })

  // Space plays and pauses, except on a focused control, which Space already presses
  useHotkeys('space', togglePlay, {
    preventDefault: true,
    ignoreEventWhen: e =>
      e.target instanceof Element && e.target.closest('button, a, input, select, textarea') !== null,
  })

  // /80 is the lightest these icons can be and still stand out 3:1 from the background, in both themes
  const baseButtonClass =
    'text-text-foreground/80 hover:text-text-foreground flex cursor-pointer items-center justify-center rounded-full transition-colors duration-200'

  const shownVolume = isMuted ? 0 : volume

  return (
    <div className={'mb-4 w-full space-y-4 bg-background px-2 py-4'}>
      {/* Song Info Row */}
      <div className={'flex flex-wrap items-center gap-3'}>
        <div className={'flex size-12 items-center justify-center rounded-lg bg-foreground'}>
          <span className={'i-mingcute-music-2-fill size-6 text-text-foreground'} />
        </div>
        <div className={'min-w-0 flex-1 space-y-0.5'}>
          {currentSong ? (
            <>
              <p className={'truncate text-sm font-medium text-text-primary'}>{currentSong.title}</p>
              <p className={'truncate text-xs text-text-foreground'}>{currentSong.artist}</p>
            </>
          ) : (
            <>
              <p className={'truncate text-sm font-medium text-text-primary'}>Not playing</p>
              <p className={'truncate text-xs text-text-foreground'}>Pick a song below, or press play</p>
            </>
          )}
          {/* Announced when a song fails to load */}
          <p role={'status'} className={'text-xs text-text-primary empty:hidden'}>
            {loadError ? "Couldn't load this song. Press play to try again." : ''}
          </p>
        </div>
      </div>

      {/* Progress Section */}
      <SongProgress />

      {/* Player Controls */}
      <div className={'flex flex-wrap items-center justify-between'}>
        {/* play control groups */}
        <div className={'flex flex-wrap items-center gap-2'}>
          <button
            type='button'
            className={cn(baseButtonClass, 'size-10')}
            aria-label={'Previous'}
            onClick={() => setAudioControls({ type: 'previous' })}
          >
            <span className={'i-mingcute-skip-previous-fill size-6'} />
          </button>

          <button
            type='button'
            className={cn(baseButtonClass, 'size-12')}
            aria-label={isActive ? 'Pause' : 'Play'}
            aria-busy={isBuffering}
            onClick={togglePlay}
          >
            {isBuffering ? (
              <span className={'i-mingcute-loading-line size-7 animate-spin motion-reduce:animate-none'} />
            ) : isPlaying ? (
              <span className={'i-mingcute-pause-fill size-8'} />
            ) : (
              <span className={'i-mingcute-play-fill size-8'} />
            )}
          </button>

          <button
            type='button'
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
            type='button'
            className={cn(baseButtonClass, 'size-10')}
            aria-label={`Play mode: ${playModeLabels[playMode]}`}
            onClick={() => setAudioControls({ type: 'togglePlayMode' })}
          >
            <PlayModeIcon />
          </button>

          <button
            type='button'
            className={cn(baseButtonClass, 'size-10')}
            aria-label={shownVolume === 0 ? 'Unmute' : 'Mute'}
            onClick={() => setAudioControls({ type: 'toggleMute' })}
          >
            {shownVolume === 0 ? (
              <span className={'i-mingcute-volume-mute-fill size-5'} />
            ) : (
              <span className={'i-mingcute-volume-fill size-5'} />
            )}
          </button>
          <RangeBar
            label={'Volume'}
            value={shownVolume}
            max={1}
            step={0.05}
            valueText={`${Math.round(shownVolume * 100)}%`}
            onChange={value => setAudioControls({ type: 'setVolume', volume: value })}
            className={'w-16'}
          />
        </div>
      </div>
    </div>
  )
}
