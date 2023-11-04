import nextMDX from '@next/mdx'
import rehypeHighlight from 'rehype-highlight'

const withMDX = nextMDX({
  extension: /\.mdx?$/,
  options: {
    remarkPlugins: [],
    rehypePlugins: [rehypeHighlight],
  }
})

/** @type {import('next').NextConfig} */
const nextConfig = {
  // Configure `pageExtensions`` to include MDX files
  pageExtensions: ['js', 'jsx', 'mdx', 'ts', 'tsx'],
  // Optionally, add any other Next.js config below
  reactStrictMode: true,
  swcMinify: true,
  async rewrites() {
    return [
      {
        source: '/rss',
        destination: '/feed.xml',
      },
      {
        source: '/rss.xml',
        destination: '/feed.xml',
      },
      {
        source: '/feed',
        destination: '/feed.xml',
      },
    ]
  }
}

export default withMDX(nextConfig)
