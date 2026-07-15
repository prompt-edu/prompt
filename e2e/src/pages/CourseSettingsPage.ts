import { Page, Locator, expect } from '@playwright/test'

// /management/course/:id/settings — general metadata form + danger zone. Each
// section is a collapsed-by-default SettingsCard (a Radix Collapsible), so the
// controls inside must be revealed by expanding the card first.
export class CourseSettingsPage {
  readonly title: Locator

  constructor(
    private readonly page: Page,
    private readonly courseId: string,
  ) {
    this.title = page.getByText('General Course Settings', { exact: true })
  }

  async goto() {
    await this.page.goto(`/management/course/${this.courseId}/settings`)
  }

  async expectLoaded() {
    await expect(this.title).toBeVisible({ timeout: 15_000 })
  }

  private async expandCard(name: RegExp) {
    const trigger = this.page.getByRole('button', { name }).first()
    await expect(trigger).toBeVisible()
    if ((await trigger.getAttribute('data-state')) === 'closed') {
      await trigger.click()
    }
  }

  async editShortDescription(text: string) {
    await this.expandCard(/General Course Settings/)
    await this.page.getByPlaceholder('One sentence summary').fill(text)
    const save = this.page.getByRole('button', { name: /save changes/i })
    await expect(save).toBeEnabled()
    await save.click()
    await expect(this.page.getByText('Successfully Updated Course')).toBeVisible()
  }

  async deleteCourse() {
    await this.expandCard(/Danger Zone/)
    await this.page.getByRole('button', { name: 'Delete', exact: true }).click()
    const dialog = this.page.getByRole('alertdialog')
    await expect(dialog).toBeVisible()
    await dialog.getByRole('button', { name: 'Delete' }).click()
    await this.page.waitForURL(/\/management\/courses/)
  }
}
