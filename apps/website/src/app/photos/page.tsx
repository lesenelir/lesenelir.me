import type { Metadata } from 'next'

import { getPhotos } from '@/lib/photos'

import { PhotoGallery } from '@/components/photos/photo-gallery'
import { PhotoProvider } from '@/components/photos/photo-provider'

// Re-list the bucket at most once an hour, so new photos show up without a redeploy
export const revalidate = 3600

export const metadata: Metadata = {
  title: 'Photos',
  description: "Lesenelir Zhou's personal website - Photos",
}

export default async function Page() {
  const photos = await getPhotos()

  return (
    <PhotoProvider photos={photos}>
      <h1 className={'mb-11 font-comic text-text-primary'}>Photos</h1>
      <PhotoGallery />
    </PhotoProvider>
  )
}
