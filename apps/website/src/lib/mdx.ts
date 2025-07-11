import fs from 'node:fs'
import path from 'node:path'

import type { TPost, TPostMetadata } from '@/types'

import { postsMetadata } from '../../content/posts/posts-metadata'

const postsDirectory = path.join(process.cwd(), 'content/posts')

export function getAllPosts(): TPost[] {
  const fileNames = fs.readdirSync(postsDirectory)
  return fileNames
    .filter((name) => name.endsWith('.mdx'))
    .map((name) => {
      const id = name.replace(/\.mdx$/, '')
      const metadata = postsMetadata[id]

      if (!metadata) {
        throw new Error(`Metadata not found for post: ${id}`)
      }

      return {
        id,
        metadata
      }
    })
    .sort((a, b) => {
      return new Date(b.metadata.date).getTime() - new Date(a.metadata.date).getTime()
    })
}

export function getPostMetadata(id: string): TPostMetadata | null {
  return postsMetadata[id] || null
}

export async function getPostComponent(id: string) {
  try {
    const component = await import(`../../content/posts/${id}.mdx`)
    return component.default
  } catch (_e) {
    return null
  }
}
