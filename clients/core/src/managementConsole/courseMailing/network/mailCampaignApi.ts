import { axiosInstance } from '@tumaet/prompt-shared-state'
import type {
  MailCampaign,
  MailCampaignDetail,
  MailCampaignRequest,
  RecipientPreview,
  SendResponse,
} from '../interfaces/mailCampaign'

const basePath = (courseID: string) => `/api/courses/${courseID}/mail-campaigns`

export const getMailCampaigns = async (courseID: string): Promise<MailCampaign[]> => {
  const response = await axiosInstance.get(basePath(courseID))
  return response.data ?? []
}

export const getMailCampaign = async (
  courseID: string,
  campaignID: string,
): Promise<MailCampaignDetail> => {
  const response = await axiosInstance.get(`${basePath(courseID)}/${campaignID}`)
  return response.data
}

export const getRecipientPreview = async (
  courseID: string,
  campaignID: string,
): Promise<RecipientPreview> => {
  const response = await axiosInstance.get(`${basePath(courseID)}/${campaignID}/recipients-preview`)
  return response.data
}

export const createMailCampaign = async (
  courseID: string,
  request: MailCampaignRequest,
): Promise<MailCampaign> => {
  const response = await axiosInstance.post(basePath(courseID), request)
  return response.data
}

export const updateMailCampaign = async (
  courseID: string,
  campaignID: string,
  request: MailCampaignRequest,
): Promise<MailCampaign> => {
  const response = await axiosInstance.put(`${basePath(courseID)}/${campaignID}`, request)
  return response.data
}

export const deleteMailCampaign = async (courseID: string, campaignID: string): Promise<void> => {
  await axiosInstance.delete(`${basePath(courseID)}/${campaignID}`)
}

export const copyMailCampaign = async (
  courseID: string,
  campaignID: string,
): Promise<MailCampaign> => {
  const response = await axiosInstance.post(`${basePath(courseID)}/${campaignID}/copy`)
  return response.data
}

export const sendMailCampaign = async (
  courseID: string,
  campaignID: string,
): Promise<SendResponse> => {
  const response = await axiosInstance.post(`${basePath(courseID)}/${campaignID}/send`)
  return response.data
}

export const resendFailedMailCampaign = async (
  courseID: string,
  campaignID: string,
): Promise<SendResponse> => {
  const response = await axiosInstance.post(`${basePath(courseID)}/${campaignID}/resend-failed`)
  return response.data
}

export const testSendMailCampaign = async (courseID: string, campaignID: string): Promise<void> => {
  await axiosInstance.post(`${basePath(courseID)}/${campaignID}/test`)
}
