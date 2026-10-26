import { getAllPosts, getPostMetadata } from '@/lib/mdx'
import { renderOgImage } from '@/lib/og'

import { siteConfig } from '@/constants'

export { ogSize as size } from '@/lib/og'

export const contentType = 'image/png'
export const alt = 'The post title, date and reading time'

// Every post's image is made at build time; any other id is a 404
export function generateStaticParams() {
  return getAllPosts().map(post => ({ id: post.id }))
}

export default async function Image({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const metadata = getPostMetadata(id)
  if (!metadata) return new Response('Not Found', { status: 404 })

  const date = new Date(metadata.date).toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })
  return renderOgImage({ title: metadata.title, footer: `${siteConfig.name} · ${date} · ${metadata.readTime}` })
}
