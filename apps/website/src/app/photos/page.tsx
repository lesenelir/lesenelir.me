import fs from 'node:fs'
import path from 'node:path'

import type { Metadata } from 'next'

import { PhotoGallery } from '@/components/photos/photo-gallery'

export const metadata: Metadata = {
  title: 'Photos',
  description: "Lesenelir Zhou's personal website - Photos"
}

export default function Page() {
  const photosDir = path.join(process.cwd(), 'public', 'photos')
  const files = fs.readdirSync(photosDir)

  const photos = files
    .filter((file) => file.endsWith('.JPG'))
    .map((file) => {
      const filename = path.parse(file).name // extract the filename without extension
      return {
        id: filename,
        src: `/photos/${file}`,
        alt: filename,
        width: 600,
        height: 400
      }
    })
    .reverse()

  return (
    <>
      <h3 className={'font-comic text-text-primary mb-11'}>Photos</h3>
      <PhotoGallery photos={photos} />
    </>
  )
}
