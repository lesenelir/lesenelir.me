import Posts from '@/app/content/pages/posts.mdx'

export default function PostPage() {
  return (
    <>
      <article className="text-justify prose max-md:prose-base dark:prose-invert">
        <Posts/>
      </article>
    </>
  )
}
