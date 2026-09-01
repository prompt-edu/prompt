import { expect, test } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import {
  ASSESSMENT_FIXTURE_PHASES,
  FULL_COURSE_STUDENT,
  SEEDED_COURSES,
} from '../../src/data/constants'
import {
  apiAsRole,
  assessmentUrl,
  createCategory,
  createSchema,
  createCompetency,
  putConfig,
  resetAssessmentPhase,
} from './helpers'

const PHASE_ID = ASSESSMENT_FIXTURE_PHASES.selfEvaluation
const SCHEMA_NAME = 'E2E Self Evaluation Rubric'
const COMPETENCY_NAME = 'I communicated openly with my team'

interface EvaluationCompletion {
  authorCourseParticipationID: string
  completed: boolean
  type: string
}

// Evaluations are student-owned: unmark the completion (rows are durable, no
// delete endpoint) and delete the evaluations as the student, so the admin
// reset can then delete categories/schemas (blocked while evaluations exist).
async function cleanUpOwnSelfEvaluation() {
  const student = await apiAsRole('student')
  try {
    await student.put(assessmentUrl(PHASE_ID, 'evaluation/completed/my-completion/unmark'), {
      data: {
        courseParticipationID: FULL_COURSE_STUDENT.courseParticipationId,
        coursePhaseID: PHASE_ID,
        authorCourseParticipationID: FULL_COURSE_STUDENT.courseParticipationId,
      },
    })
    const res = await student.get(assessmentUrl(PHASE_ID, 'evaluation/my-evaluations'))
    if (res.ok()) {
      const evaluations = (await res.json()) as { id: string }[]
      for (const evaluation of evaluations) {
        await student.delete(assessmentUrl(PHASE_ID, `evaluation/${evaluation.id}`))
      }
    }
  } finally {
    await student.dispose()
  }
}

const reset = () => resetAssessmentPhase(PHASE_ID, { schemaNames: [SCHEMA_NAME] })

test.use({ role: 'student' })

test.describe('assessment: student self evaluation', () => {
  test.beforeAll(async () => {
    await cleanUpOwnSelfEvaluation()
    await reset()
    const lecturer = await apiAsRole('lecturer')
    try {
      const schema = await createSchema(lecturer, PHASE_ID, SCHEMA_NAME)
      await putConfig(lecturer, PHASE_ID, {
        selfEvaluationEnabled: true,
        selfEvaluationSchema: schema.id,
      })
      await createCategory(lecturer, PHASE_ID, schema.id, 'Collaboration')
      const res = await lecturer.get(assessmentUrl(PHASE_ID, 'category/self/with-competencies'))
      const categories = (await res.json()) as { id: string }[]
      await createCompetency(lecturer, PHASE_ID, categories[0].id, COMPETENCY_NAME)
    } finally {
      await lecturer.dispose()
    }
  })

  test.afterAll(async () => {
    // Evaluation completions have no delete endpoint; unmark as the student
    // (the rows are durable, so assertions check the row's state, not counts).
    await cleanUpOwnSelfEvaluation()
    await reset()
  })

  // Runs before the finalizing test so the evaluation is still editable. Shares this
  // spec's phase deliberately: files run in parallel, so a separate file resetting the
  // same phase would clobber this one's schema.
  test('the evaluation header docks on scroll without shifting the content below', async ({
    page,
  }) => {
    // Short enough that the evaluation form overflows, tall enough to keep the
    // desktop layout (and so the header still starts undocked).
    await page.setViewportSize({ width: 1280, height: 520 })

    const phase = new AssessmentPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID)
    await phase.expectOverviewLoaded()
    await phase.openSelfEvaluation()

    const header = page.getByTestId('sticky-header')
    const placeholder = page.getByTestId('sticky-header-placeholder')
    const position = () => header.evaluate((el) => getComputedStyle(el).position)
    const placeholderHeight = () =>
      placeholder.evaluate((el) => el.getBoundingClientRect().height)

    // Scrolls whichever ancestor actually overflows, and reports whether one did.
    const scrollTo = (offset: 'bottom' | 'top') =>
      header.evaluate((el, target) => {
        let node: HTMLElement | null = el.parentElement
        while (node && node.scrollHeight <= node.clientHeight) node = node.parentElement
        if (!node) return false
        node.scrollTop = target === 'bottom' ? node.scrollHeight : 0
        return true
      }, offset)

    await expect(header).toBeVisible()
    expect(await position()).not.toBe('fixed')
    const undockedHeight = await placeholderHeight()

    // If nothing overflows, the rest of this test would prove nothing.
    expect(await scrollTo('bottom')).toBe(true)
    await expect.poll(position).toBe('fixed')

    // The placeholder keeps its undocked height, so nothing below the header moves.
    expect(await placeholderHeight()).toBeCloseTo(undockedHeight, 0)

    await scrollTo('top')
    await expect.poll(position).not.toBe('fixed')
    // Undocking animates the bar back to full size, so poll until the placeholder settles.
    await expect.poll(placeholderHeight).toBeCloseTo(undockedHeight, 0)
  })

  test('a student fills in and finalizes the self evaluation', async ({ page }) => {
    const phase = new AssessmentPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID)
    await phase.expectOverviewLoaded()

    await phase.openSelfEvaluation()
    await phase.scoreCompetency(COMPETENCY_NAME, 'Agree')
    await phase.markEvaluationAsFinal()

    // The overview reflects the completed self evaluation.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID)
    await phase.expectOverviewLoaded()
    await expect(page.getByText('All Evaluations Completed!')).toBeVisible()

    // The lecturer sees the completion (specific row, not a count).
    const lecturer = await apiAsRole('lecturer')
    try {
      const res = await lecturer.get(assessmentUrl(PHASE_ID, 'evaluation/completed'))
      expect(res.ok()).toBeTruthy()
      const completions = (await res.json()) as EvaluationCompletion[]
      const own = completions.find(
        (completion) =>
          completion.authorCourseParticipationID === FULL_COURSE_STUDENT.courseParticipationId &&
          completion.type === 'self',
      )
      expect(own?.completed).toBe(true)
    } finally {
      await lecturer.dispose()
    }
  })

  // The `lecturer` user holds the course lecturer role but no participation, so
  // it has no student role: every student-only endpoint behind this page would
  // answer 401 and must therefore not be requested at all.
  test.describe('previewed by a lecturer', () => {
    test.use({ role: 'lecturer' })

    test('the page stays readable instead of erroring on student-only data', async ({ page }) => {
      const phase = new AssessmentPage(page)
      await phase.gotoSelfEvaluation(SEEDED_COURSES.fullCourse.id, PHASE_ID)

      // The notice repeats the phrase in its title and its description, so the
      // assertion goes for the alert that carries both.
      await expect(
        page.getByRole('alert').filter({ hasText: 'not a student of this course' }),
      ).toBeVisible()

      // Both feedback panels show their empty state, not the error page.
      await expect(
        page.getByText('What did you do particularly well?', { exact: true }),
      ).toBeVisible()
      await expect(page.getByText('Where can you still improve?', { exact: true })).toBeVisible()
      await expect(page.getByText('Error loading feedback items')).toHaveCount(0)

      // The student-only writes behind the page are out of reach.
      const addItem = page.getByRole('button', { name: 'Add Item' })
      await expect(addItem).toHaveCount(2)
      await expect(addItem.first()).toBeDisabled()
      await expect(addItem.last()).toBeDisabled()
      await expect(page.getByRole('button', { name: 'Unmark as Final' })).toBeDisabled()
    })
  })
})
