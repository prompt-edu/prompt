import type { Announcement } from '@core/interfaces/announcement'
import { format } from 'date-fns'

export type AnnouncementStatus = 'disabled' | 'scheduled' | 'active' | 'expired'

type Schedule = Pick<Announcement, 'enabled' | 'startsAt' | 'expiresAt'>

export const formatAnnouncementDate = (value: string): string =>
  format(new Date(value), 'dd.MM.yyyy HH:mm')

export const getAnnouncementStatus = (announcement: Schedule, now: Date): AnnouncementStatus => {
  if (announcement.expiresAt && new Date(announcement.expiresAt) <= now) return 'expired'
  if (!announcement.enabled) return 'disabled'
  if (announcement.startsAt && new Date(announcement.startsAt) > now) return 'scheduled'
  return 'active'
}

export const describeVisibility = (announcement: Schedule, now: Date): string => {
  switch (getAnnouncementStatus(announcement, now)) {
    case 'active':
      return 'The announcement is now visible to all users.'
    case 'scheduled':
      return `The announcement will be visible from ${formatAnnouncementDate(announcement.startsAt ?? '')}.`
    case 'expired':
      return 'The announcement has expired and is not visible.'
    case 'disabled':
      return 'The announcement is disabled and not visible.'
  }
}

export const getDismissalKey = (announcement: Pick<Announcement, 'id' | 'updatedAt'>): string =>
  `${announcement.id}:${announcement.updatedAt}`

export const selectVisibleAnnouncements = (
  announcements: Announcement[],
  dismissedKeys: ReadonlySet<string>,
  now: Date,
): Announcement[] =>
  announcements
    .filter((announcement) => getAnnouncementStatus(announcement, now) === 'active')
    .filter((announcement) => !dismissedKeys.has(getDismissalKey(announcement)))
