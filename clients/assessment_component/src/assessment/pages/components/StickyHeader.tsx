import { cn } from '@tumaet/prompt-ui-components'
import { type ReactNode, useEffect, useRef, useState } from 'react'

// The global management header (`h-14`) stays at the top; the bar docks beneath it.
const HEADER_OFFSET_PX = 56
const UNDOCK_HYSTERESIS_PX = 8

interface StickyHeaderProps {
  children: (docked: boolean) => ReactNode
  expandedContent?: ReactNode
  className?: string
}

/**
 * Keeps page context visible below the global management header while any
 * ancestor scroll container moves. Optional expanded content collapses once
 * the header docks, matching the assessment participant header behavior.
 */
export const StickyHeader = ({ children, expandedContent, className }: StickyHeaderProps) => {
  const placeholderRef = useRef<HTMLDivElement>(null)
  const barRef = useRef<HTMLDivElement>(null)
  const dockedRef = useRef(false)
  const [docked, setDocked] = useState(false)

  // CSS `position: sticky` is unreliable here because the core's scroll
  // container (`#management-children`, `overflow-auto`) sits between this bar
  // and the viewport. A capture-phase listener observes every scroll container.
  useEffect(() => {
    const placeholder = placeholderRef.current
    const bar = barRef.current
    if (!placeholder || !bar) return

    const update = () => {
      const { top, left, width } = placeholder.getBoundingClientRect()
      const shouldDock = dockedRef.current
        ? top <= HEADER_OFFSET_PX + UNDOCK_HYSTERESIS_PX
        : top <= HEADER_OFFSET_PX

      if (shouldDock) {
        placeholder.style.height = `${bar.offsetHeight}px`
        bar.style.position = 'fixed'
        bar.style.top = `${HEADER_OFFSET_PX}px`
        bar.style.left = `${left}px`
        bar.style.width = `${width}px`
      } else {
        placeholder.style.height = ''
        bar.style.cssText = ''
      }

      dockedRef.current = shouldDock
      setDocked(shouldDock)
    }

    const resizeObserver = new ResizeObserver(update)
    resizeObserver.observe(bar)
    update()
    window.addEventListener('scroll', update, true)
    window.addEventListener('resize', update)
    return () => {
      resizeObserver.disconnect()
      window.removeEventListener('scroll', update, true)
      window.removeEventListener('resize', update)
    }
  }, [])

  return (
    <>
      <div ref={placeholderRef}>
        <div
          ref={barRef}
          className={cn(
            'z-20 transition-colors duration-300 print:!static print:!w-auto',
            docked && 'rounded-md bg-background shadow-sm',
            className,
          )}
        >
          {children(docked)}
        </div>
      </div>

      {expandedContent && (
        <div
          className={cn(
            'grid transition-all duration-300',
            docked ? 'grid-rows-[0fr] opacity-0' : 'grid-rows-[1fr] opacity-100',
          )}
        >
          <div className='overflow-hidden'>{expandedContent}</div>
        </div>
      )}
    </>
  )
}
