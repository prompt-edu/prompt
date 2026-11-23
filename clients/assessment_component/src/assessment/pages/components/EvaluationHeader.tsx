import { cn, StickyHeader } from '@tumaet/prompt-ui-components'
import type { ReactNode } from 'react'

interface EvaluationHeaderProps {
  children: ReactNode
}

/**
 * Displays an evaluation page title at full size and condenses it into a
 * compact header when it docks below the global management header.
 */
export const EvaluationHeader = ({ children }: EvaluationHeaderProps) => {
  return (
    <StickyHeader>
      {(docked) => (
        // The undocked gap is padding on the row rather than a margin on the title, so it is
        // included in the height StickyHeader freezes.
        // Only docking animates: the undocked row is in flow, so growing it back over time
        // would push the content below down while the user scrolls.
        <div
          className={cn(
            'flex min-w-0 items-center gap-2',
            docked ? 'h-10 px-3 transition-all duration-300' : 'pb-6',
          )}
        >
          {/* ManagementPageHeader's text-4xl font-bold, inlined because it takes no
              className and this title also needs a condensed docked variant. */}
          <h1
            className={cn(
              'min-w-0 flex-1 font-bold',
              docked ? 'truncate text-base text-center transition-all duration-300' : 'text-4xl',
            )}
          >
            {children}
          </h1>
        </div>
      )}
    </StickyHeader>
  )
}
