import createMDX from '@next/mdx'
import remarkGfm from 'remark-gfm'
import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  pageExtensions: ['js', 'jsx', 'mdx', 'ts', 'tsx'],
  images: {
    remotePatterns: [
      ...(process.env.S3_PUBLIC_URL
        ? [
            {
              protocol: new URL(process.env.S3_PUBLIC_URL).protocol.slice(0, -1) as
                | 'http'
                | 'https',
              hostname: new URL(process.env.S3_PUBLIC_URL).hostname,
              port: new URL(process.env.S3_PUBLIC_URL).port || '',
              pathname: '/**'
            }
          ]
        : [])
    ]
  }
}

const withMDX = createMDX({
  extension: /\.(md|mdx)$/,
  options: {
    remarkPlugins: [remarkGfm],
    rehypePlugins: []
  }
})

export default withMDX(nextConfig)
