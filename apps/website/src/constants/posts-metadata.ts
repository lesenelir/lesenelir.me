import type { TPostMetadata } from '@/types'

/**
 * Metadata key is the filename of the posts folder without the `.mdx` extension.
 */
export const postsMetadata: Record<string, TPostMetadata> = {
  'go-ts': {
    title: 'Go 与 TS 语法差异',
    date: '2024-05-07',
    readTime: '30 min'
  },
  'global-village': {
    title: '地球村',
    date: '2023-10-09',
    readTime: '08 min'
  },
  'typescript': {
    title: 'TypeScript',
    date: '2023-08-22',
    readTime: '10 min'
  },
  'recursion': {
    title: 'Recursion',
    date: '2022-12-22',
    readTime: '22 min'
  },
  'time-stopped': {
    title: '时间已经停止，剩下就是慢慢变老',
    date: '2022-11-18',
    readTime: '10 min'
  }
}
