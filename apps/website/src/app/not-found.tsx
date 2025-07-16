import Link from 'next/link'

import { cn } from '@/lib/utils'

export default function NotFound() {
  return (
    <div className={'text-text-primary/85 max-w-2xl space-y-8'}>
      <h1 className={'font-comic text-text-primary mb-8'}>404 - Page Not Found</h1>

      <p>Oops! The page you're looking for doesn't exist or has been moved.</p>

      <p>It might have been removed, renamed, or you might have typed the wrong URL.</p>

      <div className={'pt-4'}>
        <Link
          href={'/'}
          className={cn(
            'inline-flex w-fit items-center gap-1',
            'relative after:absolute after:bottom-0.5 after:left-0 after:h-px after:w-full',
            'after:bg-text-primary/50 hover:after:bg-text-primary after:transition-colors after:duration-200'
          )}
        >
          ← Back to Home
        </Link>
      </div>
    </div>
  )
}
