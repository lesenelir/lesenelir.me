'use client'

import dynamic from 'next/dynamic'
import { usePathname } from 'next/navigation'

import { postData } from '@/app/content/data/PostLists'

export default function PageComponent() {
  const pathname = usePathname()

  const url = postData.filter(item => item.href === pathname)[0].file

  // Dynamic import MDX content
  const Content = dynamic(() => import(`../../content/posts/${url}.mdx`),  {
    loading: () => <p>Loading...</p>
  })

  return (
    <>
      <Content />
    </>
  )
}
