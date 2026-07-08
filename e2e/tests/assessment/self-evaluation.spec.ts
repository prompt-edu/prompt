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
})
