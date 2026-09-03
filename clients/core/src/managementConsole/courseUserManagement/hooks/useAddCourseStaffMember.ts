import { coreApi } from '@core/network/api'
import { coreCache } from '@core/network/cache'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import type { CourseGroupName } from '../interfaces/StaffMember'

interface AddMemberArgs {
  groupName: CourseGroupName
  keycloakUserID: string
}

export const useAddCourseStaffMember = (courseId: string) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()

  return useMutation({
    mutationFn: (args: AddMemberArgs) =>
      coreApi.keycloak.addStaffMember(courseId, args.groupName, args.keycloakUserID),
    onSuccess: (_, args) => {
      coreCache.courseStaffChanged(queryClient, courseId)
      toast({ title: `Added user as ${args.groupName}` })
    },
    onError: (err: unknown, args) => {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Unknown error'
      toast({
        title: `Failed to add user as ${args.groupName}`,
        description: message,
        variant: 'destructive',
      })
    },
  })
}
