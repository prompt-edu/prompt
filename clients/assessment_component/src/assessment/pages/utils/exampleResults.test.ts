import { describe, expect, it } from 'vitest'

import type { CategoryWithCompetencies } from '../../interfaces/category'
import type { Competency } from '../../interfaces/competency'
import { buildExampleAssessmentResults, buildExampleEvaluationResults } from './exampleResults'

const buildCompetency = (id: string): Competency => ({
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
  weight: 1,
})

const categories: CategoryWithCompetencies[] = [
  {
    id: 'category-1',
    name: 'Technical',
    shortName: 'T',
    weight: 1,
    competencies: [buildCompetency('c1'), buildCompetency('c2')],
  },
]

const allVisible = {
  gradingSheetVisible: true,
  actionItemsVisible: true,
  gradeSuggestionVisible: true,
  selfEvaluationEnabled: true,
  peerEvaluationEnabled: true,
}

describe('buildExampleAssessmentResults', () => {
  it('scores every competency and fills every visible section', () => {
    const results = buildExampleAssessmentResults('phase-1', categories, allVisible)

    expect(results.assessments.map((assessment) => assessment.competencyID)).toEqual(['c1', 'c2'])
    expect(results.studentScore?.scoreNumeric).toBeGreaterThan(0)
    expect(results.assessmentCompletion.gradeSuggestion).toBeGreaterThan(0)
    expect(results.actionItems?.length).toBeGreaterThan(0)
    expect(results.selfEvaluationResults).toHaveLength(2)
    expect(results.peerEvaluationResults).toHaveLength(2)
  })

  it('leaves out what the visibility settings hide from students', () => {
    const results = buildExampleAssessmentResults('phase-1', categories, {
      gradingSheetVisible: false,
      actionItemsVisible: false,
      gradeSuggestionVisible: false,
      selfEvaluationEnabled: true,
      peerEvaluationEnabled: true,
    })

    expect(results.assessments).toEqual([])
    expect(results.studentScore).toBeUndefined()
    expect(results.selfEvaluationResults).toEqual([])
    expect(results.peerEvaluationResults).toEqual([])
    expect(results.actionItems).toEqual([])
    expect(results.assessmentCompletion.gradeSuggestion).toBe(0)
    expect(results.assessmentCompletion.comment).toBe('')
  })
})

describe('buildExampleEvaluationResults', () => {
  it('rates every self and peer competency', () => {
    const results = buildExampleEvaluationResults('phase-1', categories, categories)

    expect(results.selfResults.map((result) => result.competencyID)).toEqual(['c1', 'c2'])
    expect(results.peerResults.map((result) => result.competencyID)).toEqual(['c1', 'c2'])
  })
})
