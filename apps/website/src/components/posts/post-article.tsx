'use client'

import type { ComponentPropsWithoutRef, MouseEvent } from 'react'

// Pages left behind stay in the document, hidden (cacheComponents keeps them with <Activity>), and posts share heading
// ids such as #wrapping-up. A hash link scrolls to the first element with its id, which may be in a hidden post, so
// the links in a post (headings, the table of contents, footnotes) look for their target in the post itself.
export function PostArticle({ children, ...props }: ComponentPropsWithoutRef<'article'>) {
  const handleClick = (event: MouseEvent<HTMLElement>) => {
    // A click with a modifier key (a new tab or window) is left to the browser
    const modified = event.metaKey || event.ctrlKey || event.shiftKey || event.altKey
    if (event.defaultPrevented || event.button !== 0 || modified) return
    if (!(event.target instanceof Element)) return

    const href = event.target.closest('a')?.getAttribute('href')
    if (!href?.startsWith('#')) return

    const target = event.currentTarget.querySelector(`#${CSS.escape(decodeURIComponent(href.slice(1)))}`)
    if (!target) return

    event.preventDefault()
    target.scrollIntoView()
    window.history.pushState(null, '', href)
  }

  return (
    // biome-ignore lint/a11y/useKeyWithClickEvents: only clicks on links are handled, and Enter on a link clicks it
    <article onClick={handleClick} {...props}>
      {children}
    </article>
  )
}
