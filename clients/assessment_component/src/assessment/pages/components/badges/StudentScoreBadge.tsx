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
}

export const StudentScoreBadge = ({
  scoreLevel,
  scoreNumeric,
  showTooltip = false,
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

  const tooltipText =
    'This score is automatically generated based on the assessment input. ' +
    'It is intended to assist you in making your final grading decision.'

  if (showTooltip)
    return (
      <TooltipProvider delayDuration={250}>
        <Tooltip>
          <TooltipTrigger>
            <Badge
              className={`${config.textColor} ${config.selectedBg} hover:${config.selectedBg} cursor-help`}
              style={{ whiteSpace: 'nowrap' }}
            >
              {scoreLevel ? config.title : ''}
              {scoreLevel && scoreNumeric ? ` (${scoreNumeric.toFixed(1)})` : ''}
              {!scoreLevel && scoreNumeric ? `${scoreNumeric.toFixed(1)}` : ''}
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
      {scoreLevel ? config.title : ''}
      {scoreLevel && scoreNumeric ? ` (${scoreNumeric.toFixed(1)})` : ''}
      {!scoreLevel && scoreNumeric ? `${scoreNumeric.toFixed(1)}` : ''}
    </Badge>
  )
}
