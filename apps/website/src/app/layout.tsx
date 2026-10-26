import { Analytics } from '@vercel/analytics/next'
import type { Metadata, Viewport } from 'next'
import { Comic_Neue, EB_Garamond } from 'next/font/google'
import type React from 'react'

import '@/styles/main.css'

import { photosLayoutScript } from '@/lib/photos-layout'
import { cn } from '@/lib/utils'

import { PageTransition } from '@/components/common/page-transition'
import { Providers } from '@/components/common/providers'
import { Sidebar } from '@/components/common/sidebar'

import { siteConfig } from '@/constants'

// Stand-ins for Apple Chancery (the sidebar) and Comic Sans MS (page titles) on systems without them: Windows,
// Android, iPhone. They come after those fonts in --font-cursive and --font-comic (src/styles/tailwind.css), and a
// browser only downloads a web font it needs, so with preloading off a Mac never fetches them.
const ebGaramond = EB_Garamond({
  subsets: ['latin'],
  style: 'italic',
  variable: '--font-eb-garamond',
  preload: false,
})

const comicNeue = Comic_Neue({
  subsets: ['latin'],
  weight: '400',
  variable: '--font-comic-neue',
  preload: false,
})

// Pages inherit this, so it holds no canonical URL or og:title: each page sets its own canonical, and without an
// og:title, link previews use the page's <title>
export const metadata: Metadata = {
  metadataBase: new URL(siteConfig.url),
  title: {
    default: siteConfig.name,
    template: `%s - ${siteConfig.name}`,
  },
  description: siteConfig.description,
  authors: [{ name: siteConfig.name, url: siteConfig.url }],
  creator: siteConfig.name,
  publisher: siteConfig.name,
  openGraph: {
    type: 'website',
    siteName: siteConfig.name,
    locale: 'en_US',
  },
  twitter: { card: 'summary_large_image' },
}

// The browser UI matches the page background (src/styles/colors.css)
export const viewport: Viewport = {
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#fcfcfc' },
    { media: '(prefers-color-scheme: dark)', color: '#131413' },
  ],
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang='en' className={cn(ebGaramond.variable, comicNeue.variable)} suppressHydrationWarning>
      <head>
        <script>{photosLayoutScript}</script>
        {/* Here rather than in metadata.alternates, which a page's own alternates (its canonical URL) replace */}
        <link rel='alternate' type='application/rss+xml' title={siteConfig.name} href={'/feed.xml'} />
      </head>
      <body className={'bg-background antialiased max-sm:p-6 sm:p-16'} suppressHydrationWarning>
        <Providers>
          <Sidebar />
          <main
            className={'border-dividing/70 max-sm:mt-12 max-sm:border-t max-sm:pt-16 sm:ml-24 sm:border-l sm:pl-16'}
          >
            <PageTransition>{children}</PageTransition>
          </main>
        </Providers>
        <Analytics />
      </body>
    </html>
  )
}
