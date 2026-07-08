import { expect, test } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import {
  FULL_COURSE_PHASES,
  FULL_COURSE_STUDENT,
  SEEDED_COURSES,
  SEEDED_PHASE_STUDENTS,
} from '../../src/data/constants'
import { resetAssessmentPhase } from './helpers'

const PHASE_ID = FULL_COURSE_PHASES.assessment.id
const SCHEMA_NAME = 'E2E Rubric'
const CATEGORY_NAME = 'E2E Category'
const COMPETENCY_ONE = 'Clean Code'
const COMPETENCY_TWO = 'Teamwork'
const STUDENT_NAME = `${SEEDED_PHASE_STUDENTS.student.firstName} ${SEEDED_PHASE_STUDENTS.student.lastName}`

// The `lecturer` user holds the course-scoped ios2425-iPraktikumFull-Lecturer role.
test.use({ role: 'lecturer' })

test.describe('assessment: lecturer journey', () => {
  test.beforeAll(async () => {
    // Also sets an open assessment timeframe (the lazily created default
    // config has deadline == its creation time, which would block unmarking).
    await resetAssessmentPhase(PHASE_ID, {
      courseParticipationIds: [FULL_COURSE_STUDENT.courseParticipationId],
      schemaNames: [SCHEMA_NAME],
    })
  })

  test.afterAll(async () => {
    await resetAssessmentPhase(PHASE_ID, {
      courseParticipationIds: [FULL_COURSE_STUDENT.courseParticipationId],
      schemaNames: [SCHEMA_NAME],
    })
  })

  test('a lecturer defines a rubric and grades a student', async ({ page }) => {
    const phase = new AssessmentPage(page)

    // Create a fresh schema and bind the phase to it.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/settings')
    await phase.expectSettingsLoaded()
    await phase.createSchema(SCHEMA_NAME, 'Rubric created by the e2e lecturer journey')
    await phase.selectSchema(SCHEMA_NAME)
    await phase.saveAssessmentSettings()

    // Define the rubric: one category with two competencies.
    await phase.openSchemaDetails()
    await phase.addCategory(CATEGORY_NAME, 'E2ECat')
    await phase.addCompetency(CATEGORY_NAME, COMPETENCY_ONE, 'Code')
    await phase.addCompetency(CATEGORY_NAME, COMPETENCY_TWO, 'Team')

    // Grade the seeded student across both competencies.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/participants')
    await phase.expectParticipantsLoaded()
    await phase.openParticipant(STUDENT_NAME)
    await phase.scoreCompetency(COMPETENCY_ONE, 'Agree')
    await phase.scoreCompetency(COMPETENCY_TWO, 'Strongly Agree')
    await phase.fillGeneralRemarks('Great work throughout the phase.')
    await phase.selectGradeSuggestion('Good - 1.7')
    await phase.markAssessmentAsFinal()

    // The participants table reflects the finalized grade suggestion.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/participants')
    await phase.expectParticipantsLoaded()
    await expect(page.getByRole('row', { name: new RegExp(STUDENT_NAME) })).toContainText('1.7')

    // Release stays gated in the UI until every participant is finalized
    // (1 of 4 here); the release flow itself is covered by the visibility spec.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/settings')
    await phase.expectSettingsLoaded()
    await expect(phase.releaseButton()).toContainText('1/4 final')
    await expect(phase.releaseButton()).toBeDisabled()
  })
})
