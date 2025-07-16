import React from 'react'

export default function PostLayout({ children }: { children: React.ReactNode }) {
  return <div className={'max-w-2xl'}>{children}</div>
}
