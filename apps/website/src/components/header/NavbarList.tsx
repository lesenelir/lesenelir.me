'use client'

import { usePathname } from 'next/navigation'
import NavbarServerComponent from '@/components/header/NavbarServerComponent'

export type THeaderList = {
  name: string
  path: string
}

export const headerList: THeaderList[] = [
  {
    name: 'Home',
    path: '/',
  },
  {
    name: 'Posts',
    path: '/posts',
  },
  {
    name: 'Projects',
    path: '/projects',
  },
  {
    name: 'Photos',
    path: '/photos',
  }
]

export default function NavbarList() {
  const pathname = usePathname()

  if (pathname.includes('/posts/')) {
    return null
  }

  return (
    <>
      {
        headerList.map(item => (
          <NavbarServerComponent
            key={item.name}
            item={item}
            isActive={pathname === item.path}
          />
        ))
      }

      {/* Client Component Logic */}
      {/*{*/}
      {/*  headerList.map(item => (*/}
      {/*    <li*/}
      {/*      key={item.name}*/}
      {/*      className={`header-li ${pathname === item.path ? '' : 'header-li-unActive'}`}*/}
      {/*    >*/}
      {/*      {*/}
      {/*        pathname === item.path ? (*/}
      {/*          <p>{item.name}</p>*/}
      {/*        ) : (*/}
      {/*          <Link prefetch={false} href={item.path}>*/}
      {/*            {item.name}*/}
      {/*          </Link>*/}
      {/*        )*/}
      {/*      }*/}
      {/*    </li>*/}
      {/*  ))*/}
      {/*}*/}
    </>
  )
}
