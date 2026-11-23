import { expect, test } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import {
  ASSESSMENT_FIXTURE_PHASES,
  SEEDED_COURSES,
  TUTOR_SCOPE_PARTICIPANTS,
} from '../../src/data/constants'
import { apiAsRole, assessmentUrl } from './helpers'

// Read only: the tutor scope phase has five participants and no spec changes it.
const PHASE_ID = ASSESSMENT_FIXTURE_PHASES.tutorScope
const TOTAL = TUTOR_SCOPE_PARTICIPANTS.ownTeam.length + TUTOR_SCOPE_PARTICIPANTS.hidden.length

// The participants sorted by last name, descending: user, Test, Stan, Second, Heun.
const BY_LAST_NAME_DESC = ['New Test user', 'Test Test', 'Stan Stan', 'Selma Second', 'Niclas Heun']

interface Participation {
  courseParticipationID: string
  student: { firstName: string; lastName: string }
}

// The lecturer sees every participant, so the whole phase is navigable.
test.use({ role: 'lecturer' })

test.describe('assessment: previous / next follow the table order', () => {
  test('steps through the participants in the order the table shows them', async ({ page }) => {
    const phase = new AssessmentPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/participants')
    await phase.expectParticipantsLoaded()

    // The first click sorts ascending, the second descending.
    const sortByLastName = phase.columnHeader('Last name').getByRole('button')
    await sortByLastName.click()
    await sortByLastName.click()
    const rows = page.locator('tbody').getByRole('row')
    await expect(rows).toHaveCount(TOTAL)
    await expect(rows.first()).toHaveAccessibleName(new RegExp(BY_LAST_NAME_DESC[0]))

    await rows.first().click()
    // The participant page keeps the scroll offset of the table, which can dock the header.
    await phase.scrollToTop()
    await phase.expectOpenedParticipant(BY_LAST_NAME_DESC[0], 1, TOTAL)
    await expect(phase.previousParticipantButton()).toBeDisabled()
    await expect(phase.nextParticipantButton()).toHaveAccessibleName(
      `Next participant: ${BY_LAST_NAME_DESC[1]}`,
    )

    for (let index = 1; index < BY_LAST_NAME_DESC.length; index++) {
      await phase.nextParticipantButton().click()
      await phase.expectOpenedParticipant(BY_LAST_NAME_DESC[index], index + 1, TOTAL)
    }
    await expect(phase.nextParticipantButton()).toBeDisabled()
    await expect(phase.previousParticipantButton()).toHaveAccessibleName(
      `Previous participant: ${BY_LAST_NAME_DESC[BY_LAST_NAME_DESC.length - 2]}`,
    )

    await phase.previousParticipantButton().click()
    await phase.expectOpenedParticipant(
      BY_LAST_NAME_DESC[BY_LAST_NAME_DESC.length - 2],
      BY_LAST_NAME_DESC.length - 1,
      TOTAL,
    )
  })

  test('falls back to the participations order when opened directly', async ({ page }) => {
    const api = await apiAsRole('lecturer')
    let participations: Participation[]
    try {
      const res = await api.get(assessmentUrl(PHASE_ID, 'config/participations'))
      expect(res.status()).toBe(200)
      participations = (await res.json()) as Participation[]
    } finally {
      await api.dispose()
    }
    expect(participations).toHaveLength(TOTAL)
    const nameOf = ({ student }: Participation) => `${student.firstName} ${student.lastName}`

    // No router state, as after a reload or from a shared link.
    const phase = new AssessmentPage(page)
    await phase.goto(
      SEEDED_COURSES.fullCourse.id,
      PHASE_ID,
      `/participants/${participations[0].courseParticipationID}`,
    )
    await phase.expectOpenedParticipant(nameOf(participations[0]), 1, TOTAL)
    await expect(phase.previousParticipantButton()).toBeDisabled()
    await expect(phase.nextParticipantButton()).toHaveAccessibleName(
      `Next participant: ${nameOf(participations[1])}`,
    )

    await phase.nextParticipantButton().click()
    await phase.expectOpenedParticipant(nameOf(participations[1]), 2, TOTAL)
  })
})
