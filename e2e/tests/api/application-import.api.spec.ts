import { test, expect, request, APIRequestContext } from '@playwright/test'
import { CORE_API_URL } from '../../src/env'
import { apiContextFor } from '../../src/fixtures/api'
import { FULL_COURSE_PHASES, IMPORT_APPLICATION_PHASE_ID } from '../../src/data/constants'

// Two students imported into the import-mode phase. One deliberately has no matriculation number
// to exercise the university-login-only identity path.
const IMPORT_STUDENTS = [
  {
    firstName: 'Ingrid',
    lastName: 'Import',
    email: 'e2e-import-one@example.com',
    universityLogin: 'im01abc',
    matriculationNumber: '01900001',
  },
  {
    firstName: 'Ivan',
    lastName: 'Import',
    email: 'e2e-import-two@example.com',
    universityLogin: 'im02abc',
    matriculationNumber: '',
  },
]

const QUESTION_TITLE = 'Preferred Team'

function importBody(passStatus: 'passed' | 'not_assessed', teamAnswers: Record<string, string>) {
  return {
    passStatus,
    newQuestions: [{ columnKey: 'team', title: QUESTION_TITLE, allowedLength: 100 }],
    rows: IMPORT_STUDENTS.map((student) => ({
      student: {
        firstName: student.firstName,
        lastName: student.lastName,
        email: student.email,
        universityLogin: student.universityLogin,
        matriculationNumber: student.matriculationNumber,
        hasUniversityAccount: true,
        gender: 'prefer_not_to_say',
        nationality: 'DE',
        studyDegree: 'bachelor',
        studyProgram: 'Computer Science',
        currentSemester: 3,
      },
      answers: [{ columnKey: 'team', answer: teamAnswers[student.universityLogin] ?? '' }],
    })),
  }
}

interface ApplicationParticipation {
  courseParticipationID: string
  passStatus: string
  student?: { email?: string }
}

interface QuestionText {
  title: string
}

interface ApplicationForm {
  questionsText: QuestionText[]
}

async function participationsByEmail(
  lecturer: APIRequestContext,
  phaseID: string,
): Promise<Map<string, ApplicationParticipation>> {
  const res = await lecturer.get(`/api/applications/${phaseID}/participations`)
  expect(res.ok(), await res.text()).toBeTruthy()
  const participations = (await res.json()) as ApplicationParticipation[]
  const byEmail = new Map<string, ApplicationParticipation>()
  for (const participation of participations) {
    if (participation.student?.email) {
      byEmail.set(participation.student.email, participation)
    }
  }
  return byEmail
}

async function cleanupImportedStudents(): Promise<void> {
  const lecturer = await apiContextFor('lecturer')
  try {
    const byEmail = await participationsByEmail(lecturer, IMPORT_APPLICATION_PHASE_ID)
    const ids = IMPORT_STUDENTS.map((student) => byEmail.get(student.email)?.courseParticipationID).filter(
      (id): id is string => Boolean(id),
    )
    if (ids.length > 0) {
      await lecturer.delete(`/api/applications/${IMPORT_APPLICATION_PHASE_ID}`, { data: ids })
    }
  } finally {
    await lecturer.dispose()
  }
}

test.describe.serial('core API: CSV student import', () => {
  let anon: APIRequestContext

  test.beforeAll(async () => {
    anon = await request.newContext({ baseURL: CORE_API_URL })
  })
  test.afterAll(async () => {
    await anon.dispose()
    await cleanupImportedStudents()
  })

  test('the public application form of an import-mode phase is not served', async () => {
    const res = await anon.get(`/api/apply/${IMPORT_APPLICATION_PHASE_ID}`)
    expect(res.status()).toBe(404)
  })

  test('a public application to an import-mode phase is rejected', async () => {
    const res = await anon.post(`/api/apply/${IMPORT_APPLICATION_PHASE_ID}`, {
      data: {
        student: {
          firstName: 'Ext',
          lastName: 'Ernal',
          email: 'e2e-import-external@example.com',
          hasUniversityAccount: false,
          gender: 'female',
          nationality: 'DE',
          studyDegree: 'bachelor',
          studyProgram: 'Computer Science',
          currentSemester: 3,
        },
        answersText: [],
        answersMultiSelect: [],
        answersFileUpload: [],
      },
    })
    expect(res.status()).toBe(400)
  })

  test('importing students creates participations, questions and answers', async () => {
    const lecturer = await apiContextFor('lecturer')
    try {
      const res = await lecturer.post(`/api/applications/${IMPORT_APPLICATION_PHASE_ID}/import`, {
        data: importBody('passed', { im01abc: 'Team Rocket', im02abc: '' }),
      })
      expect(res.status(), await res.text()).toBe(201)
      const result = (await res.json()) as { created: number; updated: number; failed: number }
      expect(result.created).toBe(2)
      expect(result.updated).toBe(0)
      expect(result.failed).toBe(0)

      // Both students now participate with the chosen pass status.
      const byEmail = await participationsByEmail(lecturer, IMPORT_APPLICATION_PHASE_ID)
      for (const student of IMPORT_STUDENTS) {
        expect(byEmail.get(student.email)?.passStatus).toBe('passed')
      }

      // The mapped column became a text question on the phase.
      const formRes = await lecturer.get(`/api/applications/${IMPORT_APPLICATION_PHASE_ID}/form`)
      expect(formRes.ok(), await formRes.text()).toBeTruthy()
      const form = (await formRes.json()) as ApplicationForm
      expect(form.questionsText.some((question) => question.title === QUESTION_TITLE)).toBeTruthy()
    } finally {
      await lecturer.dispose()
    }
  })

  test('re-importing the same students updates instead of duplicating', async () => {
    const lecturer = await apiContextFor('lecturer')
    try {
      const res = await lecturer.post(`/api/applications/${IMPORT_APPLICATION_PHASE_ID}/import`, {
        data: importBody('not_assessed', { im01abc: 'Team Magma', im02abc: '' }),
      })
      expect(res.status(), await res.text()).toBe(201)
      const result = (await res.json()) as { created: number; updated: number }
      expect(result.created).toBe(0)
      expect(result.updated).toBe(2)

      // The status change from the re-import is applied to the whole batch.
      const byEmail = await participationsByEmail(lecturer, IMPORT_APPLICATION_PHASE_ID)
      for (const student of IMPORT_STUDENTS) {
        expect(byEmail.get(student.email)?.passStatus).toBe('not_assessed')
      }

      // No duplicate question was created on re-import.
      const formRes = await lecturer.get(`/api/applications/${IMPORT_APPLICATION_PHASE_ID}/form`)
      const form = (await formRes.json()) as ApplicationForm
      const matches = form.questionsText.filter((question) => question.title === QUESTION_TITLE)
      expect(matches.length).toBe(1)
    } finally {
      await lecturer.dispose()
    }
  })

  test('importing into an apply-mode phase is rejected', async () => {
    const lecturer = await apiContextFor('lecturer')
    try {
      const res = await lecturer.post(
        `/api/applications/${FULL_COURSE_PHASES.application.id}/import`,
        { data: importBody('passed', {}) },
      )
      expect(res.status()).toBe(400)
      expect(await res.text()).toContain('import mode')
    } finally {
      await lecturer.dispose()
    }
  })
})
