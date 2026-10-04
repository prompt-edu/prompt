import { Locator, Page, expect } from '@playwright/test'

export interface ApplicantData {
  firstName: string
  lastName: string
  email: string
  gender: string
  nationality: string
  studyDegree: string
  studyProgram: string
  currentSemester: string
}

// /apply/:phaseId — the public application form (no authentication). The
// external-applicant branch is used: no matriculation/login fields.
export class ApplyPage {
  constructor(private readonly page: Page) {}

  async goto(phaseId: string) {
    await this.page.goto(`/apply/${phaseId}`)
  }

  // The sanitized welcome-text container: instructor HTML is rendered into it,
  // so anything still inside it survived DOMPurify.
  get welcomeText(): Locator {
    return this.page.getByTestId('application-welcome-text-content')
  }

  welcomeLink(name: string): Locator {
    return this.welcomeText.getByRole('link', { name })
  }

  async expectCourseHeading(courseName: string) {
    await expect(this.page.getByRole('heading', { name: courseName })).toBeVisible({
      timeout: 15_000,
    })
  }

  async readGlobal(name: string): Promise<unknown> {
    return this.page.evaluate(
      (globalName) => (window as unknown as Record<string, unknown>)[globalName],
      name,
    )
  }

  async continueAsExternal() {
    await this.page.getByRole('button', { name: /Continue without a .*-Account/ }).click()
    await expect(this.page.getByRole('heading', { name: 'Personal Information' })).toBeVisible()
  }

  // The shadcn Select triggers and the nationality combobox are labeled via
  // the form's label association, so getByRole('combobox', { name }) works.
  async fillStudentForm(applicant: ApplicantData) {
    await this.page.getByLabel(/First Name/).fill(applicant.firstName)
    await this.page.getByLabel(/Last Name/).fill(applicant.lastName)
    await this.page.getByLabel(/^Email/).fill(applicant.email)
    await this.selectOption(/Gender/, applicant.gender)
    await this.selectNationality(applicant.nationality)
    await this.selectOption(/Study Degree/, applicant.studyDegree)
    await this.selectOption(/^Study Program/, applicant.studyProgram)
    await this.page.getByLabel(/Current Semester/).fill(applicant.currentSemester)
  }

  // The question form's FormControl wraps a div (not the input itself), so the
  // label has no control association; locate the field inside the label's item.
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

  private async selectOption(label: RegExp, option: string) {
    await this.page.getByRole('combobox', { name: label }).click()
    await this.page.getByRole('option', { name: option, exact: true }).click()
  }

  private async selectNationality(country: string) {
    await this.page.getByRole('combobox', { name: /Nationality/ }).click()
    await this.page.getByPlaceholder('Search nationality...').fill(country)
    await this.page.getByRole('option', { name: country, exact: true }).click()
  }
}
