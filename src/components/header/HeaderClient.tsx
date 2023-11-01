'use client'

import { usePathname } from 'next/navigation'
import { postData } from '@/app/content/data/PostLists'

export default function HeaderClient() {
  const pathname = usePathname()

  if (pathname === '/') {
    return (
      <>Lesenelir Zhou</>
    )
  }

  if (pathname === '/posts') {
    return (
      <>Posts</>
    )
  }

  if (pathname.startsWith('/posts/')) {
    const data = postData.filter(item => item.href === pathname)[0]

    return (
      <>
        <p>{data.name}</p>
        <p className={'text-xs text-textLight opacity-70 mt-2 dark:text-textDark'}>{data.date}</p>
      </>
    )
  }

  if (pathname === '/projects') {
    return (
      <>Projects</>
    )
  }

  if (pathname === '/photos') {
    return (
      <>Photos</>
    )
  }

  return (
    <>
      404
    </>
  )
}
