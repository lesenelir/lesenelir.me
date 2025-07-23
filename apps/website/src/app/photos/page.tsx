import type { Metadata } from 'next'

import { PhotoGallery } from '@/components/photos/photo-gallery'
import { getPhotos } from '@/constants'

export const metadata: Metadata = {
  title: 'Photos',
  description: "Lesenelir Zhou's personal website - Photos"
}

export default async function Page() {
  const photos = await getPhotos()

  console.log('page', photos)

  return (
    <>
      <h3 className={'font-comic text-text-primary mb-11'}>Photos</h3>
      <PhotoGallery photos={photos} />
    </>
  )
}
