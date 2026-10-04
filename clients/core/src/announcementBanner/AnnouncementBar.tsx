import type { Announcement } from '@core/interfaces/announcement'
import { useActiveAnnouncements } from '@core/network/hooks/useAnnouncements'
import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { AnnouncementBanner } from './AnnouncementBanner'
import { useDismissedAnnouncements } from './hooks/useDismissedAnnouncements'
import { getDismissalKey, selectVisibleAnnouncements } from './utils/announcementStatus'

const CLOCK_TICK_MS = 60 * 1000
const ANNOUNCEMENT_BAR_HEIGHT_VARIABLE = '--announcement-bar-height'

const setBarHeight = (height: number) =>
  document.documentElement.style.setProperty(ANNOUNCEMENT_BAR_HEIGHT_VARIABLE, `${height}px`)

export const AnnouncementBar = () => {
  const { data: announcements = [] } = useActiveAnnouncements()
  const { dismissedKeys, dismiss } = useDismissedAnnouncements()
  const [now, setNow] = useState(() => new Date())
  const [selectedIndex, setSelectedIndex] = useState(0)
  const barRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const interval = setInterval(() => setNow(new Date()), CLOCK_TICK_MS)
    return () => clearInterval(interval)
  }, [])

  const visible = selectVisibleAnnouncements(announcements, dismissedKeys, now)
  const index = Math.min(selectedIndex, Math.max(visible.length - 1, 0))
  const current: Announcement | undefined = visible[index]
  const hasBar = current !== undefined

  useLayoutEffect(() => {
    const bar = barRef.current
    if (!hasBar || !bar) {
      setBarHeight(0)
      return
    }
    let frame = 0
    let appliedHeight = -1
    const observer = new ResizeObserver(() => {
      cancelAnimationFrame(frame)
      frame = requestAnimationFrame(() => {
        const height = bar.getBoundingClientRect().height
        if (height !== appliedHeight) {
          appliedHeight = height
          setBarHeight(height)
        }
      })
    })
    setBarHeight(bar.getBoundingClientRect().height)
    observer.observe(bar)
    return () => {
      cancelAnimationFrame(frame)
      observer.disconnect()
      setBarHeight(0)
    }
  }, [hasBar])

  if (!current) return null

  const step = (offset: number) =>
    setSelectedIndex((index + offset + visible.length) % visible.length)

  return (
    <div ref={barRef} className='relative z-40 w-full shadow-sm md:sticky md:top-0 print:hidden'>
      <AnnouncementBanner
        severity={current.severity}
        title={current.title}
        message={current.message}
        linkUrl={current.linkUrl}
        linkLabel={current.linkLabel}
        pager={
          visible.length > 1
            ? {
                position: index + 1,
                total: visible.length,
                onPrevious: () => step(-1),
                onNext: () => step(1),
              }
            : undefined
        }
        onDismiss={() => dismiss(getDismissalKey(current), announcements.map(getDismissalKey))}
      />
    </div>
  )
}
