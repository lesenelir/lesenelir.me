import React from 'react'
import Link from 'next/link'
import type { Metadata } from 'next'

import Footer from '@/components/ui/Footer'
import Divide from '@/components/utils/Divide'
import Provider from '@/components/client/Provider'
import HeaderClient from '@/components/client/HeaderClient'

import './globals.css'

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
    <html lang='en' suppressHydrationWarning={true}>
      <body suppressHydrationWarning={true} className={'bg-gray-50 min-h-screen font-sans dark:bg-bgDark'}>
        <Provider>
          <div className={'max-w-screen-sm mx-auto md:py-20 max-md:py-20 max-md:px-4'}>
            {/* Header */}
            <header className={'mb-12'}>
              <h1 className={'text-3xl font-comic text-black dark:text-white max-md:text-xl'}>
                <HeaderClient/>
              </h1>
              <ul className={'header-ul max-md:text-xs'}>
                <li className={'header-li'}>
                  <Link href={'/'}>Home</Link>
                </li>
                <li className={'header-li'}>
                  <Link href={'/posts'}>Posts</Link>
                </li>
                <li className={'header-li'}>
                  <Link href={'/projects'}>Projects</Link>
                </li>
                <li className={'header-li'}>
                  <Link href={'/photos'}>Photos</Link>
                </li>
              </ul>
            </header>

            {/* Content */}
            <article className="mb-12">
              {children}
            </article>

            <Divide/>

            {/* Footer */}
            {/* Margin takes the larger value, so use padding. mb-12 + pt-4 */}
            <footer className="pt-4">
              <Footer/>
            </footer>
          </div>
        </Provider>
      </body>
    </html>
  )
}
