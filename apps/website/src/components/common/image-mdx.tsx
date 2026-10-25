import type { ImageProps } from 'next/image'
import Image from 'next/image'

import { cn } from '@/lib/utils'

interface ImageMdxProps extends Omit<ImageProps, 'src'> {
  lightSrc: string
  darkSrc: string
}

// Both images are in the server HTML and CSS shows the one for the theme, so the image holds its place from the
// first paint. next/image lazy-loads, and the hidden one (display: none) is never fetched.
// No `sizes`: these are flat-color diagrams, and the optimizer's full-size WebP is smaller than any downscaled one
// (jotai-light: 25 KB in full, 78 KB at 1200px), so next/image's default of the full-size file is the right pick
export function ImageMdx({ lightSrc, darkSrc, alt, className, ...restProps }: ImageMdxProps) {
  return (
    <>
      <Image src={lightSrc} alt={alt} className={cn(className, 'dark:hidden')} {...restProps} />
      <Image src={darkSrc} alt={alt} className={cn(className, 'hidden dark:block')} {...restProps} />
    </>
  )
}
