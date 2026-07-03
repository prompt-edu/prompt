import { mapNumberToScoreLevel } from '@tumaet/prompt-shared-state'
import { Badge, getLevelConfig } from '@tumaet/prompt-ui-components'

interface GradeSuggestionBadgeProps {
  gradeSuggestion: number | undefined
  text?: boolean
  className?: string
}

export const GradeSuggestionBadge = ({
  gradeSuggestion,
  text = false,
  className,
}: GradeSuggestionBadgeProps) => {
  if (!gradeSuggestion) {
    return undefined
  }

  const config = getLevelConfig(mapNumberToScoreLevel(gradeSuggestion))

  return (
    <Badge
      className={`${config.textColor} ${config.selectedBg} hover:${config.selectedBg} cursor-help ${className ?? ''}`}
      style={{ whiteSpace: 'nowrap' }}
    >
      {text ? 'Grade Suggestion:' : ''} {gradeSuggestion.toFixed(1)}
    </Badge>
  )
}
