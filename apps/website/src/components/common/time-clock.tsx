'use client'

import { useEffect, useState } from 'react'

export default function TimeClock() {
  const [currentTime, setCurrentTime] = useState('--:--')
  const [hours, minutes] = currentTime.split(':')

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
    }

    updateTime()
    const interval = setInterval(updateTime, 30000)

    return () => clearInterval(interval)
  }, [])

  return (
    <>
      {hours}
      <span className={'animate-blink'}>:</span>
      {minutes}
    </>
  )
}
