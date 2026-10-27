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
      placeholder.evaluate((el) => Math.round(el.getBoundingClientRect().height))
    const horizontalBounds = async () => {
      const [bar, column] = await Promise.all([header.boundingBox(), placeholder.boundingBox()])
      return {
        bar: [Math.round(bar?.x ?? -1), Math.round(bar?.width ?? -1)],
        column: [Math.round(column?.x ?? -1), Math.round(column?.width ?? -1)],
      }
    }

    // Scrolls whichever ancestor actually overflows, then records the header state on
    // every frame for half a second, so a shift that settles or a dock/undock loop shows.
    // `scrolled` is false when nothing overflows.
    const scrollTo = (offset: 'bottom' | 'top') =>
      header.evaluate(async (el, target) => {
        let node: HTMLElement | null = el.parentElement
        while (node && node.scrollHeight <= node.clientHeight) node = node.parentElement
        if (!node) return { scrolled: false, positions: [], placeholderHeights: [] }
        node.scrollTop = target === 'bottom' ? node.scrollHeight : 0
        const placeholderElement = el.parentElement as HTMLElement
        const positions: string[] = []
        const placeholderHeights: number[] = []
        for (let frame = 0; frame < 30; frame++) {
          await new Promise(requestAnimationFrame)
          positions.push(getComputedStyle(el).position)
          placeholderHeights.push(Math.round(placeholderElement.getBoundingClientRect().height))
        }
        return { scrolled: true, positions, placeholderHeights }
      }, offset)

    await expect(header).toBeVisible()
    expect(await position()).not.toBe('fixed')
    const undockedHeight = await placeholderHeight()

    const docking = await scrollTo('bottom')
    // If nothing overflows, the rest of this test would prove nothing.
    expect(docking.scrolled).toBe(true)
    // Docks and stays docked: collapsing the header must not scroll it back out.
    expect(docking.positions.slice(docking.positions.indexOf('fixed'))).not.toContain('static')
    expect(docking.positions.at(-1)).toBe('fixed')
    // The placeholder keeps its undocked height, so nothing below the header moves.
    expect(new Set(docking.placeholderHeights)).toEqual(new Set([undockedHeight]))

    // The docked bar follows the content column when the sidebar collapses.
    const dockedBounds = await horizontalBounds()
    expect(dockedBounds.bar).toEqual(dockedBounds.column)
    await page.getByRole('button', { name: 'Toggle Sidebar' }).first().click()
    await expect
      .poll(async () => (await horizontalBounds()).column)
      .not.toEqual(dockedBounds.column)
    await expect
      .poll(async () => {
        const { bar, column } = await horizontalBounds()
        return bar[0] === column[0] && bar[1] === column[1]
      })
      .toBe(true)

    // Undocking restores the full header at once instead of growing it back in flow.
    const undocking = await scrollTo('top')
    expect(undocking.positions.at(-1)).not.toBe('fixed')
    expect(new Set(undocking.placeholderHeights)).toEqual(new Set([undockedHeight]))
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
  // answer 403 and must therefore not be requested at all.
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

  // Runs after the finalizing test, whose score is the result this page reports.
  test.describe('results reviewed by a lecturer', () => {
    test.use({ role: 'lecturer' })

    test('the score level indicator scrolls beneath the management header', async ({ page }) => {
      // Short enough that the one-competency page scrolls the indicator up under the
      // management header, wide enough to keep the desktop layout.
      await page.setViewportSize({ width: 1280, height: 320 })

      const phase = new AssessmentPage(page)
      await phase.goto(
        SEEDED_COURSES.fullCourse.id,
        PHASE_ID,
        `/self-evaluations/${FULL_COURSE_STUDENT.courseParticipationId}`,
      )
      const indicator = page.getByRole('button', { name: /evaluation result: Agree$/ })
      await expect(indicator).toBeVisible({ timeout: 15_000 })

      // Scrolls the indicator into the 56px management header, then reports which
      // element the browser paints at its center.
      const hit = await indicator.evaluate(async (el) => {
        // The indicator overflows its own card, so overflowing alone does not make an
        // ancestor the scroll container.
        const scrolls = (candidate: HTMLElement) =>
          candidate.scrollHeight > candidate.clientHeight &&
          (candidate === document.scrollingElement ||
            /auto|scroll/.test(getComputedStyle(candidate).overflowY))
        let node: HTMLElement | null = el.parentElement
        while (node && !scrolls(node)) node = node.parentElement
        if (!node) return undefined
        const before = el.getBoundingClientRect()
        node.scrollTop += before.top + before.height / 2 - 28
        await new Promise(requestAnimationFrame)
        const after = el.getBoundingClientRect()
        const y = after.top + after.height / 2
        const painted = document.elementFromPoint(after.left + after.width / 2, y)
        return {
          y,
          indicatorOnTop: el.contains(painted),
          headerOnTop: !!painted?.closest('header'),
        }
      })

      expect(hit?.y).toBeGreaterThan(0)
      expect(hit?.y).toBeLessThan(56)
      expect(hit?.indicatorOnTop).toBe(false)
      expect(hit?.headerOnTop).toBe(true)
    })
  })
})
