import type { Metadata } from 'next'
import { notFound } from 'next/navigation'

import { getAllPosts, getPostComponent, getPostDescription, getPostMetadata } from '@/lib/mdx'

import { siteConfig } from '@/constants'

type TProps = {
  params: Promise<{ id: string }>
}

// Every post is prerendered at build time; any other id is a 404
export const dynamicParams = false

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

  return (
    <article className={'text-text-primary/85'}>
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

      <div lang={'zh-CN'} className={'prose prose-neutral max-w-none'}>
        <PostComponent />
      </div>
    </article>
  )
}
