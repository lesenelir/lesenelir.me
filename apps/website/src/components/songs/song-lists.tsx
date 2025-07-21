import type { TSong } from '@/types'

interface SongListsProps {
  songs: TSong[]
}

export function SongLists({ songs }: SongListsProps) {
  console.log('songs', songs)

  return <>song lists</>
}
