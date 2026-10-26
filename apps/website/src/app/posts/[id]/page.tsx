import type { Metadata } from 'next'
import Link from 'next/link'
import { notFound } from 'next/navigation'

import { getAllPosts, getPostComponent, getPostDescription, getPostHeadings, getPostMetadata } from '@/lib/mdx'

import { PostArticle } from '@/components/posts/post-article'
import { TableOfContents } from '@/components/posts/table-of-contents'

import { siteConfig } from '@/constants'

// A table of contents is only worth it for a post with a few sections
const MIN_TOC_HEADINGS = 2

type TProps = {
  params: Promise<{ id: string }>
}

// Every post is prerendered at build time; any other id renders on request, as a 404
export function generateStaticParams() {
  return getAllPosts().map(post => ({ id: post.id }))
}

export async function generateMetadata({ params }: TProps): Promise<Metadata> {
  const { id } = await params
  const metadata = getPostMetadata(id)

  if (!metadata) {
    return {
      title: 'Post Not Found',
    }
  }

  const description = getPostDescription(id)
  const url = `/posts/${id}`

  return {
    title: metadata.title,
    description,
    alternates: { canonical: url },
    // Replaces the root layout's openGraph as a whole, so siteName is repeated
    openGraph: {
      type: 'article',
      siteName: siteConfig.name,
      title: metadata.title,
      description,
      url,
      locale: 'zh_CN',
      publishedTime: metadata.date,
      authors: [siteConfig.url],
    },
  }
}

// No loading shell: a post opens whole, so a link that hasn't prefetched it waits on the page it is on
export const instant = false

export default async function Page({ params }: TProps) {
  const { id } = await params
  const metadata = getPostMetadata(id)
  const PostComponent = await getPostComponent(id)

  if (!metadata || !PostComponent) {
    notFound()
  }

  // Structured data for search engines: the post as a BlogPosting by its author
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'BlogPosting',
    headline: metadata.title,
    description: getPostDescription(id),
    datePublished: metadata.date,
    inLanguage: 'zh-CN',
    url: `${siteConfig.url}/posts/${id}`,
    image: `${siteConfig.url}/posts/${id}/opengraph-image`,
    author: { '@type': 'Person', name: siteConfig.name, url: siteConfig.url, sameAs: [siteConfig.github] },
  }

  const headings = getPostHeadings(id)
  const showToc = headings.length >= MIN_TOC_HEADINGS

  // Posts are newest first, so the one before this in the list is newer
  const posts = getAllPosts()
  const index = posts.findIndex(post => post.id === id)
  const newer = posts[index - 1]
  const older = posts[index + 1]

  return (
    <PostArticle className={'text-text-primary/85'}>
      {/* < is escaped so a title can never close the script element */}
      <script type={'application/ld+json'}>{JSON.stringify(jsonLd).replace(/</g, '\\u003c')}</script>

      <header className={'mb-8 space-y-4'}>
        {/* Posts are written in Chinese; the rest of the site is English */}
        <h1 lang={'zh-CN'} className={'font-comic text-2xl text-text-primary'}>
          {metadata.title}
        </h1>

        <div className={'flex items-center gap-3 text-sm text-text-primary/70'}>
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

      {showToc && (
        <>
          {/* Wide screens: beside the post, at the right edge of the page, following the reader down. Sticky in a
              column as tall as the post rather than fixed: while the page's enter animation (a filter) plays, a fixed
              element is placed against the page instead of the window, and jumped when it ended. */}
          <div className={'absolute inset-y-0 right-0 hidden w-56 xl:block'}>
            <TableOfContents headings={headings} className={'sticky top-16 max-h-[calc(100vh-8rem)] overflow-y-auto'} />
          </div>
          {/* Narrower screens: folded up above the post */}
          <details className={'mb-8 text-sm xl:hidden'}>
            <summary className={'cursor-pointer text-text-foreground hover:text-text-primary'}>Contents</summary>
            <TableOfContents headings={headings} className={'mt-3 border-l border-dividing pl-4'} />
          </details>
        </>
      )}

      <div lang={'zh-CN'}>
        <PostComponent />
      </div>

      {(older || newer) && (
        <nav
          aria-label={'More posts'}
          className={'mt-16 flex justify-between gap-8 border-t border-dividing pt-8 text-sm'}
        >
          {older ? (
            <Link href={`/posts/${older.id}`} prefetch className={'group min-w-0'}>
              <span className={'block text-text-foreground'}>← Previous</span>
              <span lang={'zh-CN'} className={'block text-text-primary/85 group-hover:text-text-primary'}>
                {older.metadata.title}
              </span>
            </Link>
          ) : (
            <span />
          )}
          {newer && (
            <Link href={`/posts/${newer.id}`} prefetch className={'group min-w-0 text-right'}>
              <span className={'block text-text-foreground'}>Next →</span>
              <span lang={'zh-CN'} className={'block text-text-primary/85 group-hover:text-text-primary'}>
                {newer.metadata.title}
              </span>
            </Link>
          )}
        </nav>
      )}
    </PostArticle>
  )
}
