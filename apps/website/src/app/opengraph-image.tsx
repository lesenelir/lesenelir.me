import { renderOgImage } from '@/lib/og'

import { siteConfig } from '@/constants'

export { ogSize as size } from '@/lib/og'

export const contentType = 'image/png'
export const alt = siteConfig.name

// The link preview of every page that has none of its own
export default function Image() {
  return renderOgImage({ title: siteConfig.name, footer: 'Software engineer in Beijing' })
}
