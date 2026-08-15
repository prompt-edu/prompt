import { test, expect } from '../../src/fixtures/auth'
import { InterviewPage } from '../../src/pages/InterviewPage'
import { SEEDED_COURSES, FULL_COURSE_PHASES } from '../../src/data/constants'
import {
  deleteSlotsByLocation,
  futureSlotRange,
  futureSlotTimes,
  toDatetimeLocal,
  toTimeOfDay,
} from './helpers'

const SERIES_LOCATION = 'E2E Slot Series'
const CLONE_LOCATION = 'E2E Slot Clone'
const SERIES_SLOT_COUNT = 4
const SERIES_DURATION_MINUTES = 30

test.use({ role: 'course-lecturer' })

test.describe('interview: slot series and cloning', () => {
  test.beforeAll(async () => {
    await deleteSlotsByLocation(SERIES_LOCATION)
    await deleteSlotsByLocation(CLONE_LOCATION)
  })

  test.afterAll(async () => {
    await deleteSlotsByLocation(SERIES_LOCATION)
    await deleteSlotsByLocation(CLONE_LOCATION)
  })

  test('the duration and break fields appear only while "Create multiple slots" is checked', async ({
    page,
  }) => {
    const phase = new InterviewPage(page)
    await phase.gotoSchedule(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.interview.id)
    await phase.expectScheduleLoaded()

    const dialog = await phase.openCreateDialog()
    const createMultipleSlots = dialog.getByLabel('Create multiple slots')

    await expect(createMultipleSlots).not.toBeChecked()
    await expect(dialog.getByLabel('Slot Duration (min)')).toBeHidden()

    await createMultipleSlots.check()
    await expect(dialog.getByLabel('Slot Duration (min)')).toBeVisible()
    await expect(dialog.getByLabel('Break Between (min)')).toBeVisible()

    await createMultipleSlots.uncheck()
    await expect(dialog.getByLabel('Slot Duration (min)')).toBeHidden()
    await expect(dialog.getByLabel('Break Between (min)')).toBeHidden()

    await dialog.getByRole('button', { name: 'Cancel' }).click()
    await expect(dialog).toBeHidden()
  })

  test('a lecturer divides a time range into a series of slots', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.gotoSchedule(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.interview.id)
    await phase.expectScheduleLoaded()

    const { start, end } = futureSlotRange(SERIES_SLOT_COUNT, SERIES_DURATION_MINUTES)
    await phase.createSlotSeries({
      startTime: toDatetimeLocal(start),
      endTime: toTimeOfDay(end),
      location: SERIES_LOCATION,
      capacity: 1,
      durationMinutes: SERIES_DURATION_MINUTES,
      breakMinutes: 0,
    })

    await expect(phase.slotRow(SERIES_LOCATION).first()).toBeVisible()
    expect(await phase.slotRowCount(SERIES_LOCATION)).toBe(SERIES_SLOT_COUNT)
  })

  test('cloning a slot prefills the create dialog and creates a copy', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.gotoSchedule(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.interview.id)
    await phase.expectScheduleLoaded()

    const { start, end } = futureSlotTimes()
    await phase.createSlot({
      startTime: toDatetimeLocal(start),
      endTime: toTimeOfDay(end),
      location: CLONE_LOCATION,
      capacity: 2,
    })

    const dialog = await phase.cloneSlot(CLONE_LOCATION)
    await expect(dialog.getByLabel('Start Time')).toHaveValue(toDatetimeLocal(start))
    await expect(dialog.getByLabel('End Time')).toHaveValue(toTimeOfDay(end))
    await expect(dialog.getByLabel('Location (Optional)')).toHaveValue(CLONE_LOCATION)
    await expect(dialog.getByLabel('Capacity per Slot')).toHaveValue('2')
    // Cloning must not carry the series checkbox over from a previous series creation.
    await expect(dialog.getByLabel('Create multiple slots')).not.toBeChecked()

    await dialog.getByRole('button', { name: 'Create Slot', exact: true }).click()
    await expect(dialog).toBeHidden()

    expect(await phase.slotRowCount(CLONE_LOCATION)).toBe(2)
  })
})
