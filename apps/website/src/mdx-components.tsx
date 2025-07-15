import Link from 'next/link'
import { highlight } from 'sugar-high'
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
type TableProps = ComponentPropsWithoutRef<'table'>
type TableHeadProps = ComponentPropsWithoutRef<'thead'>
type TableBodyProps = ComponentPropsWithoutRef<'tbody'>
type TableRowProps = ComponentPropsWithoutRef<'tr'>
type TableHeaderProps = ComponentPropsWithoutRef<'th'>
type TableDataProps = ComponentPropsWithoutRef<'td'>
type DelProps = ComponentPropsWithoutRef<'del'>
type InputProps = ComponentPropsWithoutRef<'input'>

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
    <p className={'text-text-primary/85 text-md my-4 leading-relaxed'} {...props} />
  ),
  hr: (props: HorizontalRuleProps) => (
    <hr className={'border-dividing my-12 w-1/3 border-t'} {...props} />
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
  pre: ({ children, ...props }: PreProps) => {
    const codeElement = Array.isArray(children) ? children[0] : children
    if (codeElement && typeof codeElement === 'object' && 'props' in codeElement) {
      const code = codeElement.props.children
      if (typeof code === 'string') {
        return (
          <pre
            className={'bg-foreground my-4 overflow-x-auto rounded-lg p-4 font-mono text-sm'}
            {...props}
          >
            <code dangerouslySetInnerHTML={{ __html: highlight(code) }} />
          </pre>
        )
      }
    }
    return (
      <pre
        className={'bg-foreground my-4 overflow-x-auto rounded-lg p-4 font-mono text-sm'}
        {...props}
      >
        {children}
      </pre>
    )
  },
  code: ({ children, ...props }: CodeProps) => {
    if (typeof children === 'string') {
      return (
        <code
          className={'bg-foreground rounded px-1 py-0.5 font-mono text-sm'}
          dangerouslySetInnerHTML={{ __html: highlight(children) }}
          {...props}
        />
      )
    }
    return (
      <code className={'bg-foreground rounded px-1 py-0.5 font-mono text-sm'} {...props}>
        {children}
      </code>
    )
  },
  a: ({ children, href, ...props }: AnchorProps) => {
    const className = cn(
      'text-link',
      'decoration-link/85 hover:decoration-link underline decoration-1 underline-offset-2',
      'transition-colors duration-200'
    )

    // Internal links (Next.js)
    if (href?.startsWith('/')) {
      return (
        <>
          {' '}
          <Link href={href} className={className} {...props}>
            {children}
          </Link>{' '}
        </>
      )
    }

    // Hash links (internal anchors)
    if (href?.startsWith('#')) {
      return (
        <>
          {' '}
          <a href={href} className={className} {...props}>
            {children}
          </a>{' '}
        </>
      )
    }

    // External links
    return (
      <>
        {' '}
        <a href={href} target="_blank" rel="noopener noreferrer" className={className} {...props}>
          {children}
          <span
            className={
              'i-mingcute-right-line relative -top-1.5 -ml-0.5 inline-block size-3.5 -rotate-45'
            }
          />
        </a>{' '}
      </>
    )
  },
  // GFM Extensions
  table: (props: TableProps) => (
    <div className={'my-6 overflow-x-auto'}>
      <table className={'border-dividing w-full border-collapse border'} {...props} />
    </div>
  ),
  thead: (props: TableHeadProps) => <thead className="bg-foreground/80" {...props} />,
  tbody: (props: TableBodyProps) => <tbody {...props} />,
  tr: (props: TableRowProps) => <tr className={'border-dividing border-b'} {...props} />,
  th: (props: TableHeaderProps) => (
    <th
      className={'text-text-primary border-dividing border-r px-4 py-2 text-left font-semibold'}
      {...props}
    />
  ),
  td: (props: TableDataProps) => (
    <td className={'text-text-primary/85 border-dividing border-r px-4 py-2'} {...props} />
  ),
  del: (props: DelProps) => <del className={'text-text-primary/60 line-through'} {...props} />,
  input: (props: InputProps) => {
    if (props.type === 'checkbox') {
      return <input className={'accent-link mr-2'} disabled {...props} />
    }
    return <input {...props} />
  }
}

export function useMDXComponents(componentsProps: MDXComponents): MDXComponents {
  return {
    ...components,
    ...componentsProps
  }
}
