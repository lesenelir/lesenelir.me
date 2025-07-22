import type { Metadata } from 'next'

import { DynamicTitle } from '@/components/songs/dynamic-title'
import { SongLists } from '@/components/songs/song-lists'
import { SongPlayer } from '@/components/songs/song-player'
import { SongsProvider } from '@/components/songs/songs-provider'
import { getSongs } from '@/constants'

export const metadata: Metadata = {
  title: 'Songs',
  description: "Lesenelir Zhou's personal website - Songs"
}

export default async function Page() {
  const songs = await getSongs()

  return (
    <SongsProvider songs={songs}>
      <DynamicTitle />
      <h3 className={'font-comic text-text-primary mb-11'}>Songs</h3>

      <SongPlayer />
      <SongLists />
    </SongsProvider>
  )
}
