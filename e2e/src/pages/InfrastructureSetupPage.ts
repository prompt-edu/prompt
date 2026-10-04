import { Page, Locator, expect } from '@playwright/test'

// /management/course/:courseId/:phaseId — the Infrastructure Setup remote
// (Module Federation) rendered inside the core shell. The overview renders the
// phase name as an <h1>, so match it by role.
export class InfrastructureSetupPage {
  readonly title: Locator

  constructor(private readonly page: Page) {
    this.title = this.page.getByRole('heading', { name: 'Infrastructure Setup', level: 1 })
  }

  async goto(courseId: string, phaseId: string, subPath = '') {
    await this.page.goto(`/management/course/${courseId}/${phaseId}${subPath}`)
  }

  async expectLoaded() {
    await expect(this.title).toBeVisible({ timeout: 15_000 })
  }
}
