import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { ApplyPage } from '../../src/pages/ApplyPage'
import { StudyProgramsPage } from '../../src/pages/StudyProgramsPage'
import { FULL_COURSE_PHASES } from '../../src/data/constants'

const PHASE_ID = FULL_COURSE_PHASES.application.id

const RUN_ID = Date.now()
const PROGRAM = `E2E Program ${RUN_ID}`

interface StudyProgram {
  id: string
  name: string
}

async function deleteStudyProgramByName(name: string): Promise<void> {
  const api = await apiContextFor('admin')
  try {
    const res = await api.get('/api/study-programs')
    const studyPrograms = (await res.json()) as StudyProgram[]
    const studyProgram = studyPrograms.find((program) => program.name === name)
    if (studyProgram) {
      await api.delete(`/api/study-programs/${studyProgram.id}`)
    }
  } finally {
    await api.dispose()
  }
}

test.use({ role: 'admin' })

test.describe('application: central study program list', () => {
  test.afterAll(async () => {
    await deleteStudyProgramByName(PROGRAM)
  })

  test('a program an admin adds is offered in the public application form', async ({
    page,
    browser,
  }) => {
    const admin = new StudyProgramsPage(page)
    await admin.goto()
    await admin.add(PROGRAM, 'E2E')

    const applicantContext = await browser.newContext({
      storageState: { cookies: [], origins: [] },
    })
    try {
      const applicantPage = await applicantContext.newPage()
      const apply = new ApplyPage(applicantPage)
      await apply.goto(PHASE_ID)
      await apply.continueAsExternal()

      await applicantPage.getByRole('combobox', { name: /^Study Program/ }).click()
      await expect(applicantPage.getByRole('option', { name: PROGRAM, exact: true })).toBeVisible()
      await expect(applicantPage.getByRole('option', { name: 'Other', exact: true })).toBeVisible()
    } finally {
      await applicantContext.close()
    }
  })
})
