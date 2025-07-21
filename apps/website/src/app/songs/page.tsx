import type { Metadata } from 'next'

import { SongLists } from '@/components/songs/song-lists'
import { SongPlayer } from '@/components/songs/song-player'
import { getSongs } from '@/constants'

export const metadata: Metadata = {
  title: 'Songs',
  description: "Lesenelir Zhou's personal website - Songs"
}

export default async function Page() {
  const songs = await getSongs()

  return (
    <>
      <h3 className={'font-comic text-text-primary mb-11'}>Songs</h3>

      <SongPlayer />
      <SongLists songs={songs} />
    </>
  )
}
