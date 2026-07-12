import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { TeamAllocationPage } from '../../src/pages/TeamAllocationPage'
import {
  FULL_COURSE_STUDENT,
  SEEDED_COURSES,
  TEAM_ALLOCATION_STUDENT_PHASE_ID,
} from '../../src/data/constants'
import { clearAllocations, createTeam, deleteTeamByName, getAllocationForParticipation, publishAllocation } from './helpers'

const PHASE_ID = TEAM_ALLOCATION_STUDENT_PHASE_ID
const TEAM_NAME = 'E2E Team Allocation Student'

// The team allocation module has no student-facing "my team" page — a student's
// only UI is the survey — so the student sees their allocated team through the
// allocation API (student role is permitted on GET /allocation).
let allocatedTeamId: string

const reset = async () => {
  await clearAllocations(PHASE_ID)
  await deleteTeamByName(TEAM_NAME, PHASE_ID)
}

test.use({ role: 'student' })

test.describe('team allocation: student journey', () => {
  test.beforeAll(async () => {
    await reset()
    const admin = await apiContextFor('admin')
    try {
      const team = await createTeam(admin, PHASE_ID, TEAM_NAME)
      allocatedTeamId = team.id
    } finally {
      await admin.dispose()
    }
    await publishAllocation(PHASE_ID, [
      { projectId: allocatedTeamId, students: [FULL_COURSE_STUDENT.courseParticipationId] },
    ])
  })

  test.afterAll(reset)

  test('a student opens the survey and can read their allocated team', async ({ page }) => {
    const phase = new TeamAllocationPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID)
    await phase.expectSurveyLoaded()

    const api = await apiContextFor('student')
    try {
      const allocation = await getAllocationForParticipation(
        api,
        PHASE_ID,
        FULL_COURSE_STUDENT.courseParticipationId,
      )
      expect(allocation?.teamAllocation).toBe(allocatedTeamId)
    } finally {
      await api.dispose()
    }
  })
})
