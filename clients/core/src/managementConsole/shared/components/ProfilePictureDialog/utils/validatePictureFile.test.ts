import { describe, expect, it } from 'vitest'
import { MAX_PICTURE_FILE_BYTES, validatePictureFile } from './validatePictureFile'

describe('validatePictureFile', () => {
  it.each(['image/jpeg', 'image/png', 'image/webp'])('accepts %s', (type) => {
    expect(validatePictureFile({ type, size: 1024 })).toBeNull()
  })

  it.each(['image/gif', 'image/svg+xml', 'application/pdf', ''])('rejects %s', (type) => {
    expect(validatePictureFile({ type, size: 1024 })).toMatch(/JPEG, PNG, or WebP/)
  })

  it('accepts a file at the size limit and rejects one above it', () => {
    expect(validatePictureFile({ type: 'image/png', size: MAX_PICTURE_FILE_BYTES })).toBeNull()
    expect(validatePictureFile({ type: 'image/png', size: MAX_PICTURE_FILE_BYTES + 1 })).toMatch(
      /smaller than 5 MB/,
    )
  })
})
