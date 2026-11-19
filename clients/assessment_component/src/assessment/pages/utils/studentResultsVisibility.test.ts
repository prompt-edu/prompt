import { describe, expect, it } from 'vitest'

import { getVisibleResultSections, hasVisibleResultSection } from './studentResultsVisibility'

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
    ).toEqual(['grade suggestions', 'action items', 'grading sheet'])
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
