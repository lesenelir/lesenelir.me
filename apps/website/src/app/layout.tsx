import React from 'react'

import { Geist, Geist_Mono } from 'next/font/google'
import type { Metadata } from 'next'

import '@/styles/main.css'

import { Providers } from '@/components/common/providers'
import { Sidebar } from '@/components/common/sidebar'
import { cn } from '@/lib/utils'

const geistSans = Geist({
  variable: '--font-geist-sans',
  subsets: ['latin']
})

const geistMono = Geist_Mono({
  variable: '--font-geist-mono',
  subsets: ['latin']
})

export const metadata: Metadata = {
  title: {
    default: 'Lesenelir Zhou',
    template: '%s - Lesenelir Zhou'
  },
  description: "Lesenelir Zhou's personal website",
  authors: [
    {
      name: 'Lesenelir Zhou',
      url: 'https://lesenelir.me'
    },
    {
      name: 'Lesenelir Zhou',
      url: 'https://github.com/lesenelir'
    }
  ],
  creator: 'Lesenelir Zhou',
  publisher: 'Lesenelir Zhou'
}

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={cn(
          geistSans.variable,
          geistMono.variable,
          'bg-background flex p-16 antialiased'
        )}
      >
        <Providers>
          <Sidebar />
          {children}
        </Providers>
      </body>
    </html>
  )
}
