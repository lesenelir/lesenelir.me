import { getPublicUrl, listSongs } from '@/lib/s3'
import type { TSong } from '@/types'

export const getSongs = async (): Promise<TSong[]> => {
  try {
    const s3Objects = await listSongs()

    return s3Objects.flatMap((obj) => {
      if (!obj.Key) return []

      const fileName = obj.Key.replace('songs/', '')
      const id = fileName.replace(/\.[^.]+$/i, '')
      // Remove track numbers (e.g., "01. ") from the beginning of the filename
      const cleanId = id.replace(/^\d+\.\s*/, '')

      const [artist, ...titleParts] = cleanId.split(' - ')
      const title = titleParts.join(' - ')

      return [
        {
          id,
          src: getPublicUrl(obj.Key),
          title: title || cleanId,
          artist: artist || 'Unknown'
        }
      ]
    })
  } catch (error) {
    console.error('Error fetching songs:', error)
    return []
  }
}
