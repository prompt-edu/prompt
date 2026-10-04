import { Page, Locator, expect } from '@playwright/test'

// /management/students — visible only to admins and lecturers.
export class StudentsPage {
  readonly heading: Locator

  constructor(private readonly page: Page) {
    // level 1 to avoid the sidebar's "STUDENTS" section label (an <h3>).
    this.heading = page.getByRole('heading', { level: 1, name: 'Students' })
  }

  async goto() {
    await this.page.goto('/management/students')
  }

  async expectLoaded() {
    await expect(this.heading).toBeVisible()
  }

  async expectBlocked() {
    // Unauthorized users never see the Students heading (PermissionRestriction
    // renders UnauthorizedPage instead).
    await expect(this.heading).toBeHidden()
  }

  rowFor(email: string): Locator {
    return this.page.locator('table tbody tr').filter({ hasText: email })
  }

  // "Not modified in N years" compares against the student's last_modified, so
  // only a deliberately old fixture can match it.
  async filterByInactivity(years: number) {
    await this.page.getByRole('button', { name: /Filter/ }).click()
    await this.page.getByRole('menuitem', { name: 'Last Modified' }).click()
    await this.page.getByRole('menuitemcheckbox', { name: `Not modified in ${years} years` }).click()
    await this.page.keyboard.press('Escape')
  }

  // PromptTable renders its row actions behind a per-row kebab, which is a bare
  // div carrying only Radix's aria-haspopup.
  async openDeletionDialogFor(email: string) {
    await this.rowFor(email).locator('[aria-haspopup="menu"]').click()
    await this.page.getByRole('menuitem', { name: 'Delete Student Data' }).click()
    await expect(this.deletionDialog).toBeVisible({ timeout: 15_000 })
  }

  get deletionDialog(): Locator {
    return this.page.getByRole('dialog', { name: 'Delete Student Data' })
  }

  // The trigger is disabled for a five-second countdown.
  async startDeletion() {
    const start = this.deletionDialog.getByRole('button', { name: /^Start Deletion$/ })
    await expect(start).toBeEnabled({ timeout: 15_000 })
    await start.click()
  }
}
