import { useEffect, useState } from 'react'

const CLOCK_TICK_MS = 60 * 1000

export const useNow = (): Date => {
  const [now, setNow] = useState(() => new Date())

  useEffect(() => {
    const interval = setInterval(() => setNow(new Date()), CLOCK_TICK_MS)
    return () => clearInterval(interval)
  }, [])

  return now
}
