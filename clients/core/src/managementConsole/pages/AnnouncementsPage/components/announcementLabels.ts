import type { AnnouncementStatus } from '@core/announcementBanner/utils/announcementStatus'
import type { AnnouncementSeverity } from '@core/interfaces/announcement'

export const SEVERITY_LABELS: Record<AnnouncementSeverity, string> = {
  info: 'Info',
  warning: 'Warning',
  critical: 'Critical',
}

export const STATUS_LABELS: Record<AnnouncementStatus, string> = {
  active: 'Active',
  scheduled: 'Scheduled',
  disabled: 'Disabled',
  expired: 'Expired',
}
