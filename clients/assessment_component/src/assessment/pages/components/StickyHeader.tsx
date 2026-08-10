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
  const undockedHeightRef = useRef(0)
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
        // Hold the height the bar had while undocked. Measuring the docked bar instead
        // shrinks the placeholder a frame later, shifting everything below by more than
        // UNDOCK_HYSTERESIS_PX, which lets a container near its bottom undock and flicker.
        placeholder.style.height = `${undockedHeightRef.current}px`
        bar.style.position = 'fixed'
        bar.style.top = `${HEADER_OFFSET_PX}px`
        bar.style.left = `${left}px`
        bar.style.width = `${width}px`
      } else {
        undockedHeightRef.current = bar.offsetHeight
        placeholder.style.height = ''
        bar.style.cssText = ''
      }

      if (dockedRef.current !== shouldDock) {
        dockedRef.current = shouldDock
        setDocked(shouldDock)
      }
    }

    // At most one measure-and-write per frame, however many scroll containers report.
    let frame: number | undefined
    const scheduleUpdate = () => {
      if (frame !== undefined) return
      frame = requestAnimationFrame(() => {
        frame = undefined
        update()
      })
    }

    const resizeObserver = new ResizeObserver(scheduleUpdate)
    resizeObserver.observe(bar)
    undockedHeightRef.current = bar.offsetHeight
    update()
    window.addEventListener('scroll', scheduleUpdate, true)
    window.addEventListener('resize', scheduleUpdate)
    return () => {
      resizeObserver.disconnect()
      if (frame !== undefined) cancelAnimationFrame(frame)
      window.removeEventListener('scroll', scheduleUpdate, true)
      window.removeEventListener('resize', scheduleUpdate)
    }
  }, [])

  return (
    <>
      {/* print:h-auto! drops the inline height written while docked, which would
          otherwise leave a stray gap on the printed page. */}
      <div ref={placeholderRef} data-testid='sticky-header-placeholder' className='print:h-auto!'>
        <div
          ref={barRef}
          data-testid='sticky-header'
          className={cn(
            'z-20 transition-colors duration-300 print:static! print:w-auto!',
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
