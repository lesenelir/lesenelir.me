import Link from 'next/link'
import type { MDXComponents } from 'mdx/types'
import type { ComponentPropsWithoutRef } from 'react'

import { cn } from '@/lib/utils'

type HeadingProps = ComponentPropsWithoutRef<'h1'>
type ParagraphProps = ComponentPropsWithoutRef<'p'>
type HorizontalRuleProps = ComponentPropsWithoutRef<'hr'>
type StrongProps = ComponentPropsWithoutRef<'strong'>
type ListProps = ComponentPropsWithoutRef<'ul'>
type BlockquoteProps = ComponentPropsWithoutRef<'blockquote'>
type PreProps = ComponentPropsWithoutRef<'pre'>
type CodeProps = ComponentPropsWithoutRef<'code'>
type AnchorProps = ComponentPropsWithoutRef<'a'>

const components = {
  h1: (props: HeadingProps) => (
    <h1 className={'text-text-primary mb-4 text-2xl font-bold'} {...props} />
  ),
  h2: (props: HeadingProps) => (
    <h2 className={'text-text-primary mb-3 text-xl font-semibold'} {...props} />
  ),
  h3: (props: HeadingProps) => (
    <h3 className={'text-text-primary mb-2 text-lg font-medium'} {...props} />
  ),
  h4: (props: HeadingProps) => <h4 className={'text-text-primary mb-1'} {...props} />,
  p: (props: ParagraphProps) => (
    <p className={'text-text-primary/85 my-4 leading-relaxed'} {...props} />
  ),
  hr: (props: HorizontalRuleProps) => (
    <hr className={'border-dividing my-12 border-t'} {...props} />
  ),
  strong: (props: StrongProps) => (
    <strong className={'text-text-primary font-semibold'} {...props} />
  ),
  ol: (props: ListProps) => (
    <ol className={'text-text-primary/85 my-4 list-inside list-decimal space-y-4'} {...props} />
  ),
  ul: (props: ListProps) => (
    <ul className={'text-text-primary/85 my-4 list-inside list-disc space-y-4'} {...props} />
  ),
  blockquote: (props: BlockquoteProps) => (
    <blockquote
      className={'text-text-primary/75 border-l-3 border-dividing my-4 pl-4 italic'}
      {...props}
    />
  ),
  pre: (props: PreProps) => (
    <pre
      className={'bg-foreground my-4 overflow-x-auto rounded-lg p-4 font-mono text-sm'}
      {...props}
    />
  ),
  code: (props: CodeProps) => (
    <code className={'bg-foreground rounded px-1 py-0.5 font-mono text-sm'} {...props} />
  ),
  a: ({ children, href, ...props }: AnchorProps) => {
    const className = cn(
      'text-link',
      'relative after:absolute after:bottom-0 after:left-0 after:h-px after:w-full',
      'after:bg-link/85 hover:after:bg-link after:transition-colors after:duration-200'
    )

    // Internal links (Next.js)
    if (href?.startsWith('/')) {
      return (
        <Link href={href} className={className} {...props}>
          {children}
        </Link>
      )
    }

    // Hash links (internal anchors)
    if (href?.startsWith('#')) {
      return (
        <a href={href} className={className} {...props}>
          {children}
        </a>
      )
    }

    // External links
    return (
      <a href={href} target="_blank" rel="noopener noreferrer" className={className} {...props}>
        {children}
        <span
          className={'i-mingcute-right-line absolute -right-2.5 -top-1.5 size-3.5 -rotate-45'}
        />
      </a>
    )
  }
}

export function useMDXComponents(componentsProps: MDXComponents): MDXComponents {
  return {
    ...components,
    ...componentsProps
  }
}
