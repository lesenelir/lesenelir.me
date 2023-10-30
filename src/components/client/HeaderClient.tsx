"use client"

import { usePathname } from 'next/navigation'

export default function HeaderClient() {
  const pathname = usePathname()

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