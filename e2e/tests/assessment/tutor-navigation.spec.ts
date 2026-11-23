import { expect, test } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import { ASSESSMENT_FIXTURE_PHASES, SEEDED_COURSES } from '../../src/data/constants'
import { apiAsRole, assessmentUrl, putConfig } from './helpers'

// Its teams come from the tutor scope Team Allocation phase: course-editor (Eduard Eduard)
// leads Team Lovelace and Team Turing, and Carla Coach tutors Team Lovelace too. So the
// tutor table has three rows but only two tutors to step through.
const PHASE_ID = ASSESSMENT_FIXTURE_PHASES.tutorNavigation
const COURSE_ID = SEEDED_COURSES.fullCourse.id
const COACH = 'Carla Coach'
const LEAD = 'Eduard Eduard'

interface Team {
  tutors: { id?: string; firstName: string; lastName: string }[]
}

test.use({ role: 'lecturer' })

test.describe('assessment: tutor previous / next follow the tutor table order', () => {
  test.beforeAll(async () => {
    const lecturer = await apiAsRole('lecturer')
    try {
      // Rows only open the results page while the tutor evaluation is on.
      await putConfig(lecturer, PHASE_ID, { tutorEvaluationEnabled: true })
    } finally {
      await lecturer.dispose()
    }
  })

  for (const { direction, clicks, order } of [
    { direction: 'ascending', clicks: 1, order: [COACH, LEAD] },
    { direction: 'descending', clicks: 2, order: [LEAD, COACH] },
  ]) {
    test(`steps through each tutor once in ${direction} last name order`, async ({ page }) => {
      const phase = new AssessmentPage(page)
      await phase.goto(COURSE_ID, PHASE_ID, '/tutors')
      const rows = page.locator('tbody').getByRole('row')
      await expect(rows).toHaveCount(3)

      const sortByLastName = phase.columnHeader('Last Name').getByRole('button')
      for (let click = 0; click < clicks; click++) {
        await sortByLastName.click()
      }
      await expect(rows.first()).toHaveAccessibleName(new RegExp(order[0]))

      await rows.first().click()
      await phase.scrollToTop()
      await phase.expectOpenedParticipant(order[0], 1, order.length)
      await expect(phase.previousParticipantButton()).toBeDisabled()
      await expect(phase.nextParticipantButton()).toHaveAccessibleName(
        `Next participant: ${order[1]}`,
      )

      await phase.nextParticipantButton().click()
      await phase.expectOpenedParticipant(order[1], 2, order.length)
      // No wrap-around at the end of the table.
      await expect(phase.nextParticipantButton()).toBeDisabled()
      await expect(phase.previousParticipantButton()).toHaveAccessibleName(
        `Previous participant: ${order[0]}`,
      )

      await phase.previousParticipantButton().click()
      await phase.expectOpenedParticipant(order[0], 1, order.length)
    })
  }

  test('falls back to the team order when opened directly', async ({ page }) => {
    const api = await apiAsRole('lecturer')
    let teams: Team[]
    try {
      const res = await api.get(assessmentUrl(PHASE_ID, 'config/teams'))
      expect(res.status()).toBe(200)
      teams = (await res.json()) as Team[]
    } finally {
      await api.dispose()
    }
    // Each tutor once, at their first team.
    const tutors = [
      ...new Map(
        teams
          .flatMap((team) => team.tutors)
          .filter((tutor) => tutor.id)
          .map((tutor) => [tutor.id, tutor]),
      ).values(),
    ]
    expect(tutors).toHaveLength(2)
    const nameOf = (tutor: Team['tutors'][number]) => `${tutor.firstName} ${tutor.lastName}`

    // No router state, as after a reload or from a shared link.
    const phase = new AssessmentPage(page)
    await phase.goto(COURSE_ID, PHASE_ID, `/tutors/${tutors[0].id}`)
    await phase.expectOpenedParticipant(nameOf(tutors[0]), 1, 2)
    await expect(phase.previousParticipantButton()).toBeDisabled()
    await expect(phase.nextParticipantButton()).toHaveAccessibleName(
      `Next participant: ${nameOf(tutors[1])}`,
    )

    await phase.nextParticipantButton().click()
    await phase.expectOpenedParticipant(nameOf(tutors[1]), 2, 2)
    await expect(phase.nextParticipantButton()).toBeDisabled()
  })
})
