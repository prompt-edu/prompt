import { Page, Locator, expect } from '@playwright/test'

export class MatchingPage {
  constructor(private readonly page: Page) {}

  async goto(courseId: string, phaseId: string, subpath = '') {
    await this.page.goto(`/management/course/${courseId}/${phaseId}${subpath}`)
  }

  async expectOverviewLoaded() {
    await expect(
      this.page.getByRole('heading', { name: 'Matching Data Export and Import' }),
    ).toBeVisible({ timeout: 15_000 })
  }

  private reimportInput(): Locator {
    return this.page.locator('input[type="file"][accept=".csv"]')
  }

  async uploadReimportCsv(csv: string) {
    await this.reimportInput().setInputFiles({
      name: 'assigned-students.csv',
      mimeType: 'text/csv',
      buffer: Buffer.from(csv, 'utf-8'),
    })
    await expect(this.page.getByRole('heading', { name: 'Import Students' })).toBeVisible({
      timeout: 15_000,
    })
  }

  async expectMatchedStudent(matriculationNumber: string) {
    await expect(this.page.getByText('Successfully Matched Students')).toBeVisible()
    await expect(this.page.getByRole('cell', { name: matriculationNumber })).toBeVisible()
  }

  async importStudents() {
    await this.page.getByRole('button', { name: 'Import Students' }).click()
  }

  async expectImportSuccess() {
    await expect(this.page.getByText('Students were successfully imported.')).toBeVisible({
      timeout: 15_000,
    })
  }

  async expectAccessDenied() {
    await expect(this.page.getByText('Access Denied')).toBeVisible({ timeout: 15_000 })
    await expect(
      this.page.getByText('You do not have permission to access this page.'),
    ).toBeVisible()
    await expect(
      this.page.getByRole('heading', { name: 'Matching Data Export and Import' }),
    ).toHaveCount(0)
  }
}
