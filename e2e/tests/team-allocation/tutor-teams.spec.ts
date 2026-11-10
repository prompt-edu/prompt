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

interface TeamWithTutors {
  id: string
  name: string
  tutors: { id: string; firstName: string; lastName: string }[]
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
          universityLogin: 'co69edt',
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

  test.describe('as the lecturer', () => {
    test.use({ role: 'course-lecturer' })

    test('imports a course editor from the course staff as the lead of several teams', async ({
      page,
    }) => {
      const phase = new TeamAllocationPage(page)
      await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/allocations')
      await phase.expectAllocationsLoaded()

      await page.getByRole('button', { name: 'Import Tutors' }).click()
      const dialog = page.getByRole('dialog', { name: 'Import Tutors' })
      await dialog.getByRole('tab', { name: 'Course Staff' }).click()

      await dialog.getByRole('checkbox', { name: EDITOR_NAME }).click()
      await dialog.getByRole('button', { name: `Teams of ${EDITOR_NAME}` }).click()
      const options = page.getByRole('listbox')
      await options.getByRole('option', { name: teams.lovelace.name }).click()
      await options.getByRole('option', { name: teams.hamilton.name }).click()
      await options.getByRole('option', { name: 'Close' }).click()

      await dialog.getByRole('button', { name: 'Import Selected Tutors' }).click()
      await expect(dialog).toBeHidden({ timeout: 15_000 })

      // The import sets exactly the selected teams and keeps the id the tutor already has.
      const editorTeams = await teamsAs('course-editor')
      expect(editorTeams.map((t) => t.name).sort()).toEqual([
        teams.hamilton.name,
        teams.lovelace.name,
      ])
      const hamilton = (await teamsAs('course-lecturer')).find((t) => t.id === teams.hamilton.id)
      expect(hamilton?.tutors.map((t) => t.id)).toEqual([ledTutorId])

      await expect(
        page.getByRole('heading', { level: 3, name: teams.hamilton.name, exact: true }),
      ).toBeVisible()
    })
  })
})
