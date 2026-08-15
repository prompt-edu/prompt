import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { InterviewPage } from '../../src/pages/InterviewPage'
import { SEEDED_COURSES, FULL_COURSE_PHASES } from '../../src/data/constants'
import { createSlot, deleteSlotsByLocation } from './helpers'

const LOCATION = 'E2E Student Slot'

test.use({ role: 'student' })

test.describe('interview: student journey', () => {
  let slotId: string

  test.beforeAll(async () => {
    await deleteSlotsByLocation(LOCATION)
    const admin = await apiContextFor('admin')
    try {
      const slot = await createSlot(admin, LOCATION, 1)
      slotId = slot.id
    } finally {
      await admin.dispose()
    }
  })

  test.afterAll(async () => {
    await deleteSlotsByLocation(LOCATION)
  })

  test('a student books a slot and then cancels the booking', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.interview.id)
    await phase.expectLoaded()

    await phase.bookSlot(slotId)
    await expect(phase.slotCard(slotId).getByText('1 / 1 booked')).toBeVisible()

    await phase.cancelBooking(slotId)
  })
})
