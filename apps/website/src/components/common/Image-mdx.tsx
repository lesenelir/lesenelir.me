'use client'

import { useEffect, useState } from 'react'

import { useTheme } from 'next-themes'
import Image from 'next/image'
import type { ImageProps } from 'next/image'

interface ImageMdxProps extends Omit<ImageProps, 'src'> {
  lightSrc: string
  darkSrc: string
}

export function ImageMdx(props: ImageMdxProps) {
  const { resolvedTheme } = useTheme()
  const { alt, width, height, className, lightSrc, darkSrc, ...restProps } = props
  const [mounted, setMounted] = useState(false)

  const imageSrc = resolvedTheme === 'light' ? lightSrc : darkSrc

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return null

  return (
    <Image
      src={imageSrc}
      alt={alt}
      width={width}
      height={height}
      className={className}
      {...restProps}
    />
  )
}
