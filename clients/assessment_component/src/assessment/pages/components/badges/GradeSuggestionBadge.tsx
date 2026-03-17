import { Badge } from '@tumaet/prompt-ui-components'

import { getLevelConfig } from '../../utils/getLevelConfig'
import { mapNumberToScoreLevel } from '@tumaet/prompt-shared-state'

interface GradeSuggestionBadgeProps {
  gradeSuggestion: number | undefined
  text?: boolean
}

export const GradeSuggestionBadge = ({
  gradeSuggestion,
  text = false,
}: GradeSuggestionBadgeProps) => {
  if (!gradeSuggestion) {
    return undefined
  }

  const config = getLevelConfig(mapNumberToScoreLevel(gradeSuggestion))

  return (
    <Badge
      className={`${config.textColor} ${config.selectedBg} hover:${config.selectedBg} cursor-help`}
      style={{ whiteSpace: 'nowrap' }}
    >
      {text ? 'Grade Suggestion:' : ''} {gradeSuggestion.toFixed(1)}
    </Badge>
  )
}
