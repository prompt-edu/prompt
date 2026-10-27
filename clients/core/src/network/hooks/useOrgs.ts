import type { CreateOrg, UpdateOrg } from '@core/interfaces/org'
import { coreApi } from '@core/network/api'
import { coreCache, coreKeys } from '@core/network/cache'
import { serverErrorMessage } from '@core/network/serverErrorMessage'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import { isAxiosError } from 'axios'

interface UpdateOrgArgs {
  orgId: string
  org: UpdateOrg
}

interface UpdateOrgParentArgs {
  orgId: string
  parentOrgID: string | null
}

const MAX_RETRIES = 3

export const isNotFoundError = (error: unknown): boolean =>
  isAxiosError(error) && error.response?.status === 404

const isClientError = (error: unknown): boolean => {
  const status = isAxiosError(error) ? error.response?.status : undefined
  return status !== undefined && status >= 400 && status < 500
}

export const useOrgs = () => {
  return useQuery({
    queryKey: coreKeys.orgs.all(),
    queryFn: coreApi.orgs.list,
  })
}

export const useOrg = (orgId: string | undefined) => {
  return useQuery({
    queryKey: coreKeys.orgs.byId(orgId),
    queryFn: () => coreApi.orgs.get(orgId ?? ''),
    enabled: orgId !== undefined,
    retry: (failureCount, error) => !isClientError(error) && failureCount < MAX_RETRIES,
  })
}

export const useCreateOrg = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (org: CreateOrg) => coreApi.orgs.create(org),
    onSuccess: (org) => {
      toast({ title: 'Org created', description: `${org.name} (${org.slug})` })
      coreCache.orgsChanged(queryClient)
    },
    onError: (error) => {
      toast({
        title: 'Failed to create org',
        description: serverErrorMessage(error),
        variant: 'destructive',
      })
    },
  })
}

export const useUpdateOrg = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ orgId, org }: UpdateOrgArgs) => coreApi.orgs.update(orgId, org),
    onSuccess: () => {
      toast({ title: 'Org updated' })
      coreCache.orgsChanged(queryClient)
    },
    onError: (error) => {
      toast({
        title: 'Failed to update org',
        description: serverErrorMessage(error),
        variant: 'destructive',
      })
    },
  })
}

export const useUpdateOrgParent = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ orgId, parentOrgID }: UpdateOrgParentArgs) =>
      coreApi.orgs.updateParent(orgId, { parentOrgID }),
    onSuccess: () => {
      toast({ title: 'Parent org updated' })
      coreCache.orgsChanged(queryClient)
    },
    onError: (error) => {
      toast({
        title: 'Failed to change the parent org',
        description: serverErrorMessage(error),
        variant: 'destructive',
      })
    },
  })
}

export const useDeleteOrg = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (orgId: string) => coreApi.orgs.remove(orgId),
    onSuccess: () => {
      toast({ title: 'Org deleted' })
      coreCache.orgDeleted(queryClient)
    },
    onError: (error) => {
      toast({
        title: 'Failed to delete org',
        description: serverErrorMessage(error),
        variant: 'destructive',
      })
    },
  })
}
