import React from 'react'

import { ThemeProvider } from 'next-themes'
import { Geist, Geist_Mono } from 'next/font/google'
import type { Metadata } from 'next'

import '@/styles/main.css'

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
  title: 'Lesenelir Zhou',
  description: "Lesenelir Zhou's personal website"
}

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={cn(geistSans.variable, geistMono.variable, 'bg-background antialiased')}>
        <ThemeProvider>{children}</ThemeProvider>
      </body>
    </html>
  )
}
