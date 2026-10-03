import { describe, expect, it } from 'vitest'
import { collapseBreadcrumbs } from './collapseBreadcrumbs'

describe('collapseBreadcrumbs', () => {
  it('keeps the first and the last crumb and hides the middle', () => {
    expect(collapseBreadcrumbs(['course', 'phase', 'participants', 'name'])).toEqual({
      first: 'course',
      hidden: ['phase', 'participants'],
      last: 'name',
    })
  })

  it('has nothing to hide for up to two crumbs', () => {
    expect(collapseBreadcrumbs(['course', 'phase'])).toEqual({
      first: 'course',
      hidden: [],
      last: 'phase',
    })
    expect(collapseBreadcrumbs(['course'])).toEqual({
      first: 'course',
      hidden: [],
      last: 'course',
    })
  })
})
