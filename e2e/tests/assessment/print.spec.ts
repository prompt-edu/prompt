import { authFile, expect, test } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import {
  ASSESSMENT_FIXTURE_PHASES,
  FULL_COURSE_STUDENT,
  SEEDED_COURSES,
  SEEDED_PHASE_STUDENTS,
} from '../../src/data/constants'
import {
  apiAsRole,
  assessmentUrl,
  createCategory,
  createCompetency,
  createSchema,
  deleteOwnEvaluations,
  getAssessmentCategories,
  gradeCompetency,
  markAssessmentComplete,
  putConfig,
  releaseResults,
  resetAssessmentPhase,
  submitSelfEvaluation,
} from './helpers'
import type { CategoryWithCompetencies } from './helpers'

const PHASE_ID = ASSESSMENT_FIXTURE_PHASES.print
const SCHEMA_NAME = 'E2E Print Rubric'
const SELF_SCHEMA_NAME = 'E2E Print Self Rubric'
const COMPETENCY_NAME = 'Print Competency'
const SELF_COMPETENCY_NAME = 'I documented my work'
const REMARKS = 'Consistent, well-documented contributions across the phase.'
const STUDENT_NAME = `${SEEDED_PHASE_STUDENTS.student.firstName} ${SEEDED_PHASE_STUDENTS.student.lastName}`

const reset = async () => {
  await deleteOwnEvaluations(PHASE_ID)
  await resetAssessmentPhase(PHASE_ID, {
    courseParticipationIds: [FULL_COURSE_STUDENT.courseParticipationId],
    schemaNames: [SCHEMA_NAME, SELF_SCHEMA_NAME],
  })
}

// The `lecturer` user holds the course-scoped ios2425-iPraktikumFull-Lecturer
// role; the student results test switches to Stan's session via storageState.
test.use({ role: 'lecturer' })

