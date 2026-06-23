import { useMutation, useQueryClient } from '@tanstack/react-query'
import { axiosInstance } from '@tumaet/prompt-shared-state'
import { useToast } from '@tumaet/prompt-ui-components'
import { CourseGroupName } from '../interfaces/StaffMember'
import { courseStaffQueryKey } from './useCourseStaff'

interface RemoveMemberArgs {
  groupName: CourseGroupName
  keycloakUserID: string
}

const removeCourseStaffMember = async (courseId: string, args: RemoveMemberArgs): Promise<void> => {
  await axiosInstance.delete(
    `/api/keycloak/${courseId}/group/${args.groupName}/members/${args.keycloakUserID}`,
  )
}

export const useRemoveCourseStaffMember = (courseId: string) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()

  return useMutation({
    mutationFn: (args: RemoveMemberArgs) => removeCourseStaffMember(courseId, args),
    onSuccess: (_, args) => {
      queryClient.invalidateQueries({ queryKey: courseStaffQueryKey(courseId) })
      queryClient.invalidateQueries({ queryKey: ['keycloakUserSearch'] })
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
