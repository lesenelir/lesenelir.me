import { ImageResponse } from 'next/og'
import type { ReactElement } from 'react'

export const ogSize = { width: 1200, height: 630 }

// The light theme's background and text colors (src/styles/colors.css)
const colors = { background: '#fcfcfc', primary: '#303030', foreground: '#6a727d' }

// Satori only draws glyphs from the fonts it is given, and post titles are Chinese, so fetch Noto Sans SC from
// Google Fonts with just the characters drawn. Images are made at build time; a failed fetch fails the build rather
// than shipping a preview of empty boxes.
async function loadFont(text: string, weight: number) {
  const params = new URLSearchParams({ family: `Noto Sans SC:wght@${weight}`, text })
  const css = await fetch(`https://fonts.googleapis.com/css2?${params}`).then(res => res.text())
  const url = css.match(/src: url\((.+?)\) format\('(opentype|truetype)'\)/)?.[1]
  if (!url) throw new Error(`Noto Sans SC ${weight} not found in the Google Fonts response`)
  return fetch(url).then(res => res.arrayBuffer())
}

interface OgImageProps {
  title: string
  footer: string
}

export async function renderOgImage({ title, footer }: OgImageProps) {
  const site = 'lesenelir.me'
  const [regular, bold] = await Promise.all([loadFont(`${site}${footer}`, 400), loadFont(title, 600)])

  const element: ReactElement = (
    <div
      style={{
        width: '100%',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
        padding: 80,
        backgroundColor: colors.background,
        fontFamily: 'Noto Sans SC',
      }}
    >
      <div style={{ fontSize: 28, color: colors.foreground }}>{site}</div>
      <div style={{ fontSize: 68, fontWeight: 600, lineHeight: 1.3, color: colors.primary }}>{title}</div>
      <div style={{ fontSize: 28, color: colors.foreground }}>{footer}</div>
    </div>
  )

  return new ImageResponse(element, {
    ...ogSize,
    fonts: [
      { name: 'Noto Sans SC', data: regular, weight: 400, style: 'normal' },
      { name: 'Noto Sans SC', data: bold, weight: 600, style: 'normal' },
    ],
  })
}
