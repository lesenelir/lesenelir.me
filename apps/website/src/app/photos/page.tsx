import type { Metadata } from 'next'

import { PhotoGallery } from '@/components/photos/photo-gallery'
import { PhotoProvider } from '@/components/photos/photo-provider'

import { getPhotos } from '@/constants'

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
      <h3 className={'mb-11 font-comic text-text-primary'}>Photos</h3>
      <PhotoGallery />
    </PhotoProvider>
  )
}
