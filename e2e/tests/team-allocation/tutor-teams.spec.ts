import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { BASE_URL, TEAM_ALLOCATION_API } from '../../src/env'
import { TeamAllocationPage } from '../../src/pages/TeamAllocationPage'
import {
  SEEDED_COURSES,
  TEAM_ALLOCATION_TUTOR_SCOPE,
  TUTOR_SCOPE_PARTICIPANTS,
} from '../../src/data/constants'
import { getAllocations } from './helpers'

// course-editor is a project lead of two of the three teams of the tutor scope phase.
// They are imported from the course staff, which works because the e2e realm puts
// them in the Editor group of iPraktikumFull.
const { phaseId: PHASE_ID, ledTutorId, teams } = TEAM_ALLOCATION_TUTOR_SCOPE
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

const teamUrl = (path = '') =>
  `${BASE_URL}${TEAM_ALLOCATION_API}/course_phase/${PHASE_ID}/team${path}`

async function teamsAs(role: 'course-editor' | 'course-lecturer'): Promise<TeamWithTutors[]> {
  const api = await apiContextFor(role)
  try {
    const res = await api.get(teamUrl())
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
    const res = await api.post(teamUrl('/tutors'), {
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

test.describe('team allocation: a tutor leading several teams', () => {
  test.afterAll(restoreSeededLead)

  test('reads exactly the teams and allocations the tutor leads', async () => {
    const editorTeams = await teamsAs('course-editor')
    expect(editorTeams.map((t) => t.name).sort()).toEqual([teams.lovelace.name, teams.turing.name])

    const editor = await apiContextFor('course-editor')
    try {
      const allocations = await getAllocations(editor, PHASE_ID)
      expect(allocations.map((a) => a.courseParticipationID).sort()).toEqual(
        TUTOR_SCOPE_PARTICIPANTS.ownTeam.map((p) => p.courseParticipationId).sort(),
      )

      expect((await editor.get(teamUrl(`/${teams.turing.id}`))).status()).toBe(200)
      expect((await editor.get(teamUrl(`/${teams.hamilton.id}`))).status()).toBe(403)
    } finally {
      await editor.dispose()
    }

    const lecturerTeams = await teamsAs('course-lecturer')
    for (const team of [teams.lovelace, teams.turing]) {
      const tutors = lecturerTeams.find((t) => t.id === team.id)?.tutors ?? []
      expect(tutors.map((t) => t.id)).toContain(ledTutorId)
    }
  })

  test('lists the tutor logins to lecturers only', async () => {
    const lecturer = await apiContextFor('course-lecturer')
    try {
      const res = await lecturer.get(teamUrl('/tutors'))
      expect(res.status()).toBe(200)
      const assignments = (await res.json()) as TutorAssignment[]
      expect(
        assignments
          .filter((a) => a.id === ledTutorId)
          .map((a) => a.teamID)
          .sort(),
      ).toEqual([teams.lovelace.id, teams.turing.id].sort())
      for (const assignment of assignments.filter((a) => a.id === ledTutorId)) {
        expect(assignment.universityLogin).toBe(EDITOR_LOGIN)
      }
    } finally {
      await lecturer.dispose()
    }

    for (const role of ['course-editor', 'student'] as const) {
      const api = await apiContextFor(role)
      try {
        expect((await api.get(teamUrl('/tutors'))).status(), role).toBe(403)
      } finally {
        await api.dispose()
      }
    }
  })

  test.describe('as the lecturer', () => {
    test.use({ role: 'course-lecturer' })

    test('adds a team to a course editor who already leads several', async ({ page }) => {
      const phase = new TeamAllocationPage(page)
      await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/allocations')
      await phase.expectAllocationsLoaded()

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

      await expect(
        page.getByRole('heading', { level: 3, name: teams.hamilton.name, exact: true }),
      ).toBeVisible()
    })
  })
})
