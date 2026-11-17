import {
  type CreateOrg,
  ORG_ROLE_LABELS,
  type OrgMember,
  type OrgRole,
  type UpdateOrg,
} from '@core/interfaces/org'
import { isOrgId } from '@core/managementConsole/orgAdministration/utils/isOrgId'
import { orgRolesOf } from '@core/managementConsole/shared/utils/orgRoles'
import { coreApi } from '@core/network/api'
import { coreCache, coreKeys } from '@core/network/cache'
import { serverErrorMessage } from '@core/network/serverErrorMessage'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '@tumaet/prompt-shared-state'
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

interface SetOrgMemberRoleArgs {
  // A listed member, or a user picked from the search to add to the org
  user: Pick<OrgMember, 'keycloakUserID' | 'username'>
  role: OrgRole
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

// The orgs whose roles the caller's token holds. The list is keyed by those roles, so a token refresh
// that changes them loads it anew, and a caller without org roles never requests it.
export const useOwnOrgs = () => {
  const { permissions } = useAuthStore()
  const orgRoles = orgRolesOf(permissions)

  return useQuery({
    queryKey: coreKeys.orgs.own(orgRoles),
    queryFn: coreApi.orgs.own,
    enabled: orgRoles.length > 0,
  })
}

export const useOrgMembers = (orgId: string) => {
  return useQuery({
    queryKey: coreKeys.orgs.members(orgId),
    queryFn: () => coreApi.orgs.members(orgId),
    retry: (failureCount, error) => !isClientError(error) && failureCount < MAX_RETRIES,
  })
}

// Adds a user picked from the search. The server refuses a user who is already a member, so a stale
// pick never changes their role.
export const useAddOrgMember = (orgId: string) => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ user, role }: SetOrgMemberRoleArgs) =>
      coreApi.orgs.addMember(orgId, user.keycloakUserID, { role }),
    onSuccess: (_data, { user, role }) => {
      toast({
        title: 'Member added',
        description: `${user.username} is now an org ${ORG_ROLE_LABELS[role]}.`,
      })
    },
    onError: (error) => {
      toast({
        title: 'Failed to add the member',
        description: serverErrorMessage(error),
        variant: 'destructive',
      })
    },
    // Also after an error: the user may have become a member in the meantime
    onSettled: () => coreCache.orgMembersChanged(queryClient, orgId),
  })
}

// Changes the role of a listed member by their Keycloak ID.
export const useSetOrgMemberRole = (orgId: string) => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ user, role }: SetOrgMemberRoleArgs) =>
      coreApi.orgs.setMemberRole(orgId, user.keycloakUserID, { role }),
    onSuccess: (_data, { user, role }) => {
      toast({
        title: 'Role changed',
        description: `${user.username} is now an org ${ORG_ROLE_LABELS[role]}.`,
      })
    },
    onError: (error) => {
      toast({
        title: 'Failed to change the role',
        description: serverErrorMessage(error),
        variant: 'destructive',
      })
    },
    // Also after an error: a role change that fails halfway has already added the new role
    onSettled: () => coreCache.orgMembersChanged(queryClient, orgId),
  })
}

export const useRemoveOrgMember = (orgId: string) => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (keycloakUserID: string) => coreApi.orgs.removeMember(orgId, keycloakUserID),
    onSuccess: () => {
      toast({ title: 'Member removed' })
    },
    onError: (error) => {
      toast({
        title: 'Failed to remove the member',
        description: serverErrorMessage(error),
        variant: 'destructive',
      })
    },
    // Also after an error: a removal that fails halfway has already left one of the groups
    onSettled: () => coreCache.orgMembersChanged(queryClient, orgId),
  })
}

export const useOrg = (orgId: string | undefined) => {
  return useQuery({
    queryKey: coreKeys.orgs.byId(orgId),
    queryFn: () => coreApi.orgs.get(orgId ?? ''),
    enabled: isOrgId(orgId),
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
    onSuccess: (_data, orgId) => {
      toast({ title: 'Org deleted' })
      coreCache.orgDeleted(queryClient, orgId)
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
