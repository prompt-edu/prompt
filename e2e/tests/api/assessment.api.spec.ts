import { request } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { BASE_URL, ASSESSMENT_API } from '../../src/env'
import {
  ASSESSMENT_FOREIGN_PHASE_ID,
  FULL_COURSE_PHASES,
} from '../../src/data/constants'

// The assessment server is reached on the browser origin through the e2e
// nginx proxy (same path prefix as prod Traefik). prompt-sdk auth answers 401
// both for missing tokens and for valid tokens lacking the required role.
// All checks are side-effect-free (GET only) on the graph-tail assessment
// phase, so this file cannot interfere with the journey/visibility specs.
const phaseUrl = (phaseId: string, path: string) =>
  `${BASE_URL}${ASSESSMENT_API}/course_phase/${phaseId}/${path}`

const PHASE_ID = FULL_COURSE_PHASES.assessment.id

test.describe('assessment API auth', () => {
  test('rejects unauthenticated requests', async () => {
    const anon = await request.newContext()
    try {
      const res = await anon.get(phaseUrl(PHASE_ID, 'student-assessment'))
      expect(res.status()).toBe(401)
    } finally {
      await anon.dispose()
    }
  })

  test('rejects a student on a lecturer endpoint', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(PHASE_ID, 'student-assessment'))
    expect(res.status()).toBe(401)
  })

  test('rejects a lecturer on a student-only endpoint', async ({ apiAs }) => {
    // my-results requires CourseStudent; lecturers hold the Lecturer role only.
    const api = await apiAs('lecturer')
    const res = await api.get(phaseUrl(PHASE_ID, 'student-assessment/my-results'))
    expect(res.status()).toBe(401)
  })

  test('rejects a course editor on a lecturer-only endpoint', async ({ apiAs }) => {
    // Editors may grade but must not manage reminders/releases.
    const api = await apiAs('course-editor')
    const res = await api.get(phaseUrl(PHASE_ID, 'config/reminders/incomplete'))
    expect(res.status()).toBe(401)
  })

  test('accepts a course editor on a grading read endpoint', async ({ apiAs }) => {
    const api = await apiAs('course-editor')
    const res = await api.get(phaseUrl(PHASE_ID, 'student-assessment'))
    expect(res.status()).toBe(200)
  })

  test('answers an enrolled student with 204 while results are unreleased', async ({ apiAs }) => {
    // The lecturer journey never releases results on this phase, so this stays
    // deterministic regardless of spec-file ordering.
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(PHASE_ID, 'student-assessment/my-results'))
    expect(res.status()).toBe(204)
  })

  test('rejects a student on a phase of a course they are not enrolled in', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(ASSESSMENT_FOREIGN_PHASE_ID, 'student-assessment/my-results'))
    expect(res.status()).toBe(401)
  })
})
