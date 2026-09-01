import { Page, Locator, expect } from '@playwright/test'

// /management/course/:id/audit-log (course-scoped) and /management/admin/audit-log
// (platform-wide). Both render a page header plus a filterable table; the table
// renders even when empty. The log is paged server-side one keyset page at a
// time, navigated with the Newer/Older buttons.
export class AuditLogPage {
  readonly heading: Locator
  readonly accessDenied: Locator
  readonly rows: Locator
  readonly newer: Locator
  readonly older: Locator
  readonly rowsPerPage: Locator
  readonly deniedRows: Locator

  constructor(private readonly page: Page) {
    this.heading = page.getByRole('heading', { name: 'Audit Log' })
    this.accessDenied = page.getByText('Access Denied', { exact: true })
    this.rows = page.locator('table tbody tr')
    this.newer = page.getByRole('button', { name: 'Newer entries' })
    this.older = page.getByRole('button', { name: 'Older entries' })
    this.rowsPerPage = page.locator('div', { hasText: /^Rows per page/ }).getByRole('combobox')
    this.deniedRows = this.rows.filter({ hasText: 'denied' })
  }

  async gotoCourse(courseId: string) {
    await this.page.goto(`/management/course/${courseId}/audit-log`)
  }

  async gotoGlobal() {
    await this.page.goto('/management/admin/audit-log')
  }

  async expectLoaded() {
    await expect(this.heading).toBeVisible({ timeout: 15_000 })
  }

  async expectUnauthorized() {
    await expect(this.accessDenied).toBeVisible({ timeout: 15_000 })
  }

  async selectOutcome(label: string) {
    await this.page.getByRole('combobox').filter({ hasText: 'All outcomes' }).click()
    await this.page.getByRole('option', { name: label }).click()
  }
}
