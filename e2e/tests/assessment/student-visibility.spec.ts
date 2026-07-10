import { expect, test, authFile } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import {
  ASSESSMENT_FIXTURE_PHASES,
  FULL_COURSE_STUDENT2,
  SEEDED_COURSES,
} from '../../src/data/constants'
import {
  apiAsRole,
  assessmentUrl,
  createCategory,
  createCompetency,
  createSchema,
  getAssessmentCategories,
  gradeCompetency,
  markAssessmentComplete,
  putConfig,
  releaseResults,
  resetAssessmentPhase,
} from './helpers'

const PHASE_ID = ASSESSMENT_FIXTURE_PHASES.visibility
const SCHEMA_NAME = 'E2E Visibility Rubric'
const REMARKS = 'Solid contribution to the visibility fixture.'

const reset = () =>
  resetAssessmentPhase(PHASE_ID, {
    courseParticipationIds: [FULL_COURSE_STUDENT2.courseParticipationId],
    schemaNames: [SCHEMA_NAME],
  })

// Selma (student2) is the graded student; Stan (student) is the enrolled
// bystander whose results must stay empty.
test.use({ role: 'student2' })

test.describe('assessment: student result visibility', () => {
  test.beforeAll(async () => {
    await reset()
    const lecturer = await apiAsRole('lecturer')
    try {
      const schema = await createSchema(lecturer, PHASE_ID, SCHEMA_NAME)
      await putConfig(lecturer, PHASE_ID, { assessmentSchemaId: schema.id })
      await createCategory(lecturer, PHASE_ID, schema.id, 'Visibility Category')
      const category = (await getAssessmentCategories(lecturer, PHASE_ID))[0]
      await createCompetency(lecturer, PHASE_ID, category.id, 'Visibility Competency')
      const competency = (await getAssessmentCategories(lecturer, PHASE_ID))[0].competencies[0]
      await gradeCompetency(
        lecturer,
        PHASE_ID,
        FULL_COURSE_STUDENT2.courseParticipationId,
        competency.id,
        'good',
      )
      await markAssessmentComplete(
        lecturer,
        PHASE_ID,
        FULL_COURSE_STUDENT2.courseParticipationId,
        1.7,
        REMARKS,
      )
    } finally {
      await lecturer.dispose()
    }
  })

  test.afterAll(async () => {
    await reset()
  })

  test('released scores reach the assessed student and stay hidden from others', async ({
    page,
    browser,
  }) => {
    const phase = new AssessmentPage(page)

    // Before release: the student sees no results, the API returns 204.
    await phase.gotoResults(SEEDED_COURSES.fullCourse.id, PHASE_ID)
    await phase.expectResultsNotReleased()
    const selmaApi = await apiAsRole('student2')
    try {
      const before = await selmaApi.get(assessmentUrl(PHASE_ID, 'student-assessment/my-results'))
      expect(before.status()).toBe(204)

      const lecturer = await apiAsRole('lecturer')
      try {
        await releaseResults(lecturer, PHASE_ID)
      } finally {
        await lecturer.dispose()
      }

      // After release: grade suggestion and remarks are visible to Selma. The
      // remarks render as a (read-only) textarea VALUE, so assert on the value;
      // the grade select renders the option label "Good - 1.7" as text.
      await phase.gotoResults(SEEDED_COURSES.fullCourse.id, PHASE_ID)
      await expect(
        page.getByPlaceholder('What did this person do particularly well?'),
      ).toHaveValue(REMARKS)
      await expect(page.getByText('Good - 1.7')).toBeVisible()
      const after = await selmaApi.get(assessmentUrl(PHASE_ID, 'student-assessment/my-results'))
      expect(after.status()).toBe(200)
      const results = (await after.json()) as { assessmentCompletion: { gradeSuggestion: number } }
      expect(results.assessmentCompletion.gradeSuggestion).toBe(1.7)
    } finally {
      await selmaApi.dispose()
    }

    // Stan is enrolled in the same phase but has no completed assessment:
    // his results stay empty and Selma's remarks never leak into his view.
    const stanContext = await browser.newContext({ storageState: authFile('student') })
    try {
      const stanPage = await stanContext.newPage()
      const stanPhase = new AssessmentPage(stanPage)
      await stanPhase.gotoResults(SEEDED_COURSES.fullCourse.id, PHASE_ID)
      // Without a completed assessment the results section renders nothing:
      // no remarks textarea, no grade label.
      await expect(
        stanPage.getByPlaceholder('What did this person do particularly well?'),
      ).toHaveCount(0)
      await expect(stanPage.getByText('Good - 1.7')).toHaveCount(0)
    } finally {
      await stanContext.close()
    }
    const stanApi = await apiAsRole('student')
    try {
      const res = await stanApi.get(assessmentUrl(PHASE_ID, 'student-assessment/my-results'))
      expect(res.status()).toBe(204)
    } finally {
      await stanApi.dispose()
    }
  })
})
