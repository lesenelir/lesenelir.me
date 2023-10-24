import React from 'react'
import type { Metadata } from 'next'

import './globals.css'

export const metadata: Metadata = {
  title: 'Lesenelir Zhou',
  description: 'Lesenelir Zhou personal website',
}

interface IProps {
  children: React.ReactNode
}

export default function RootLayout(props: IProps) {
  const {children} = props

  return (
    <html lang="en">
      <body className={''}>{children}</body>
    </html>
  )
}
