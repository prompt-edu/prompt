import type { Student } from '@tumaet/prompt-shared-state'
import { describe, expect, it } from 'vitest'
import type { CourseStaffMember } from '../../../interfaces/courseStaff'
import {
  buildTutorImport,
  isSelectableTutor,
  staffMemberToTutorCandidate,
  studentToTutorCandidate,
} from './tutorImport'

const editor: CourseStaffMember = {
  keycloakUserID: 'kc-1',
  username: 'course-editor',
  email: 'editor@example.com',
  firstName: 'Eduard',
  lastName: 'Editor',
  universityLogin: 'co69edt',
}

describe('tutor import', () => {
  it('imports a course editor under their Keycloak user id', () => {
    expect(staffMemberToTutorCandidate(editor)).toEqual({
      id: 'kc-1',
      firstName: 'Eduard',
      lastName: 'Editor',
      universityLogin: 'co69edt',
    })
  })

  it('imports a student under their id', () => {
    const student = { id: 'st-1', firstName: 'Stan', lastName: 'Stan', universityLogin: 'no42tum' }
    expect(studentToTutorCandidate(student as Student)).toEqual({
      id: 'st-1',
      firstName: 'Stan',
      lastName: 'Stan',
      universityLogin: 'no42tum',
    })
  })

  it('only selects people with a university login', () => {
    expect(
      isSelectableTutor({ id: 'a', firstName: '', lastName: '', universityLogin: 'ab12cde' }),
    ).toBe(true)
    expect(isSelectableTutor({ id: 'a', firstName: '', lastName: '', universityLogin: ' ' })).toBe(
      false,
    )
    expect(isSelectableTutor({ id: 'a', firstName: '', lastName: '' })).toBe(false)
  })

  it('builds one entry per selected person with every team assigned to them', () => {
    const candidates = [
      staffMemberToTutorCandidate(editor),
      { id: 'st-1', firstName: 'Stan', lastName: 'Stan', universityLogin: 'no42tum' },
      { id: 'st-2', firstName: 'No', lastName: 'Login' },
    ]

    expect(
      buildTutorImport(candidates, ['kc-1', 'st-2'], {
        'kc-1': ['team-a', 'team-b'],
        'st-1': ['team-a'],
      }),
    ).toEqual([
      {
        courseParticipationID: 'kc-1',
        firstName: 'Eduard',
        lastName: 'Editor',
        universityLogin: 'co69edt',
        teamIDs: ['team-a', 'team-b'],
      },
    ])
  })
})
