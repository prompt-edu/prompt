import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { TeamAllocationPage } from '../../src/pages/TeamAllocationPage'
import {
  FULL_COURSE_STUDENT,
  SEEDED_COURSES,
  SEEDED_PHASE_STUDENTS,
  TEAM_ALLOCATION_JOURNEY_PHASE_ID,
} from '../../src/data/constants'
import {
  clearAllocations,
  deleteTeamByName,
  getAllocations,
  getTeams,
  publishAllocation,
} from './helpers'

const PHASE_ID = TEAM_ALLOCATION_JOURNEY_PHASE_ID
const TEAM_NAME = 'E2E Team Allocation Lecturer'
const stan = SEEDED_PHASE_STUDENTS.student

const reset = async () => {
  await clearAllocations(PHASE_ID)
  await deleteTeamByName(TEAM_NAME, PHASE_ID)
}

test.use({ role: 'course-lecturer' })

test.describe('team allocation: lecturer journey', () => {
  test.beforeAll(reset)
  test.afterAll(reset)

  test('lecturer sets up a team, the allocation is published, and the result shows', async ({
    page,
  }) => {
    const phase = new TeamAllocationPage(page)

    // Set up teams through the settings UI.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/settings')
    await phase.expectSettingsLoaded()
    await phase.createTeam(TEAM_NAME)

    // Run the allocation: TEASE is external, so publish the computed assignment
    // through the save endpoint (assign Stan to the team the lecturer created).
    const lecturer = await apiContextFor('course-lecturer')
    let teamId: string
    try {
      const team = (await getTeams(lecturer, PHASE_ID)).find((t) => t.name === TEAM_NAME)
      expect(team).toBeTruthy()
      teamId = team!.id
    } finally {
      await lecturer.dispose()
    }
    await publishAllocation(PHASE_ID, [
      { projectId: teamId, students: [FULL_COURSE_STUDENT.courseParticipationId] },
    ])

    // Verify the result: the allocations view lists the team with Stan as member.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/allocations')
    await phase.expectAllocationsLoaded()
    await phase.expectAllocatedMember(TEAM_NAME, `${stan.firstName} ${stan.lastName}`)

    // ...and the participants table shows Stan's allocated team.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/participants')
    await phase.expectParticipantAllocatedTeam(stan.lastName, TEAM_NAME)

    // ...and the allocation is readable through the API.
    const api = await apiContextFor('course-lecturer')
    try {
      const allocations = await getAllocations(api, PHASE_ID)
      const stanAllocation = allocations.find(
        (a) => a.courseParticipationID === FULL_COURSE_STUDENT.courseParticipationId,
      )
      expect(stanAllocation?.teamAllocation).toBe(teamId)
    } finally {
      await api.dispose()
    }
  })
})
