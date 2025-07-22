import type { ReactNode } from 'react'

import { cn } from '@/lib/utils'

interface ExternalLinkProps {
  href: string
  children: ReactNode
  icon?: ReactNode
}

export default function ExternalLink({ href, children, icon }: ExternalLinkProps) {
  return (
    <a
      target={'_blank'}
      href={href}
      rel={'noopener noreferrer'}
      className={cn(
        'text-link inline-flex w-fit items-center gap-1',
        'decoration-link/85 hover:decoration-link underline decoration-1 underline-offset-2',
        'transition-colors duration-200'
      )}
    >
      {icon}
      {children}
      <span
        className={
          'i-mingcute-right-line relative -top-1.5 -ml-1.5 inline-block size-3.5 -rotate-45'
        }
      />
    </a>
  )
}
