import { describe, expect, it } from 'vitest'
import { toHttpUrl } from './toHttpUrl'

describe('toHttpUrl', () => {
  it('returns the normalized URL for http and https links', () => {
    expect(toHttpUrl('https://prompt.aet.cit.tum.de')).toBe('https://prompt.aet.cit.tum.de/')
    expect(toHttpUrl('http://example.com/status?page=1')).toBe('http://example.com/status?page=1')
  })

  it('rejects links with any other scheme', () => {
    expect(toHttpUrl('javascript:alert(1)')).toBeUndefined()
    expect(toHttpUrl('JavaScript:alert(1)')).toBeUndefined()
    expect(toHttpUrl('data:text/html,<script>alert(1)</script>')).toBeUndefined()
    expect(toHttpUrl('mailto:prompt@example.com')).toBeUndefined()
  })

  it('rejects values that are not absolute URLs', () => {
    expect(toHttpUrl('')).toBeUndefined()
    expect(toHttpUrl('https://')).toBeUndefined()
    expect(toHttpUrl('/relative/path')).toBeUndefined()
    expect(toHttpUrl('not a url')).toBeUndefined()
  })
})
