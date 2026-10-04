import type { MailingReport, SendStatusMail } from '@tumaet/prompt-shared-state'
import { API_PREFIX, coreRequest } from '../client'

export interface SendStatusMailRequest extends SendStatusMail {
  recipientCourseParticipationIDs?: string[]
}

export const mailing = {
  sendStatusMail: (coursePhaseID: string, status: SendStatusMailRequest): Promise<MailingReport> =>
    coreRequest.put(`${API_PREFIX}/mailing/${coursePhaseID}`, status),
}
