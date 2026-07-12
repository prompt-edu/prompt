import { test, expect } from '../../src/fixtures/auth'
import { InterviewPage } from '../../src/pages/InterviewPage'
import { SEEDED_COURSES, FULL_COURSE_PHASES } from '../../src/data/constants'
import { deleteSlotsByLocation, futureSlotTimes, toDatetimeLocal } from './helpers'

const LOCATION = 'E2E Lecturer Slot'
// The only uniquely-named interview participant: the phase has two distinct
// "Niclas Heun" students, and Stan Stan is reserved for the student journey.
const STUDENT_NAME = 'Max Mustermann'

test.use({ role: 'course-lecturer' })

test.describe('interview: lecturer journey', () => {
  test.beforeAll(async () => {
    await deleteSlotsByLocation(LOCATION)
  })

  test.afterAll(async () => {
    await deleteSlotsByLocation(LOCATION)
  })

  test('a lecturer creates a slot and manually assigns a student', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.gotoSchedule(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.interview.id)
    await phase.expectScheduleLoaded()

    const { start, end } = futureSlotTimes()
    await phase.createSlot({
      startTime: toDatetimeLocal(start),
      endTime: toDatetimeLocal(end),
      location: LOCATION,
      capacity: 1,
    })

    const row = phase.slotRow(LOCATION)
    await expect(row.getByText('No bookings yet')).toBeVisible()
    await expect(row.getByText('0 / 1')).toBeVisible()

    await phase.assignStudent(LOCATION, STUDENT_NAME)

    await expect(row.getByText(STUDENT_NAME)).toBeVisible()
    await expect(row.getByText('1 / 1')).toBeVisible()
    await expect(row.getByText('Full')).toBeVisible()
  })
})
