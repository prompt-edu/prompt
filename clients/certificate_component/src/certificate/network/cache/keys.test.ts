import { describe, expect, it } from 'vitest'

import { certificateKeys } from './keys'

const PHASE = 'phase-1'

describe('certificateKeys', () => {
  it('builds the phase keys', () => {
    expect(certificateKeys.config(PHASE)).toEqual(['config', PHASE])
    expect(certificateKeys.myStatus(PHASE)).toEqual(['certificateStatus', PHASE])
  })

  it('keeps the participant list off the shared-state participants entry', () => {
    expect(certificateKeys.participants(PHASE)).toEqual(['certificate-participants', PHASE])
    expect(certificateKeys.participants(PHASE)[0]).not.toBe('participants')
  })

  it('keeps a missing id in the key rather than coercing it', () => {
    expect(certificateKeys.config(undefined)).toEqual(['config', undefined])
  })
})
