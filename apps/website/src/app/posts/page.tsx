import type { Metadata } from 'next'
import Link from 'next/link'

import { getAllPosts } from '@/lib/mdx'

export const metadata: Metadata = {
  title: 'Posts',
  description: "Lesenelir Zhou's personal website - Posts",
  alternates: { canonical: '/posts' },
}

export default function Page() {
  const posts = getAllPosts()

  return (
    <>
      <h1 className={'mb-11 font-comic text-text-primary'}>Posts</h1>

      <div className={'flex flex-col gap-3 text-md text-text-primary/85'}>
        {posts.map(post => (
          <Link
            key={post.id}
            href={`/posts/${post.id}`}
            prefetch
            className={'group flex items-start justify-between gap-4'}
          >
            <span
              lang={'zh-CN'}
              className={
                'min-w-0 flex-1 leading-relaxed wrap-break-word transition-all duration-200 group-hover:text-text-primary'
              }
            >
              {post.metadata.title}
            </span>
            <div
              className={
                'flex shrink-0 items-center gap-1.5 text-[13px] text-text-primary/70 transition-all duration-200 group-hover:text-text-primary'
              }
            >
              <time dateTime={post.metadata.date}>
                {new Date(post.metadata.date).toLocaleDateString('en-US', {
                  month: 'short',
                  day: 'numeric',
                  year: 'numeric',
                })}
              </time>
              <span>·</span>
              <span>{post.metadata.readTime}</span>
            </div>
          </Link>
        ))}
      </div>

      {posts.length === 0 && (
        <div className={'p-12 text-center text-text-primary/85'}>
          <p>No posts yet. Coming soon!</p>
        </div>
      )}
    </>
  )
}
