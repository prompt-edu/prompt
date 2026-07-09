import { Page, Locator, expect } from '@playwright/test'

export type ApplicationStatusBadge = 'Not Assessed' | 'Accepted' | 'Rejected'

// /management/course/:courseId/:phaseId/participants — the application phase's
// applications table plus the per-application details page (Accept/Reject).
export class ApplicationAdminPage {
  constructor(private readonly page: Page) {}

  async gotoParticipants(courseId: string, phaseId: string) {
    await this.page.goto(`/management/course/${courseId}/${phaseId}/participants`)
    await expect(
      this.page.getByRole('heading', { name: 'Application Participants' }),
    ).toBeVisible({ timeout: 15_000 })
  }

  applicantRow(email: string): Locator {
    return this.page.getByRole('row').filter({ hasText: email })
  }

  async expectStatus(email: string, status: ApplicationStatusBadge) {
    await expect(this.applicantRow(email)).toContainText(status)
  }

  async openApplication(email: string) {
    await this.applicantRow(email).getByText(email).click()
    await expect(this.acceptButton()).toBeVisible({ timeout: 15_000 })
  }

  async expectAnswerVisible(answer: string) {
    await expect(this.page.getByText(answer)).toBeVisible()
  }

  // Accept flips the phase participation's pass status to 'passed'; the button
  // disables once the updated status is reflected back into the page.
  async accept() {
    await this.acceptButton().click()
    await expect(this.acceptButton()).toBeDisabled()
  }

  private acceptButton(): Locator {
    return this.page.getByRole('button', { name: 'Accept', exact: true })
  }
}
