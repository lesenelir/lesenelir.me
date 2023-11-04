import RSS from 'rss'

import { postData } from '@/app/content/data/PostLists'

export async function GET() {
  const feed = new RSS({
    title: 'Lesenelir Zhou',
    description: 'Lesenelir Posts',
    site_url: 'https://lesenelir-cc.vercel.app', // 你的网站域名
    feed_url: 'https://lesenelir-cc.vercel.app/feed.xml', // 尽可能用绝对 URL
    language: 'en-US',
  })

  postData.forEach(post => {
    feed.item({
      title: post.name,
      url: `https://lesenelir-cc.vercel.app${post.href}`,
      description: post.name,
      date: post.date,
      // enclosure: {
      //   url: post.imageUrl
      // }
    })
  })

  return new Response(feed.xml(), {
    headers: {
      'content-type': 'application/xml'
    }
  })
}
