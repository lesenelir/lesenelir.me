'use client'

import { useAtom, useAtomValue } from 'jotai'

import { currentSongAtom, isPlayingAtom } from '@/atoms/songs'

export function SongPlayer() {
  const currentSong = useAtomValue(currentSongAtom)
  const [isPlaying, setIsPlaying] = useAtom(isPlayingAtom)

  const handlePlayPause = () => {
    setIsPlaying(!isPlaying)
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
        <span className={'text-text-foreground/85 text-xs'}>0:43</span>
        <div className={'flex-1'}>
          <div className={'bg-foreground/20 h-1 w-full rounded-full'}>
            <div className={'bg-foreground h-1 w-1/6 rounded-full'} />
          </div>
        </div>
        <span className={'text-text-foreground/85 text-xs'}>4:13</span>
      </div>

      {/* Player Controls */}
      <div className={'flex flex-wrap items-center justify-between'}>
        <button
          className={
            'text-text-foreground/70 hover:text-text-foreground flex size-10 cursor-pointer items-center justify-center rounded-full transition-colors'
          }
          aria-label={'Previous'}
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
        >
          <span className={'i-mingcute-skip-forward-fill size-6'} />
        </button>

        <button
          className={
            'text-text-foreground/70 hover:text-text-foreground flex size-10 cursor-pointer items-center justify-center rounded-full transition-colors'
          }
          aria-label={'Shuffle'}
        >
          <span className={'i-mingcute-shuffle-fill size-5'} />
        </button>

        <button
          className={
            'text-text-foreground/70 hover:text-text-foreground flex size-10 cursor-pointer items-center justify-center rounded-full transition-colors'
          }
          aria-label={'Repeat'}
        >
          <span className={'i-mingcute-repeat-fill size-5'} />
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
