import { ScoreLevel } from '@tumaet/prompt-shared-state'
import { describe, expect, it } from 'vitest'

import type { CategoryWithCompetencies } from '../../interfaces/category'
import type { Competency } from '../../interfaces/competency'
import type { CompetencyScore } from '../../interfaces/competencyScore'
import { getWeightedScoreLevel } from './getWeightedScoreLevel'

const buildCompetency = (id: string, weight: number): Competency => ({
  id,
  categoryID: 'category-1',
  name: id,
  shortName: id,
  description: '',
  descriptionVeryBad: '',
  descriptionBad: '',
  descriptionOk: '',
  descriptionGood: '',
  descriptionVeryGood: '',
  weight,
})

const buildCategory = (
  id: string,
  weight: number,
  competencies: Competency[],
): CategoryWithCompetencies => ({
  id,
  name: id,
  shortName: id,
  weight,
  competencies,
})

const buildScore = (competencyID: string, scoreLevel: ScoreLevel): CompetencyScore => ({
  id: `score-${competencyID}-${scoreLevel}`,
  courseParticipationID: 'participation-1',
  coursePhaseID: 'phase-1',
  competencyID,
  scoreLevel,
})

describe('getWeightedScoreLevel', () => {
  it('returns 0 without scores or without categories', () => {
    const category = buildCategory('category-1', 1, [buildCompetency('competency-1', 1)])

    expect(getWeightedScoreLevel([], [category])).toBe(0)
    expect(getWeightedScoreLevel([buildScore('competency-1', ScoreLevel.Good)], [])).toBe(0)
  })

  it('maps a single score to its numeric level', () => {
    const category = buildCategory('category-1', 1, [buildCompetency('competency-1', 1)])

    expect(
      getWeightedScoreLevel([buildScore('competency-1', ScoreLevel.VeryGood)], [category]),
    ).toBe(1)
  })

  it('averages equally weighted competencies', () => {
    const category = buildCategory('category-1', 1, [
      buildCompetency('competency-1', 1),
      buildCompetency('competency-2', 1),
    ])
    const scores = [
      buildScore('competency-1', ScoreLevel.VeryGood),
      buildScore('competency-2', ScoreLevel.Ok),
    ]

    expect(getWeightedScoreLevel(scores, [category])).toBe(2)
  })

  it('weights competencies within a category', () => {
    const category = buildCategory('category-1', 1, [
      buildCompetency('competency-1', 3),
      buildCompetency('competency-2', 1),
    ])
    const scores = [
      buildScore('competency-1', ScoreLevel.VeryGood),
      buildScore('competency-2', ScoreLevel.VeryBad),
    ]

    expect(getWeightedScoreLevel(scores, [category])).toBe(2)
  })

  it('weights categories against each other', () => {
    const categories = [
      buildCategory('category-1', 3, [buildCompetency('competency-1', 1)]),
      buildCategory('category-2', 1, [buildCompetency('competency-2', 1)]),
    ]
    const scores = [
      buildScore('competency-1', ScoreLevel.VeryGood),
      buildScore('competency-2', ScoreLevel.VeryBad),
    ]

    expect(getWeightedScoreLevel(scores, categories)).toBe(2)
  })

  it('ignores the weight of a category that has no score at all', () => {
    const categories = [
      buildCategory('category-1', 1, [buildCompetency('competency-1', 1)]),
      buildCategory('category-2', 9, [buildCompetency('competency-2', 1)]),
    ]

    expect(getWeightedScoreLevel([buildScore('competency-1', ScoreLevel.Good)], categories)).toBe(2)
  })

  it('ignores an unscored competency inside a scored category', () => {
    const category = buildCategory('category-1', 1, [
      buildCompetency('competency-1', 1),
      buildCompetency('competency-2', 1),
    ])

    expect(getWeightedScoreLevel([buildScore('competency-1', ScoreLevel.Bad)], [category])).toBe(4)
  })

  it('averages several scores for the same competency', () => {
    const category = buildCategory('category-1', 1, [buildCompetency('competency-1', 1)])
    const scores = [
      buildScore('competency-1', ScoreLevel.VeryGood),
      buildScore('competency-1', ScoreLevel.Ok),
    ]

    expect(getWeightedScoreLevel(scores, [category])).toBe(2)
  })
})
