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
      const { data } = matter(fileContents)

      return {
        id,
        metadata: data as TPost['metadata']
      }
    })
    .sort((a, b) => {
      return new Date(b.metadata.date).getTime() - new Date(a.metadata.date).getTime()
    })
}

export function getPostMetadata(id: string): TPost['metadata'] | null {
  try {
    const fullPath = path.join(postsDirectory, `${id}.mdx`)
    const fileContents = fs.readFileSync(fullPath, 'utf8')
    const { data } = matter(fileContents)

    return data as TPost['metadata']
  } catch (_e) {
    return null
  }
}

export async function getPostComponent(id: string) {
  try {
    const component = await import(`../../content/posts/${id}.mdx`)
    return component.default
  } catch (_e) {
    return null
  }
}
