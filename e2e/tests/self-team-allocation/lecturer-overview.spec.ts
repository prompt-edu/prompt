import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import {
  SEEDED_COURSES,
  SEEDED_PHASE_STUDENTS,
  SELF_TEAM_ALLOCATION_PHASE_ID,
} from '../../src/data/constants'
import { deleteTeamByName, getTeams, teamsUrl } from './helpers'

const TEAM_NAME = 'E2E Lecturer Overview Team'

test.use({ role: 'course-lecturer' })

test.describe('self team allocation: lecturer overview', () => {
  // Form a team through the API as the student: create it, then join it
  // (creation does not assign the creator).
  test.beforeAll(async () => {
    await deleteTeamByName(TEAM_NAME)
    const student = await apiContextFor('student')
    try {
      const created = await student.post(teamsUrl(), { data: { teamNames: [TEAM_NAME] } })
      expect(created.status()).toBe(201)
      const team = (await getTeams(student)).find((t) => t.name === TEAM_NAME)
      if (!team) throw new Error('created team not found')
      const joined = await student.put(`${teamsUrl()}/${team.id}/assignment`)
      expect(joined.ok()).toBeTruthy()
    } finally {
      await student.dispose()
    }
  })

  test.afterAll(async () => {
    await deleteTeamByName(TEAM_NAME)
  })

  test('the participants page lists students with their allocated team', async ({ page }) => {
    await page.goto(
      `/management/course/${SEEDED_COURSES.iPraktikum.id}/${SELF_TEAM_ALLOCATION_PHASE_ID}/participants`,
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
    await page.goto(
      `/management/course/${SEEDED_COURSES.iPraktikum.id}/${SELF_TEAM_ALLOCATION_PHASE_ID}`,
    )
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
