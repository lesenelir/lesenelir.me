import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Posts',
  description: "Lesenelir Zhou's personal website - Posts"
}

export default function Page() {
  return <main className={'border-dividing ml-24 min-h-96 border-l pl-16'}>posts</main>
}
