import type { Metadata } from 'next'

import { getSongs } from '@/lib/songs'

import { DynamicTitle } from '@/components/songs/dynamic-title'
import { SongLists } from '@/components/songs/song-lists'
import { SongPlayer } from '@/components/songs/song-player'
import { HydrateSongs } from '@/components/songs/songs-provider'

export const metadata: Metadata = {
  title: 'Songs',
  description: "Lesenelir Zhou's personal website - Songs",
  alternates: { canonical: '/songs' },
}

export default async function Page() {
  const songs = await getSongs()

  return (
    <HydrateSongs songs={songs}>
      <DynamicTitle />
      <h1 className={'mb-11 font-comic text-text-primary'}>Songs</h1>

      <SongPlayer />
      <SongLists />
    </HydrateSongs>
  )
}
