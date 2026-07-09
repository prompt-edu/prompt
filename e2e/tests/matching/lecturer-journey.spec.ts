import { expect, test } from '../../src/fixtures/auth'
import { MatchingPage } from '../../src/pages/MatchingPage'
import {
  FULL_COURSE_STUDENT,
  FULL_COURSE_STUDENT2,
  MATCHING_JOURNEY_PHASE_ID,
  SEEDED_COURSES,
  SEEDED_PHASE_STUDENTS,
} from '../../src/data/constants'
import { apiAsRole, buildReimportCsv, getParticipations, resetMatchingPhase } from './helpers'

const PHASE_ID = MATCHING_JOURNEY_PHASE_ID

const stan = {
  ...SEEDED_PHASE_STUDENTS.student,
  matriculationNumber: FULL_COURSE_STUDENT.matriculationNumber,
  courseParticipationId: FULL_COURSE_STUDENT.courseParticipationId,
}
const selma = {
  ...SEEDED_PHASE_STUDENTS.student2,
  matriculationNumber: FULL_COURSE_STUDENT2.matriculationNumber,
  courseParticipationId: FULL_COURSE_STUDENT2.courseParticipationId,
}

const reset = () =>
  resetMatchingPhase(PHASE_ID, [
    stan.courseParticipationId,
    selma.courseParticipationId,
  ])

test.use({ role: 'course-lecturer' })

test.describe('matching: lecturer re-import journey', () => {
  test.beforeAll(reset)
  test.afterAll(reset)

  test('re-importing the assigned students marks them as passed', async ({ page }) => {
    const phase = new MatchingPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID)
    await phase.expectOverviewLoaded()

    await phase.uploadReimportCsv(buildReimportCsv([stan, selma]))
    await phase.expectMatchedStudent(stan.matriculationNumber)
    await phase.expectMatchedStudent(selma.matriculationNumber)

    await phase.importStudents()
    await phase.expectImportSuccess()

    const api = await apiAsRole('course-lecturer')
    try {
      const participations = await getParticipations(api, PHASE_ID)
      const passed = participations
        .filter((p) => p.passStatus === 'passed')
        .map((p) => p.courseParticipationID)
      expect(passed).toContain(stan.courseParticipationId)
      expect(passed).toContain(selma.courseParticipationId)
    } finally {
      await api.dispose()
    }
  })
})
