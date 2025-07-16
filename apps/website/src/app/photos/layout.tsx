import React from 'react'

export default function PhotoLayout({ children }: { children: React.ReactNode }) {
  return <div className={'w-full max-w-none'}>{children}</div>
}
