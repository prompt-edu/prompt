import type { UpsertAnnouncement } from '@core/interfaces/announcement'

type AnnouncementFormErrors = Partial<Record<keyof UpsertAnnouncement, string>>

const isHttpUrl = (value: string): boolean => {
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

export const validateAnnouncementForm = (values: UpsertAnnouncement): AnnouncementFormErrors => {
  const errors: AnnouncementFormErrors = {}
  if (!values.message) {
    errors.message = 'A message is required.'
  }
  if (values.linkUrl && !isHttpUrl(values.linkUrl)) {
    errors.linkUrl = 'Enter a valid http or https URL.'
  }
  if (values.linkLabel && !values.linkUrl) {
    errors.linkLabel = 'A link label needs a link URL.'
  }
  if (values.startsAt && values.expiresAt && values.expiresAt <= values.startsAt) {
    errors.expiresAt = 'The expiry must be after the start.'
  }
  return errors
}
