import { type Locator, type Page, expect } from '@playwright/test'

// /management/privacy and its two subpages. These routes carry no permission
// guard: any authenticated user may manage their own data.
export class PrivacyPage {
  constructor(private readonly page: Page) {}

  async gotoOverview() {
    await this.page.goto('/management/privacy')
    await expect(this.page.getByRole('heading', { level: 1, name: 'Privacy' })).toBeVisible({
      timeout: 15_000,
    })
  }

  async gotoExport() {
    await this.page.goto('/management/privacy/data-export')
    await expect(this.page.getByRole('heading', { level: 1, name: 'Data Export' })).toBeVisible({
      timeout: 15_000,
    })
  }

  async gotoDeletion() {
    await this.page.goto('/management/privacy/data-deletion')
    await expect(this.page.getByRole('heading', { level: 1, name: 'Data Deletion' })).toBeVisible({
      timeout: 15_000,
    })
  }

  get overviewCards(): Locator {
    return this.page.getByRole('button').filter({ hasText: /Data Export|Data Deletion/ })
  }

  // The banner's `data-state` is its presentation state, not the backend enum:
  // a pending export renders `in_progress` and a complete one `success`.
  get statusBanner(): Locator {
    return this.page.getByTestId('privacy-status-banner')
  }

  async expectBannerState(state: string, timeout = 15_000) {
    await expect(this.statusBanner).toHaveAttribute('data-state', state, { timeout })
  }

  get requestExportButton(): Locator {
    return this.page.getByRole('button', { name: 'Request data export' })
  }

  get requestDeletionButton(): Locator {
    return this.page.getByRole('button', { name: /^Request data deletion/ })
  }

  async requestExport() {
    await this.requestExportButton.click()
    const dialog = this.page.getByRole('alertdialog')
    await expect(dialog).toBeVisible()
    await dialog.getByRole('button', { name: 'Request Data Export' }).click()
  }

  async requestDeletion() {
    await this.requestDeletionButton.click()
    const dialog = this.page.getByRole('alertdialog')
    await expect(dialog).toContainText('An administrator must approve your request')
    await dialog.getByRole('button', { name: 'Request Deletion' }).click()
  }

  documentDownloadButton(sourceName: string): Locator {
    return this.page.getByRole('button', { name: `Download ${sourceName} export` })
  }
}
