import { getPublicUrl, listPhotos } from '@/lib/s3'
import type { TPhoto } from '@/types'

const photoMetadata: Record<string, { desc: string }> = {
  R0000466: { desc: '2025/07/01 (Tues) Nagoya, Japan @ Male Companion' },
  R0000454: { desc: '2025/07/01 (Tues) Nagoya, Japan @ Walking (2)' },
  R0000450: { desc: '2025/07/01 (Tues) Nagoya, Japan @ Walking (1)' },
  R0000443: { desc: '2025/07/01 (Tues) Nagoya, Japan @ Mirroring (2)' },
  R0000442: { desc: '2025/07/01 (Tues) Nagoya, Japan @ Mirroring (1)' },
  R0000438: { desc: '2025/07/01 (Tues) Nagoya, Japan @ Street' },
  R0000434: { desc: '2025/07/01 (Tues) Nagoya, Japan @ Railway' },
  R0000423: { desc: '2025/07/01 (Tues) Nagoya, Japan @ Bicycle' },
  R0000422: { desc: '2025/07/01 (Tues) Nagoya, Japan @ Street-side house' },
  R0000411: { desc: '2025/07/01 (Tues) Nagoya, Japan @ Nagoya Castle (2)' },
  R0000408: { desc: '2025/07/01 (Tues) Nagoya, Japan @ Nagoya Castle (1)' },
  R0000391: { desc: '2025/06/28 (Sat) Nagoya, Japan @ Ice Cube (2)' },
  R0000389: { desc: '2025/06/28 (Sat) Nagoya, Japan @ Ice Cube (1)' },
  R0000386: { desc: '2025/06/28 (Sat) Nagoya, Japan @ Japan Style' },
  R0000351: { desc: '2025/06/27 (Fri) Nagoya, Japan @ Weary' },
  R0000345: { desc: '2025/06/27 (Fri) Nagoya, Japan @ Lady of the Railway' },
  R0000332: { desc: '2025/02/06 (Thur) Guangzhou, China @ JNU School' },
  R0000297: { desc: '2025/02/05 (Wed) Hong Kong, China @ Reading' },
  R0000296: { desc: "2025/02/05 (Wed) Hong Kong, China @ Mr Jiang's Gesture" },
  R0000255: { desc: '2025/02/04 (Tues) Hong Kong, China @ The Crypto Shadow' },
  R0000235: { desc: '2025/02/04 (Tues) Hong Kong, China @ GitHub Wall' },
  R0000192: { desc: '2025/01/31 (Fri) Quzhou, China @ Give Happiness Chance' },
  R0000166: { desc: '2025/01/23 (Thur) Beijing, China @ Train Station (11)' },
  R0000164: { desc: '2025/01/23 (Thur) Beijing, China @ Train Station (10)' },
  R0000161: { desc: '2025/01/23 (Thur) Beijing, China @ Train Station (9)' },
  R0000160: { desc: '2025/01/23 (Thur) Beijing, China @ Train Station (8)' },
  R0000155: { desc: '2025/01/23 (Thur) Beijing, China @ Train Station (7)' },
  R0000154: { desc: '2025/01/23 (Thur) Beijing, China @ Train Station (6)' },
  R0000153: { desc: '2025/01/23 (Thur) Beijing, China @ Train Station (5)' },
  R0000138: { desc: '2025/01/23 (Thur) Beijing, China @ Train Station (4)' },
  R0000135: { desc: '2025/01/23 (Thur) Beijing, China @ Train Station (3)' },
  R0000112: { desc: '2025/01/23 (Thur) Beijing, China @ Train Station (2)' },
  R0000110: { desc: '2025/01/23 (Thur) Beijing, China @ Train Station (1)' },
  R0000093: { desc: '2025/01/23 (Thur) Beijing, China @ Bro' },
  R0000064: { desc: '2024/12/12 (Sun) Beijing, China @ Beneath Your Window' },
  R0000052: { desc: '2024/12/12 (Sun) Beijing, China @ Apples And Bananas Left Behind' },
  R0000041: { desc: '2024/12/12 (Sun) Beijing, China @ Yummy' }
}

export const getPhotos = async (): Promise<TPhoto[]> => {
  try {
    const s3Objects = await listPhotos()

    const photos: TPhoto[] = s3Objects.flatMap((obj) => {
      if (!obj.Key) return []

      const fileName = obj.Key.replace('photos/', '')
      const id = fileName.replace(/\.[^.]+$/i, '')

      return [
        {
          id,
          src: getPublicUrl(obj.Key),
          alt: id,
          desc: photoMetadata[id]?.desc ?? ''
        }
      ]
    })

    return photos.reverse()
  } catch (error) {
    console.error('Error fetching photos:', error)
    return []
  }
}
