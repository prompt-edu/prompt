import { useCallback, useState } from 'react'

const STORAGE_KEY = 'prompt-dismissed-announcements'

const readDismissedKeys = (): Set<string> => {
  try {
    const stored: unknown = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]')
    return new Set(
      Array.isArray(stored) ? stored.filter((key): key is string => typeof key === 'string') : [],
    )
  } catch {
    return new Set()
  }
}

const persistDismissedKeys = (keys: Set<string>): boolean => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify([...keys]))
    return true
  } catch {
    return false
  }
}

export const useDismissedAnnouncements = () => {
  const [dismissedKeys, setDismissedKeys] = useState(readDismissedKeys)

  const dismiss = useCallback((key: string, activeKeys: string[]) => {
    setDismissedKeys((current) => {
      const next = new Set([...current].filter((existing) => activeKeys.includes(existing)))
      next.add(key)
      persistDismissedKeys(next)
      return next
    })
  }, [])

  return { dismissedKeys, dismiss }
}
