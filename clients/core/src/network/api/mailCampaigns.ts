import type {
  MailCampaign,
  MailCampaignDetail,
  MailCampaignRequest,
  RecipientPreview,
  SendResponse,
} from '@core/managementConsole/courseMailing/interfaces/mailCampaign'
import { API_PREFIX, coreRequest } from '../client'

const path = (courseID: string) => `${API_PREFIX}/courses/${courseID}/mail-campaigns`
const ofCampaign = (courseID: string, campaignID: string) => `${path(courseID)}/${campaignID}`

export const mailCampaigns = {
  // The server answers with null rather than an empty list when the course has no campaigns
  list: async (courseID: string): Promise<MailCampaign[]> =>
    (await coreRequest.get<MailCampaign[] | null>(path(courseID))) ?? [],

  byID: (courseID: string, campaignID: string): Promise<MailCampaignDetail> =>
    coreRequest.get(ofCampaign(courseID, campaignID)),

  recipientPreview: (courseID: string, campaignID: string): Promise<RecipientPreview> =>
    coreRequest.get(`${ofCampaign(courseID, campaignID)}/recipients-preview`),

  create: (courseID: string, request: MailCampaignRequest): Promise<MailCampaign> =>
    coreRequest.post(path(courseID), request),

  update: (
    courseID: string,
    campaignID: string,
    request: MailCampaignRequest,
  ): Promise<MailCampaign> => coreRequest.put(ofCampaign(courseID, campaignID), request),

  remove: (courseID: string, campaignID: string): Promise<void> =>
    coreRequest.del(ofCampaign(courseID, campaignID)),

  copy: (courseID: string, campaignID: string): Promise<MailCampaign> =>
    coreRequest.post(`${ofCampaign(courseID, campaignID)}/copy`),

  send: (courseID: string, campaignID: string): Promise<SendResponse> =>
    coreRequest.post(`${ofCampaign(courseID, campaignID)}/send`),

  resendFailed: (courseID: string, campaignID: string): Promise<SendResponse> =>
    coreRequest.post(`${ofCampaign(courseID, campaignID)}/resend-failed`),

  testSend: (courseID: string, campaignID: string): Promise<void> =>
    coreRequest.post(`${ofCampaign(courseID, campaignID)}/test`),
}
