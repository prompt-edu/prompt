import { expect, Page } from '@playwright/test'

export class PresentationPage {
  constructor(private readonly page: Page) {}

  async goto(courseId: string, phaseId: string) {
    await this.page.goto(`/management/course/${courseId}/${phaseId}`)
  }

  async expectOverviewLoaded() {
    await expect(this.page.getByRole('heading', { name: 'Presentations' })).toBeVisible({
      timeout: 15_000,
    })
    await expect(this.page.getByText('No presentation is assigned yet')).toBeVisible()
  }
}