test.describe('assessment: print report', () => {
  test.beforeAll(async () => {
    await reset()
    const lecturer = await apiAsRole('lecturer')
    try {
      // A graded, finalized and released assessment for Stan, so both the
      // grading view and the student results view render a print report.
      const schema = await createSchema(lecturer, PHASE_ID, SCHEMA_NAME)
      await putConfig(lecturer, PHASE_ID, { assessmentSchemaId: schema.id })
      await createCategory(lecturer, PHASE_ID, schema.id, 'Print Category')
      const category = (await getAssessmentCategories(lecturer, PHASE_ID))[0]
      await createCompetency(lecturer, PHASE_ID, category.id, COMPETENCY_NAME)
      const competency = (await getAssessmentCategories(lecturer, PHASE_ID))[0].competencies[0]
      await gradeCompetency(
        lecturer,
        PHASE_ID,
        FULL_COURSE_STUDENT.courseParticipationId,
        competency.id,
        'good',
      )
      await markAssessmentComplete(
        lecturer,
        PHASE_ID,
        FULL_COURSE_STUDENT.courseParticipationId,
        1.7,
        REMARKS,
      )
      await releaseResults(lecturer, PHASE_ID)

      // A separate self-evaluation rubric, so the self report renders its own
      // categories rather than the assessment ones.
      const selfSchema = await createSchema(lecturer, PHASE_ID, SELF_SCHEMA_NAME)
      await putConfig(lecturer, PHASE_ID, {
        selfEvaluationEnabled: true,
        selfEvaluationSchema: selfSchema.id,
      })
      await createCategory(lecturer, PHASE_ID, selfSchema.id, 'Print Self Category')
      const selfRes = await lecturer.get(
        assessmentUrl(PHASE_ID, 'category/self/with-competencies'),
      )
      const selfCategories = (await selfRes.json()) as CategoryWithCompetencies[]
      await createCompetency(
        lecturer,
        PHASE_ID,
        selfCategories[0].id,
        SELF_COMPETENCY_NAME,
      )
      const selfCompetency = (
        (await (
          await lecturer.get(assessmentUrl(PHASE_ID, 'category/self/with-competencies'))
        ).json()) as CategoryWithCompetencies[]
      )[0].competencies[0]

      const student = await apiAsRole('student')
      try {
        await submitSelfEvaluation(
          student,
          PHASE_ID,
          FULL_COURSE_STUDENT.courseParticipationId,
          selfCompetency.id,
          'good',
        )
      } finally {
        await student.dispose()
      }
    } finally {
      await lecturer.dispose()
    }
  })

  test.afterAll(async () => {
    await reset()
  })

  test('the lecturer grading view prints the report full width', async ({ page }) => {
    const phase = new AssessmentPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/participants')
    await phase.expectParticipantsLoaded()
    await phase.openParticipant(STUDENT_NAME)

    // The export control (with the "PDF / Print" item) lives bottom-right.
    await expect(page.getByRole('button', { name: 'Export' })).toBeVisible()

    // The regression guard: under print media the sidebar is gone and the
    // report spans the page instead of being squeezed to the right half.
    await phase.expectPrintReportFillsPage()
    await expect(phase.printReport()).toContainText(STUDENT_NAME)
  })

  test('the student results view shows a bottom-right print button that prints full width', async ({
    browser,
  }) => {
    const context = await browser.newContext({ storageState: authFile('student') })
    try {
      const studentPage = await context.newPage()
      // window.print() opens an un-scriptable native dialog; stub it so the
      // click is observable via a console message.
      await studentPage.addInitScript(() => {
        window.print = () => console.info('E2E_PRINT_INVOKED')
      })

      const phase = new AssessmentPage(studentPage)
      await phase.gotoResults(SEEDED_COURSES.fullCourse.id, PHASE_ID)
      await expect(studentPage.getByText('Good - 1.7')).toBeVisible()

      // The print button moved out of the top header to the bottom-right,
      // matching the assessment page convention.
      const printButton = studentPage.getByRole('button', { name: 'PDF / Print' })
      await expect(printButton).toBeVisible()
      const heading = studentPage.getByRole('heading', { name: 'Assessment Results' })
      const buttonBox = await printButton.boundingBox()
      const headingBox = await heading.boundingBox()
      const viewport = studentPage.viewportSize()
      expect(buttonBox, 'print button box').not.toBeNull()
      expect(headingBox, 'results heading box').not.toBeNull()
      expect(buttonBox!.y, 'button below the results content').toBeGreaterThan(headingBox!.y)
      expect(buttonBox!.x, 'button right-aligned').toBeGreaterThan(viewport!.width / 2)

      // Clicking it actually triggers printing (via printPage -> window.print).
      const printInvoked = studentPage.waitForEvent('console', {
        predicate: (message) => message.text().includes('E2E_PRINT_INVOKED'),
        timeout: 5_000,
      })
      await printButton.click()
      await printInvoked

      await phase.expectPrintReportFillsPage()
    } finally {
      await context.close()
    }
  })

  test('the self evaluation results view prints a clean report', async ({ page }) => {
    const phase = new AssessmentPage(page)
    await phase.goto(
      SEEDED_COURSES.fullCourse.id,
      PHASE_ID,
      `/self-evaluations/${FULL_COURSE_STUDENT.courseParticipationId}`,
    )
    await expect(
      page.getByRole('heading', { name: `Self Evaluation Results for ${STUDENT_NAME}` }),
    ).toBeVisible({ timeout: 15_000 })
    await expect(page.getByRole('button', { name: 'PDF / Print' })).toBeVisible()

    // The report is a purpose-built twin, not the interactive tree: it carries
    // the rubric description instead of the five-option score selector.
    await phase.expectPrintReportFillsPage()
    await expect(phase.printReport()).toContainText(STUDENT_NAME)
    await expect(phase.printReport()).toContainText(SELF_COMPETENCY_NAME)
  })

  test('the self evaluations overview bulk-prints one report per participant', async ({ page }) => {
    // window.print() opens an un-scriptable native dialog; stub it so the click
    // is observable via a console message.
    await page.addInitScript(() => {
      window.print = () => console.info('E2E_PRINT_INVOKED')
    })

    const phase = new AssessmentPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/self-evaluations')
    const printAll = page.getByRole('button', { name: 'PDF / Print All' })
    await expect(printAll).toBeVisible({ timeout: 15_000 })

    // The reports are mounted on demand, so nothing is in the DOM up front.
    await expect(phase.printReport()).toHaveCount(0)

    const printInvoked = page.waitForEvent('console', {
      predicate: (message) => message.text().includes('E2E_PRINT_INVOKED'),
      timeout: 10_000,
    })
    await printAll.click()
    await printInvoked

    // Only Stan has a self evaluation in this fixture; participants without one
    // are skipped rather than printed empty.
    await expect(phase.printReport()).toHaveCount(1)
    await expect(phase.printReport()).toContainText(STUDENT_NAME)
  })
})
