import createMDX from '@next/mdx'
import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  pageExtensions: ['js', 'jsx', 'mdx', 'ts', 'tsx'],
  images: {
    // In dev, let the browser fetch photos straight from the CDN. Proxies in fake-ip mode (e.g. Clash) resolve it
    // to 198.18.x.x, which Next 16's optimizer rejects as a private IP. Production keeps optimization.
    unoptimized: process.env.NODE_ENV === 'development',
    formats: ['image/avif', 'image/webp'],
    // Photos and post images never change under the same URL, so keep each optimized file for 31 days
    minimumCacheTTL: 2678400,
    remotePatterns: [
      ...(process.env.S3_PUBLIC_URL
        ? [
            {
              protocol: new URL(process.env.S3_PUBLIC_URL).protocol.slice(0, -1) as 'http' | 'https',
              hostname: new URL(process.env.S3_PUBLIC_URL).hostname,
              port: new URL(process.env.S3_PUBLIC_URL).port || '',
              pathname: '/**',
            },
          ]
        : []),
    ],
  },
}

const withMDX = createMDX({
  extension: /\.(md|mdx)$/,
  options: {
    // Turbopack can't pass JS functions to Rust, so plugins are referenced by package name
    // remark-frontmatter keeps a post's frontmatter out of the rendered page; src/lib/mdx.ts reads it
    remarkPlugins: ['remark-gfm', 'remark-frontmatter'],
    rehypePlugins: [
      // Heading ids for anchors and the table of contents; src/lib/mdx.ts makes the same ids with github-slugger
      'rehype-slug',
      // Code highlighted at build time. Options must be plain data here; with defaultColor off, tokens carry
      // --shiki-light and --shiki-dark, and src/styles/colors.css picks one per theme
      [
        '@shikijs/rehype',
        {
          themes: { light: 'vitesse-light', dark: 'vitesse-dark' },
          defaultColor: false,
          langs: ['go', 'typescript', 'tsx', 'javascript', 'bash'],
          fallbackLanguage: 'text',
          addLanguageClass: true,
        },
      ],
    ],
  },
})

export default withMDX(nextConfig)
