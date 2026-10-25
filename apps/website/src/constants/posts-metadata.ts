import type { TPostMetadata } from '@/types'

/**
 * Metadata key is the filename of the posts folder without the `.mdx` extension.
 */
export const postsMetadata: Record<string, TPostMetadata> = {
  'goodbye-youth': {
    title: '再见，青春',
    date: '2024-06-28',
    readTime: '05 min',
    // The opening lines; taken automatically, the description would start with the editor's note
    description:
      '走在北京三元桥上，迎面而来的汽车，怔怔的走在只允许一个人通过的人行道上。霎那间 ... 我想到了广州很多美好往事，想到了很多生活片段 ... 而这些过往回忆、我的青春伴随着北京忙碌的车流消逝在下一个路口。',
  },
  'go-concurrency': {
    title: 'Go 并发编程浅析',
    date: '2024-05-20',
    readTime: '20 min',
  },
  'go-ts': {
    title: 'Go 与 TS 语法差异',
    date: '2024-05-07',
    readTime: '30 min',
  },
  jotai: {
    title: 'Jotai 理念与实践',
    date: '2024-04-01',
    readTime: '20 min',
  },
  'global-village': {
    title: '地球村',
    date: '2023-10-09',
    readTime: '08 min',
  },
  typescript: {
    title: 'TS 类型系统',
    date: '2023-08-22',
    readTime: '10 min',
  },
  recursion: {
    title: 'Recursion',
    date: '2022-12-22',
    readTime: '22 min',
  },
  'time-stopped': {
    title: '时间已经停止，剩下就是慢慢变老',
    date: '2022-11-18',
    readTime: '10 min',
  },
}
