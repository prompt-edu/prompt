import { describe, expect, it } from 'vitest'

import { presentationKeys } from './keys'

const PHASE = 'phase-1'
const PRESENTATION = 'presentation-1'

describe('presentationKeys', () => {
  it('makes the own presentation a descendant of the phase presentations', () => {
    expect(presentationKeys.presentations.inPhase(PHASE)).toEqual(['presentations', PHASE])
    expect(presentationKeys.presentations.own(PHASE)).toEqual(['presentations', PHASE, 'own'])
  })

  it('keeps a missing id in the key rather than coercing it', () => {
    expect(presentationKeys.slots(undefined)).toEqual(['presentation-slots', undefined])
  })

  it('builds the phase keys', () => {
    expect(presentationKeys.slots(PHASE)).toEqual(['presentation-slots', PHASE])
    expect(presentationKeys.targets(PHASE)).toEqual(['presentation-targets', PHASE])
    expect(presentationKeys.config(PHASE)).toEqual(['presentation-config', PHASE])
    expect(presentationKeys.categories(PHASE)).toEqual(['presentation-categories', PHASE])
  })

  it('builds the per-presentation keys as a prefix hierarchy', () => {
    expect(presentationKeys.materials.inPhase(PHASE)).toEqual(['presentation-materials', PHASE])
    expect(presentationKeys.materials.ofPresentation(PHASE, PRESENTATION)).toEqual([
      'presentation-materials',
      PHASE,
      PRESENTATION,
    ])
    expect(presentationKeys.feedback.inPhase(PHASE)).toEqual(['presentation-feedback', PHASE])
    expect(presentationKeys.feedback.ofPresentation(PHASE, PRESENTATION)).toEqual([
      'presentation-feedback',
      PHASE,
      PRESENTATION,
    ])
  })
})
