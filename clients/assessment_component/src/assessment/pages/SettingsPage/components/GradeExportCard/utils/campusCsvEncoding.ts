import type { CampusCsvEncoding, DecodedCsvFile } from '../interfaces/campusGradeExport'

/**
 * CampusOnline exports are usually Windows-1252 and occasionally UTF-8 with a
 * BOM. `TextDecoder` reads both, but `TextEncoder` only writes UTF-8, so the
 * download needs an explicit Windows-1252 encoder to keep umlauts in student
 * names intact across the round trip.
 */

const UTF8_BOM = [0xef, 0xbb, 0xbf]

/**
 * The 27 printable characters Windows-1252 places in the C1 range 0x80-0x9F,
 * as `[byte, code point]`. Every other byte maps to the identical code point.
 */
const WINDOWS_1252_C1_RANGE: [number, number][] = [
  [0x80, 0x20ac], // €
  [0x82, 0x201a], // ‚
  [0x83, 0x0192], // ƒ
  [0x84, 0x201e], // „
  [0x85, 0x2026], // …
  [0x86, 0x2020], // †
  [0x87, 0x2021], // ‡
  [0x88, 0x02c6], // ˆ
  [0x89, 0x2030], // ‰
  [0x8a, 0x0160], // Š
  [0x8b, 0x2039], // ‹
  [0x8c, 0x0152], // Œ
  [0x8e, 0x017d], // Ž
  [0x91, 0x2018], // '
  [0x92, 0x2019], // '
  [0x93, 0x201c], // "
  [0x94, 0x201d], // "
  [0x95, 0x2022], // •
  [0x96, 0x2013], // –
  [0x97, 0x2014], // —
  [0x98, 0x02dc], // ˜
  [0x99, 0x2122], // ™
  [0x9a, 0x0161], // š
  [0x9b, 0x203a], // ›
  [0x9c, 0x0153], // œ
  [0x9e, 0x017e], // ž
  [0x9f, 0x0178], // Ÿ
]

const CODE_POINT_TO_WINDOWS_1252_C1: Map<number, number> = new Map(
  WINDOWS_1252_C1_RANGE.map(([byte, codePoint]) => [codePoint, byte]),
)

const QUESTION_MARK_BYTE = 0x3f

const startsWith = (bytes: Uint8Array, prefix: number[]): boolean =>
  prefix.every((byte, index) => bytes[index] === byte)

/**
 * Decodes an uploaded CSV and remembers which encoding it was in so the
 * download can be written back the same way.
 *
 * Throws for UTF-16, which we deliberately do not support: round-tripping it is
 * not worth the code, and a clear error beats silent corruption.
 */
export const decodeCsvFile = (buffer: ArrayBuffer): DecodedCsvFile => {
  const bytes = new Uint8Array(buffer)

  if (startsWith(bytes, [0xff, 0xfe]) || startsWith(bytes, [0xfe, 0xff])) {
    throw new Error(
      'UTF-16 encoded files are not supported. Please re-export the assessment list as CSV from CampusOnline.',
    )
  }

  if (startsWith(bytes, UTF8_BOM)) {
    return {
      text: new TextDecoder('utf-8').decode(bytes.subarray(UTF8_BOM.length)),
      encoding: 'utf-8-bom',
    }
  }

  try {
    return { text: new TextDecoder('utf-8', { fatal: true }).decode(bytes), encoding: 'utf-8' }
  } catch {
    return { text: new TextDecoder('windows-1252').decode(bytes), encoding: 'windows-1252' }
  }
}

const encodeWindows1252 = (text: string): Uint8Array<ArrayBuffer> => {
  const bytes = new Uint8Array(text.length)

  for (let index = 0; index < text.length; index++) {
    const codePoint = text.charCodeAt(index)

    if (codePoint < 0x80 || (codePoint >= 0xa0 && codePoint <= 0xff)) {
      bytes[index] = codePoint
      continue
    }

    bytes[index] = CODE_POINT_TO_WINDOWS_1252_C1.get(codePoint) ?? QUESTION_MARK_BYTE
  }

  return bytes
}

/** Exact inverse of `decodeCsvFile` for the given encoding. */
export const encodeCsvText = (
  text: string,
  encoding: CampusCsvEncoding,
): Uint8Array<ArrayBuffer> => {
  if (encoding === 'windows-1252') {
    return encodeWindows1252(text)
  }

  const utf8 = new TextEncoder().encode(text)
  if (encoding === 'utf-8') {
    return utf8
  }

  const withBom = new Uint8Array(UTF8_BOM.length + utf8.length)
  withBom.set(UTF8_BOM, 0)
  withBom.set(utf8, UTF8_BOM.length)
  return withBom
}

export const getCsvMimeType = (encoding: CampusCsvEncoding): string =>
  encoding === 'windows-1252' ? 'text/csv;charset=windows-1252' : 'text/csv;charset=utf-8'
