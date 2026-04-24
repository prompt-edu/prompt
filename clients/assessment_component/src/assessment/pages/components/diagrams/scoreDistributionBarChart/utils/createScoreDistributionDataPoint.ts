import { ScoreLevel, mapNumberToScoreLevel } from '@tumaet/prompt-shared-state'

import { ScoreDistributionDataPoint } from '../interfaces/ScoreDistributionDataPoint'

import { computeQuartile } from '../../utils/computeQuartile'

export const createScoreDistributionDataPoint = (
  shortLabel: string,
  label: string,
  scores: number[],
  scoreLevels: ScoreLevel[],
): ScoreDistributionDataPoint => {
  if (scores.length === 0 || scoreLevels.length === 0) {
    return {
      shortLabel,
      label,
      average: 0,
      lowerQuartile: 0,
      median: ScoreLevel.VeryBad,
      upperQuartile: 0,
      counts: {
        veryBad: 0,
        bad: 0,
        ok: 0,
        good: 0,
        veryGood: 0,
      },
    }
  }

  const counts = (Object.values(ScoreLevel) as ScoreLevel[]).reduce<Record<ScoreLevel, number>>(
    (acc, scoreLevel) => {
      acc[scoreLevel] = scoreLevels.filter((level) => level === scoreLevel).length
      return acc
    },
    {} as Record<ScoreLevel, number>,
  )

  return {
    shortLabel,
    label,
    average: scores.reduce((sum, score) => sum + score, 0) / scores.length,
    lowerQuartile: computeQuartile(scores, 0.25),
    median: mapNumberToScoreLevel(computeQuartile(scores, 0.5)),
    upperQuartile: computeQuartile(scores, 0.75),
    counts,
  }
}
