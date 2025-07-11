export type TPostMetadata = {
  title: string
  date: string
  readTime: string
}

export type TPost = {
  id: string
  metadata: TPostMetadata
  content: string
}
