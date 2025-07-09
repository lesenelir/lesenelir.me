import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Photos - Lesenelir Zhou',
  description: "Lesenelir Zhou's personal website - Photos"
}

export default function Page() {
  return <main className={'border-l pl-16'}>photos</main>
}
