import { Metadata } from 'next'

import Cd from '@/components/utils/Cd'
import PageComponent from '@/app/posts/[id]/page-component'

export const metadata: Metadata = {
  title: '',
  description: '',
}

export default function page() {
  return (
    <>
      <article className="mb-12 custom-typography text-justify prose max-md:prose-stone dark:prose-invert">
        <PageComponent/>
      </article>

      <Cd/>
    </>
  )
}
