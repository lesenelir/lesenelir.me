'use client'

import Link from 'next/link'
import { useEffect } from 'react'

import { cn } from '@/lib/utils'

interface ErrorPageProps {
  error: Error & { digest?: string }
  reset: () => void
}

const linkClassName = cn(
  'inline-flex w-fit cursor-pointer items-center gap-1',
  'relative after:absolute after:bottom-0.5 after:left-0 after:h-px after:w-full',
  'after:bg-text-primary/50 after:transition-colors after:duration-200 hover:after:bg-text-primary'
)

// Shown in place of a page that failed to render, with the sidebar still around it; laid out like the 404 page
export default function ErrorPage({ error, reset }: ErrorPageProps) {
  useEffect(() => {
    console.error(error)
  }, [error])

  return (
    <div className={'max-w-2xl space-y-8 text-text-primary/85'}>
      <h1 className={'mb-8 font-comic text-text-primary'}>Something Went Wrong</h1>

      <p>This page couldn't be loaded. It's likely a temporary problem, so trying again usually helps.</p>

      <div className={'flex gap-6 pt-4'}>
        <button type='button' onClick={reset} className={linkClassName}>
          ↻ Try again
        </button>
        <Link href={'/'} className={linkClassName}>
          ← Back to Home
        </Link>
      </div>
    </div>
  )
}
