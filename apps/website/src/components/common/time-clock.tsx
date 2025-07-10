'use client'

import { useEffect, useState } from 'react'

export default function TimeClock() {
  const [currentTime, setCurrentTime] = useState('')
  const [showColon, setShowColon] = useState(true)

  useEffect(() => {
    const updateTime = () => {
      const now = new Date()
      const beijingTime = new Intl.DateTimeFormat('en-US', {
        timeZone: 'Asia/Shanghai',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
      }).format(now)

      setCurrentTime(beijingTime)
      setShowColon(now.getSeconds() % 2 === 0)
    }

    updateTime()
    const interval = setInterval(updateTime, 1000)

    return () => clearInterval(interval)
  }, [])

  const formatTimeWithBlinkingColon = (time: string) => {
    const [hours, minutes] = time.split(':')
    return (
      <>
        {hours}
        {showColon ? ':' : ' '}
        {minutes}
      </>
    )
  }

  return <>{formatTimeWithBlinkingColon(currentTime)}</>
}
