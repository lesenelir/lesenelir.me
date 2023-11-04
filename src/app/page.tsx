import Home from '@/app/content/pages/home.mdx'
import { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Lesenelir Zhou',
  description: 'Lesenelir Zhou Home Page',
}

export default function HomePage() {
  return (
    <>
      <article className="text-justify prose max-md:prose-base dark:prose-invert">
        <Home/>
      </article>
    </>
  )
}
