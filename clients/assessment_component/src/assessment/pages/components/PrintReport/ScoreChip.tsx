import type { ScoreLevel } from '@tumaet/prompt-shared-state'
import { getLevelConfig } from '@tumaet/prompt-ui-components'

export const ScoreChip = ({ scoreLevel }: { scoreLevel: ScoreLevel }) => {
  const config = getLevelConfig(scoreLevel)
  return (
    <span
      className={`inline-block rounded-sm px-2 py-0.5 text-xs font-medium ${config.textColor} ${config.selectedBg}`}
    >
      {config.title}
    </span>
  )
}
