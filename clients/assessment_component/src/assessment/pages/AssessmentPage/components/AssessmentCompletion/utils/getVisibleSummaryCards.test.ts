import { describe, expect, it } from 'vitest'

import { getVisibleSummaryCards } from './getVisibleSummaryCards'

const config = (gradeSuggestionVisible: boolean, actionItemsVisible: boolean) => ({
  gradeSuggestionVisible,
  actionItemsVisible,
})

describe('getVisibleSummaryCards', () => {
  it('shows every card to tutors, whatever students see', () => {
    const all = { generalRemarks: true, grade: true, actionItems: true }
    expect(getVisibleSummaryCards(false, config(false, false))).toEqual(all)
    expect(getVisibleSummaryCards(false, undefined)).toEqual(all)
  })

  it('hides the grade and the general remarks from students when grade suggestions are hidden', () => {
    expect(getVisibleSummaryCards(true, config(false, true))).toEqual({
      generalRemarks: false,
      grade: false,
      actionItems: true,
    })
  })

  it('hides the action items from students when action items are hidden', () => {
    expect(getVisibleSummaryCards(true, config(true, false))).toEqual({
      generalRemarks: true,
      grade: true,
      actionItems: false,
    })
  })

  it('shows students every section that is visible', () => {
    expect(getVisibleSummaryCards(true, config(true, true))).toEqual({
      generalRemarks: true,
      grade: true,
      actionItems: true,
    })
  })

  it('hides every card from students until the config has loaded', () => {
    expect(getVisibleSummaryCards(true, undefined)).toEqual({
      generalRemarks: false,
      grade: false,
      actionItems: false,
    })
  })
})
