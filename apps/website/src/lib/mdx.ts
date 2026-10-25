import fs from 'node:fs'
import path from 'node:path'

import type { TPost, TPostMetadata } from '@/types'

import { postsMetadata } from '@/constants'

const postsDirectory = path.join(process.cwd(), 'content/posts')

export function getAllPosts(): TPost[] {
  const fileNames = fs.readdirSync(postsDirectory)
  return fileNames
    .filter(name => name.endsWith('.mdx'))
    .map(name => {
      const id = name.replace(/\.mdx$/, '')
      const metadata = postsMetadata[id]

      if (!metadata) {
        throw new Error(`Metadata not found for post: ${id}`)
      }

      return {
        id,
        metadata,
      }
    })
    .sort((a, b) => {
      return new Date(b.metadata.date).getTime() - new Date(a.metadata.date).getTime()
    })
}

export function getPostMetadata(id: string): TPostMetadata | null {
  return postsMetadata[id] || null
}

// Enough text for a search result or a link preview, cut on a character so Chinese stays whole
const EXCERPT_MIN_LENGTH = 80
const EXCERPT_MAX_LENGTH = 120

// Plain text from the opening paragraphs of a post. Blockquotes (epigraphs, disclaimers) are only used when a post
// has no other paragraphs.
function getPostExcerpt(id: string): string {
  const source = fs
    .readFileSync(path.join(postsDirectory, `${id}.mdx`), 'utf8')
    .replace(/^import .*$/gm, '')
    .replace(/```[\s\S]*?```/g, '')
    .replace(/<[A-Z][\s\S]*?\/>/g, '')

  // Paragraphs, without headings, rules, tables and lists
  const blocks = source
    .split(/\n\s*\n/)
    .map(block => block.trim())
    .filter(block => block && !/^(#|---|\||[-*] |\d+\. )/.test(block))
  const paragraphs = blocks.filter(block => !block.startsWith('>'))

  const toPlainText = (block: string) =>
    block
      .replace(/^>\s?/gm, '')
      .replace(/!\[[^\]]*\]\([^)]*\)/g, '')
      .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
      .replace(/[*_`~]/g, '')
      .replace(/\s+/g, ' ')
      .trim()

  let excerpt = ''
  for (const block of paragraphs.length > 0 ? paragraphs : blocks) {
    excerpt = excerpt ? `${excerpt} ${toPlainText(block)}` : toPlainText(block)
    if ([...excerpt].length >= EXCERPT_MIN_LENGTH) break
  }

  const chars = [...excerpt]
  return chars.length > EXCERPT_MAX_LENGTH ? `${chars.slice(0, EXCERPT_MAX_LENGTH).join('').trimEnd()}…` : excerpt
}

export function getPostDescription(id: string): string {
  return postsMetadata[id]?.description ?? getPostExcerpt(id)
}

export async function getPostComponent(id: string) {
  try {
    const component = await import(`../../content/posts/${id}.mdx`)
    return component.default
  } catch (_e) {
    return null
  }
}
