import { authFile, expect, test } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import {
  ASSESSMENT_FIXTURE_PHASES,
  FULL_COURSE_STUDENT,
  FULL_COURSE_STUDENT2,
  SEEDED_COURSES,
  TUTOR_SCOPE_TUTOR,
} from '../../src/data/constants'
import type { Role } from '../../src/data/roles'
import {
  apiAsRole,
  assessmentUrl,
  type CategoryWithCompetencies,
  createCategory,
  createCompetency,
  createSchema,
  putConfig,
  releaseResults,
  resetAssessmentPhase,
} from './helpers'

const PHASE_ID = ASSESSMENT_FIXTURE_PHASES.tutorResults
const COURSE_ID = SEEDED_COURSES.fullCourse.id
const SCHEMA_NAME = 'E2E Tutor Evaluation Rubric'
// Rated by both students, so its average is shown.
const SHARED_COMPETENCY = 'Explains concepts clearly'
// Rated by Stan only, so it stays hidden.
const LONE_COMPETENCY = 'Reviews code thoroughly'
// Both students praise the tutor, so both texts are shown.
const PRAISE = 'Explains things patiently'
const SECOND_PRAISE = 'Always prepared for the meetings'
// Only Stan criticizes, so the text would point to him and stays hidden.
const CRITIQUE = 'Could answer questions faster'

const STUDENTS: { role: Role; courseParticipationId: string; name: string }[] = [
  { role: 'student', courseParticipationId: FULL_COURSE_STUDENT.courseParticipationId, name: 'Stan' },
  {
    role: 'student2',
    courseParticipationId: FULL_COURSE_STUDENT2.courseParticipationId,
    name: 'Selma',
  },
]

interface TutorEvaluationResults {
  tutorID: string
  evaluatorCount: number
  results: { competencyID: string; averageScoreNumeric: number }[]
  feedbackItems: { feedbackType: string; feedbackText: string }[]
}

// Evaluations and feedback items are student-owned, so each student removes their own
// before the admin reset can delete the categories they reference.
async function cleanUpStudentRows() {
  for (const { role } of STUDENTS) {
    const student = await apiAsRole(role)
    try {
      const evaluations = await student.get(assessmentUrl(PHASE_ID, 'evaluation/my-evaluations'))
      if (evaluations.ok()) {
        for (const { id } of (await evaluations.json()) as { id: string }[]) {
          await student.delete(assessmentUrl(PHASE_ID, `evaluation/${id}`))
        }
      }
      const feedback = await student.get(
        assessmentUrl(PHASE_ID, 'evaluation/feedback-items/my-feedback'),
      )
      if (feedback.ok()) {
        for (const { id } of (await feedback.json()) as { id: string }[]) {
          await student.delete(assessmentUrl(PHASE_ID, `evaluation/feedback-items/${id}`))
        }
      }
    } finally {
      await student.dispose()
    }
  }
}

async function evaluateTutor(
  role: Role,
  authorId: string,
  scores: { competencyID: string; scoreLevel: string }[],
  feedback: { feedbackType: 'positive' | 'negative'; feedbackText: string }[],
) {
  const student = await apiAsRole(role)
  try {
    for (const score of scores) {
      const res = await student.post(assessmentUrl(PHASE_ID, 'evaluation'), {
        data: {
          ...score,
          courseParticipationID: TUTOR_SCOPE_TUTOR.id,
          authorCourseParticipationID: authorId,
          type: 'tutor',
        },
      })
      expect(res.status(), await res.text()).toBe(201)
    }
    for (const item of feedback) {
      const res = await student.post(assessmentUrl(PHASE_ID, 'evaluation/feedback-items'), {
        data: {
          ...item,
          courseParticipationID: TUTOR_SCOPE_TUTOR.id,
          coursePhaseID: PHASE_ID,
          authorCourseParticipationID: authorId,
          type: 'tutor',
        },
      })
      expect(res.ok(), await res.text()).toBe(true)
    }
  } finally {
    await student.dispose()
  }
}

async function getMyTutorResults(role: Role) {
  const api = await apiAsRole(role)
  try {
    const res = await api.get(assessmentUrl(PHASE_ID, 'evaluation/my-tutor-results'))
    return { status: res.status(), body: await res.text() }
  } finally {
    await api.dispose()
  }
}

async function unrelease() {
  const admin = await apiAsRole('admin')
  try {
    await admin.post(assessmentUrl(PHASE_ID, 'config/unrelease'))
  } finally {
    await admin.dispose()
  }
}

// Moves the tutor evaluation deadline into the past: tutors only see their results once it passed.
async function closeTutorEvaluation() {
  const lecturer = await apiAsRole('lecturer')
  try {
    await putConfig(lecturer, PHASE_ID, {
      tutorEvaluationEnabled: true,
      tutorEvaluationDeadline: new Date(Date.now() - 60 * 1000).toISOString(),
    })
  } finally {
    await lecturer.dispose()
  }
}

const reset = async () => {
  await unrelease()
  await cleanUpStudentRows()
  await resetAssessmentPhase(PHASE_ID, { schemaNames: [SCHEMA_NAME] })
}

test.use({ role: 'course-editor' })
test.describe.configure({ mode: 'serial' })

