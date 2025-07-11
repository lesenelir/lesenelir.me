import type { MDXComponents } from 'mdx/types'

export function useMDXComponents(components: MDXComponents): MDXComponents {
  return {
    h1: ({ children }) => <h1 className="text-text-primary mb-4 text-2xl font-bold">{children}</h1>,
    h2: ({ children }) => (
      <h2 className="text-text-primary mb-3 text-xl font-semibold">{children}</h2>
    ),
    h3: ({ children }) => (
      <h3 className="text-text-primary mb-2 text-lg font-medium">{children}</h3>
    ),
    p: ({ children }) => <p className="text-text-primary/85 mb-4 leading-relaxed">{children}</p>,
    code: ({ children }) => (
      <code className="rounded bg-gray-100 px-1 py-0.5 font-mono text-sm dark:bg-gray-800">
        {children}
      </code>
    ),
    pre: ({ children }) => (
      <pre className="mb-4 overflow-x-auto rounded-lg bg-gray-100 p-4 dark:bg-gray-800">
        {children}
      </pre>
    ),
    a: ({ children, href }) => (
      <a
        href={href}
        className="text-blue-600 hover:underline dark:text-blue-400"
        target={href?.startsWith('http') ? '_blank' : undefined}
        rel={href?.startsWith('http') ? 'noopener noreferrer' : undefined}
      >
        {children}
      </a>
    ),
    ul: ({ children }) => (
      <ul className="text-text-primary/85 mb-4 list-inside list-disc space-y-1">{children}</ul>
    ),
    ol: ({ children }) => (
      <ol className="text-text-primary/85 mb-4 list-inside list-decimal space-y-1">{children}</ol>
    ),
    blockquote: ({ children }) => (
      <blockquote className="text-text-primary/75 mb-4 border-l-4 border-gray-300 pl-4 italic dark:border-gray-600">
        {children}
      </blockquote>
    ),
    ...components
  }
}
