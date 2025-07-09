import Link from 'next/link'

import { sidebarData } from '@/constants'

export function Sidebar() {
  return (
    <aside className={'mr-16 flex flex-col gap-4'}>
      {sidebarData.map((item) => (
        <Link key={item.id} href={item.href}>
          {item.title}
        </Link>
      ))}
    </aside>
  )
}
