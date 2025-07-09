import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Posts',
  description: "Lesenelir Zhou's personal website - Posts"
}

export default function Page() {
  return <main className={'border-dividing border-l pl-16'}>posts</main>
}
