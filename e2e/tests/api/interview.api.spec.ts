import { request } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { BASE_URL, INTERVIEW_API } from '../../src/env'
import { FULL_COURSE_PHASES, INTERVIEW_FOREIGN_PHASE_ID } from '../../src/data/constants'

// The phase server is reached on the browser origin through the e2e nginx
// proxy (same path prefix as prod Traefik). prompt-sdk auth answers 401 both
// for missing tokens and for valid tokens lacking the required role.
const phaseUrl = (phaseId: string, path: string) =>
  `${BASE_URL}${INTERVIEW_API}/course_phase/${phaseId}/${path}`

const FULL_PHASE = FULL_COURSE_PHASES.interview.id

test.describe('interview API auth', () => {
  test('rejects unauthenticated requests', async () => {
    const anon = await request.newContext()
    try {
      const res = await anon.get(phaseUrl(FULL_PHASE, 'interview-slots'))
      expect(res.status()).toBe(401)
    } finally {
      await anon.dispose()
    }
  })

  test('rejects a student creating a slot (staff-only endpoint)', async ({ apiAs }) => {
    // interview-slots POST is Admin/CourseLecturer/CourseEditor only.
    const api = await apiAs('student')
    const res = await api.post(phaseUrl(FULL_PHASE, 'interview-slots'), { data: {} })
    expect(res.status()).toBe(401)
  })

  test('rejects a student on the admin-assign endpoint', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.post(phaseUrl(FULL_PHASE, 'interview-assignments/admin'), { data: {} })
    expect(res.status()).toBe(401)
  })

  test('rejects a token without a course-scoped role', async ({ apiAs }) => {
    // `lecturer` holds PROMPT_Lecturer only; it is CourseLecturer on fullCourse
    // but has no role on TestCourse, which owns the foreign phase.
    const api = await apiAs('lecturer')
    const res = await api.post(phaseUrl(INTERVIEW_FOREIGN_PHASE_ID, 'interview-slots'), {
      data: {},
    })
    expect(res.status()).toBe(401)
  })

  test('accepts an enrolled student reading slots (is_student round-trip via core)', async ({
    apiAs,
  }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(FULL_PHASE, 'interview-slots'))
    expect(res.status()).toBe(200)
  })

  test('rejects a student on a phase of a course they are not enrolled in', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(phaseUrl(INTERVIEW_FOREIGN_PHASE_ID, 'interview-slots'))
    expect(res.status()).toBe(401)
  })
})
