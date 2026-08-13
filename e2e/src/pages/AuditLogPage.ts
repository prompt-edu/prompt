import { Page, Locator, expect } from '@playwright/test'

// /management/course/:id/audit-log (course-scoped) and /management/admin/audit-log
// (platform-wide). Both render a page header plus a filterable table; the table
// renders even when empty.
export class AuditLogPage {
  readonly heading: Locator
  readonly accessDenied: Locator

  constructor(private readonly page: Page) {
    this.heading = page.getByRole('heading', { name: 'Audit Log' })
    this.accessDenied = page.getByText('Access Denied', { exact: true })
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
}
