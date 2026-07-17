import { test, expect, authFile } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { ApplyPage } from '../../src/pages/ApplyPage'
import { ApplicationAdminPage } from '../../src/pages/ApplicationAdminPage'
import {
  FULL_COURSE_APPLICATION_QUESTION,
  FULL_COURSE_PHASES,
  SEEDED_COURSES,
} from '../../src/data/constants'

const PHASE_ID = FULL_COURSE_PHASES.application.id
const MOTIVATION_ANSWER = 'I want to build a real iOS app with a real team.'

// Unique identity per run: assertions filter by it (never by row counts), so
// parallel spec files and leftovers from crashed runs cannot collide.
const RUN_ID = Date.now()
const APPLICANT = {
  firstName: 'Erika',
  lastName: `External${RUN_ID}`,
  email: `e2e-applicant-${RUN_ID}@example.com`,
  gender: 'Female',
  nationality: 'Germany',
  studyDegree: 'Bachelor',
  studyProgram: 'Computer Science',
  currentSemester: '3',
}

interface ApplicationParticipation {
  courseParticipationID: string
  passStatus: string
  student?: { email?: string }
}

async function findApplicationByEmail(
  email: string,
): Promise<ApplicationParticipation | undefined> {
  const api = await apiContextFor('lecturer')
  try {
    const res = await api.get(`/api/applications/${PHASE_ID}/participations`)
    if (!res.ok()) {
      throw new Error(`GET participations failed: ${res.status()} ${await res.text()}`)
    }
    const participations = (await res.json()) as ApplicationParticipation[]
    return participations.find((p) => p.student?.email === email)
  } finally {
    await api.dispose()
  }
}

async function deleteApplicationByEmail(email: string): Promise<void> {
  const participation = await findApplicationByEmail(email)
  if (!participation) return
  const api = await apiContextFor('lecturer')
  try {
    await api.delete(`/api/applications/${PHASE_ID}`, {
      data: [participation.courseParticipationID],
    })
  } finally {
    await api.dispose()
  }
}

// The applicant is not logged in: the public form must work without a session.
test.use({ storageState: { cookies: [], origins: [] } })

test.describe('application: applicant journey', () => {
  test.afterAll(async () => {
    await deleteApplicationByEmail(APPLICANT.email)
  })

  test('an external applicant applies and the lecturer accepts them', async ({
    page,
    browser,
  }) => {
    // Applicant: public form at /apply, external branch (no university account).
    const apply = new ApplyPage(page)
    await apply.goto(PHASE_ID)
    await apply.continueAsExternal()
    await apply.fillStudentForm(APPLICANT)
    await apply.answerTextQuestion(FULL_COURSE_APPLICATION_QUESTION.title, MOTIVATION_ANSWER)
    await apply.submit()
    await apply.expectSaved()

    // Lecturer in their own browser context (own session/storage).
    const lecturerContext = await browser.newContext({ storageState: authFile('lecturer') })
    try {
      const lecturerPage = await lecturerContext.newPage()
      const admin = new ApplicationAdminPage(lecturerPage)

      await admin.gotoParticipants(SEEDED_COURSES.fullCourse.id, PHASE_ID)
      await admin.expectStatus(APPLICANT.email, 'Not Assessed')

      await admin.openApplication(APPLICANT.email)
      await admin.expectQuestionVisible(FULL_COURSE_APPLICATION_QUESTION.title)
      await admin.expectAnswerVisible(MOTIVATION_ANSWER)
      await admin.accept()

      await admin.gotoParticipants(SEEDED_COURSES.fullCourse.id, PHASE_ID)
      await admin.expectStatus(APPLICANT.email, 'Accepted')
    } finally {
      await lecturerContext.close()
    }

    // Acceptance lands in the API: the application is passed and the applicant
    // exists as a course participant (the participation row is created at
    // submission; acceptance flips its pass status).
    const accepted = await findApplicationByEmail(APPLICANT.email)
    expect(accepted?.passStatus).toBe('passed')

    const api = await apiContextFor('lecturer')
    try {
      const res = await api.get(`/api/courses/${SEEDED_COURSES.fullCourse.id}/participations`)
      expect(res.ok(), await res.text()).toBeTruthy()
      const courseParticipations = (await res.json()) as { id: string }[]
      expect(
        courseParticipations.some((p) => p.id === accepted?.courseParticipationID),
      ).toBeTruthy()
    } finally {
      await api.dispose()
    }
  })
})
