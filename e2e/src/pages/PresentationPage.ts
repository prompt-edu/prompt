import { expect, Page } from '@playwright/test'

export class PresentationPage {
  constructor(private readonly page: Page) {}

  async goto(courseId: string, phaseId: string) {
    await this.page.goto(`/management/course/${courseId}/${phaseId}`)
  }

  // Instructors get a disabled preview of the presenter page, filled with sample data, so the
  // disclaimer is what proves the remote rendered and reached its phase API.
  async expectOverviewLoaded() {
    await expect(this.page.getByRole('heading', { name: 'Presentations' })).toBeVisible({
      timeout: 15_000,
    })
    await expect(this.page.getByText('You are not a student of this course.')).toBeVisible()
    await expect(this.page.getByText('Presentation materials')).toBeVisible()
  }
}
