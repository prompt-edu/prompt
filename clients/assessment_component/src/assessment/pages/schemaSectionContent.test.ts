import { describe, expect, it } from 'vitest'
import { AssessmentType } from '../interfaces/assessmentType'
import { getSchemaSectionContent } from './schemaSectionContent'

describe('getSchemaSectionContent', () => {
  it('lowercases the default tutor label mid-sentence', () => {
    const content = getSchemaSectionContent({ title: 'Tutor', text: 'tutor' })[AssessmentType.TUTOR]
    expect(content.title).toBe('Tutor-Evaluation')
    expect(content.inlineTitle).toBe('tutor-evaluation')
    expect(content.timeframeHint).toBe(
      'This controls when students can evaluate their tutor and when submissions close.',
    )
  })

  it('keeps the configured name as entered mid-sentence', () => {
    const content = getSchemaSectionContent({ title: 'PL', text: 'PL' })[AssessmentType.TUTOR]
    expect(content.title).toBe('PL-Evaluation')
    expect(content.inlineTitle).toBe('PL-evaluation')
  })
})
