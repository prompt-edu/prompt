import { Page, Locator, expect } from '@playwright/test'

// /apply/:phaseId/authenticated — the application form for an already
// logged-in student. It is the one surface where a student views and edits
// their own profile (the "Personal Information" section prefilled from their
// account + student record). University-account fields (name, matriculation)
// are read-only here; the remaining personal fields are editable.
export class AuthenticatedApplicationPage {
  constructor(private readonly page: Page) {}

  async goto(phaseId: string) {
    await this.page.goto(`/apply/${phaseId}/authenticated`)
  }

  async expectLoaded() {
    await expect(
      this.page.getByRole('heading', { name: 'Personal Information' }),
    ).toBeVisible({ timeout: 15_000 })
  }

  firstName(): Locator {
    return this.page.getByLabel(/First Name/)
  }

  matriculationNumber(): Locator {
    return this.page.getByLabel(/Matriculation Number/)
  }

  currentSemester(): Locator {
    return this.page.getByLabel(/Current Semester/)
  }

  // Identity is immutable (name/matriculation are read-only here), so this is
  // safe to assert on the initial view regardless of prior edits or retries.
  async expectIdentity(firstName: string, matriculationNumber: string) {
    await expect(this.firstName()).toHaveValue(firstName)
    await expect(this.matriculationNumber()).toHaveValue(matriculationNumber)
  }

  async expectCurrentSemester(semester: string) {
    await expect(this.currentSemester()).toHaveValue(semester)
  }

  async setCurrentSemester(value: string) {
    await this.currentSemester().fill(value)
  }

  // The question FormControl wraps a div (not the input), so the label has no
  // control association; locate the field inside the label's item.
  async answerTextQuestion(title: string, answer: string) {
    const formItem = this.page.locator('label', { hasText: title }).locator('..')
    await formItem.getByRole('textbox').fill(answer)
  }

  async submit() {
    await this.page.getByRole('button', { name: 'Submit' }).click()
  }

  async expectSaved() {
    await expect(this.page.getByRole('dialog')).toContainText('Application Saved', {
      timeout: 15_000,
    })
  }

  // Persist a profile edit and its required-question answer, then wait for the
  // save confirmation. Used both to drive and to restore the edit.
  async saveProfile(semester: string, questionTitle: string, answer: string) {
    await this.setCurrentSemester(semester)
    await this.answerTextQuestion(questionTitle, answer)
    await this.submit()
    await this.expectSaved()
  }
}
