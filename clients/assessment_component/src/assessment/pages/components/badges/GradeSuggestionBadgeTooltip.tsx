import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'

import { GradeSuggestionBadge } from './GradeSuggestionBadge'

interface GradeSuggestionBadgeProps {
  gradeSuggestion: number | undefined
  text?: boolean
}

export const GradeSuggestionBadgeWithTooltip = ({
  gradeSuggestion,
  text = false,
}: GradeSuggestionBadgeProps) => {
  if (!gradeSuggestion) {
    return undefined
  }

  const tooltipText = 'This is the grade you propose to the course instructor.'

  return (
    <TooltipProvider delayDuration={250}>
      <Tooltip>
        <TooltipTrigger>
          <GradeSuggestionBadge gradeSuggestion={gradeSuggestion} text={text} />
        </TooltipTrigger>
        <TooltipContent side='top'>
          <p className='max-w-lg text-center'>{tooltipText}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
