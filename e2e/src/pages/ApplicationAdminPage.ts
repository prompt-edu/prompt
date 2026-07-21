import { Page, Locator, expect } from '@playwright/test'

export type ApplicationStatusBadge = 'Not Assessed' | 'Accepted' | 'Rejected'

// The lecturer-facing admin surfaces of an Application phase under
// /management/course/:courseId/:phaseId — the participants table (+ per-
// application Accept/Reject details), the questions page's student preview,
// and the mailing configuration page.
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

  async expectParticipantListed(email: string, name: string) {
    await expect(this.applicantRow(email)).toContainText(name)
  }

  async gotoQuestions(courseId: string, phaseId: string) {
    await this.page.goto(`/management/course/${courseId}/${phaseId}/questions`)
    await expect(
      this.page.getByRole('heading', { name: 'Application Questions' }),
    ).toBeVisible({ timeout: 15_000 })
  }

  // The "student preview" is a dialog rendering the student-facing application
  // form (the same personal-information section a student fills in).
  async openStudentPreview() {
    await this.page.getByRole('button', { name: 'Preview Application' }).click()
  }

  async expectPreviewLoaded() {
    const dialog = this.page.getByRole('dialog', { name: 'Application Preview' })
    await expect(dialog).toBeVisible({ timeout: 15_000 })
    await expect(dialog.getByRole('heading', { name: 'Personal Information' })).toBeVisible()
  }

  async gotoMailing(courseId: string, phaseId: string) {
    await this.page.goto(`/management/course/${courseId}/${phaseId}/mailing`)
  }

  async expectMailingLoaded() {
    await expect(
      this.page.getByRole('heading', { name: 'Application Mailing Settings' }),
    ).toBeVisible({ timeout: 15_000 })
    await expect(this.page.getByRole('heading', { name: 'Mailing Templates' })).toBeVisible()
  }

  async openApplication(email: string) {
    await this.applicantRow(email).getByText(email).click()
    await expect(this.acceptButton()).toBeVisible({ timeout: 15_000 })
  }

  async expectAnswerVisible(answer: string) {
    await expect(this.page.getByText(answer)).toBeVisible()
  }

  async expectQuestionVisible(title: string) {
    await expect(
      this.page.getByTestId('application-answers').getByText(title, { exact: true }),
    ).toBeVisible()
  }

  async openFilterMenu() {
    await this.page.getByRole('button', { name: 'Filter' }).click()
    await expect(this.page.getByRole('menu')).toBeVisible()
  }

  studyProgramOption(name: string): Locator {
    return this.page.getByRole('menuitemcheckbox', { name })
  }

  // The checkbox item calls preventDefault, so the menu stays open on select;
  // close it explicitly so the filtered table is not covered by the overlay.
  async filterByStudyProgram(name: string) {
    await this.studyProgramOption(name).click()
    await this.page.keyboard.press('Escape')
    await expect(this.page.getByRole('menu')).toBeHidden()
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
