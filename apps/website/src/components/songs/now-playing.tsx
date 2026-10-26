'use client'

import { useAtomValue, useSetAtom } from 'jotai'

import { cn } from '@/lib/utils'

import { audioControlsAtom, currentSongAtom, isBufferingAtom, isPlayingAtom } from '@/atoms/songs'

// Each bar's height while paused, and its delay while playing, staggered so the bars move out of step. The delay is
// an inline style: the animate-* utility's animation shorthand would reset a delay class.
const bars = [
  { rest: 'scale-y-40', delay: '-0.6s' },
  { rest: 'scale-y-70', delay: '-0.3s' },
  { rest: 'scale-y-50', delay: '-0.45s' },
]

// Three bars beside "songs" in the sidebar once a song has been picked: moving while it plays, still while paused.
// Pressing them pauses or resumes, from any page.
export function NowPlaying() {
  const currentSong = useAtomValue(currentSongAtom)
  const isPlaying = useAtomValue(isPlayingAtom)
  const isBuffering = useAtomValue(isBufferingAtom)
  const setAudioControls = useSetAtom(audioControlsAtom)

  if (!currentSong) return null

  // Loading counts as playing, as for the player's own button
  const isActive = isPlaying || isBuffering
  const song = `${currentSong.title} by ${currentSong.artist}`

  return (
    // /80 is the lightest these bars can be and still stand out 3:1 from the background, in both themes
    <button
      type='button'
      onClick={() => setAudioControls({ type: isActive ? 'pause' : 'continue' })}
      aria-label={`${isActive ? 'Pause' : 'Play'} ${song}`}
      title={`${currentSong.title} — ${currentSong.artist}`}
      className={
        'flex size-4 cursor-pointer items-end justify-center gap-0.5 pb-0.5 text-text-foreground/80 transition-colors duration-200 hover:text-text-primary'
      }
    >
      {bars.map(bar => (
        <span
          key={bar.delay}
          className={cn(
            'h-3 w-0.5 origin-bottom rounded-full bg-current',
            // With reduced motion, a playing song stands the bars up at full height instead
            isPlaying ? 'motion-safe:animate-equalizer' : bar.rest
          )}
          style={isPlaying ? { animationDelay: bar.delay } : undefined}
        />
      ))}
    </button>
  )
}
