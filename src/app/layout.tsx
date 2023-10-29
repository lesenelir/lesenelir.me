import React from 'react'
import type { Metadata } from 'next'

import Footer from '@/components/ui/Footer'
import Divide from '@/components/utils/Divide'

import './globals.css'
import Link from 'next/link'

export const metadata: Metadata = {
  title: 'Lesenelir 3000',
  description: 'Lesenelir Zhou personal website',
}

interface IProps {
  children: React.ReactNode
}

export default function RootLayout(props: IProps) {
  const {children} = props

  return (
    <html lang="en">
    <body className="bg-gray-100 min-h-screen font-sans">
      <div className="max-w-screen-md my-20 mx-auto px-6 max-md:px-4">
        {/* Header */}
        <header className="pb-4 mb-8">
          <h1 className="text-3xl font-comic text-black dark:text-white">
            <Link href={'/'}>Lesenelir Zhou</Link>
          </h1>
          <ul className="flex flex-row justify-end gap-4 text-navbarLight underline underline-offset-1 dark:text-navbarDark">
            <li className="header-li">
              <Link href={'/posts'}>Posts</Link>
            </li>
            <li className="header-li">
              <Link href={'/projects'}>Projects</Link>
            </li>
            <li className="header-li">
              <Link href={'/photos'}>Photos</Link>
            </li>
          </ul>
        </header>

        {/* Content */}
        <article className="mb-8 text-textLight dark:text-textDark">
          {children}
        </article>

        <Divide/>

        {/* Footer */}
        <footer className="pt-4">
          <Footer/>
        </footer>
      </div>
    </body>
    </html>
  )
}
