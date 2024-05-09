'use client'

import { useRouter } from 'next/navigation'

export default function Cd() {
  const router = useRouter()

  return (
    <div className={'cd-container'}>
      <span onClick={() => router.push('/posts')}>
        <span className={'mr-2 font-medium'}>&gt;</span>
        <a className={'underline underline-offset-2'}>
          cd {' '}.{' '}.
        </a>
      </span>
    </div>
  )
}
