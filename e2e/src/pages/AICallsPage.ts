import { Page, Locator, expect } from '@playwright/test'

// /management/course/:id/ai (admins only), served by the ai_component remote. Lists
// the AI calls of one course phase at a time; a row opens the call's content in a dialog.
export class AICallsPage {
  readonly heading: Locator
  readonly accessDenied: Locator
  readonly rows: Locator
  readonly phaseSelect: Locator

  constructor(private readonly page: Page) {
    this.heading = page.getByRole('heading', { name: 'AI Calls' })
    this.accessDenied = page.getByText('Access Denied', { exact: true })
    this.rows = page.locator('table tbody tr')
    this.phaseSelect = page.getByRole('combobox', { name: 'Course phase' })
  }

  async goto(courseId: string) {
    await this.page.goto(`/management/course/${courseId}/ai`)
  }

  async expectLoaded() {
    await expect(this.heading).toBeVisible({ timeout: 15_000 })
  }

  async selectPhase(phaseName: string) {
    await this.phaseSelect.click()
    await this.page.getByRole('option', { name: phaseName, exact: true }).click()
  }
}
