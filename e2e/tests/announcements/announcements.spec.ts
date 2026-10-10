import { expect } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { test } from '../../src/fixtures/auth'
import { AnnouncementsPage, announcementBanner } from '../../src/pages/AnnouncementsPage'
import { uniqueSuffix } from '../courses/helpers'

interface Announcement {
  id: string
  title: string
}

const deleteAnnouncementsTitled = async (title: string) => {
  const admin = await apiContextFor('admin')
  try {
    const response = await admin.get('/api/announcements', { params: { includeExpired: true } })
    const announcements = (await response.json()) as Announcement[]
    for (const announcement of announcements.filter((a) => a.title === title)) {
      await admin.delete(`/api/announcements/${announcement.id}`)
    }
  } finally {
    await admin.dispose()
  }
}

test.describe('announcement banners', () => {
  test.use({ role: 'admin' })

  let title: string

  test.beforeEach(({}, testInfo) => {
    title = `Maintenance ${uniqueSuffix(testInfo.workerIndex)}`
  })

  test.afterEach(async () => {
    await deleteAnnouncementsTitled(title)
  })

  test('an enabled announcement is shown to anonymous visitors until they dismiss it', async ({
    page,
    browser,
  }) => {
    const announcements = new AnnouncementsPage(page)
    await announcements.goto()
    await announcements.expectLoaded()
    await announcements.createEnabledAnnouncement(title, 'PROMPT is unavailable tonight.')
    await expect(announcements.rowFor(title)).toContainText('Active')

    const visitorContext = await browser.newContext({ storageState: { cookies: [], origins: [] } })
    try {
      const visitor = await visitorContext.newPage()
      await visitor.goto('/')
      const banner = announcementBanner(visitor, title)
      await expect(banner).toBeVisible({ timeout: 15_000 })
      await expect(banner).toContainText('PROMPT is unavailable tonight.')

      await banner.getByRole('button', { name: 'Dismiss announcement' }).click()
      await expect(banner).toBeHidden()

      await visitor.reload()
      await expect(visitor.getByRole('heading').first()).toBeVisible({ timeout: 15_000 })
      await expect(announcementBanner(visitor, title)).toBeHidden()
    } finally {
      await visitorContext.close()
    }
  })
})
