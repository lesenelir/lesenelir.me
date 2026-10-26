import type { Metadata } from 'next'

import { getPhotos } from '@/lib/photos'

import { PhotoGallery } from '@/components/photos/photo-gallery'
import { PhotoProvider } from '@/components/photos/photo-provider'

export const metadata: Metadata = {
  title: 'Photos',
  description: "Lesenelir Zhou's personal website - Photos",
  alternates: { canonical: '/photos' },
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
