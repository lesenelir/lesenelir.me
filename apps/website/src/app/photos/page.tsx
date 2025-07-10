import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Photos',
  description: "Lesenelir Zhou's personal website - Photos"
}

export default function Page() {
  return <main className={'border-dividing ml-24 min-h-96 border-l pl-16'}>photos</main>
}
