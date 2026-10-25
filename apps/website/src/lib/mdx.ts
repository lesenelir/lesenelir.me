import fs from 'node:fs'
import path from 'node:path'
import { parse as parseYaml } from 'yaml'

import type { TPost, TPostMetadata } from '@/types'

const postsDirectory = path.join(process.cwd(), 'content/posts')

const FRONTMATTER = /^---\n([\s\S]*?)\n---\n/

// Reading speed: Chinese characters and English words a minute, and code lines a minute when skimmed
const CHARS_PER_MINUTE = 300
const WORDS_PER_MINUTE = 200
const CODE_LINES_PER_MINUTE = 100

// Enough text for a search result or a link preview, cut on a character so Chinese stays whole
const EXCERPT_MIN_LENGTH = 80
const EXCERPT_MAX_LENGTH = 120

interface TPostSource {
  metadata: TPostMetadata
  // The post without its frontmatter
  body: string
}

// Each post is read once per process; pages, the sitemap, the feed and the images all ask for the same posts
const sources = new Map<string, TPostSource | null>()

// A post's title, date and optional description are in its frontmatter; its reading time is worked out from its text
function readPost(id: string): TPostSource | null {
  if (sources.has(id)) return sources.get(id) ?? null

  const file = path.join(postsDirectory, `${id}.mdx`)
  if (!fs.existsSync(file)) {
    sources.set(id, null)
    return null
  }

  const source = fs.readFileSync(file, 'utf8')
  const match = source.match(FRONTMATTER)
  const frontmatter = match ? parseYaml(match[1]) : null
  if (typeof frontmatter?.title !== 'string' || typeof frontmatter?.date !== 'string') {
    throw new Error(`${id}.mdx needs a title and a date ('YYYY-MM-DD') in its frontmatter`)
  }

  const body = match ? source.slice(match[0].length) : source
  const post = {
    metadata: {
      title: frontmatter.title,
      date: frontmatter.date,
      description: typeof frontmatter.description === 'string' ? frontmatter.description : undefined,
      readTime: getReadTime(body),
    },
    body,
  }

  sources.set(id, post)
  return post
}

function getReadTime(body: string): string {
  const codeBlocks = body.match(/```[\s\S]*?```/g) ?? []
  const codeLines = codeBlocks.reduce((lines, block) => lines + block.split('\n').length - 2, 0)
  const prose = body.replace(/```[\s\S]*?```/g, '').replace(/<[A-Z][\s\S]*?\/>/g, '')
  const chars = prose.match(/\p{Script=Han}/gu)?.length ?? 0
  const words = prose.match(/[A-Za-z0-9]+/g)?.length ?? 0

  const minutes = Math.max(
    1,
    Math.ceil(chars / CHARS_PER_MINUTE + words / WORDS_PER_MINUTE + codeLines / CODE_LINES_PER_MINUTE)
  )
  return `${String(minutes).padStart(2, '0')} min`
}

export function getAllPosts(): TPost[] {
  return fs
    .readdirSync(postsDirectory)
    .filter(name => name.endsWith('.mdx'))
    .flatMap(name => {
      const id = name.replace(/\.mdx$/, '')
      const post = readPost(id)
      return post ? [{ id, metadata: post.metadata }] : []
    })
    .sort((a, b) => new Date(b.metadata.date).getTime() - new Date(a.metadata.date).getTime())
}

export function getPostMetadata(id: string): TPostMetadata | null {
  return readPost(id)?.metadata ?? null
}

// Plain text from the opening paragraphs of a post. Blockquotes (epigraphs, disclaimers) are only used when a post
// has no other paragraphs.
function getPostExcerpt(body: string): string {
  const source = body
    .replace(/```[\s\S]*?```/g, '')
    .replace(/^import .*$/gm, '')
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
  const post = readPost(id)
  if (!post) return ''
  return post.metadata.description ?? getPostExcerpt(post.body)
}

export async function getPostComponent(id: string) {
  try {
    const component = await import(`../../content/posts/${id}.mdx`)
    return component.default
  } catch (_e) {
    return null
  }
}
