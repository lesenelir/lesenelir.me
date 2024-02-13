import Image from 'next/image'
import { useTheme } from 'next-themes'

interface IProps {
  dark: {
    src: string
    alt: string
  }
  light: {
    src: string
    alt: string
  }
}

export default function ImageInMDX(props: IProps) {
  const { theme } = useTheme()
  const { dark, light } = props

  return (
    <>
      {
        theme === 'dark' ? (
          <Image
            src={dark.src}
            alt={dark.alt}
          />
        ) : (
          <Image
            src={light.src}
            alt={light.alt}
          />
        )
      }
    </>
  )
}
