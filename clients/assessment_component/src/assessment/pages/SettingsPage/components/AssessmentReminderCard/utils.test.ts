import { describe, expect, it } from 'vitest'
import type { CoursePhaseConfig } from '../../../../interfaces/coursePhaseConfig'
import { getReminderTypes } from './utils'

const tutorEvaluationConfig = (tutorDisplayName: string) =>
  ({ tutorEvaluationEnabled: true, tutorDisplayName }) as CoursePhaseConfig

describe('getReminderTypes', () => {
  it('lowercases the default tutor label mid-sentence', () => {
    const [tutor] = getReminderTypes(tutorEvaluationConfig(''))
    expect(tutor.label).toBe('Tutor Evaluation')
    expect(tutor.inlineLabel).toBe('tutor evaluation')
  })

  it('keeps the configured name as entered mid-sentence', () => {
    const [tutor] = getReminderTypes(tutorEvaluationConfig('PL'))
    expect(tutor.label).toBe('PL Evaluation')
    expect(tutor.inlineLabel).toBe('PL evaluation')
  })
})
