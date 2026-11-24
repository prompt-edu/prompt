import { describe, expect, it } from 'vitest'

import {
  describeVisibleResultSections,
  getVisibleResultSections,
  hasVisibleResultSection,
} from './studentResultsVisibility'

const hidden = {
  gradeSuggestionVisible: false,
  actionItemsVisible: false,
  gradingSheetVisible: false,
}

describe('getVisibleResultSections', () => {
  it('lists every visible section in release dialog order', () => {
    expect(
      getVisibleResultSections({
        gradeSuggestionVisible: true,
        actionItemsVisible: true,
        gradingSheetVisible: true,
      }),
    ).toEqual(['grade suggestions', 'general remarks', 'action items', 'grading sheet'])
  })

  it('lists the general remarks with the grade suggestions, not the action items', () => {
    expect(getVisibleResultSections({ ...hidden, gradeSuggestionVisible: true })).toEqual([
      'grade suggestions',
      'general remarks',
    ])
    expect(getVisibleResultSections({ ...hidden, actionItemsVisible: true })).toEqual([
      'action items',
    ])
  })

  it('leaves out hidden sections', () => {
    expect(getVisibleResultSections({ ...hidden, gradingSheetVisible: true })).toEqual([
      'grading sheet',
    ])
  })

  it('is empty when every section is hidden or the config is missing', () => {
    expect(getVisibleResultSections(hidden)).toEqual([])
    expect(getVisibleResultSections(undefined)).toEqual([])
  })
})

describe('describeVisibleResultSections', () => {
  it('joins one or two sections with "and"', () => {
    expect(describeVisibleResultSections({ ...hidden, actionItemsVisible: true })).toBe(
      'action items',
    )
    expect(describeVisibleResultSections({ ...hidden, gradeSuggestionVisible: true })).toBe(
      'grade suggestions and general remarks',
    )
  })

  it('separates three or more sections with commas', () => {
    expect(
      describeVisibleResultSections({
        gradeSuggestionVisible: true,
        actionItemsVisible: true,
        gradingSheetVisible: true,
      }),
    ).toBe('grade suggestions, general remarks, action items, and grading sheet')
  })

  it('is empty when nothing is visible', () => {
    expect(describeVisibleResultSections(hidden)).toBe('')
  })
})

describe('hasVisibleResultSection', () => {
  it('is true when any one section is visible', () => {
    expect(hasVisibleResultSection({ ...hidden, gradeSuggestionVisible: true })).toBe(true)
    expect(hasVisibleResultSection({ ...hidden, actionItemsVisible: true })).toBe(true)
    expect(hasVisibleResultSection({ ...hidden, gradingSheetVisible: true })).toBe(true)
  })

  it('is false when every section is hidden or the config is missing', () => {
    expect(hasVisibleResultSection(hidden)).toBe(false)
    expect(hasVisibleResultSection(undefined)).toBe(false)
  })
})
