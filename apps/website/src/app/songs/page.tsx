import type { Metadata } from 'next'

import { getSongs } from '@/lib/songs'

import { DynamicTitle } from '@/components/songs/dynamic-title'
import { SongLists } from '@/components/songs/song-lists'
import { SongPlayer } from '@/components/songs/song-player'
import { SongsProvider } from '@/components/songs/songs-provider'

// Re-list the bucket at most once an hour, so new songs show up without a redeploy
export const revalidate = 3600

export const metadata: Metadata = {
  title: 'Songs',
  description: "Lesenelir Zhou's personal website - Songs",
}

export default async function Page() {
  const songs = await getSongs()

  return (
    <SongsProvider songs={songs}>
      <DynamicTitle />
      <h1 className={'mb-11 font-comic text-text-primary'}>Songs</h1>

      <SongPlayer />
      <SongLists />
    </SongsProvider>
  )
}
