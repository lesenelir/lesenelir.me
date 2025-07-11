import Link from 'next/link'
import type { Metadata } from 'next'

import { getAllPosts } from '@/lib/mdx'

export const metadata: Metadata = {
  title: 'Posts',
  description: "Lesenelir Zhou's personal website - Posts"
}

export default function Page() {
  const posts = getAllPosts()

  return (
    <>
      <h3 className={'font-comic text-text-primary mb-11'}>Posts</h3>

      <div className={'text-text-primary/85 flex flex-col gap-3'}>
        {posts.map((post) => (
          <Link
            key={post.id}
            href={`/posts/${post.id}`}
            className={'group flex items-center justify-between'}
          >
            <span className={'group-hover:text-text-primary transition-all duration-200'}>
              {post.metadata.title}
            </span>
            <div
              className={
                'text-text-primary/60 group-hover:text-text-primary flex items-center gap-1.5 text-sm transition-all duration-200'
              }
            >
              <time dateTime={post.metadata.date}>
                {new Date(post.metadata.date).toLocaleDateString('en-US', {
                  month: 'short',
                  day: 'numeric',
                  year: 'numeric'
                })}
              </time>
              <span>·</span>
              <span>{post.metadata.readTime}</span>
            </div>
          </Link>
        ))}
      </div>

      {posts.length === 0 && (
        <div className={'text-text-primary/85 p-12 text-center'}>
          <p>No posts yet. Coming soon!</p>
        </div>
      )}
    </>
  )
}
