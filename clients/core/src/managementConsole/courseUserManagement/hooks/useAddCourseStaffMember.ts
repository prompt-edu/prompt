import { useMutation, useQueryClient } from '@tanstack/react-query'
import { axiosInstance } from '@tumaet/prompt-shared-state'
import { useToast } from '@tumaet/prompt-ui-components'
import { CourseGroupName } from '../interfaces/StaffMember'
import { courseStaffQueryKey } from './useCourseStaff'

interface AddMemberArgs {
  groupName: CourseGroupName
  keycloakUserID: string
}

const addCourseStaffMember = async (courseId: string, args: AddMemberArgs): Promise<void> => {
  await axiosInstance.put(
    `/api/keycloak/${courseId}/group/${args.groupName}/members/${args.keycloakUserID}`,
  )
}

export const useAddCourseStaffMember = (courseId: string) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()

  return useMutation({
    mutationFn: (args: AddMemberArgs) => addCourseStaffMember(courseId, args),
    onSuccess: (_, args) => {
      queryClient.invalidateQueries({ queryKey: courseStaffQueryKey(courseId) })
      queryClient.invalidateQueries({ queryKey: ['keycloakUserSearch'] })
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
