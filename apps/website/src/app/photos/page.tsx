import type { Metadata } from 'next'

import { PhotoGallery } from '@/components/photos/photo-gallery'
import { PhotoProvider } from '@/components/photos/photo-provider'
import { getPhotos } from '@/constants'

export const metadata: Metadata = {
  title: 'Photos',
  description: "Lesenelir Zhou's personal website - Photos"
}

export default async function Page() {
  const photos = await getPhotos()

  return (
    <PhotoProvider photos={photos}>
      <h3 className={'font-comic text-text-primary mb-11'}>Photos</h3>
      <PhotoGallery />
    </PhotoProvider>
  )
}
