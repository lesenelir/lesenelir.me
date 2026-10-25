export type TPostMetadata = {
  title: string
  date: string
  readTime: string
  // For search results, link previews and the feed; taken from the post's opening paragraphs when left out
  description?: string
}

export type TPost = {
  id: string
  metadata: TPostMetadata
}
