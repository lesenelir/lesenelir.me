'use client'

import { useRouter } from 'next/navigation'

export default function Back() {
  const router = useRouter()

  return (
    <>
      <span className={'back-span'} onClick={() => router.back()}>
        Back
      </span>
    </>
  )
}
