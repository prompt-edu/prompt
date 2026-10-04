export type AnnouncementSeverity = 'info' | 'warning' | 'critical'

export interface Announcement {
  id: string
  severity: AnnouncementSeverity
  title: string
  message: string
  linkUrl: string
  linkLabel: string
  startsAt: string | null
  expiresAt: string | null
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export interface UpsertAnnouncement {
  severity: AnnouncementSeverity
  title: string
  message: string
  linkUrl: string
  linkLabel: string
  startsAt: string | null
  expiresAt: string | null
  enabled: boolean
}
