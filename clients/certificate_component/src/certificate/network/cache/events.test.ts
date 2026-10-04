import { QueryClient } from '@tanstack/react-query'
import { beforeEach, describe, expect, it } from 'vitest'

import { certificateCache } from './events'
import { certificateKeys } from './keys'

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

const phaseKeys = (phaseId: string) => [
  certificateKeys.config(phaseId),
  certificateKeys.participants(phaseId),
  certificateKeys.myStatus(phaseId),
]

beforeEach(() => {
  queryClient = new QueryClient()
  seed(...phaseKeys(PHASE), ...phaseKeys(OTHER_PHASE))
})

describe('certificateCache', () => {
  it.each([
    ['configChanged', certificateCache.configChanged, certificateKeys.config(PHASE)],
    [
      'certificateDownloaded',
      certificateCache.certificateDownloaded,
      certificateKeys.participants(PHASE),
    ],
    [
      'myCertificateDownloaded',
      certificateCache.myCertificateDownloaded,
      certificateKeys.myStatus(PHASE),
    ],
  ] as const)('%s invalidates exactly its own cache in this phase', (_, fire, expected) => {
    fire(queryClient, PHASE)

    for (const key of [...phaseKeys(PHASE), ...phaseKeys(OTHER_PHASE)]) {
      expect(isInvalidated(key), JSON.stringify(key)).toBe(
        key[0] === expected[0] && key[1] === PHASE,
      )
    }
  })

  it('leaves the shared-state participants entry alone', () => {
    seed(['participants', PHASE])

    certificateCache.certificateDownloaded(queryClient, PHASE)

    expect(isInvalidated(['participants', PHASE])).toBe(false)
  })
})

describe('a key whose scoping id is missing', () => {
  it('is truncated at the missing segment rather than matching nothing', () => {
    certificateCache.configChanged(queryClient, undefined)

    expect(isInvalidated(certificateKeys.config(PHASE))).toBe(true)
    expect(isInvalidated(certificateKeys.config(OTHER_PHASE))).toBe(true)
    expect(isInvalidated(certificateKeys.myStatus(PHASE))).toBe(false)
  })
})
