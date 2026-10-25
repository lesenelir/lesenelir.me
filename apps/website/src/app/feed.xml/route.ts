import { getAllPosts, getPostDescription } from '@/lib/mdx'

import { siteConfig } from '@/constants'

// Built once with the site; posts only change with a deploy
export const dynamic = 'force-static'

const escapeXml = (text: string) =>
  text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')

// RSS 2.0 with the posts' descriptions; readers link through to the full post
export function GET() {
  const items = getAllPosts().map(post => {
    const url = `${siteConfig.url}/posts/${post.id}`
    return [
      '<item>',
      `<title>${escapeXml(post.metadata.title)}</title>`,
      `<link>${url}</link>`,
      `<guid isPermaLink="true">${url}</guid>`,
      `<pubDate>${new Date(post.metadata.date).toUTCString()}</pubDate>`,
      `<description>${escapeXml(getPostDescription(post.id))}</description>`,
      '</item>',
    ].join('')
  })

  const xml = [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">',
    '<channel>',
    `<title>${escapeXml(siteConfig.name)}</title>`,
    `<link>${siteConfig.url}</link>`,
    `<description>${escapeXml(siteConfig.description)}</description>`,
    '<language>zh-CN</language>',
    `<atom:link href="${siteConfig.url}/feed.xml" rel="self" type="application/rss+xml"/>`,
    ...items,
    '</channel>',
    '</rss>',
  ].join('\n')

  return new Response(xml, { headers: { 'Content-Type': 'application/rss+xml; charset=utf-8' } })
}
