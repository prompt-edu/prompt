import { describe, expect, it } from 'vitest'
import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'
import { getTutorLabel } from './useTutorLabel'

const configWithTutorDisplayName = (tutorDisplayName: string) =>
  ({ tutorDisplayName }) as CoursePhaseConfig

describe('getTutorLabel', () => {
  it('falls back to Tutor in titles and tutor in running text while no name is configured', () => {
    const expected = { title: 'Tutor', text: 'tutor' }
    expect(getTutorLabel(undefined)).toEqual(expected)
    expect(getTutorLabel(configWithTutorDisplayName(''))).toEqual(expected)
  })

  it('uses the configured name unchanged in titles and running text', () => {
    expect(getTutorLabel(configWithTutorDisplayName('Coach'))).toEqual({
      title: 'Coach',
      text: 'Coach',
    })
    expect(getTutorLabel(configWithTutorDisplayName('PL'))).toEqual({ title: 'PL', text: 'PL' })
  })
})
