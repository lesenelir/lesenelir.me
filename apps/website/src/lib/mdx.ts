import fs from 'node:fs'
import path from 'node:path'

import matter from 'gray-matter'

import type { TPost } from '@/types'

const postsDirectory = path.join(process.cwd(), 'content/posts')

export function getAllPosts(): TPost[] {
  const fileNames = fs.readdirSync(postsDirectory)
  return fileNames
    .filter((name) => name.endsWith('.mdx'))
    .map((name) => {
      const id = name.replace(/\.mdx$/, '')
      const fullPath = path.join(postsDirectory, name)
      const fileContents = fs.readFileSync(fullPath, 'utf8')
      const { data, content } = matter(fileContents)

      return {
        id,
        metadata: data as TPost['metadata'],
        content
      }
    })
    .sort((a, b) => {
      return new Date(b.metadata.date).getTime() - new Date(a.metadata.date).getTime()
    })
}

export function getPost(id: string): TPost | null {
  try {
    const fullPath = path.join(postsDirectory, `${id}.mdx`)
    const fileContents = fs.readFileSync(fullPath, 'utf8')
    const { data, content } = matter(fileContents)

    return {
      id,
      metadata: data as TPost['metadata'],
      content
    }
  } catch (_e) {
    // If the post is not found, return null
    return null
  }
}
