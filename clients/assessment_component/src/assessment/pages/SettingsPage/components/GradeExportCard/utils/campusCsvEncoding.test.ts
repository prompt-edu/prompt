import { describe, expect, it } from 'vitest'

import { decodeCsvFile, encodeCsvText, getCsvMimeType } from './campusCsvEncoding'

const toBuffer = (bytes: number[]): ArrayBuffer => new Uint8Array(bytes).buffer

const UTF8_BOM = [0xef, 0xbb, 0xbf]
const WINDOWS_1252_MUELLER = [0x4d, 0xfc, 0x6c, 0x6c, 0x65, 0x72]

describe('decodeCsvFile', () => {
  it('rejects UTF-16 rather than corrupting it', () => {
    expect(() => decodeCsvFile(toBuffer([0xff, 0xfe, 0x4d, 0x00]))).toThrow(/UTF-16/)
    expect(() => decodeCsvFile(toBuffer([0xfe, 0xff, 0x00, 0x4d]))).toThrow(/UTF-16/)
  })

  it('strips a UTF-8 BOM and remembers it was there', () => {
    const bytes = [...UTF8_BOM, ...new TextEncoder().encode('Müller')]

    expect(decodeCsvFile(toBuffer(bytes))).toEqual({ text: 'Müller', encoding: 'utf-8-bom' })
  })

  it('reads a plain UTF-8 file', () => {
    const bytes = [...new TextEncoder().encode('Müller')]

    expect(decodeCsvFile(toBuffer(bytes))).toEqual({ text: 'Müller', encoding: 'utf-8' })
  })

  it('falls back to Windows-1252 for bytes that are not valid UTF-8', () => {
    expect(decodeCsvFile(toBuffer(WINDOWS_1252_MUELLER))).toEqual({
      text: 'Müller',
      encoding: 'windows-1252',
    })
  })
})

describe('encodeCsvText', () => {
  it('round trips Windows-1252 umlauts', () => {
    const encoded = encodeCsvText('Müller', 'windows-1252')

    expect([...encoded]).toEqual(WINDOWS_1252_MUELLER)
    expect(decodeCsvFile(encoded.buffer)).toEqual({ text: 'Müller', encoding: 'windows-1252' })
  })

  it('round trips the printable characters Windows-1252 keeps in the C1 range', () => {
    const encoded = encodeCsvText('€Š™', 'windows-1252')

    expect([...encoded]).toEqual([0x80, 0x8a, 0x99])
  })

  it('writes an unmappable character as a question mark', () => {
    expect([...encodeCsvText('中', 'windows-1252')]).toEqual([0x3f])
  })

  it('round trips UTF-8 without a BOM', () => {
    const encoded = encodeCsvText('Müller', 'utf-8')

    expect(decodeCsvFile(encoded.buffer)).toEqual({ text: 'Müller', encoding: 'utf-8' })
  })

  it('round trips UTF-8 with a BOM', () => {
    const encoded = encodeCsvText('Müller', 'utf-8-bom')

    expect([...encoded.slice(0, 3)]).toEqual(UTF8_BOM)
    expect(decodeCsvFile(encoded.buffer)).toEqual({ text: 'Müller', encoding: 'utf-8-bom' })
  })
})

describe('getCsvMimeType', () => {
  it('names the charset the file was written in', () => {
    expect(getCsvMimeType('windows-1252')).toBe('text/csv;charset=windows-1252')
    expect(getCsvMimeType('utf-8')).toBe('text/csv;charset=utf-8')
    expect(getCsvMimeType('utf-8-bom')).toBe('text/csv;charset=utf-8')
  })
})
