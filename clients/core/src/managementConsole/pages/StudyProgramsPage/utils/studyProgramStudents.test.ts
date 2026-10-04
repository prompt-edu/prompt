import { describe, expect, it } from 'vitest'

import {
  deleteWarning,
  renameWarning,
  reservedStudyProgramError,
  withStudentCounts,
} from './studyProgramStudents'

const computerScience = { id: 'cs', name: 'Computer Science', shortName: 'CS' }
const physics = { id: 'ph', name: 'Physics', shortName: null }

describe('withStudentCounts', () => {
  it('attaches each count to its program and defaults a missing count to zero', () => {
    const rows = withStudentCounts(
      [computerScience, physics],
      [{ studyProgramID: 'cs', studentCount: 3 }],
    )

    expect(rows).toEqual([
      { ...computerScience, studentCount: 3 },
      { ...physics, studentCount: 0 },
    ])
  })

  it('marks every count as unknown while the counts are not loaded', () => {
    const rows = withStudentCounts([computerScience], undefined)

    expect(rows).toEqual([{ ...computerScience, studentCount: null }])
  })
})

describe('reservedStudyProgramError', () => {
  it('reserves "Other" in any casing and with surrounding spaces', () => {
    const error = '"Other" is reserved for study programs applicants enter as free text.'
    expect(reservedStudyProgramError('Other')).toBe(error)
    expect(reservedStudyProgramError('  oTHER ')).toBe(error)
  })

  it('reserves "Unknown" in any casing and with surrounding spaces', () => {
    const error = '"Unknown" is reserved for applications without a study program.'
    expect(reservedStudyProgramError('Unknown')).toBe(error)
    expect(reservedStudyProgramError(' uNKNOWN  ')).toBe(error)
  })

  it('allows values that only contain a reserved word', () => {
    expect(reservedStudyProgramError('Other Sciences')).toBeUndefined()
    expect(reservedStudyProgramError('Unknown Studies')).toBeUndefined()
    expect(reservedStudyProgramError('')).toBeUndefined()
  })
})

describe('renameWarning', () => {
  it('names both programs and pluralizes the count', () => {
    expect(renameWarning(1, 'Informatics', 'Computer Science')).toBe(
      '1 student with "Informatics" will be updated to "Computer Science".',
    )
    expect(renameWarning(4, 'Informatics', 'Computer Science')).toBe(
      '4 students with "Informatics" will be updated to "Computer Science".',
    )
  })
})

describe('deleteWarning', () => {
  it('sums the students of every program being deleted', () => {
    expect(
      deleteWarning([
        { ...computerScience, studentCount: 2 },
        { ...physics, studentCount: 1 },
      ]),
    ).toBe('3 students keep their study program, which will then count as "Other".')
  })

  it('says so when no student is affected', () => {
    expect(deleteWarning([{ ...physics, studentCount: 0 }])).toBe(
      'No student has this study program.',
    )
  })

  it('does not claim zero students when the counts are unknown', () => {
    expect(deleteWarning([{ ...physics, studentCount: null }])).toBe(
      'Student counts could not be loaded. Students keep their study program, which will then count as "Other".',
    )
  })
})
