import { imageSize } from 'image-size'

import type { TPhoto } from '@/types'

import { getObjectStart, getPublicUrl, listPhotos } from '@/lib/s3'

import { photoMetadata } from '@/constants'

// The dimensions sit in the JPEG header, after the camera's EXIF preview: about 100 KB into a 10-15 MB file. A
// photo with a bigger preview or more metadata (e.g. exported from Lightroom) gets a second, larger read.
const HEADER_BYTES = [256 * 1024, 2 * 1024 * 1024]

// Used when no header can be parsed: the camera's 3:2 landscape
const FALLBACK_DIMENSIONS = { width: 6000, height: 4000 }

const getPhotoDimensions = async (key: string) => {
  let lastError: unknown

  for (const bytes of HEADER_BYTES) {
    const header = await getObjectStart(key, bytes)
    try {
      const { width, height, orientation } = imageSize(header)
      // EXIF orientations 5-8 are turned a quarter, and browsers display them that way
      return orientation && orientation >= 5 ? { width: height, height: width } : { width, height }
    } catch (error) {
      lastError = error
    }
  }

  console.warn(`Could not read the dimensions of ${key}:`, lastError)
  return FALLBACK_DIMENSIONS
}

export const getPhotos = async (): Promise<TPhoto[]> => {
  const s3Objects = await listPhotos()

  const photos: TPhoto[] = await Promise.all(
    s3Objects.flatMap(obj => {
      const key = obj.Key
      if (!key) return []

      const fileName = key.replace('photos/', '')
      const id = fileName.replace(/\.[^.]+$/i, '')

      return [
        getPhotoDimensions(key).then(({ width, height }) => ({
          id,
          src: getPublicUrl(key),
          // The caption tells a screen reader what the photo is; the filename (R0000466) doesn't
          alt: photoMetadata[id]?.desc || id,
          desc: photoMetadata[id]?.desc ?? '',
          width,
          height,
        })),
      ]
    })
  )

  return photos.reverse()
}
