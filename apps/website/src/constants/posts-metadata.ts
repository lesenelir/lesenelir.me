import type { TPostMetadata } from '@/types'

/**
 * Metadata key is the filename of the posts folder without the `.mdx` extension.
 */
export const postsMetadata: Record<string, TPostMetadata> = {
  'hello-world': {
    title: 'Hello World',
    date: '2024-02-01',
    readTime: '3 min read'
  },
  'recursion': {
    title: 'Recursion',
    date: '2025-01-15',
    readTime: '3 min read'
  },
  'web-development-tips': {
    title: 'Web Development Tips',
    date: '2024-01-10',
    readTime: '5 min read'
  },
  'photography-journey': {
    title: 'My Photography Journey',
    date: '2024-07-05',
    readTime: '4 min read'
  }
}
