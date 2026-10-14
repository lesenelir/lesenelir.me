import type { Metadata } from 'next'
import { notFound } from 'next/navigation'

import { getPostComponent, getPostMetadata } from '@/lib/mdx'

type TProps = {
  params: Promise<{ id: string }>
}

export async function generateMetadata({ params }: TProps): Promise<Metadata> {
  const { id } = await params
  const metadata = getPostMetadata(id)

  if (!metadata) {
    return {
      title: 'Post Not Found',
    }
  }

  return {
    title: metadata.title,
  }
}

export default async function Page({ params }: TProps) {
  const { id } = await params
  const metadata = getPostMetadata(id)
  const PostComponent = await getPostComponent(id)

  if (!metadata || !PostComponent) {
    notFound()
  }

  return (
    <article className={'text-text-primary/85'}>
      <header className={'mb-8 space-y-4'}>
        <h1 className={'font-comic text-2xl text-text-primary'}>{metadata.title}</h1>

        <div className={'flex items-center gap-3 text-sm text-text-primary/60'}>
          <time dateTime={metadata.date}>
            {new Date(metadata.date).toLocaleDateString('en-US', {
              month: 'long',
              day: 'numeric',
              year: 'numeric',
            })}
          </time>
          <span>·</span>
          <span>{metadata.readTime}</span>
        </div>
      </header>

      <div className={'prose prose-neutral max-w-none'}>
        <PostComponent />
      </div>
    </article>
  )
}
