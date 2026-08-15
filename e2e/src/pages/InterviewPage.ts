import { Page, Locator, expect } from '@playwright/test'

export interface SlotFormValues {
  startTime: string // datetime-local format: YYYY-MM-DDTHH:mm
  endTime: string // time-of-day format: HH:mm; an end at or before the start means the next day
  location: string
  capacity: number
}

export interface SlotSeriesFormValues extends SlotFormValues {
  durationMinutes: number
  breakMinutes: number
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

  createDialog(): Locator {
    return this.page.getByRole('dialog', { name: 'Create Interview Slots' })
  }

  async openCreateDialog(): Promise<Locator> {
    await this.page.getByRole('button', { name: 'Create Slots', exact: true }).click()
    return this.createDialog()
  }

  private async fillSlotFields(
    dialog: Locator,
    { startTime, endTime, location, capacity }: SlotFormValues,
  ) {
    await dialog.getByLabel('Start Time').fill(startTime)
    await dialog.getByLabel('End Time').fill(endTime)
    await dialog.getByLabel('Location (Optional)').fill(location)
    await dialog.getByLabel('Capacity per Slot').fill(String(capacity))
  }

  async createSlot(values: SlotFormValues) {
    const dialog = await this.openCreateDialog()
    await this.fillSlotFields(dialog, values)
    await dialog.getByRole('button', { name: 'Create Slot', exact: true }).click()
    await expect(dialog).toBeHidden()
    await expect(this.slotRow(values.location)).toBeVisible()
  }

  // Divides the time range into slots via the "Create multiple slots" checkbox.
  async createSlotSeries({ durationMinutes, breakMinutes, ...values }: SlotSeriesFormValues) {
    const dialog = await this.openCreateDialog()
    await this.fillSlotFields(dialog, values)
    await dialog.getByLabel('Create multiple slots').check()
    await dialog.getByLabel('Slot Duration (min)').fill(String(durationMinutes))
    await dialog.getByLabel('Break Between (min)').fill(String(breakMinutes))
    await dialog.getByRole('button', { name: /^Create \d+ Slots?$/ }).click()
    await expect(dialog).toBeHidden()
  }

  // Opens the create dialog prefilled from an existing slot.
  async cloneSlot(location: string): Promise<Locator> {
    await this.slotRow(location).first().getByRole('button', { name: 'Clone slot' }).click()
    return this.createDialog()
  }

  slotRow(location: string): Locator {
    return this.page.getByRole('row', { name: new RegExp(location) })
  }

  async slotRowCount(location: string): Promise<number> {
    return this.slotRow(location).count()
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
