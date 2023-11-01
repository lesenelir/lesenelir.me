import PageComponent from '@/app/posts/[id]/page-component'
import Cd from '@/components/utils/Cd'

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
