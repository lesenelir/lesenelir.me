import type React from 'react'

// The text keeps to a 672px column; the full width around it is where a post's table of contents goes
export default function PostLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className={'relative'}>
      <div className={'max-w-2xl'}>{children}</div>
    </div>
  )
}
