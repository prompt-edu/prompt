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
}
