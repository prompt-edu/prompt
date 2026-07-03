import React from 'react'
import {
  Badge,
  getLevelConfig,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'
import { mapNumberToScoreLevel, ScoreLevel } from '@tumaet/prompt-shared-state'

interface ScoreLevelBadgeProps {
  scoreLevel?: ScoreLevel
  scoreNumeric?: number
  showTooltip?: boolean
  compact?: boolean
  className?: string
}

export const StudentScoreBadge = ({
  scoreLevel,
  scoreNumeric,
  showTooltip = false,
  compact = false,
  className,
}: ScoreLevelBadgeProps) => {
  if (!scoreLevel && !scoreNumeric) {
    return undefined // No score provided, nothing to display
  }

  const config = getLevelConfig(
    scoreLevel
      ? scoreLevel
      : scoreNumeric
        ? mapNumberToScoreLevel(scoreNumeric)
        : ScoreLevel.VeryBad,
  )

  const fullText =
    (scoreLevel ? config.title : '') +
    (scoreLevel && scoreNumeric ? ` (${scoreNumeric.toFixed(1)})` : '') +
    (!scoreLevel && scoreNumeric ? `${scoreNumeric.toFixed(1)}` : '')
  const compactText = scoreLevel ? config.title : scoreNumeric ? `${scoreNumeric.toFixed(1)}` : ''
  const content = compact ? compactText : fullText

  const tooltipText =
    'This score is automatically generated based on the assessment input. ' +
    'It is intended to assist you in making your final grading decision.'

  if (showTooltip)
    return (
      <TooltipProvider delayDuration={250}>
        <Tooltip>
          <TooltipTrigger>
            <Badge
              className={`${config.textColor} ${config.selectedBg} hover:${config.selectedBg} cursor-help ${className ?? ''}`}
              style={{ whiteSpace: 'nowrap' }}
            >
              {content}
            </Badge>
          </TooltipTrigger>
          <TooltipContent side='top'>
            <p className='max-w-lg text-center'>{tooltipText}</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )

  return (
    <Badge
      className={`${config.textColor} ${config.selectedBg} hover:${config.selectedBg} cursor-help`}
      style={{ whiteSpace: 'nowrap' }}
    >
      {content}
    </Badge>
  )
}
