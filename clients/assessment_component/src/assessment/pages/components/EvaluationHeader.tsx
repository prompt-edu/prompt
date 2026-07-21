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
  const hasActions = Boolean(previousAction || nextAction)

  return (
    <StickyHeader>
      {(docked) => (
        <div
          className={cn(
            'flex min-w-0 items-center gap-2 transition-all duration-300',
            docked && 'h-10 px-3',
          )}
        >
          {previousAction}
          <h1
            className={cn(
              'min-w-0 flex-1 font-bold transition-all duration-300',
              docked ? 'truncate text-base text-center' : 'text-4xl',
              !docked && !hasActions && 'mb-6',
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
