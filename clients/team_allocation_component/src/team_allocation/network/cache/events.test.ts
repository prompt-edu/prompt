import { QueryClient } from '@tanstack/react-query'
import { beforeEach, describe, expect, it } from 'vitest'

import { teamAllocationCache } from './events'
import { teamAllocationKeys } from './keys'

const PHASE = 'phase-1'
const OTHER_PHASE = 'phase-2'

let queryClient: QueryClient

const seed = (...keys: readonly (readonly unknown[])[]): void => {
  for (const key of keys) {
    queryClient.setQueryData(key, 'seeded')
  }
}

const isInvalidated = (key: readonly unknown[]): boolean =>
  queryClient.getQueryState(key)?.isInvalidated === true

beforeEach(() => {
  queryClient = new QueryClient()
})

describe('teamsChanged', () => {
  it('invalidates both team caches, the allocations and everything the survey derives', () => {
    const keys = [
      teamAllocationKeys.teams(PHASE),
      teamAllocationKeys.tease.teams(PHASE),
      teamAllocationKeys.allocations(PHASE),
      teamAllocationKeys.survey.form(PHASE),
      teamAllocationKeys.survey.statistics(PHASE),
      teamAllocationKeys.tease.students(PHASE),
      teamAllocationKeys.config(PHASE),
    ]
    seed(...keys)

    teamAllocationCache.teamsChanged(queryClient, PHASE)

    for (const key of keys) {
      expect(isInvalidated(key)).toBe(true)
    }
  })

  it('leaves the skills, the timeframe and another phase alone', () => {
    seed(
      teamAllocationKeys.skills(PHASE),
      teamAllocationKeys.survey.timeframe(PHASE),
      teamAllocationKeys.teams(OTHER_PHASE),
    )

    teamAllocationCache.teamsChanged(queryClient, PHASE)

    expect(isInvalidated(teamAllocationKeys.skills(PHASE))).toBe(false)
    expect(isInvalidated(teamAllocationKeys.survey.timeframe(PHASE))).toBe(false)
    expect(isInvalidated(teamAllocationKeys.teams(OTHER_PHASE))).toBe(false)
  })
})

describe('skillsChanged', () => {
  it('invalidates both skill caches and everything the survey derives', () => {
    const keys = [
      teamAllocationKeys.skills(PHASE),
      teamAllocationKeys.tease.skills(PHASE),
      teamAllocationKeys.survey.form(PHASE),
      teamAllocationKeys.survey.statistics(PHASE),
      teamAllocationKeys.tease.students(PHASE),
      teamAllocationKeys.config(PHASE),
    ]
    seed(...keys)

    teamAllocationCache.skillsChanged(queryClient, PHASE)

    for (const key of keys) {
      expect(isInvalidated(key)).toBe(true)
    }
  })

  it('leaves the teams and the allocations alone', () => {
    seed(teamAllocationKeys.teams(PHASE), teamAllocationKeys.allocations(PHASE))

    teamAllocationCache.skillsChanged(queryClient, PHASE)

    expect(isInvalidated(teamAllocationKeys.teams(PHASE))).toBe(false)
    expect(isInvalidated(teamAllocationKeys.allocations(PHASE))).toBe(false)
  })
})

describe('surveyTimeframeChanged', () => {
  it('invalidates the timeframe, the survey form carrying the deadline, and the config', () => {
    seed(
      teamAllocationKeys.survey.timeframe(PHASE),
      teamAllocationKeys.survey.form(PHASE),
      teamAllocationKeys.config(PHASE),
      teamAllocationKeys.survey.statistics(PHASE),
    )

    teamAllocationCache.surveyTimeframeChanged(queryClient, PHASE)

    expect(isInvalidated(teamAllocationKeys.survey.timeframe(PHASE))).toBe(true)
    expect(isInvalidated(teamAllocationKeys.survey.form(PHASE))).toBe(true)
    expect(isInvalidated(teamAllocationKeys.config(PHASE))).toBe(true)
    expect(isInvalidated(teamAllocationKeys.survey.statistics(PHASE))).toBe(false)
  })
})

describe('tutorsImported', () => {
  it('invalidates the team caches, which carry the tutors', () => {
    seed(teamAllocationKeys.teams(PHASE), teamAllocationKeys.tease.teams(PHASE))

    teamAllocationCache.tutorsImported(queryClient, PHASE)

    expect(isInvalidated(teamAllocationKeys.teams(PHASE))).toBe(true)
    expect(isInvalidated(teamAllocationKeys.tease.teams(PHASE))).toBe(true)
  })

  it('leaves the allocations and the survey alone', () => {
    seed(teamAllocationKeys.allocations(PHASE), teamAllocationKeys.survey.form(PHASE))

    teamAllocationCache.tutorsImported(queryClient, PHASE)

    expect(isInvalidated(teamAllocationKeys.allocations(PHASE))).toBe(false)
    expect(isInvalidated(teamAllocationKeys.survey.form(PHASE))).toBe(false)
  })
})

describe('a key whose scoping id is missing', () => {
  it('is truncated at the missing segment rather than matching nothing', () => {
    seed(teamAllocationKeys.teams(PHASE), teamAllocationKeys.teams(OTHER_PHASE))

    teamAllocationCache.tutorsImported(queryClient, undefined)

    expect(isInvalidated(teamAllocationKeys.teams(PHASE))).toBe(true)
    expect(isInvalidated(teamAllocationKeys.teams(OTHER_PHASE))).toBe(true)
  })

  it('keeps the namespace, so other caches stay untouched', () => {
    seed(teamAllocationKeys.allocations(PHASE))

    teamAllocationCache.tutorsImported(queryClient, undefined)

    expect(isInvalidated(teamAllocationKeys.allocations(PHASE))).toBe(false)
  })
})
