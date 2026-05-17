import { useEffect, useState } from 'react'

export function useIsLargeScreen() {
  const [isLarge, setIsLarge] = useState(() => window.matchMedia('(min-width: 1024px)').matches)

  useEffect(() => {
    const mq = window.matchMedia('(min-width: 1024px)')
    const handler = (e: MediaQueryListEvent) => setIsLarge(e.matches)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [])

  return isLarge
}
