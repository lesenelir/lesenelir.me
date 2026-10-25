import type { MetadataRoute } from 'next'

import { getAllPosts } from '@/lib/mdx'

import { siteConfig } from '@/constants'

export default function sitemap(): MetadataRoute.Sitemap {
  const posts = getAllPosts()

  return [
    { url: siteConfig.url },
    { url: `${siteConfig.url}/posts`, lastModified: posts[0]?.metadata.date },
    { url: `${siteConfig.url}/photos` },
    { url: `${siteConfig.url}/songs` },
    ...posts.map(post => ({ url: `${siteConfig.url}/posts/${post.id}`, lastModified: post.metadata.date })),
  ]
}
