import { coreApi } from '@core/network/api'
import { coreCache, coreKeys } from '@core/network/cache'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { MailCampaignRequest } from '../interfaces/mailCampaign'
import { MailCampaignStatus } from '../interfaces/mailCampaign'

// A send dispatches in the background and can run for a while, so poll while a
// campaign is still "sending" instead of leaving the overview/detail view stale.
const SENDING_POLL_INTERVAL_MS = 4000

export const useGetMailCampaigns = (courseID: string) =>
  useQuery({
    queryKey: coreKeys.mailCampaigns.inCourse(courseID),
    queryFn: () => coreApi.mailCampaigns.list(courseID),
    refetchInterval: (query) =>
      query.state.data?.some((campaign) => campaign.status === MailCampaignStatus.Sending)
        ? SENDING_POLL_INTERVAL_MS
        : false,
  })

export const useGetMailCampaign = (courseID: string, campaignID: string | undefined) =>
  useQuery({
    queryKey: coreKeys.mailCampaigns.byId(courseID, campaignID ?? ''),
    queryFn: () => coreApi.mailCampaigns.byID(courseID, campaignID ?? ''),
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
    queryKey: coreKeys.mailCampaigns.recipientPreview(courseID, campaignID ?? ''),
    queryFn: () => coreApi.mailCampaigns.recipientPreview(courseID, campaignID ?? ''),
    enabled: enabled && !!campaignID,
  })

export const useCreateMailCampaign = (courseID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (request: MailCampaignRequest) => coreApi.mailCampaigns.create(courseID, request),
    onSuccess: () => coreCache.mailCampaignListChanged(queryClient, courseID),
  })
}

export const useUpdateMailCampaign = (courseID: string, campaignID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (request: MailCampaignRequest) =>
      coreApi.mailCampaigns.update(courseID, campaignID, request),
    onSuccess: () => {
      coreCache.mailCampaignChanged(queryClient, courseID, campaignID)
    },
  })
}

export const useDeleteMailCampaign = (courseID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (campaignID: string) => coreApi.mailCampaigns.remove(courseID, campaignID),
    onSuccess: () => coreCache.mailCampaignListChanged(queryClient, courseID),
  })
}

export const useCopyMailCampaign = (courseID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (campaignID: string) => coreApi.mailCampaigns.copy(courseID, campaignID),
    onSuccess: () => coreCache.mailCampaignListChanged(queryClient, courseID),
  })
}

export const useSendMailCampaign = (courseID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (campaignID: string) => coreApi.mailCampaigns.send(courseID, campaignID),
    onSuccess: (_data, campaignID) => {
      coreCache.mailCampaignSent(queryClient, courseID, campaignID)
    },
  })
}

export const useResendFailedMailCampaign = (courseID: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (campaignID: string) => coreApi.mailCampaigns.resendFailed(courseID, campaignID),
    onSuccess: (_data, campaignID) => {
      coreCache.mailCampaignSent(queryClient, courseID, campaignID)
    },
  })
}

export const useTestSendMailCampaign = (courseID: string) =>
  useMutation({
    mutationFn: (campaignID: string) => coreApi.mailCampaigns.testSend(courseID, campaignID),
  })
