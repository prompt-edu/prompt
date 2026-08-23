import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { MailCampaignRequest } from '../interfaces/mailCampaign'
import { MailCampaignStatus } from '../interfaces/mailCampaign'
import {
  copyMailCampaign,
  createMailCampaign,
  deleteMailCampaign,
  getMailCampaign,
  getMailCampaigns,
  getRecipientPreview,
  resendFailedMailCampaign,
  sendMailCampaign,
  testSendMailCampaign,
  updateMailCampaign,
} from '../network/mailCampaignApi'

// A send dispatches in the background and can run for a while, so poll while a
// campaign is still "sending" instead of leaving the overview/detail view stale.
const SENDING_POLL_INTERVAL_MS = 4000

const campaignsKey = (courseID: string) => ['mailCampaigns', courseID]
const campaignKey = (courseID: string, campaignID: string) => ['mailCampaign', courseID, campaignID]
const recipientPreviewKey = (courseID: string, campaignID: string) => [
  'mailCampaignRecipientPreview',
  courseID,
  campaignID,
]

export const useGetMailCampaigns = (courseID: string) =>
  useQuery({
    queryKey: campaignsKey(courseID),
    queryFn: () => getMailCampaigns(courseID),
    refetchInterval: (query) =>
      query.state.data?.some((campaign) => campaign.status === MailCampaignStatus.Sending)
        ? SENDING_POLL_INTERVAL_MS
        : false,
  })

export const useGetMailCampaign = (courseID: string, campaignID: string | undefined) =>
  useQuery({
    queryKey: campaignKey(courseID, campaignID ?? ''),
    queryFn: () => getMailCampaign(courseID, campaignID ?? ''),
    enabled: !!campaignID,
    refetchInterval: (query) =>
      query.state.data?.status === MailCampaignStatus.Sending ? SENDING_POLL_INTERVAL_MS : false,
  })

export const useGetRecipientPreview = (
  courseID: string,
  campaignID: string | undefined,
  enabled: boolean,
) =>
  useQuery({
    queryKey: ['mailCampaignRecipientPreview', courseID, campaignID ?? ''],
    queryFn: () => getRecipientPreview(courseID, campaignID ?? ''),
    enabled: enabled && !!campaignID,
  })

export const useCreateMailCampaign = (courseID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (request: MailCampaignRequest) => createMailCampaign(courseID, request),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: campaignsKey(courseID) }),
  })
}

export const useUpdateMailCampaign = (courseID: string, campaignID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (request: MailCampaignRequest) => updateMailCampaign(courseID, campaignID, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: campaignsKey(courseID) })
      queryClient.invalidateQueries({ queryKey: campaignKey(courseID, campaignID) })
      // Targeting (phase/statuses) may have changed, so the cached preview is stale.
      queryClient.invalidateQueries({ queryKey: recipientPreviewKey(courseID, campaignID) })
    },
  })
}

export const useDeleteMailCampaign = (courseID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (campaignID: string) => deleteMailCampaign(courseID, campaignID),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: campaignsKey(courseID) }),
  })
}

export const useCopyMailCampaign = (courseID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (campaignID: string) => copyMailCampaign(courseID, campaignID),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: campaignsKey(courseID) }),
  })
}

export const useSendMailCampaign = (courseID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (campaignID: string) => sendMailCampaign(courseID, campaignID),
    onSuccess: (_data, campaignID) => {
      queryClient.invalidateQueries({ queryKey: campaignsKey(courseID) })
      queryClient.invalidateQueries({ queryKey: campaignKey(courseID, campaignID) })
    },
  })
}

export const useResendFailedMailCampaign = (courseID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (campaignID: string) => resendFailedMailCampaign(courseID, campaignID),
    onSuccess: (_data, campaignID) => {
      queryClient.invalidateQueries({ queryKey: campaignsKey(courseID) })
      queryClient.invalidateQueries({ queryKey: campaignKey(courseID, campaignID) })
    },
  })
}

export const useTestSendMailCampaign = (courseID: string) =>
  useMutation({
    mutationFn: (campaignID: string) => testSendMailCampaign(courseID, campaignID),
  })
