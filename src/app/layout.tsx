import React from 'react'
import type { Metadata } from 'next'

import Footer from '@/components/ui/Footer'
import Divide from '@/components/utils/Divide'
import NavbarList from '@/components/header/NavbarList'
import Provider from '@/components/theme/Provider'
import HeaderClient from '@/components/header/HeaderClient'

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
      <body
        suppressHydrationWarning={true}
        className={'bg-gray-50 min-h-screen font-sans dark:bg-bgDark transition-all duration-300 ease-in'}
      >
        <Provider>
          <div className={'max-w-screen-sm mx-auto md:py-20 max-md:py-20 max-md:px-4'}>
            {/* Header */}
            <header className={'mb-12'}>
              <h1 className={'text-3xl font-comic text-black dark:text-white max-md:text-2xl'}>
                <HeaderClient/>
              </h1>
              <ul className={'header-ul max-md:text-base'}>
                <NavbarList/>
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
