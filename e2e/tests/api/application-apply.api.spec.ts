import { test, expect, request, APIRequestContext } from '@playwright/test'
import { CORE_API_URL } from '../../src/env'
import { apiContextFor } from '../../src/fixtures/api'
import {
  CLOSED_APPLICATION_PHASE_ID,
  FULL_COURSE_APPLICATION_QUESTION,
  FULL_COURSE_PHASES,
} from '../../src/data/constants'

const OPEN_PHASE_ID = FULL_COURSE_PHASES.application.id
const APPLICANT_EMAIL = `e2e-api-applicant-${Date.now()}@example.com`

function applicationBody(overrides: { answersText?: { applicationQuestionID: string; answer: string }[] } = {}) {
  return {
    student: {
      firstName: 'Alex',
      lastName: 'ApiApplicant',
      email: APPLICANT_EMAIL,
      hasUniversityAccount: false,
      gender: 'female',
      nationality: 'DE',
      studyDegree: 'bachelor',
      studyProgram: 'Computer Science',
      currentSemester: 3,
    },
    answersText: overrides.answersText ?? [
      { applicationQuestionID: FULL_COURSE_APPLICATION_QUESTION.id, answer: 'Motivated via API.' },
    ],
    answersMultiSelect: [],
    answersFileUpload: [],
  }
}

interface ApplicationParticipation {
  courseParticipationID: string
  passStatus: string
  student?: { email?: string }
}

async function deleteApplication(email: string): Promise<void> {
  const lecturer = await apiContextFor('lecturer')
  try {
    const res = await lecturer.get(`/api/applications/${OPEN_PHASE_ID}/participations`)
    if (!res.ok()) return
    const participations = (await res.json()) as ApplicationParticipation[]
    const mine = participations.find((p) => p.student?.email === email)
    if (!mine) return
    await lecturer.delete(`/api/applications/${OPEN_PHASE_ID}`, {
      data: [mine.courseParticipationID],
    })
  } finally {
    await lecturer.dispose()
  }
}

// The public /apply endpoints reject invalid submissions server-side — closed
// phases, missing required answers, duplicate external applications.
test.describe('core API: public application submission', () => {
  let ctx: APIRequestContext

  test.beforeAll(async () => {
    ctx = await request.newContext({ baseURL: CORE_API_URL })
  })
  test.afterAll(async () => {
    await ctx.dispose()
    await deleteApplication(APPLICANT_EMAIL)
  })

  test('the application form of a closed phase is not served', async () => {
    const res = await ctx.get(`/api/apply/${CLOSED_APPLICATION_PHASE_ID}`)
    expect(res.status()).toBe(404)
  })

  test('applying to a closed phase is rejected', async () => {
    const res = await ctx.post(`/api/apply/${CLOSED_APPLICATION_PHASE_ID}`, {
      data: applicationBody(),
    })
    expect(res.status()).toBe(400)
    expect(await res.text()).toContain('deadline')
  })

  test('an application missing a required answer is rejected', async () => {
    const res = await ctx.post(`/api/apply/${OPEN_PHASE_ID}`, {
      data: applicationBody({ answersText: [] }),
    })
    expect(res.status()).toBe(400)
  })

  test('a valid application creates a participation; applying twice is rejected', async () => {
    const created = await ctx.post(`/api/apply/${OPEN_PHASE_ID}`, { data: applicationBody() })
    expect(created.status(), await created.text()).toBe(201)

    const lecturer = await apiContextFor('lecturer')
    try {
      const res = await lecturer.get(`/api/applications/${OPEN_PHASE_ID}/participations`)
      expect(res.ok(), await res.text()).toBeTruthy()
      const participations = (await res.json()) as ApplicationParticipation[]
      const mine = participations.find((p) => p.student?.email === APPLICANT_EMAIL)
      expect(mine?.passStatus).toBe('not_assessed')
    } finally {
      await lecturer.dispose()
    }

    const duplicate = await ctx.post(`/api/apply/${OPEN_PHASE_ID}`, { data: applicationBody() })
    expect(duplicate.status()).toBe(405)
  })
})
