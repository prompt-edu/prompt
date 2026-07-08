import {
  mapNumberToScoreLevel,
  mapScoreLevelToNumber,
  type ScoreLevel,
} from '@tumaet/prompt-shared-state'

import type { Evaluation } from '../../interfaces/evaluation'

export function getAverageScoreLevel(evaluations: Evaluation[]): ScoreLevel | undefined {
  if (evaluations.length === 0) return undefined
  const average =
    evaluations.reduce((sum, evaluation) => sum + mapScoreLevelToNumber(evaluation.scoreLevel), 0) /
    evaluations.length
  return mapNumberToScoreLevel(average)
}
