import { PassStatus } from '@tumaet/prompt-shared-state'
import { describe, expect, it } from 'vitest'

import type { ParsedCampusCsv, StudentGradeEntry } from '../interfaces/campusGradeExport'
import { fillCampusGrades } from './fillCampusGrades'

const HEADER_ROW = [
  'REGISTRATION_NUMBER',
  'FAMILY_NAME_OF_STUDENT',
  'FIRST_NAME_OF_STUDENT',
  'GRADE',
  'DATE_OF_ASSESSMENT',
]

const COLUMN_INDICES = {
  registrationNumber: 0,
  familyName: 1,
  firstName: 2,
  grade: 3,
  dateOfAssessment: 4,
}

const buildCsv = (
  dataRows: string[][],
  columnIndices: Partial<typeof COLUMN_INDICES> = {},
): ParsedCampusCsv => ({
  fileName: 'assessment-list.csv',
  rows: [HEADER_ROW, ...dataRows],
  columnIndices: { ...COLUMN_INDICES, ...columnIndices },
  delimiter: ';',
  newline: '\r\n',
  hasTrailingNewline: true,
  encoding: 'windows-1252',
})

const buildEntry = (overrides: Partial<StudentGradeEntry> = {}): StudentGradeEntry => ({
  courseParticipationID: 'participation-1',
  firstName: 'Anna',
  lastName: 'Müller',
  matriculationNumber: '03712345',
  grade: 1.7,
  completedAt: '2026-02-17T12:00:00Z',
  passStatus: PassStatus.PASSED,
  ...overrides,
})

describe('fillCampusGrades', () => {
  it('matches a padded registration number against an unpadded one', () => {
    const csv = buildCsv([['3712345', 'Müller', 'Anna', '', '']])

    const { csv: filledCsv, report } = fillCampusGrades(csv, [buildEntry()])

    expect(filledCsv.rows[1]).toEqual(['3712345', 'Müller', 'Anna', '1.7', '17.02.2026'])
    expect(report.filled).toEqual([
      {
        csvRowNumber: 2,
        registrationNumber: '3712345',
        studentName: 'Anna Müller',
        grade: '1.7',
        dateOfAssessment: '17.02.2026',
        matchedBy: 'registrationNumber',
      },
    ])
    expect(report.totalDataRows).toBe(1)
  })

  it('falls back to a case- and diacritic-insensitive name match', () => {
    const csv = buildCsv([['', 'MÜLLER', 'ANNA', '', '']])

    const { report } = fillCampusGrades(csv, [buildEntry()])

    expect(report.filled[0].matchedBy).toBe('name')
    expect(report.nameFallbackCount).toBe(1)
  })

  it('leaves a row untouched when its registration number matches two students', () => {
    const csv = buildCsv([['3712345', 'Müller', 'Anna', '', '']])
    const entries = [
      buildEntry(),
      buildEntry({ courseParticipationID: 'participation-2', firstName: 'Anne' }),
    ]

    const { csv: filledCsv, report } = fillCampusGrades(csv, entries)

    expect(filledCsv.rows[1]).toEqual(['3712345', 'Müller', 'Anna', '', ''])
    expect(report.filled).toHaveLength(0)
    expect(report.skipped).toEqual([
      {
        csvRowNumber: 2,
        registrationNumber: '3712345',
        studentName: 'Anna Müller',
        reason: 'ambiguous',
      },
    ])
  })

  it('leaves a row untouched when its name matches two students', () => {
    const csv = buildCsv([['', 'Müller', 'Anna', '', '']])
    const entries = [
      buildEntry({ matriculationNumber: '03712345' }),
      buildEntry({ courseParticipationID: 'participation-2', matriculationNumber: '03799999' }),
    ]

    const { report } = fillCampusGrades(csv, entries)

    expect(report.skipped[0].reason).toBe('ambiguous')
  })

  it('grows a short row so the grade and date cells exist', () => {
    const csv = buildCsv([['3712345', 'Müller', 'Anna']])

    const { csv: filledCsv } = fillCampusGrades(csv, [buildEntry()])

    expect(filledCsv.rows[1]).toEqual(['3712345', 'Müller', 'Anna', '1.7', '17.02.2026'])
  })

  it('does not mutate the rows it was given', () => {
    const csv = buildCsv([['3712345', 'Müller', 'Anna', '', '']])
    const originalRow = [...csv.rows[1]]

    fillCampusGrades(csv, [buildEntry()])

    expect(csv.rows[1]).toEqual(originalRow)
  })

  it('reports a replaced grade instead of silently overwriting it', () => {
    const csv = buildCsv([['3712345', 'Müller', 'Anna', '3.0', '01.01.2026']])

    const { report } = fillCampusGrades(csv, [buildEntry()])

    expect(report.overwrittenCount).toBe(1)
    expect(report.filled[0].overwrittenGrade).toBe('3.0')
  })

  it('skips a matched student who has no grade in PROMPT', () => {
    const csv = buildCsv([['3712345', 'Müller', 'Anna', '', '']])

    const { csv: filledCsv, report } = fillCampusGrades(csv, [buildEntry({ grade: null })])

    expect(filledCsv.rows[1][COLUMN_INDICES.grade]).toBe('')
    expect(report.skipped[0].reason).toBe('noGradeInPrompt')
  })

  it('reports a row that matches no student as unmatched', () => {
    const csv = buildCsv([['9999999', 'Schmidt', 'Bernd', '', '']])

    const { report } = fillCampusGrades(csv, [buildEntry()])

    expect(report.skipped[0].reason).toBe('unmatched')
  })

  it('reports graded students that no row consumed', () => {
    const csv = buildCsv([['3712345', 'Müller', 'Anna', '', '']])
    const missing = buildEntry({
      courseParticipationID: 'participation-2',
      firstName: 'Bernd',
      lastName: 'Schmidt',
      matriculationNumber: '03799999',
      grade: 2.3,
    })

    const { report } = fillCampusGrades(csv, [buildEntry(), missing])

    expect(report.gradedStudentsMissingFromCsv).toEqual([
      {
        courseParticipationID: 'participation-2',
        studentName: 'Bernd Schmidt',
        matriculationNumber: '03799999',
        grade: 2.3,
      },
    ])
  })

  it('counts a failed student whose grade is not 5.0', () => {
    const csv = buildCsv([['3712345', 'Müller', 'Anna', '', '']])

    const { report } = fillCampusGrades(csv, [buildEntry({ passStatus: PassStatus.FAILED })])

    expect(report.passStatusMismatchCount).toBe(1)
  })

  it('keeps the date format the uploaded file already uses', () => {
    const csv = buildCsv([
      ['3799999', 'Schmidt', 'Bernd', '2.0', '2026-01-01'],
      ['3712345', 'Müller', 'Anna', '', ''],
    ])

    const { csv: filledCsv } = fillCampusGrades(csv, [buildEntry()])

    expect(filledCsv.rows[2][COLUMN_INDICES.dateOfAssessment]).toBe('2026-02-17')
  })

  it('ignores blank rows', () => {
    const csv = buildCsv([[''], ['3712345', 'Müller', 'Anna', '', '']])

    const { report } = fillCampusGrades(csv, [buildEntry()])

    expect(report.totalDataRows).toBe(1)
  })

  it('does not match by name when the file has no name columns', () => {
    const csv = buildCsv([['', '', '', '', '']], { familyName: -1, firstName: -1 })

    const { report } = fillCampusGrades(csv, [buildEntry()])

    expect(report.skipped[0].reason).toBe('unmatched')
  })
})
