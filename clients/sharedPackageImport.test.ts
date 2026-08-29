import { mapScoreLevelToNumber, ScoreLevel } from '@tumaet/prompt-shared-state'
import { describe, expect, it } from 'vitest'

describe('shared package imports', () => {
  it('resolves @tumaet/prompt-shared-state from the package root', () => {
    expect(mapScoreLevelToNumber(ScoreLevel.VeryGood)).toBe(1)
  })
})
