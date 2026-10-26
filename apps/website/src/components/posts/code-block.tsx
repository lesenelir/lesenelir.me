'use client'

import type { ComponentPropsWithoutRef } from 'react'
import { useRef, useState } from 'react'

import { cn } from '@/lib/utils'

const languageNames: Record<string, string> = {
  go: 'Go',
  typescript: 'TypeScript',
  tsx: 'TSX',
  javascript: 'JavaScript',
  bash: 'Bash',
}

// How long the button says "Copied"
const COPIED_DURATION = 1500

interface CodeBlockProps extends ComponentPropsWithoutRef<'pre'> {
  language?: string
}

// A highlighted code block with its language and a copy button in the corner. The button shows on hover or focus,
// and always on touch screens, which have no hover.
export function CodeBlock({ language, className, children, ...props }: CodeBlockProps) {
  const preRef = useRef<HTMLPreElement>(null)
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    await navigator.clipboard.writeText(preRef.current?.textContent ?? '')
    setCopied(true)
    setTimeout(() => setCopied(false), COPIED_DURATION)
  }

  return (
    <div className={'group relative my-4'}>
      <pre
        ref={preRef}
        className={cn(className, 'overflow-x-auto rounded-lg bg-foreground p-4 font-mono text-sm leading-relaxed')}
        {...props}
      >
        {children}
      </pre>

      <div className={'absolute top-2 right-2 flex items-center gap-2 text-xs text-text-foreground'}>
        {language && <span>{languageNames[language] ?? language}</span>}
        <button
          type='button'
          aria-label={copied ? 'Copied' : 'Copy code'}
          onClick={copy}
          className={cn(
            'flex size-7 cursor-pointer items-center justify-center rounded-md bg-background/80 transition-opacity duration-200',
            'opacity-0 group-hover:opacity-100 hover:text-text-primary focus-visible:opacity-100 pointer-coarse:opacity-100'
          )}
        >
          <span className={cn(copied ? 'i-mingcute-check-line' : 'i-mingcute-copy-2-line', 'size-4')} />
        </button>
      </div>
    </div>
  )
}