test.describe('assessment: tutors see their tutor evaluation results', () => {
  let sharedCompetencyId = ''
  let loneCompetencyId = ''

  test.beforeAll(async () => {
    await reset()
    const lecturer = await apiAsRole('lecturer')
    try {
      const schema = await createSchema(lecturer, PHASE_ID, SCHEMA_NAME)
      await putConfig(lecturer, PHASE_ID, {
        tutorEvaluationEnabled: true,
        tutorEvaluationSchema: schema.id,
      })
      await createCategory(lecturer, PHASE_ID, schema.id, 'Teaching')
      const categoriesUrl = assessmentUrl(PHASE_ID, 'category/tutor/with-competencies')
      const [category] = (await (
        await lecturer.get(categoriesUrl)
      ).json()) as CategoryWithCompetencies[]
      await createCompetency(lecturer, PHASE_ID, category.id, SHARED_COMPETENCY)
      await createCompetency(lecturer, PHASE_ID, category.id, LONE_COMPETENCY)
      const [withCompetencies] = (await (
        await lecturer.get(categoriesUrl)
      ).json()) as CategoryWithCompetencies[]
      const idOf = (name: string) =>
        withCompetencies.competencies.find((competency) => competency.name === name)?.id ?? ''
      sharedCompetencyId = idOf(SHARED_COMPETENCY)
      loneCompetencyId = idOf(LONE_COMPETENCY)
    } finally {
      await lecturer.dispose()
    }

    await evaluateTutor(
      'student',
      FULL_COURSE_STUDENT.courseParticipationId,
      [
        { competencyID: sharedCompetencyId, scoreLevel: 'veryGood' },
        { competencyID: loneCompetencyId, scoreLevel: 'good' },
      ],
      [
        { feedbackType: 'positive', feedbackText: PRAISE },
        { feedbackType: 'negative', feedbackText: CRITIQUE },
      ],
    )
    await evaluateTutor(
      'student2',
      FULL_COURSE_STUDENT2.courseParticipationId,
      [{ competencyID: sharedCompetencyId, scoreLevel: 'good' }],
      [{ feedbackType: 'positive', feedbackText: SECOND_PRAISE }],
    )
  })

  test.afterAll(async () => {
    await reset()
  })

  test('keeps the results from the tutor until they are released', async ({ page }) => {
    expect((await getMyTutorResults('course-editor')).status).toBe(204)

    await new AssessmentPage(page).goto(COURSE_ID, PHASE_ID)
    await expect(page.getByRole('heading', { name: 'Assessment Results & Evaluation' })).toBeVisible()
    await expect(page.getByText('Your Tutor Evaluation Results')).toHaveCount(0)
  })

  test('keeps released results from the tutor while the tutor evaluation is open', async ({
    page,
  }) => {
    const admin = await apiAsRole('admin')
    try {
      await releaseResults(admin, PHASE_ID)
    } finally {
      await admin.dispose()
    }

    expect((await getMyTutorResults('course-editor')).status).toBe(204)

    const phase = new AssessmentPage(page)
    await phase.goto(COURSE_ID, PHASE_ID)
    await expect(page.getByRole('heading', { name: 'Assessment Results & Evaluation' })).toBeVisible()
    await expect(page.getByText('Your Tutor Evaluation Results')).toHaveCount(0)

    await phase.goto(COURSE_ID, PHASE_ID, '/tutor-results')
    await expect(
      page.getByText(
        'Tutor evaluation results are shown once the tutor evaluation deadline has passed.',
      ),
    ).toBeVisible()
  })

  test('serves the tutor anonymous averages and feedback once the evaluation closed', async () => {
    await closeTutorEvaluation()

    const { status, body } = await getMyTutorResults('course-editor')
    expect(status).toBe(200)
    const results = JSON.parse(body) as TutorEvaluationResults
    expect(results.tutorID).toBe(TUTOR_SCOPE_TUTOR.id)
    expect(results.evaluatorCount).toBe(2)
    expect(results.results).toEqual([
      { competencyID: sharedCompetencyId, averageScoreNumeric: 1.5 },
    ])
    expect(results.feedbackItems).toEqual([
      { feedbackType: 'positive', feedbackText: SECOND_PRAISE },
      { feedbackType: 'positive', feedbackText: PRAISE },
    ])
    for (const { courseParticipationId } of STUDENTS) {
      expect(body).not.toContain(courseParticipationId)
    }

    // The lecturer tutors no team of this phase.
    expect((await getMyTutorResults('lecturer')).status).toBe(204)
  })

  test('shows the tutor their results from the phase page', async ({ page }) => {
    await new AssessmentPage(page).goto(COURSE_ID, PHASE_ID)
    await expect(page.getByText('Your Tutor Evaluation Results')).toBeVisible()
    await expect(page.getByText('2 students evaluated you.')).toBeVisible()

    await page.getByRole('button', { name: 'View your results' }).click()
    await expect(page).toHaveURL(new RegExp(`/${PHASE_ID}/tutor-results$`))
    await expect(page.getByRole('heading', { name: 'Your Tutor Evaluation Results' })).toBeVisible()
    await expect(page.getByText(SHARED_COMPETENCY)).toBeVisible()
    await expect(page.getByText(LONE_COMPETENCY)).toHaveCount(0)
    await expect(page.getByText(PRAISE)).toBeVisible()
    await expect(page.getByText(SECOND_PRAISE)).toBeVisible()
    await expect(page.getByText(CRITIQUE)).toHaveCount(0)
    for (const { name } of STUDENTS) {
      await expect(page.getByText(new RegExp(`\\b${name}\\b`))).toHaveCount(0)
    }
  })

  test('shows nothing to staff who tutor no team', async ({ browser }) => {
    const context = await browser.newContext({
      storageState: authFile('lecturer'),
    })
    try {
      const page = await context.newPage()
      await new AssessmentPage(page).goto(COURSE_ID, PHASE_ID)
      await expect(
        page.getByRole('heading', { name: 'Assessment Results & Evaluation' }),
      ).toBeVisible()
      await expect(page.getByText('Your Tutor Evaluation Results')).toHaveCount(0)
    } finally {
      await context.close()
    }
  })
})
