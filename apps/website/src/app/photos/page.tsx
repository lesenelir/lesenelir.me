import fs from 'node:fs'
import path from 'node:path'

import type { Metadata } from 'next'

import { PhotoGallery } from '@/components/photos/photo-gallery'

export const metadata: Metadata = {
  title: 'Photos',
  description: "Lesenelir Zhou's personal website - Photos"
}

function getPhotos() {
  const photosDir = path.join(process.cwd(), 'public', 'photos')
  const files = fs.readdirSync(photosDir)

  return files
    .filter((file) => file.endsWith('.JPG'))
    .map((file) => {
      const filename = path.parse(file).name
      return {
        id: filename,
        src: `/photos/${file}`,
        alt: filename
      }
    })
    .reverse()
}

export default function Page() {
  const photos = getPhotos()

  return (
    <>
      <h3 className={'font-comic text-text-primary mb-11'}>Photos</h3>
      <PhotoGallery photos={photos} />
    </>
  )
}

export async function generateStaticParams() {
  return [{}]
}
