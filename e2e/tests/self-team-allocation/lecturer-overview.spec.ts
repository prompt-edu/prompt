import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import {
  SEEDED_COURSES,
  SEEDED_PHASE_STUDENTS,
  SELF_TEAM_ALLOCATION_OVERVIEW_PHASE_ID,
} from '../../src/data/constants'
import { deleteTeamByName, getTeams, teamsUrl } from './helpers'

const TEAM_NAME = 'E2E Lecturer Overview Team'
// This spec owns a dedicated phase: the team it forms (containing `student`)
// would otherwise disable team creation for the student-journey spec running
// on the shared phase in a parallel worker.
const PHASE_ID = SELF_TEAM_ALLOCATION_OVERVIEW_PHASE_ID

test.use({ role: 'course-lecturer' })

test.describe('self team allocation: lecturer overview', () => {
  // Form a team through the API as the student: create it, then join it
  // (creation does not assign the creator).
  test.beforeAll(async () => {
    await deleteTeamByName(TEAM_NAME, PHASE_ID)
    const student = await apiContextFor('student')
    try {
      const created = await student.post(teamsUrl(PHASE_ID), { data: { teamNames: [TEAM_NAME] } })
      expect(created.status()).toBe(201)
      const team = (await getTeams(student, PHASE_ID)).find((t) => t.name === TEAM_NAME)
      if (!team) throw new Error('created team not found')
      const joined = await student.put(`${teamsUrl(PHASE_ID)}/${team.id}/assignment`)
      expect(joined.ok()).toBeTruthy()
    } finally {
      await student.dispose()
    }
  })

  test.afterAll(async () => {
    await deleteTeamByName(TEAM_NAME, PHASE_ID)
  })

  test('the participants page lists students with their allocated team', async ({ page }) => {
    await page.goto(
      `/management/course/${SEEDED_COURSES.iPraktikum.id}/${PHASE_ID}/participants`,
    )
    await expect(
      page.getByRole('heading', { level: 1, name: 'Self Team Allocation Participants' }),
    ).toBeVisible({ timeout: 15_000 })

    const table = page.locator('#table-view')
    await expect(table.getByText(SEEDED_PHASE_STUDENTS.student.lastName).first()).toBeVisible()
    await expect(table.getByText(SEEDED_PHASE_STUDENTS.student2.lastName).first()).toBeVisible()
    await expect(table.getByText(TEAM_NAME).first()).toBeVisible()
  })

  test('the teams overview shows the formed team', async ({ page }) => {
    await page.goto(`/management/course/${SEEDED_COURSES.iPraktikum.id}/${PHASE_ID}`)
    await expect(page.getByRole('heading', { level: 1, name: 'Team Allocation' })).toBeVisible({
      timeout: 15_000,
    })
    const card = page.getByTestId(`team-card-${TEAM_NAME}`)
    await expect(card).toBeVisible()
    await expect(card.getByText('1/3 Members')).toBeVisible()
    await expect(
      card.getByText(
        `${SEEDED_PHASE_STUDENTS.student.firstName} ${SEEDED_PHASE_STUDENTS.student.lastName}`,
      ),
    ).toBeVisible()
  })
})
