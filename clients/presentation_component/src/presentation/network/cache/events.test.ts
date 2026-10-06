import { QueryClient } from '@tanstack/react-query'
import { beforeEach, describe, expect, it } from 'vitest'

import { presentationCache } from './events'
import { presentationKeys } from './keys'

const PHASE = 'phase-1'
const OTHER_PHASE = 'phase-2'
const PRESENTATION = 'presentation-1'
const OTHER_PRESENTATION = 'presentation-2'

let queryClient: QueryClient

const seed = (...keys: readonly (readonly unknown[])[]): void => {
  for (const key of keys) {
    queryClient.setQueryData(key, 'seeded')
  }
}

const isInvalidated = (key: readonly unknown[]): boolean =>
  queryClient.getQueryState(key)?.isInvalidated === true

const presentationData = (phaseId: string) => [
  presentationKeys.presentations.own(phaseId),
  presentationKeys.materials.ofPresentation(phaseId, PRESENTATION),
  presentationKeys.feedback.ofPresentation(phaseId, PRESENTATION),
]

beforeEach(() => {
  queryClient = new QueryClient()
})

describe('scheduleChanged', () => {
  it('invalidates the schedule and everything hanging off a presentation', () => {
    const keys = [
      presentationKeys.slots(PHASE),
      presentationKeys.targets(PHASE),
      ...presentationData(PHASE),
    ]
    seed(...keys)

    presentationCache.scheduleChanged(queryClient, PHASE)

    for (const key of keys) {
      expect(isInvalidated(key)).toBe(true)
    }
  })

  it('leaves the settings and another phase alone', () => {
    seed(
      presentationKeys.config(PHASE),
      presentationKeys.categories(PHASE),
      presentationKeys.slots(OTHER_PHASE),
    )

    presentationCache.scheduleChanged(queryClient, PHASE)

    expect(isInvalidated(presentationKeys.config(PHASE))).toBe(false)
    expect(isInvalidated(presentationKeys.categories(PHASE))).toBe(false)
    expect(isInvalidated(presentationKeys.slots(OTHER_PHASE))).toBe(false)
  })
})

describe('settingsChanged', () => {
  it('invalidates the settings, the schedule and everything a reset deletes', () => {
    const keys = [
      presentationKeys.config(PHASE),
      presentationKeys.categories(PHASE),
      presentationKeys.slots(PHASE),
      presentationKeys.targets(PHASE),
      ...presentationData(PHASE),
    ]
    seed(...keys)

    presentationCache.settingsChanged(queryClient, PHASE)

    for (const key of keys) {
      expect(isInvalidated(key)).toBe(true)
    }
  })

  it('leaves another phase alone', () => {
    seed(...presentationData(OTHER_PHASE), presentationKeys.config(OTHER_PHASE))

    presentationCache.settingsChanged(queryClient, PHASE)

    for (const key of [...presentationData(OTHER_PHASE), presentationKeys.config(OTHER_PHASE)]) {
      expect(isInvalidated(key)).toBe(false)
    }
  })
})

describe('materialsChanged', () => {
  it('invalidates the materials of this presentation and the presentation summaries', () => {
    seed(
      presentationKeys.materials.ofPresentation(PHASE, PRESENTATION),
      presentationKeys.presentations.own(PHASE),
    )

    presentationCache.materialsChanged(queryClient, PHASE, PRESENTATION)

    expect(isInvalidated(presentationKeys.materials.ofPresentation(PHASE, PRESENTATION))).toBe(true)
    expect(isInvalidated(presentationKeys.presentations.own(PHASE))).toBe(true)
  })

  it('leaves the materials of another presentation and the feedback alone', () => {
    seed(
      presentationKeys.materials.ofPresentation(PHASE, OTHER_PRESENTATION),
      presentationKeys.feedback.ofPresentation(PHASE, PRESENTATION),
    )

    presentationCache.materialsChanged(queryClient, PHASE, PRESENTATION)

    expect(
      isInvalidated(presentationKeys.materials.ofPresentation(PHASE, OTHER_PRESENTATION)),
    ).toBe(false)
    expect(isInvalidated(presentationKeys.feedback.ofPresentation(PHASE, PRESENTATION))).toBe(false)
  })
})

describe('feedbackChanged', () => {
  it('invalidates the feedback of this presentation only', () => {
    seed(
      presentationKeys.feedback.ofPresentation(PHASE, PRESENTATION),
      presentationKeys.feedback.ofPresentation(PHASE, OTHER_PRESENTATION),
      presentationKeys.presentations.own(PHASE),
    )

    presentationCache.feedbackChanged(queryClient, PHASE, PRESENTATION)

    expect(isInvalidated(presentationKeys.feedback.ofPresentation(PHASE, PRESENTATION))).toBe(true)
    expect(isInvalidated(presentationKeys.feedback.ofPresentation(PHASE, OTHER_PRESENTATION))).toBe(
      false,
    )
    expect(isInvalidated(presentationKeys.presentations.own(PHASE))).toBe(false)
  })
})

describe('feedbackStatusChanged', () => {
  it('invalidates the feedback and the presentation summaries carrying its status', () => {
    seed(
      presentationKeys.feedback.ofPresentation(PHASE, PRESENTATION),
      presentationKeys.presentations.own(PHASE),
    )

    presentationCache.feedbackStatusChanged(queryClient, PHASE, PRESENTATION)

    expect(isInvalidated(presentationKeys.feedback.ofPresentation(PHASE, PRESENTATION))).toBe(true)
    expect(isInvalidated(presentationKeys.presentations.own(PHASE))).toBe(true)
  })
})

describe('a key whose scoping id is missing', () => {
  it('is truncated at the missing segment rather than matching nothing', () => {
    seed(
      presentationKeys.feedback.ofPresentation(PHASE, PRESENTATION),
      presentationKeys.feedback.ofPresentation(PHASE, OTHER_PRESENTATION),
    )

    presentationCache.feedbackChanged(queryClient, PHASE, undefined)

    expect(isInvalidated(presentationKeys.feedback.ofPresentation(PHASE, PRESENTATION))).toBe(true)
    expect(isInvalidated(presentationKeys.feedback.ofPresentation(PHASE, OTHER_PRESENTATION))).toBe(
      true,
    )
  })

  it('keeps the segments before the missing one, so other phases stay untouched', () => {
    seed(presentationKeys.feedback.ofPresentation(OTHER_PHASE, PRESENTATION))

    presentationCache.feedbackChanged(queryClient, PHASE, undefined)

    expect(isInvalidated(presentationKeys.feedback.ofPresentation(OTHER_PHASE, PRESENTATION))).toBe(
      false,
    )
  })
})
