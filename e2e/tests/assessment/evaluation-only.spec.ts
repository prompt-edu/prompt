import { ASSESSMENT_FIXTURE_PHASES, SEEDED_COURSES } from '../../src/data/constants'
import { authFile, expect, test } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import { apiAsRole, assessmentUrl, getConfig, putConfig, resetAssessmentPhase } from './helpers'

const PHASE_ID = ASSESSMENT_FIXTURE_PHASES.evaluationOnly

// The `lecturer` user holds the course-scoped ios2425-iPraktikumFull-Lecturer role.
test.use({ role: 'lecturer' })

test.describe('assessment: evaluation-only phase', () => {
  test.beforeAll(async () => {
    // Also re-enables the assessment, so a previous run's disabled state never
    // makes the first assertion pass for the wrong reason.
    await resetAssessmentPhase(PHASE_ID)
    // The participants table only carries an evaluation column for an enabled
    // evaluation type, so the self evaluation is switched on for this phase.
    const admin = await apiAsRole('admin')
    try {
      await putConfig(admin, PHASE_ID, { assessmentEnabled: true, selfEvaluationEnabled: true })
    } finally {
      await admin.dispose()
    }
  })

  test.afterAll(async () => {
    await resetAssessmentPhase(PHASE_ID)
  })

  test('a lecturer turns the assessment off and releases evaluation results', async ({
    page,
    browser,
  }) => {
    const phase = new AssessmentPage(page)

    // The grading half of the phase is present while the assessment is enabled.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/settings')
    await phase.expectSettingsLoaded()
    await phase.expectAssessmentControls(true)

    await phase.setAssessmentEnabled(false)
    await phase.saveAssessmentSettings()

    // Schema, timeframe and the release-visibility switches go with it.
    await phase.expectAssessmentControls(false)

    // The mode is persisted, not just local card state.
    await page.reload()
    await phase.expectSettingsLoaded()
    await expect(phase.assessmentEnabledSwitch()).toHaveAttribute('aria-checked', 'false')
    await phase.expectAssessmentControls(false)

    const lecturerApi = await apiAsRole('lecturer')
    try {
      expect((await getConfig(lecturerApi, PHASE_ID)).assessmentEnabled).toBe(false)
    } finally {
      await lecturerApi.dispose()
    }

    // The participants table keeps the evaluation columns and drops the
    // assessment ones.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/participants')
    await phase.expectEvaluationParticipantsLoaded()
    await expect(phase.columnHeader('Score Level')).toHaveCount(0)
    await expect(phase.columnHeader('Grade')).toHaveCount(0)
    await expect(phase.columnHeader('Self Eval')).toBeVisible()

    // Statistics has nothing to show without grading.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/statistics')
    await phase.expectAssessmentDisabledNotice()

    // Nothing has to be marked as final first, so the release gate is open and
    // the button drops the "n/m final" counter.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/settings')
    await phase.expectSettingsLoaded()
    await expect(phase.releaseButton()).toBeEnabled()
    await expect(phase.releaseButton()).toHaveText('Release Results')

    const studentApi = await apiAsRole('student')
    try {
      const before = await studentApi.get(assessmentUrl(PHASE_ID, 'evaluation/my-results'))
      expect(before.status()).toBe(204)

      await phase.confirmRelease('Evaluation')

      // Released evaluation results reach the student; the assessment report
      // stays closed on an evaluation-only phase.
      const after = await studentApi.get(assessmentUrl(PHASE_ID, 'evaluation/my-results'))
      expect(after.status()).toBe(200)
      const results = (await after.json()) as {
        selfResults: unknown[]
        peerResults: unknown[]
      }
      expect(results.selfResults).toEqual([])
      expect(results.peerResults).toEqual([])

      const assessmentResults = await studentApi.get(
        assessmentUrl(PHASE_ID, 'student-assessment/my-results'),
      )
      expect(assessmentResults.status()).toBe(204)
    } finally {
      await studentApi.dispose()
    }

    // Stan gets the evaluation results page rather than the assessment report.
    // He submitted nothing here, so it says so instead of spinning.
    const stanContext = await browser.newContext({ storageState: authFile('student') })
    try {
      const stanPage = await stanContext.newPage()
      await new AssessmentPage(stanPage).goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/results')
      await expect(stanPage.getByRole('heading', { name: 'Evaluation Results' })).toBeVisible({
        timeout: 15_000,
      })
      await expect(
        stanPage.getByText('No evaluation results are available for you in this phase yet.'),
      ).toBeVisible()
      await expect(stanPage.getByRole('heading', { name: 'Assessment Results' })).toHaveCount(0)
    } finally {
      await stanContext.close()
    }
  })
})
