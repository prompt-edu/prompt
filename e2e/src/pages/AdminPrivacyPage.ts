import { type Locator, type Page, expect } from '@playwright/test'

// /management/admin/privacy — the PROMPT_Admin-only console listing every
// deletion request and every export.
export class AdminPrivacyPage {
  constructor(private readonly page: Page) {}

  async goto() {
    await this.page.goto('/management/admin/privacy')
  }

  async expectLoaded() {
    await expect(this.page.getByRole('heading', { level: 1, name: 'Privacy' })).toBeVisible({
      timeout: 15_000,
    })
  }

  async expectUnauthorized() {
    await expect(this.page.getByText('Access Denied')).toBeVisible({ timeout: 15_000 })
  }

  async openDeletionTab() {
    await this.page.getByRole('tab', { name: 'Deletion' }).click()
    await expect(this.page.getByRole('heading', { name: 'Deletion Requests' })).toBeVisible({
      timeout: 15_000,
    })
  }

  async openExportTab() {
    await this.page.getByRole('tab', { name: 'Export' }).click()
    await expect(this.page.getByRole('heading', { name: 'Data Exports' })).toBeVisible({
      timeout: 15_000,
    })
  }

  get rows(): Locator {
    return this.page.locator('table tbody tr')
  }

  // The requester column renders a StudentAvatar, which shows "First Last" and
  // no email, so rows are found by name.
  rowFor(requesterName: string): Locator {
    return this.rows.filter({ hasText: requesterName })
  }

  // The review dialog only opens for a request still awaiting approval.
  async openReview(requester: string) {
    await this.rowFor(requester).click()
    await expect(this.reviewDialog).toBeVisible({ timeout: 15_000 })
  }

  get reviewDialog(): Locator {
    return this.page.getByRole('dialog', { name: 'Review Deletion Request' })
  }

  async reject() {
    await this.reviewDialog.getByRole('button', { name: 'Reject', exact: true }).click()
    await expect(this.reviewDialog).toBeHidden({ timeout: 15_000 })
  }

  // The approve button is disabled for a five-second countdown.
  async approve() {
    const approve = this.reviewDialog.getByRole('button', { name: /^Approve and Start Deletion$/ })
    await expect(approve).toBeEnabled({ timeout: 15_000 })
    await approve.click()
    await expect(this.reviewDialog).toBeHidden({ timeout: 15_000 })
  }

  async selectStatus(label: string) {
    await this.page.getByRole('combobox').filter({ hasText: 'Status' }).click()
    await this.page.getByRole('option', { name: label }).click()
  }

  // PromptTable renders its row actions behind a per-row kebab, which is a bare
  // div carrying only Radix's aria-haspopup.
  async runRowAction(row: Locator, action: string) {
    await row.locator('[aria-haspopup="menu"]').click()
    await this.page.getByRole('menuitem', { name: action }).click()
  }

  async archiveExportAndResetRateLimit(requester: string) {
    await this.runRowAction(this.rowFor(requester), 'Archive + reset rate limit')
  }
}
