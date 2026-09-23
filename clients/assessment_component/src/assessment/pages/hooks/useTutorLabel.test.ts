import { describe, expect, it } from 'vitest'
import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'
import { getTutorLabel } from './useTutorLabel'

const configWithTutorDisplayName = (tutorDisplayName: string) =>
  ({ tutorDisplayName }) as CoursePhaseConfig

describe('getTutorLabel', () => {
  it('falls back to Tutor while no name is configured', () => {
    expect(getTutorLabel(undefined)).toBe('Tutor')
    expect(getTutorLabel(configWithTutorDisplayName(''))).toBe('Tutor')
  })

  it('uses the configured name', () => {
    expect(getTutorLabel(configWithTutorDisplayName('Coach'))).toBe('Coach')
  })
})
