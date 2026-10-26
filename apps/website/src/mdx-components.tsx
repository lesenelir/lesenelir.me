import type { MDXComponents } from 'mdx/types'
import Image from 'next/image'
import Link from 'next/link'
import type { ComponentPropsWithoutRef, ElementType } from 'react'
import { isValidElement } from 'react'

import { cn } from '@/lib/utils'

import { ImageMdx } from '@/components/common/image-mdx'
import { CodeBlock } from '@/components/posts/code-block'

type HeadingProps = ComponentPropsWithoutRef<'h1'>
type ParagraphProps = ComponentPropsWithoutRef<'p'>
type HorizontalRuleProps = ComponentPropsWithoutRef<'hr'>
type StrongProps = ComponentPropsWithoutRef<'strong'>
type ListProps = ComponentPropsWithoutRef<'ul'>
type BlockquoteProps = ComponentPropsWithoutRef<'blockquote'>
type PreProps = ComponentPropsWithoutRef<'pre'>
type CodeProps = ComponentPropsWithoutRef<'code'>
type AnchorProps = ComponentPropsWithoutRef<'a'>
type TableProps = ComponentPropsWithoutRef<'table'>
type TableHeadProps = ComponentPropsWithoutRef<'thead'>
type TableBodyProps = ComponentPropsWithoutRef<'tbody'>
type TableRowProps = ComponentPropsWithoutRef<'tr'>
type TableHeaderProps = ComponentPropsWithoutRef<'th'>
type TableDataProps = ComponentPropsWithoutRef<'td'>
type DelProps = ComponentPropsWithoutRef<'del'>
type InputProps = ComponentPropsWithoutRef<'input'>

// A section heading that links to itself, with a # on hover. rehype-slug gives it its id; scroll-mt keeps it off the
// window's edge when a link jumps to it. The # is decoration, so the link and the heading are named by the text.
const sectionHeading =
  (Tag: ElementType, className: string) =>
  ({ id, children, ...props }: HeadingProps) => (
    <Tag id={id} className={cn('group scroll-mt-8 text-text-primary', className)} {...props}>
      {id ? (
        <a href={`#${id}`}>
          {children}
          <span
            aria-hidden={'true'}
            className={'ml-2 text-text-foreground opacity-0 transition-opacity duration-200 group-hover:opacity-100'}
          >
            #
          </span>
        </a>
      ) : (
        children
      )}
    </Tag>
  )

const components = {
  // Available in every post without an import
  Image,
  ImageMdx,
  // The page title is the h1, so a post's sections start at ##
  h1: sectionHeading('h1', 'mb-4 text-2xl font-bold'),
  h2: sectionHeading('h2', 'mb-2 text-lg font-medium'),
  h3: sectionHeading('h3', 'mb-1'),
  h4: sectionHeading('h4', 'mb-1'),
  p: (props: ParagraphProps) => <p className={'my-4 text-md leading-relaxed text-text-primary/85'} {...props} />,
  hr: (props: HorizontalRuleProps) => <hr className={'my-12 w-1/3 border-t border-dividing'} {...props} />,
  strong: (props: StrongProps) => <strong className={'font-semibold text-text-primary'} {...props} />,
  // Markers sit outside the text, so a wrapped line lines up with the first; a list inside an item sits closer
  ol: (props: ListProps) => (
    <ol
      className={'my-4 list-outside list-decimal space-y-2 pl-6 text-md text-text-primary/85 [li>&]:my-2'}
      {...props}
    />
  ),
  ul: (props: ListProps) => (
    <ul className={'my-4 list-outside list-disc space-y-2 pl-6 text-md text-text-primary/85 [li>&]:my-2'} {...props} />
  ),
  blockquote: (props: BlockquoteProps) => (
    <blockquote className={'my-4 border-l-3 border-dividing pl-4 text-text-primary/75 italic'} {...props} />
  ),
  // Code blocks come highlighted from Shiki (next.config.ts), with a language-* class on their <code>
  pre: ({ children, ...props }: PreProps) => {
    const codeClassName = isValidElement<CodeProps>(children) ? children.props.className : undefined
    const language = codeClassName?.match(/language-(\S+)/)?.[1]
    return (
      <CodeBlock language={language === 'text' ? undefined : language} {...props}>
        {children}
      </CodeBlock>
    )
  },
  // Inline code is plain; the <code> of a block keeps Shiki's markup as it is
  code: ({ className, ...props }: CodeProps) =>
    className?.includes('language-') ? (
      <code className={className} {...props} />
    ) : (
      <code className={cn('rounded bg-foreground px-1 py-0.5 font-mono text-sm', className)} {...props} />
    ),
  a: ({ children, href, ...props }: AnchorProps) => {
    const className = cn(
      'text-link',
      'underline decoration-link/85 decoration-1 underline-offset-2 hover:decoration-link',
      'transition-colors duration-200'
    )

    // Spacing around a link is the post's own: the source has spaces where they belong, and none before punctuation

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
      <a href={href} target='_blank' rel='noopener noreferrer' className={className} {...props}>
        {children}
        <span className={'i-mingcute-right-line relative -top-1.5 -ml-0.5 inline-block size-3.5 -rotate-45'} />
      </a>
    )
  },
  // GFM Extensions
  table: (props: TableProps) => (
    <div className={'my-6 overflow-x-auto'}>
      <table className={'w-full border-collapse border border-dividing'} {...props} />
    </div>
  ),
  thead: (props: TableHeadProps) => <thead className='bg-foreground/80' {...props} />,
  tbody: (props: TableBodyProps) => <tbody {...props} />,
  tr: (props: TableRowProps) => <tr className={'border-b border-dividing'} {...props} />,
  th: (props: TableHeaderProps) => (
    <th className={'border-r border-dividing px-4 py-2 text-left font-semibold text-text-primary'} {...props} />
  ),
  td: (props: TableDataProps) => (
    <td className={'border-r border-dividing px-4 py-2 text-text-primary/85'} {...props} />
  ),
  del: (props: DelProps) => <del className={'text-text-primary/70 line-through'} {...props} />,
  input: (props: InputProps) => {
    if (props.type === 'checkbox') {
      return <input className={'mr-2 accent-link'} disabled {...props} />
    }
    return <input {...props} />
  },
}

export function useMDXComponents(componentsProps: MDXComponents): MDXComponents {
  return {
    ...components,
    ...componentsProps,
  }
}
