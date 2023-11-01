import PageComponent from '@/app/posts/[id]/page-component'

export default function page() {
  return (
    <>
      <article className="custom-typography text-justify prose max-md:prose-stone dark:prose-invert">
        <PageComponent/>
      </article>
    </>
  )
}
