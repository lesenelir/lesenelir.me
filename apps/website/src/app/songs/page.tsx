import type { Metadata } from 'next'

import { getSongs } from '@/constants'

export const metadata: Metadata = {
  title: 'Songs',
  description: "Lesenelir Zhou's personal website - Songs"
}

export default async function Page() {
  const songs = await getSongs()

  console.log('songs', songs)

  return (
    <>
      <h3 className={'font-comic text-text-primary mb-11'}>Songs</h3>

      <span>WIP</span>
    </>
  )
}
