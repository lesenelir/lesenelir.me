"use client"

import React from 'react'
import { ThemeProvider } from 'next-themes'

interface IProps {
  children: React.ReactNode
}

export default function Provider(props: IProps) {
  const {children} = props

  return (
    <ThemeProvider enableSystem={true} defaultTheme={'system'}>
      {children}
    </ThemeProvider>
  )
}
