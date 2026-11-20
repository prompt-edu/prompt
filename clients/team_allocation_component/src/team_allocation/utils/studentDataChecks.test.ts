import { describe, expect, it } from 'vitest'
import { Device } from '../interfaces/tease/device'
import { Gender } from '../interfaces/tease/gender'
import { LanguageProficiency } from '../interfaces/tease/languageProficiency'
import { SkillProficiency, UNKNOWN_SKILL_PROFICIENCY } from '../interfaces/tease/skillProficiency'
import type { TeaseStudent } from '../interfaces/tease/student'
import { getLaunchWarnings, runStudentDataChecks } from './studentDataChecks'
import {
  getSurveyStatus,
  hasSubmittedSurvey,
  NOT_SUBMITTED_FILTER_SEARCH,
  SURVEY_NOT_SUBMITTED,
  SURVEY_SUBMITTED,
} from './surveyStatus'

// Shaped like the GET /tease/course_phase/:id/students response, which never fills the comments.
const completeStudent = (overrides: Partial<TeaseStudent> = {}): TeaseStudent => ({
  id: 'student-1',
  firstName: 'Ada',
  lastName: 'Lovelace',
  email: 'ada@example.com',
  gender: Gender.Female,
  nationality: 'GB',
  semester: 3,
  studyDegree: 'Master',
  studyProgram: 'Informatics',
  devices: [Device.Mac],
  introCourseProficiency: SkillProficiency.Advanced,
  introSelfEvaluation: SkillProficiency.Intermediate,
  languages: [
    { language: 'en', proficiency: LanguageProficiency.C1C2 },
    { language: 'de', proficiency: LanguageProficiency.B1B2 },
  ],
  projectPreferences: [{ projectId: 'team-1', priority: 0 }],
  skills: [],
  studentComments: [],
  tutorComments: [],
  ...overrides,
})

describe('survey status', () => {
  it('counts a student with a team preference as submitted', () => {
    expect(hasSubmittedSurvey(completeStudent())).toBe(true)
    expect(getSurveyStatus(completeStudent())).toBe(SURVEY_SUBMITTED)
  })

  it('counts a student without preferences or without data as not submitted', () => {
    expect(hasSubmittedSurvey(completeStudent({ projectPreferences: [] }))).toBe(false)
    expect(getSurveyStatus(undefined)).toBe(SURVEY_NOT_SUBMITTED)
  })

  it('builds the table filter for students without a submission', () => {
    const filters = new URLSearchParams(NOT_SUBMITTED_FILTER_SEARCH).get('filters')
    expect(JSON.parse(filters ?? '')).toEqual([
      { id: 'surveyStatus', value: [SURVEY_NOT_SUBMITTED] },
    ])
  })
})

describe('runStudentDataChecks', () => {
  it('marks every check valid when all students have all data', () => {
    const checks = runStudentDataChecks([completeStudent(), completeStudent({ id: 'student-2' })])
    expect(checks.every((check) => check.isValid)).toBe(true)
  })

  it('reports how many students have the data when only some do', () => {
    const checks = runStudentDataChecks([
      completeStudent(),
      completeStudent({ id: 'student-2', devices: [] }),
    ])
    const devices = checks.find((check) => check.label === 'Devices')
    expect(devices?.isValid).toBe(false)
    expect(devices?.completionRate).toBe(50)
    expect(devices?.problemDescription).toBe('1 of 2 students have provided devices information.')
  })

  it('does not check the comments, which the server never sends', () => {
    const labels = runStudentDataChecks([completeStudent()]).map((check) => check.label)
    expect(labels).not.toContain('Tutor Comments')
    expect(labels).not.toContain('Student Comments')
  })

  it('counts the unknown score level from the server as missing', () => {
    const checks = runStudentDataChecks([
      completeStudent({
        introCourseProficiency: UNKNOWN_SKILL_PROFICIENCY,
        introSelfEvaluation: UNKNOWN_SKILL_PROFICIENCY,
      }),
    ])
    expect(checks.find((check) => check.label === 'Score Level')?.isValid).toBe(false)
    expect(checks.find((check) => check.label === 'Self Evaluation')?.isValid).toBe(false)
  })

  it('does not round a nearly complete check up to valid', () => {
    const students = Array.from({ length: 300 }, (_, i) => completeStudent({ id: `student-${i}` }))
    students[0] = completeStudent({ id: 'student-0', devices: [] })
    const devices = runStudentDataChecks(students).find((check) => check.label === 'Devices')
    expect(devices?.completionRate).toBe(100)
    expect(devices?.isValid).toBe(false)
  })
})

describe('getLaunchWarnings', () => {
  it('returns no warnings for complete data', () => {
    expect(getLaunchWarnings([completeStudent()])).toEqual({
      incompleteChecks: [],
      studentsWithoutSurvey: 0,
      totalStudents: 1,
    })
  })

  it('lists the incomplete checks and counts the missing survey submissions', () => {
    const warnings = getLaunchWarnings([
      completeStudent(),
      completeStudent({ id: 'student-2', devices: [], projectPreferences: [] }),
    ])
    expect(warnings.incompleteChecks).toEqual(['Devices'])
    expect(warnings.studentsWithoutSurvey).toBe(1)
    expect(warnings.totalStudents).toBe(2)
  })

  it('returns no warnings when the phase has no students', () => {
    expect(getLaunchWarnings([])).toEqual({
      incompleteChecks: [],
      studentsWithoutSurvey: 0,
      totalStudents: 0,
    })
  })
})
