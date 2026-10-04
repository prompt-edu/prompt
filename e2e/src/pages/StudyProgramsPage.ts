import { type Locator, type Page, expect } from '@playwright/test'

export class StudyProgramsPage {
  constructor(private readonly page: Page) {}

  async goto() {
    await this.page.goto('/management/admin/study-programs')
    await expect(
      this.page.getByRole('heading', { level: 1, name: 'Study Programs' }),
    ).toBeVisible({ timeout: 15_000 })
  }

  row(name: string): Locator {
    return this.page.locator('table tbody tr', { hasText: name })
  }

  async add(name: string, shortName: string) {
    await this.page.getByRole('button', { name: 'Add Study Program' }).click()
    const dialog = this.page.getByRole('dialog')
    await dialog.getByLabel('Name', { exact: true }).fill(name)
    await dialog.getByLabel(/Short Name/).fill(shortName)
    await dialog.getByRole('button', { name: 'Save' }).click()
    await expect(dialog).toBeHidden()
    await expect(this.row(name)).toBeVisible()
  }
}
