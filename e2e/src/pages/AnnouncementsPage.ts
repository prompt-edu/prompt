import { type Locator, type Page, expect } from '@playwright/test'

export class AnnouncementsPage {
  constructor(private readonly page: Page) {}

  async goto() {
    await this.page.goto('/management/admin/announcements')
  }

  async expectLoaded() {
    await expect(this.page.getByRole('heading', { level: 1, name: 'Announcements' })).toBeVisible({
      timeout: 15_000,
    })
  }

  async createEnabledAnnouncement(title: string, message: string) {
    await this.page.getByRole('button', { name: 'Create Announcement' }).click()
    const dialog = this.page.getByRole('dialog')
    await dialog.getByLabel('Title (optional)').fill(title)
    await dialog.getByLabel('Message').fill(message)
    await dialog.getByRole('switch', { name: 'Enabled' }).click()
    await dialog.getByRole('button', { name: 'Save' }).click()
    await expect(dialog).toBeHidden({ timeout: 15_000 })
  }

  rowFor(title: string): Locator {
    return this.page.locator('table tbody tr').filter({ hasText: title })
  }
}

export const announcementBanner = (page: Page, title: string): Locator =>
  page.getByRole('status').filter({ hasText: title })
