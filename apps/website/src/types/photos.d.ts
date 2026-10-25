export type LayoutMode = 'grid' | 'row' | 'single'

export interface TPhoto {
  id: string
  src: string
  alt: string
  desc?: string
  // As displayed, with the EXIF orientation applied
  width: number
  height: number
}
