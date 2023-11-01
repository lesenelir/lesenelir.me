'use client'

import { useParams } from 'next/navigation'
import dynamic from 'next/dynamic'

export default function PageComponent() {
  const params = useParams()
  const id = params.id as string // { id: 'time-stop' }

  // Dynamic import MDX content
  const Content = dynamic(() => import(`../../content/posts/${id}.mdx`))

  return <Content />
}
