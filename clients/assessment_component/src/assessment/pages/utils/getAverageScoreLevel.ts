import {
  mapNumberToScoreLevel,
  mapScoreLevelToNumber,
  type ScoreLevel,
} from '@tumaet/prompt-shared-state'

import type { CompetencyScore } from '../../interfaces/competencyScore'

export function getAverageScoreLevel(scores: CompetencyScore[]): ScoreLevel | undefined {
  if (scores.length === 0) return undefined
  const average =
    scores.reduce((sum, score) => sum + mapScoreLevelToNumber(score.scoreLevel), 0) / scores.length
  return mapNumberToScoreLevel(average)
}
