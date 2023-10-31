"use client"

import { usePathname } from 'next/navigation'

export default function HeaderClient() {
  const pathname = usePathname()

  console.log(pathname)

  if (pathname === '/') {
    return (
      <>Lesenelir Zhou</>
    )
  }

  if (pathname === '/posts') {
    return (
      <>Posts</>
    )
  }

  if (pathname.startsWith('/posts/')) {
    // TODO: specific post title
    return (
      <>Posts22</>
    )
  }

  if (pathname === '/projects') {
    return (
      <>Projects</>
    )
  }

  if (pathname === '/photos') {
    return (
      <>Photos</>
    )
  }

  return (
    <>
      404
    </>
  )
}
