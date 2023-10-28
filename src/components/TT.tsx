"use client"

import React from "react"

interface IProps {
  children: React.ReactNode
}

export default function TTPage(props: IProps) {
  const {children} = props

  return (
    <div className={'text-xl bg-green-200'}>
      {children}
    </div>
  )
}
