import { Page, Locator, expect } from '@playwright/test'

export interface SlotFormValues {
  startTime: string // datetime-local format: YYYY-MM-DDTHH:mm
  endTime: string // time-of-day format: HH:mm (interviews never span days)
  location: string
  capacity: number
}

// /management/course/:courseId/:phaseId — the Interview remote (Module
// Federation) rendered inside the core shell. The phase root shows the
// student booking view (staff see it with a disclaimer); /schedule is the
// lecturer-only slot management page.
export class InterviewPage {
  readonly heading: Locator

  constructor(private readonly page: Page) {
    this.heading = this.page.getByRole('heading', { level: 1, name: 'Interview Scheduling' })
  }

  async goto(courseId: string, phaseId: string, subpath = '') {
    await this.page.goto(`/management/course/${courseId}/${phaseId}${subpath}`)
  }

  async expectLoaded() {
    await expect(this.heading).toBeVisible({ timeout: 15_000 })
  }

  // ── Schedule management (lecturer) ───────────────────────────────────────

  async gotoSchedule(courseId: string, phaseId: string) {
    await this.goto(courseId, phaseId, '/schedule')
  }

  async expectScheduleLoaded() {
    await expect(
      this.page.getByRole('heading', { name: 'Interview Schedule Management' }),
    ).toBeVisible({ timeout: 15_000 })
  }

  async createSlot({ startTime, endTime, location, capacity }: SlotFormValues) {
    await this.page.getByRole('button', { name: 'Create Slots', exact: true }).click()
    const dialog = this.page.getByRole('dialog', { name: 'Create Interview Slots' })
    const createMultipleSlots = dialog.getByLabel('Create multiple slots')
    await expect(createMultipleSlots).not.toBeChecked()
    await createMultipleSlots.check()
    await expect(dialog.getByLabel('Slot Duration (min)')).toBeVisible()
    await expect(dialog.getByLabel('Break Between (min)')).toBeVisible()
    await createMultipleSlots.uncheck()
    await expect(dialog.getByLabel('Slot Duration (min)')).toBeHidden()
    await dialog.getByLabel('Start Time').fill(startTime)
    await dialog.getByLabel('End Time').fill(endTime)
    await dialog.getByLabel('Location (Optional)').fill(location)
    await dialog.getByLabel('Capacity per Slot').fill(String(capacity))
    await dialog.getByRole('button', { name: 'Create Slot', exact: true }).click()
    await expect(dialog).toBeHidden()
    await expect(this.slotRow(location)).toBeVisible()
  }

  slotRow(location: string): Locator {
    return this.page.getByRole('row', { name: new RegExp(location) })
  }

  async assignStudent(location: string, studentName: string) {
    await this.slotRow(location).getByRole('button', { name: 'Assign student' }).click()
    const dialog = this.page.getByRole('dialog', { name: 'Assign Student to Interview Slot' })
    await dialog.getByRole('combobox').click()
    await this.page.getByRole('option', { name: studentName }).click()
    await dialog.getByRole('button', { name: 'Assign Student' }).click()
    await expect(dialog).toBeHidden()
  }

  // ── Student booking view (phase root) ────────────────────────────────────

  slotCard(slotId: string): Locator {
    return this.page.getByTestId(`interview-slot-${slotId}`)
  }

  // Selecting the card reveals the confirm button; booking flips the badge.
  async bookSlot(slotId: string) {
    const card = this.slotCard(slotId)
    await card.click()
    await card.getByRole('button', { name: 'Confirm Booking' }).click()
    await expect(card.getByText('Booked', { exact: true })).toBeVisible()
  }

  async cancelBooking(slotId: string, capacity = 1) {
    const card = this.slotCard(slotId)
    await card.getByRole('button', { name: 'Cancel Booking' }).click()
    await expect(card.getByText('Available', { exact: true })).toBeVisible()
    await expect(card.getByText(`0 / ${capacity} booked`)).toBeVisible()
  }
}
