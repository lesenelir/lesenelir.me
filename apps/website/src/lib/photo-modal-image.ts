import { getImageProps } from 'next/image'
import { preload } from 'react-dom'

import type { TPhoto } from '@/types'

// The modal shows a photo as large as the viewport allows: the full width when the viewport is narrower than the
// photo's aspect ratio, otherwise the full height, which makes it (width / height) × 100vh wide
export const getModalImageProps = (photo: TPhoto) => ({
  src: photo.src,
  width: photo.width,
  height: photo.height,
  sizes: `(max-aspect-ratio: ${photo.width}/${photo.height}) 100vw, ${Math.ceil((photo.width / photo.height) * 100)}vh`,
})

// Starts downloading a photo before the modal shows it. It asks for the same srcset and sizes as the modal's
// <Image>, so the browser picks the same file and the modal finds it in the cache.
export const preloadModalImage = (photo: TPhoto) => {
  const { props } = getImageProps({ ...getModalImageProps(photo), alt: photo.alt })
  preload(props.src, { as: 'image', imageSrcSet: props.srcSet, imageSizes: props.sizes })
}
