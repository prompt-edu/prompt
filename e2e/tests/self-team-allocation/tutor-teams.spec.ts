import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { BASE_URL, SELF_TEAM_ALLOCATION_API } from '../../src/env'
import { SelfTeamAllocationPage } from '../../src/pages/SelfTeamAllocationPage'
import { SEEDED_COURSES, SELF_TEAM_ALLOCATION_TUTOR_SCOPE } from '../../src/data/constants'

// course-editor is a project lead of two of the three teams of the tutor scope phase. They are
// imported from the course staff, which works because the e2e realm puts them in the Editor
// group of iPraktikumFull.
const { phaseId: PHASE_ID, ledTutorId, teams } = SELF_TEAM_ALLOCATION_TUTOR_SCOPE
const EDITOR_NAME = 'Eduard Eduard'
const EDITOR_LOGIN = 'co69edt'

interface TeamWithTutors {
  id: string
  name: string
  tutors: { id: string; firstName: string; lastName: string }[]
}

interface TutorAssignment {
  id: string
  universityLogin: string
  teamID: string
}

const phaseUrl = (path: string) =>
  `${BASE_URL}${SELF_TEAM_ALLOCATION_API}/course_phase/${PHASE_ID}${path}`

async function teamsAs(role: 'course-editor' | 'course-lecturer'): Promise<TeamWithTutors[]> {
  const api = await apiContextFor(role)
  try {
    const res = await api.get(phaseUrl('/team'))
    expect(res.status()).toBe(200)
    return ((await res.json()) as { teams: TeamWithTutors[] }).teams
  } finally {
    await api.dispose()
  }
}

// Puts the seeded assignment back: course-editor leads Team Lovelace and Team Turing.
async function restoreSeededLead() {
  const api = await apiContextFor('course-lecturer')
  try {
    const res = await api.post(phaseUrl('/team/tutors'), {
      data: [
        {
          courseParticipationID: ledTutorId,
          firstName: 'Eduard',
          lastName: 'Eduard',
          universityLogin: EDITOR_LOGIN,
          teamIDs: [teams.lovelace.id, teams.turing.id],
        },
      ],
    })
    expect(res.status()).toBe(201)
  } finally {
    await api.dispose()
  }
}

test.describe('self team allocation: a tutor leading several teams', () => {
  test.afterAll(restoreSeededLead)

  test('reads exactly the teams and allocations the tutor leads', async () => {
    const editorTeams = await teamsAs('course-editor')
    expect(editorTeams.map((t) => t.name).sort()).toEqual([teams.lovelace.name, teams.turing.name])

    const editor = await apiContextFor('course-editor')
    try {
      const allocations = (await (await editor.get(phaseUrl('/allocation'))).json()) as {
        teamAllocation: string
      }[]
      expect([...new Set(allocations.map((a) => a.teamAllocation))].sort()).toEqual(
        [teams.lovelace.id, teams.turing.id].sort(),
      )
      expect((await editor.get(phaseUrl(`/team/${teams.turing.id}`))).status()).toBe(200)
      expect((await editor.get(phaseUrl(`/team/${teams.hamilton.id}`))).status()).toBe(403)

      const me = await editor.get(phaseUrl('/team/tutors/me'))
      expect(me.status()).toBe(200)
      expect(((await me.json()) as { id: string }).id).toBe(ledTutorId)
    } finally {
      await editor.dispose()
    }
  })

  test('lists the tutor logins to lecturers only', async () => {
    const lecturer = await apiContextFor('course-lecturer')
    try {
      const res = await lecturer.get(phaseUrl('/team/tutors'))
      expect(res.status()).toBe(200)
      const assignments = ((await res.json()) as TutorAssignment[]).filter(
        (a) => a.id === ledTutorId,
      )
      expect(assignments.map((a) => a.teamID).sort()).toEqual(
        [teams.lovelace.id, teams.turing.id].sort(),
      )
      for (const assignment of assignments) {
        expect(assignment.universityLogin).toBe(EDITOR_LOGIN)
      }
    } finally {
      await lecturer.dispose()
    }

    const editor = await apiContextFor('course-editor')
    try {
      expect((await editor.get(phaseUrl('/team/tutors'))).status()).toBe(403)
    } finally {
      await editor.dispose()
    }
  })

  test.describe('as the lecturer', () => {
    test.use({ role: 'course-lecturer' })

    test('adds a team to a course editor who already leads several', async ({ page }) => {
      const phase = new SelfTeamAllocationPage(page)
      await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID)
      await phase.expectLoaded()

      await page.getByRole('button', { name: 'Import Tutors' }).click()
      const dialog = page.getByRole('dialog', { name: 'Import Tutors' })
      await dialog.getByRole('tab', { name: 'Course Staff' }).click()

      // Course Staff offers the editor under their Keycloak id while their rows carry ledTutorId,
      // so the picker finds the teams they lead by login.
      await dialog.getByRole('checkbox', { name: EDITOR_NAME }).click()
      const picker = dialog.getByRole('button', { name: `Teams of ${EDITOR_NAME}` })
      await expect(picker).toContainText(teams.lovelace.name)
      await expect(picker).toContainText(teams.turing.name)
      await expect(picker).not.toContainText(teams.hamilton.name)

      await picker.click()
      const options = page.getByRole('listbox')
      await options.getByRole('option', { name: teams.hamilton.name }).click()
      await options.getByRole('option', { name: 'Close' }).click()

      await dialog.getByRole('button', { name: 'Import Selected Tutors' }).click()
      await expect(dialog).toBeHidden({ timeout: 15_000 })

      // The import sets exactly the selected teams and keeps the id the tutor already has.
      const editorTeams = await teamsAs('course-editor')
      expect(editorTeams.map((t) => t.name).sort()).toEqual([
        teams.hamilton.name,
        teams.lovelace.name,
        teams.turing.name,
      ])
      const hamilton = (await teamsAs('course-lecturer')).find((t) => t.id === teams.hamilton.id)
      expect(hamilton?.tutors.map((t) => t.id)).toEqual([ledTutorId])

      await expect(phase.teamCard(teams.hamilton.name).getByText(EDITOR_NAME)).toBeVisible()
    })
  })
})
