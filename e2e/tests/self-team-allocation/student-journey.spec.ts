import { test } from '../../src/fixtures/auth'
import { authFile } from '../../src/fixtures/auth'
import { SelfTeamAllocationPage } from '../../src/pages/SelfTeamAllocationPage'
import {
  SEEDED_COURSES,
  SEEDED_PHASE_STUDENTS,
  SELF_TEAM_ALLOCATION_PHASE_ID,
} from '../../src/data/constants'
import { deleteTeamByName } from './helpers'

const TEAM_NAME = 'E2E Journey Team'

test.use({ role: 'student' })

test.describe('self team allocation: student journey', () => {
  test.beforeAll(async () => {
    await deleteTeamByName(TEAM_NAME)
  })

  test.afterAll(async () => {
    await deleteTeamByName(TEAM_NAME)
  })

  test('a student creates a team and a second student joins it', async ({ page, browser }) => {
    const phase = new SelfTeamAllocationPage(page)
    await phase.goto(SEEDED_COURSES.iPraktikum.id, SELF_TEAM_ALLOCATION_PHASE_ID)
    await phase.expectLoaded()

    // Creating a team does not assign the creator; joining is a second step.
    await phase.createTeam(TEAM_NAME)
    await phase.expectMemberCount(TEAM_NAME, 0)
    await phase.joinTeam(TEAM_NAME)
    await phase.expectMemberCount(TEAM_NAME, 1)

    // Second student in their own browser context (own session/storage).
    const context2 = await browser.newContext({ storageState: authFile('student2') })
    try {
      const page2 = await context2.newPage()
      const phase2 = new SelfTeamAllocationPage(page2)
      await phase2.goto(SEEDED_COURSES.iPraktikum.id, SELF_TEAM_ALLOCATION_PHASE_ID)
      await phase2.expectLoaded()

      await phase2.joinTeam(TEAM_NAME)
      await phase2.expectMemberCount(TEAM_NAME, 2)
      await phase2.expectMemberVisible(
        TEAM_NAME,
        `${SEEDED_PHASE_STUDENTS.student.firstName} ${SEEDED_PHASE_STUDENTS.student.lastName}`,
      )
    } finally {
      await context2.close()
    }

    // First student sees the second member after a refresh.
    await page.reload()
    await phase.expectLoaded()
    await phase.expectMemberCount(TEAM_NAME, 2)
    await phase.expectMemberVisible(
      TEAM_NAME,
      `${SEEDED_PHASE_STUDENTS.student2.firstName} ${SEEDED_PHASE_STUDENTS.student2.lastName}`,
    )
  })
})
