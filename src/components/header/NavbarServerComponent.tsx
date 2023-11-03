import Link from 'next/link'

import type { THeaderList } from '@/components/header/NavbarList'

interface IProps {
  item: THeaderList
  isActive: boolean
}

export default function NavbarServerComponent(props: IProps) {
  const {item, isActive} = props

  return (
    <>
      <li className={`header-li ${isActive ? '' : 'header-li-unActive'}`}>
        {
          isActive ? (
            <p>{item.name}</p>
          ) : (
            // [Bug]: Not working width Link component
            <Link href={item.path}>
              <p>{item.name}</p>
            </Link>
          )
        }
      </li>
    </>
  )
}
