import { cn } from '@tumaet/prompt-ui-components'
import type { ReactNode } from 'react'

import { StickyHeader } from './StickyHeader'

interface EvaluationHeaderProps {
  children: ReactNode
  previousAction?: ReactNode
  nextAction?: ReactNode
}

/**
 * Displays an evaluation page title at full size and condenses it into a
 * compact header when it docks below the global management header.
 */
export const EvaluationHeader = ({
  children,
  previousAction,
  nextAction,
}: EvaluationHeaderProps) => {
  return (
    <StickyHeader>
      {(docked) => (
        // The undocked gap is padding on the row rather than a margin on the title, so it
        // is the same whether or not there are nav actions, it does not push the actions
        // out of line with the title, and it is included in the height StickyHeader freezes.
        <div
          className={cn(
            'flex min-w-0 items-center gap-2 transition-all duration-300',
            docked ? 'h-10 px-3' : 'pb-6',
          )}
        >
          {previousAction}
          {/* ManagementPageHeader's text-4xl font-bold, inlined because it takes no
              className and this title also needs a condensed docked variant. */}
          <h1
            className={cn(
              'min-w-0 flex-1 font-bold transition-all duration-300',
              docked ? 'truncate text-base text-center' : 'text-4xl',
            )}
          >
            {children}
          </h1>
          {nextAction}
        </div>
      )}
    </StickyHeader>
  )
}
