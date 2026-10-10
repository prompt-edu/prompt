import type { Announcement, UpsertAnnouncement } from '@core/interfaces/announcement'
import { API_PREFIX, coreRequest, publicRequest } from '../client'

const path = `${API_PREFIX}/announcements`

export const announcements = {
  active: (): Promise<Announcement[]> => publicRequest.get(`${path}/active`),

  list: (includeExpired: boolean): Promise<Announcement[]> =>
    coreRequest.get(path, { params: { includeExpired } }),

  create: (announcement: UpsertAnnouncement): Promise<Announcement> =>
    coreRequest.post(path, announcement),

  update: (announcementID: string, announcement: UpsertAnnouncement): Promise<Announcement> =>
    coreRequest.put(`${path}/${announcementID}`, announcement),

  remove: (announcementID: string): Promise<void> => coreRequest.del(`${path}/${announcementID}`),
}
