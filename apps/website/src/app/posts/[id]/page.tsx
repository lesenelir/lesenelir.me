import { notFound } from 'next/navigation'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { Metadata } from 'next'

import { getPost } from '@/lib/mdx'

type TProps = {
  params: { id: string }
}

export async function generateMetadata({ params }: TProps): Promise<Metadata> {
  const post = getPost(params.id)

  if (!post) {
    return {
      title: 'Post Not Found'
    }
  }

  return {
    title: post.metadata.title
  }
}

export default function Page({ params }: TProps) {
  const post = getPost(params.id)

  if (!post) {
    notFound()
  }

  return (
    <article className={'text-text-primary/85'}>
      <header className={'mb-8 space-y-4'}>
        <h1 className={'font-comic text-text-primary text-2xl'}>{post.metadata.title}</h1>

        <div className={'text-text-primary/60 flex items-center gap-3 text-sm'}>
          <time dateTime={post.metadata.date}>
            {new Date(post.metadata.date).toLocaleDateString('en-US', {
              month: 'long',
              day: 'numeric',
              year: 'numeric'
            })}
          </time>
          <span>·</span>
          <span>{post.metadata.readTime}</span>
        </div>
      </header>

      <div className={'prose prose-neutral max-w-none'}>
        <ReactMarkdown remarkPlugins={[remarkGfm]}>{post.content}</ReactMarkdown>
      </div>
    </article>
  )
}
