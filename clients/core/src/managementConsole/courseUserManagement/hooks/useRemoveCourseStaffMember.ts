import { coreApi } from '@core/network/api'
import { coreCache } from '@core/network/cache'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import type { CourseGroupName } from '../interfaces/StaffMember'

interface RemoveMemberArgs {
  groupName: CourseGroupName
  keycloakUserID: string
}

export const useRemoveCourseStaffMember = (courseId: string) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()

  return useMutation({
    mutationFn: (args: RemoveMemberArgs) =>
      coreApi.keycloak.removeStaffMember(courseId, args.groupName, args.keycloakUserID),
    onSuccess: (_, args) => {
      coreCache.courseStaffChanged(queryClient, courseId)
      toast({ title: `Removed user from ${args.groupName}` })
    },
    onError: (err: unknown, args) => {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Unknown error'
      toast({
        title: `Failed to remove user from ${args.groupName}`,
        description: message,
        variant: 'destructive',
      })
    },
  })
}
