import type { Student } from '@tumaet/prompt-shared-state'
import { describe, expect, it } from 'vitest'
import type { CourseStaffMember } from '../../../interfaces/courseStaff'
import {
  buildTutorImport,
  currentTeamsOf,
  currentTutorTeams,
  isSelectableTutor,
  normalizeLogin,
  staffMemberToTutorCandidate,
  staffToTutorCandidates,
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

  it('offers lecturers and editors, a person in both groups once', () => {
    const lecturer = { ...editor, keycloakUserID: 'kc-2', universityLogin: 'ab12lec' }
    const candidates = staffToTutorCandidates({ lecturers: [lecturer, editor], editors: [editor] })
    expect(candidates.map((candidate) => candidate.id)).toEqual(['kc-2', 'kc-1'])
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

  it('normalizes a login like the server does', () => {
    expect(normalizeLogin(' CO69edt ')).toBe('co69edt')
    expect(normalizeLogin(undefined)).toBe('')
  })

  it('builds one entry per selected person with every team assigned to them', () => {
    const candidates = [
      staffMemberToTutorCandidate(editor),
      { id: 'st-1', firstName: 'Stan', lastName: 'Stan', universityLogin: 'no42tum' },
      { id: 'st-2', firstName: 'No', lastName: 'Login' },
    ]
    const teams: Record<string, string[]> = { 'kc-1': ['team-a', 'team-b'], 'st-1': ['team-a'] }

    expect(
      buildTutorImport(candidates, ['co69edt', ''], (candidate) => teams[candidate.id] ?? []),
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

  it('imports a person offered twice under the same login once', () => {
    const candidates = [
      { id: 'st-1', firstName: 'Stan', lastName: 'Stan', universityLogin: 'no42tum' },
      { id: 'st-9', firstName: 'Stan', lastName: 'Stan', universityLogin: 'NO42TUM' },
    ]

    expect(
      buildTutorImport(candidates, ['no42tum'], () => ['team-a']).map(
        (t) => t.courseParticipationID,
      ),
    ).toEqual(['st-1'])
  })

  describe('current teams', () => {
    const current = currentTutorTeams([
      { id: 'a0000005', universityLogin: 'co69edt', teamID: 'lovelace' },
      { id: 'a0000005', universityLogin: 'co69edt', teamID: 'turing' },
      { id: 'ce000001', universityLogin: 'cc12oac', teamID: 'lovelace' },
      { id: 'legacy-1', universityLogin: '', teamID: 'hamilton' },
    ])

    it('finds a person by login when their rows carry a different id', () => {
      // Course Staff offers Eduard under his Keycloak id, his rows carry a course participation id.
      expect(currentTeamsOf(current, staffMemberToTutorCandidate(editor))).toEqual([
        'lovelace',
        'turing',
      ])
    })

    it('ignores the case and surrounding whitespace of the login', () => {
      const candidate = { id: 'st-5', firstName: 'Carla', lastName: 'Coach' }
      expect(currentTeamsOf(current, { ...candidate, universityLogin: ' CC12OAC ' })).toEqual([
        'lovelace',
      ])
    })

    it('falls back to the id for rows stored without a login', () => {
      const candidate = {
        id: 'legacy-1',
        firstName: 'Bob',
        lastName: 'B',
        universityLogin: 'bo01bbb',
      }
      expect(currentTeamsOf(current, candidate)).toEqual(['hamilton'])
    })

    it('starts someone who leads no team with none', () => {
      const candidate = {
        id: 'st-7',
        firstName: 'New',
        lastName: 'Tutor',
        universityLogin: 'nt01uto',
      }
      expect(currentTeamsOf(current, candidate)).toEqual([])
    })
  })
})
