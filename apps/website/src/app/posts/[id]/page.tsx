import type { Metadata } from 'next'

import Cd from '@/components/utils/Cd'
import { postData } from '@/app/content/data/PostLists'
import PageComponent from '@/app/posts/[id]/page-component'

type Props = {
  params: { id: string }
}

export const generateMetadata = ({ params }: Props): Metadata => {
  const title = postData.filter(item => item.href === `/posts/${params.id}`)[0].name

  return {
    title
  }
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
