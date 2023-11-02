'use client'

import { usePathname } from 'next/navigation'

import Back from '@/components/utils/Back'
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
      <div className={'-mb-8'}>
        <p>{data.name}</p>
        <div className={'flex flex-row justify-between'}>
          <span className={'text-base text-textLight opacity-50 mt-2 dark:text-textDark'}>{data.date}</span>
          <span className={'text-base text-textLight opacity-70 mt-2 dark:text-textDark'}><Back/></span>
        </div>
      </div>
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
