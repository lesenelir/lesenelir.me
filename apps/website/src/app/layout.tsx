import React from 'react'
import { Analytics } from '@vercel/analytics/react'

import './globals.css'
import Favicon from '/public/favicon.ico'

import Mask from '@/components/header/Mask'
import Footer from '@/components/ui/Footer'
import Divide from '@/components/utils/Divide'
import Provider from '@/components/theme/Provider'
import NavbarList from '@/components/header/NavbarList'
import HeaderClient from '@/components/header/HeaderClient'
import MotionWrapper from '@/components/utils/motionWrapper'

interface IProps {
  children: React.ReactNode
}

export const metadata = {
  icons: [{ rel: 'icon', url: Favicon.src }]
}

export default function RootLayout(props: IProps) {
  const {children} = props

  return (
    <html lang='en' suppressHydrationWarning={true}>
      <body
        suppressHydrationWarning={true}
        className={'custom-scrollbar bg-gray-50 min-h-screen font-sans dark:bg-bgDark transition-all duration-300 ease-in'}
      >
        <Provider>
          <Mask/>
          <div className={'max-w-screen-sm mx-auto md:py-20 max-md:py-20 max-md:px-4'}>
            {/* Header */}
            <header className={'mb-10'}>
              <h1 className={'text-3xl font-comic text-black dark:text-white max-md:text-2xl'}>
                <HeaderClient/>
              </h1>
              <ul className={'header-ul max-md:text-base'}>
                <NavbarList/>
              </ul>
            </header>

            {/* Content */}
            <article className="mb-12">
              <MotionWrapper>
                {children}
              </MotionWrapper>
            </article>

            <Divide/>

            {/* Footer */}
            {/* Margin takes the larger value, so use padding. mb-12 + pt-4 */}
            <footer className="pt-4">
              <Footer/>
            </footer>
          </div>

          {/* Analytics */}
          <Analytics />
        </Provider>
      </body>
    </html>
  )
}
