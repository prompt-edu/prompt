import { cn } from '@tumaet/prompt-ui-components'
import { type ReactNode, useEffect, useRef, useState } from 'react'
import { flushSync } from 'react-dom'

// The bar docks beneath the global management header, which an announcement banner can push down.
const MANAGEMENT_HEADER_ID = 'management-header'
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
 *
 * The page layout never depends on the docked state: the placeholder keeps the
 * undocked height of the bar (expanded content included), and only the docked
 * bar, which is out of flow, animates. Anything else shifts the content below
 * and, on a page that barely overflows, scrolls it back past the docking point.
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
      const headerBottom =
        document.getElementById(MANAGEMENT_HEADER_ID)?.getBoundingClientRect().bottom ?? 0
      const shouldDock = dockedRef.current
        ? top <= headerBottom + UNDOCK_HYSTERESIS_PX
        : top <= headerBottom

      if (dockedRef.current !== shouldDock) {
        dockedRef.current = shouldDock
        // Render the new variant before writing styles, so the undocked height below is
        // measured from the undocked layout rather than from the docked one it replaces.
        flushSync(() => setDocked(shouldDock))
      }

      if (shouldDock) {
        placeholder.style.height = `${undockedHeightRef.current}px`
        bar.style.position = 'fixed'
        bar.style.top = `${headerBottom}px`
        bar.style.left = `${left}px`
        bar.style.width = `${width}px`
      } else {
        placeholder.style.height = ''
        bar.style.cssText = ''
        undockedHeightRef.current = bar.offsetHeight
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

    // The placeholder follows the content column, so observing it keeps the docked bar
    // aligned while the sidebar opens, collapses, or animates after a window resize.
    const resizeObserver = new ResizeObserver(scheduleUpdate)
    resizeObserver.observe(bar)
    resizeObserver.observe(placeholder)
    const bannerHeightObserver = new MutationObserver(scheduleUpdate)
    bannerHeightObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['style'],
    })
    undockedHeightRef.current = bar.offsetHeight
    scheduleUpdate()
    window.addEventListener('scroll', scheduleUpdate, true)
    window.addEventListener('resize', scheduleUpdate)
    return () => {
      resizeObserver.disconnect()
      bannerHeightObserver.disconnect()
      if (frame !== undefined) cancelAnimationFrame(frame)
      window.removeEventListener('scroll', scheduleUpdate, true)
      window.removeEventListener('resize', scheduleUpdate)
    }
  }, [])

  return (
    // print:h-auto! drops the inline height written while docked, which would
    // otherwise leave a stray gap on the printed page.
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

        {/* Collapses inside the docked bar, so the placeholder still reserves its height.
            Only the collapse animates: the undocked bar is in flow, and growing it back
            over time would push the content below down while the user scrolls. */}
        {expandedContent && (
          <div
            className={cn(
              'grid',
              docked
                ? 'grid-rows-[0fr] opacity-0 transition-all duration-300'
                : 'grid-rows-[1fr] opacity-100',
            )}
          >
            <div className='overflow-hidden'>
              {/* The gap sits inside the overflow so it collapses with the content. */}
              <div className='pt-4'>{expandedContent}</div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
