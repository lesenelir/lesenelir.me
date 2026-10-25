import type { TPhoto } from '@/types'

import { getPublicUrl, listPhotos } from '@/lib/s3'

import { photoMetadata } from '@/constants'

export const getPhotos = async (): Promise<TPhoto[]> => {
  const s3Objects = await listPhotos()

  const photos: TPhoto[] = s3Objects.flatMap(obj => {
    if (!obj.Key) return []

    const fileName = obj.Key.replace('photos/', '')
    const id = fileName.replace(/\.[^.]+$/i, '')

    return [
      {
        id,
        src: getPublicUrl(obj.Key),
        alt: id,
        desc: photoMetadata[id]?.desc ?? '',
      },
    ]
  })

  return photos.reverse()
}
