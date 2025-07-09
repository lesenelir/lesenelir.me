import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Posts - Lesenelir Zhou',
  description: "Lesenelir Zhou's personal website - Posts"
}

export default function Page() {
  return <main className={'border-l pl-16'}>posts</main>
}
