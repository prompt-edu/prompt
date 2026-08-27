import { ScoreLevel } from '@tumaet/prompt-shared-state'
import { describe, expect, it } from 'vitest'

import type { CompetencyScore } from '../../interfaces/competencyScore'
import { getAverageScoreLevel } from './getAverageScoreLevel'

const buildScore = (scoreLevel: ScoreLevel): CompetencyScore => ({
  id: `score-${scoreLevel}`,
  courseParticipationID: 'participation-1',
  coursePhaseID: 'phase-1',
  competencyID: 'competency-1',
  scoreLevel,
})

describe('getAverageScoreLevel', () => {
  it('has no average without scores', () => {
    expect(getAverageScoreLevel([])).toBeUndefined()
  })

  it('returns the level of a single score', () => {
    expect(getAverageScoreLevel([buildScore(ScoreLevel.Ok)])).toBe(ScoreLevel.Ok)
  })

  it('rounds an average down to the better level at the boundary', () => {
    const scores = [buildScore(ScoreLevel.VeryGood), buildScore(ScoreLevel.Good)]

    expect(getAverageScoreLevel(scores)).toBe(ScoreLevel.VeryGood)
  })

  it('maps an average between two levels to the better one', () => {
    const scores = [buildScore(ScoreLevel.Good), buildScore(ScoreLevel.Ok)]

    expect(getAverageScoreLevel(scores)).toBe(ScoreLevel.Good)
  })

  it('maps the worst average to the worst level', () => {
    const scores = [buildScore(ScoreLevel.VeryBad), buildScore(ScoreLevel.VeryBad)]

    expect(getAverageScoreLevel(scores)).toBe(ScoreLevel.VeryBad)
  })
})
