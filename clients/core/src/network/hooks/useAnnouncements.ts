import { describeVisibility } from '@core/announcementBanner/utils/announcementStatus'
import type { Announcement, UpsertAnnouncement } from '@core/interfaces/announcement'
import { coreApi } from '@core/network/api'
import { coreCache, coreKeys } from '@core/network/cache'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import { isAxiosError } from 'axios'

const ACTIVE_ANNOUNCEMENTS_REFETCH_INTERVAL_MS = 5 * 60 * 1000

const describeError = (error: Error): string =>
  (isAxiosError<{ error?: string }>(error) && error.response?.data?.error) ||
  'Please try again later'

export const useActiveAnnouncements = () => {
  return useQuery({
    queryKey: coreKeys.announcements.active(),
    queryFn: coreApi.announcements.active,
    refetchInterval: ACTIVE_ANNOUNCEMENTS_REFETCH_INTERVAL_MS,
    refetchOnWindowFocus: true,
    retry: false,
  })
}

export const useAnnouncements = (includeExpired: boolean) => {
  return useQuery({
    queryKey: coreKeys.announcements.list(includeExpired),
    queryFn: () => coreApi.announcements.list(includeExpired),
  })
}

const useAnnouncementMutation = <TVariables>(
  mutationFn: (variables: TVariables) => Promise<Announcement>,
  successTitle: string,
  errorTitle: string,
) => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn,
    onSuccess: (announcement) => {
      toast({ title: successTitle, description: describeVisibility(announcement, new Date()) })
      coreCache.announcementsChanged(queryClient)
    },
    onError: (error) => {
      toast({ title: errorTitle, description: describeError(error), variant: 'destructive' })
    },
  })
}

export const useCreateAnnouncement = () =>
  useAnnouncementMutation(
    (announcement: UpsertAnnouncement) => coreApi.announcements.create(announcement),
    'Announcement created',
    'Failed to create announcement',
  )

export const useUpdateAnnouncement = () =>
  useAnnouncementMutation(
    ({ id, announcement }: { id: string; announcement: UpsertAnnouncement }) =>
      coreApi.announcements.update(id, announcement),
    'Announcement updated',
    'Failed to update announcement',
  )

export const useDeleteAnnouncement = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => coreApi.announcements.remove(id),
    onSuccess: () => {
      toast({ title: 'Announcement deleted' })
      coreCache.announcementsChanged(queryClient)
    },
    onError: (error) => {
      toast({
        title: 'Failed to delete announcement',
        description: describeError(error),
        variant: 'destructive',
      })
    },
  })
}
