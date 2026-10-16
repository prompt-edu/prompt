import type { Announcement } from '@core/interfaces/announcement'
import { describe, expect, it } from 'vitest'
import {
  getAnnouncementStatus,
  getDismissalKey,
  selectVisibleAnnouncements,
} from './announcementStatus'

const NOW = new Date('2026-10-04T12:00:00Z')
const PAST = '2026-10-03T12:00:00Z'
const FUTURE = '2026-10-05T12:00:00Z'

const announcement = (overrides: Partial<Announcement> = {}): Announcement => ({
  id: 'announcement-1',
  severity: 'info',
  title: '',
  message: 'Message',
  linkUrl: '',
  linkLabel: '',
  startsAt: null,
  expiresAt: null,
  enabled: true,
  createdAt: PAST,
  updatedAt: PAST,
  ...overrides,
})

describe('getAnnouncementStatus', () => {
  it('derives the status from the enabled flag and the schedule', () => {
    expect(getAnnouncementStatus(announcement(), NOW)).toBe('active')
    expect(getAnnouncementStatus(announcement({ enabled: false }), NOW)).toBe('disabled')
    expect(getAnnouncementStatus(announcement({ startsAt: FUTURE }), NOW)).toBe('scheduled')
    expect(getAnnouncementStatus(announcement({ expiresAt: PAST }), NOW)).toBe('expired')
  })

  it('reports an expired announcement as expired even when it is disabled', () => {
    expect(getAnnouncementStatus(announcement({ enabled: false, expiresAt: PAST }), NOW)).toBe(
      'expired',
    )
  })
})

describe('selectVisibleAnnouncements', () => {
  it('hides dismissed announcements until they are edited', () => {
    const dismissed = announcement()
    const dismissedKeys = new Set([getDismissalKey(dismissed)])

    expect(selectVisibleAnnouncements([dismissed], dismissedKeys, NOW)).toEqual([])
    expect(
      selectVisibleAnnouncements([{ ...dismissed, updatedAt: FUTURE }], dismissedKeys, NOW),
    ).toHaveLength(1)
  })

  it('drops announcements that expired since they were fetched and keeps the server order', () => {
    const list = [
      announcement({ id: 'critical', severity: 'critical' }),
      announcement({ id: 'expired', expiresAt: PAST }),
      announcement({ id: 'info' }),
    ]

    expect(selectVisibleAnnouncements(list, new Set(), NOW).map(({ id }) => id)).toEqual([
      'critical',
      'info',
    ])
  })
})
