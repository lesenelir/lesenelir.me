import { getPublicUrl, listSongs } from '@/lib/s3'
import type { TSong } from '@/types'

export const getSongs = async (): Promise<TSong[]> => {
  try {
    const s3Objects = await listSongs()

    return s3Objects.flatMap((obj) => {
      if (!obj.Key) return []

      let fileName = obj.Key.replace('songs/', '')
      fileName = fileName.replace(/\.(mp3|ogg|webm|m4a)$/i, '')
      fileName = fileName.replace(/^\d+\.\s*/, '')

      const [artist, ...titleParts] = fileName.split(' - ')
      const title = titleParts.join(' - ')

      return [
        {
          id: fileName,
          src: getPublicUrl(obj.Key),
          title: title || fileName,
          artist: artist || 'Unknown'
        }
      ]
    })
  } catch (error) {
    console.error('Error fetching songs:', error)
    return []
  }
}
