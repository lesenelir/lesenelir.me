import type { Metadata } from 'next'

import Projects from '@/app/content/pages/projects.mdx'

export const metadata: Metadata = {
  title: 'Lesenelir Projects',
  description: 'Lesenelir Zhou Projects Page',
}

export default function ProjectPage() {
  return (
    <article className="text-justify prose max-md:prose-base dark:prose-invert">
      <Projects/>
    </article>
  )
}
