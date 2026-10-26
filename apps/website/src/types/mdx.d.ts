// From a post's frontmatter (title, date, description) and its text (readTime)
export type TPostMetadata = {
  title: string
  // YYYY-MM-DD
  date: string
  // Worked out from the post's length, e.g. '08 min'
  readTime: string
  // For search results, link previews and the feed; taken from the post's opening paragraphs when left out
  description?: string
}

export type TPost = {
  id: string
  metadata: TPostMetadata
}

// A section heading of a post, for its table of contents
export type TPostHeading = {
  id: string
  text: string
  level: number
}
