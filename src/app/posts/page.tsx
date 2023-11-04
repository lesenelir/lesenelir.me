import { Metadata } from 'next'

import Posts from '@/app/content/pages/posts.mdx'

export const metadata: Metadata = {
  title: 'Lesenelir Posts',
  description: 'Lesenelir Zhou Posts Page',
}

export default function PostPage() {
  return (
    <>
      <article className="text-justify prose max-md:prose-base dark:prose-invert">
        <Posts/>
      </article>
    </>
  )
}
